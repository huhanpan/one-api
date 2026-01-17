package controller

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/songquanpeng/one-api/common"
	"github.com/songquanpeng/one-api/common/config"
	"github.com/songquanpeng/one-api/common/helper"
	"github.com/songquanpeng/one-api/common/logger"
	"github.com/songquanpeng/one-api/model"
	"github.com/songquanpeng/one-api/relay"
	"github.com/songquanpeng/one-api/relay/adaptor"
	"github.com/songquanpeng/one-api/relay/adaptor/openai"
	"github.com/songquanpeng/one-api/relay/billing"
	billingratio "github.com/songquanpeng/one-api/relay/billing/ratio"
	"github.com/songquanpeng/one-api/relay/meta"
	relaymodel "github.com/songquanpeng/one-api/relay/model"
)

func RelayResponsesHelper(c *gin.Context) *relaymodel.ErrorWithStatusCode {
	ctx := c.Request.Context()
	meta := meta.GetByContext(c)

	// get & validate response request
	responseRequest, err := getAndValidateResponseRequest(c)
	if err != nil {
		logger.Errorf(ctx, "getAndValidateResponseRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "invalid_response_request", http.StatusBadRequest)
	}
	meta.IsStream = responseRequest.Stream

	// map model name
	meta.OriginModelName = responseRequest.Model
	responseRequest.Model, _ = getMappedModelName(responseRequest.Model, meta.ModelMapping)
	meta.ActualModelName = responseRequest.Model

	// get model ratio & group ratio
	modelRatio := billingratio.GetModelRatio(responseRequest.Model, meta.ChannelType)
	groupRatio := billingratio.GetGroupRatio(meta.Group)
	ratio := modelRatio * groupRatio

	// pre-consume quota - estimate tokens based on input
	promptTokens := estimatePromptTokensFromResponseInput(responseRequest.Input)
	meta.PromptTokens = promptTokens
	preConsumedQuota := int64(0)
	if responseRequest.MaxOutputTokens != nil {
		preConsumedQuota = getPreConsumedQuotaForResponse(promptTokens, *responseRequest.MaxOutputTokens, ratio)
	} else {
		// default max tokens estimation
		preConsumedQuota = getPreConsumedQuotaForResponse(promptTokens, 4096, ratio)
	}

	userQuota, err := model.CacheGetUserQuota(ctx, meta.UserId)
	if err != nil {
		return openai.ErrorWrapper(err, "get_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota-preConsumedQuota < 0 {
		return openai.ErrorWrapper(errors.New("user quota is not enough"), "insufficient_user_quota", http.StatusForbidden)
	}
	err = model.CacheDecreaseUserQuota(meta.UserId, preConsumedQuota)
	if err != nil {
		return openai.ErrorWrapper(err, "decrease_user_quota_failed", http.StatusInternalServerError)
	}
	if userQuota > 100*preConsumedQuota {
		preConsumedQuota = 0
		logger.Info(ctx, fmt.Sprintf("user %d has enough quota %d, trusted and no need to pre-consume", meta.UserId, userQuota))
	}
	if preConsumedQuota > 0 {
		err := model.PreConsumeTokenQuota(meta.TokenId, preConsumedQuota)
		if err != nil {
			return openai.ErrorWrapper(err, "pre_consume_token_quota_failed", http.StatusForbidden)
		}
	}

	adaptorImpl := relay.GetAdaptor(meta.APIType)
	if adaptorImpl == nil {
		return openai.ErrorWrapper(fmt.Errorf("invalid api type: %d", meta.APIType), "invalid_api_type", http.StatusBadRequest)
	}
	adaptorImpl.Init(meta)

	// get request body
	requestBody, err := getRequestBodyForResponse(c, meta, responseRequest, adaptorImpl)
	if err != nil {
		return openai.ErrorWrapper(err, "convert_request_failed", http.StatusInternalServerError)
	}

	// print request log
	requestURL, err := adaptorImpl.GetRequestURL(meta)
	if err != nil {
		logger.Warnf(ctx, "Failed to get full request URL: %s", err.Error())
		requestURL = meta.BaseURL
		if requestURL == "" {
			requestURL = c.Request.URL.String()
		}
	}
	logger.Infof(ctx, "Request URL: %s", requestURL)

	// read and print request body
	var requestBytes []byte
	requestText := ""
	if requestBody != nil {
		requestBytes, err = io.ReadAll(requestBody)
		if err != nil {
			logger.Errorf(ctx, "Failed to read request body: %s", err.Error())
		} else {
			requestText = string(requestBytes)
			logger.Infof(ctx, "Request Body: %s", string(requestBytes))
			// recreate request body for DoRequest
			requestBody = bytes.NewReader(requestBytes)
		}
	}

	// do request
	c.Set("request_start_time", time.Now())
	resp, err := adaptorImpl.DoRequest(c, meta, requestBody)
	if err != nil {
		logger.Errorf(ctx, "DoRequest failed: %s", err.Error())
		return openai.ErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}
	if isErrorHappened(meta, resp) {
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return RelayErrorHandler(resp)
	}

	// do response
	usage, respErr := adaptorImpl.DoResponse(c, resp, meta)
	if respErr != nil {
		logger.Errorf(ctx, "respErr is not nil: %+v", respErr)
		billing.ReturnPreConsumedQuota(ctx, preConsumedQuota, meta.TokenId)
		return respErr
	}
	responseText := ""
	if rt, ok := c.Get("response_text"); ok {
		if rtStr, ok := rt.(string); ok {
			responseText = rtStr
		}
	}
	if ft, ok := c.Get("first_token_time"); ok {
		logger.Infof(ctx, "first token latency: %s", ft)
		go postConsumeQuotaForResponse(ctx, usage, meta, responseRequest, ratio, preConsumedQuota, modelRatio, groupRatio, ft.(time.Duration).Milliseconds(), requestText, responseText)
	} else {
		logger.Infof(ctx, "first token latency: %v", nil)
		go postConsumeQuotaForResponse(ctx, usage, meta, responseRequest, ratio, preConsumedQuota, modelRatio, groupRatio, 0, requestText, responseText)
	}

	return nil
}

func getAndValidateResponseRequest(c *gin.Context) (*relaymodel.ResponseRequest, error) {
	responseRequest := &relaymodel.ResponseRequest{}
	err := c.BindJSON(responseRequest)
	if err != nil {
		return nil, err
	}
	if responseRequest.Model == "" {
		return nil, errors.New("model is required")
	}
	if responseRequest.Input == nil {
		return nil, errors.New("input is required")
	}
	return responseRequest, nil
}

func estimatePromptTokensFromResponseInput(input any) int {
	// Simple estimation: count characters and divide by 4 (rough token estimate)
	// This is a basic approximation; in production you might want to use proper tokenization
	inputStr := ""
	switch v := input.(type) {
	case string:
		inputStr = v
	case []any:
		for _, item := range v {
			if str, ok := item.(string); ok {
				inputStr += str
			} else if itemMap, ok := item.(map[string]any); ok {
				// Handle structured input items
				if content, ok := itemMap["content"]; ok {
					if contentStr, ok := content.(string); ok {
						inputStr += contentStr
					}
				}
			}
		}
	case map[string]any:
		// Handle object input
		jsonBytes, err := json.Marshal(v)
		if err == nil {
			inputStr = string(jsonBytes)
		}
	}
	// Rough estimation: ~4 characters per token
	if len(inputStr) == 0 {
		return 10 // minimum default
	}
	return len(inputStr) / 4
}

func getPreConsumedQuotaForResponse(promptTokens int, maxTokens int, ratio float64) int64 {
	preConsumedTokens := config.PreConsumedQuota + int64(promptTokens) + int64(maxTokens)
	return int64(float64(preConsumedTokens) * ratio)
}

func getRequestBodyForResponse(c *gin.Context, meta *meta.Meta, responseRequest *relaymodel.ResponseRequest, adaptorImpl adaptor.Adaptor) (io.Reader, error) {
	// Get the original request body from context (it was saved by common.GetRequestBody in relay.go)
	requestBodyBytes, err := common.GetRequestBody(c)
	if err != nil {
		return nil, err
	}
	return bytes.NewBuffer(requestBodyBytes), nil
}

func postConsumeQuotaForResponse(ctx context.Context, usage *relaymodel.Usage, meta *meta.Meta, responseRequest *relaymodel.ResponseRequest, ratio float64, preConsumedQuota int64, modelRatio float64, groupRatio float64, firstTokenTime int64, requestText string, responseText string) {
	// Similar to postConsumeQuota but for Response API
	if usage == nil {
		logger.Error(ctx, "usage is nil, which is unexpected")
		return
	}

	// For Response API, we use the usage from the response directly
	var quota int64
	promptTokens := usage.PromptTokens
	completionTokens := usage.CompletionTokens
	completionRatio := billingratio.GetCompletionRatio(responseRequest.Model, meta.ChannelType)

	quota = int64(float64(promptTokens+completionTokens*int(completionRatio)) * ratio)
	if ratio != 0 && quota <= 0 {
		quota = 1
	}

	totalTokens := promptTokens + completionTokens
	if totalTokens == 0 {
		quota = 0
	}

	quotaDelta := quota - preConsumedQuota
	err := model.PostConsumeTokenQuota(meta.TokenId, quotaDelta)
	if err != nil {
		logger.Error(ctx, "error consuming token remain quota: "+err.Error())
	}
	err = model.CacheUpdateUserQuota(ctx, meta.UserId)
	if err != nil {
		logger.Error(ctx, "error update user quota cache: "+err.Error())
	}
	logContent := fmt.Sprintf("倍率：%.2f × %.2f × %.2f", modelRatio, groupRatio, completionRatio)
	model.RecordConsumeLog(ctx, &model.Log{
		UserId:            meta.UserId,
		ChannelId:         meta.ChannelId,
		PromptTokens:      promptTokens,
		CompletionTokens:  completionTokens,
		ModelName:         responseRequest.Model,
		TokenName:         meta.TokenName,
		Quota:             int(quota),
		Content:           logContent,
		RequestText:       requestText,
		ResponseText:      responseText,
		IsStream:          meta.IsStream,
		ElapsedTime:       helper.CalcElapsedTime(meta.StartTime),
		FirstTokenTime:    firstTokenTime,
		SystemPromptReset: false,
	})
	model.UpdateUserUsedQuotaAndRequestCount(meta.UserId, quota)
	model.UpdateChannelUsedQuota(meta.ChannelId, quota)
}
