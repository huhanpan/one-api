package ratio

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"

	"github.com/songquanpeng/one-api/common/logger"
)

const (
	USD2RMB   = 7
	USD       = 500 // $0.002 = 1 -> $1 = 500
	MILLI_USD = 1.0 / 1000 * USD
	RMB       = USD / USD2RMB
)

var modelRatioLock sync.RWMutex

// ModelRatio
// https://platform.openai.com/docs/models/model-endpoint-compatibility
// https://cloud.baidu.com/doc/WENXINWORKSHOP/s/Blfmc9dlf
// https://openai.com/pricing
// 1 === $0.002 / 1K tokens
// 1 === ￥0.014 / 1k tokens
var ModelRatio = map[string]float64{
	// https://openai.com/pricing
}

var CompletionRatio = map[string]float64{}

var (
	DefaultModelRatio      map[string]float64
	DefaultCompletionRatio map[string]float64
)

func init() {
	DefaultModelRatio = make(map[string]float64)
	for k, v := range ModelRatio {
		DefaultModelRatio[k] = v
	}
	DefaultCompletionRatio = make(map[string]float64)
	for k, v := range CompletionRatio {
		DefaultCompletionRatio[k] = v
	}
}

func AddNewMissingRatio(oldRatio string) string {
	newRatio := make(map[string]float64)
	err := json.Unmarshal([]byte(oldRatio), &newRatio)
	if err != nil {
		logger.SysError("error unmarshalling old ratio: " + err.Error())
		return oldRatio
	}
	for k, v := range DefaultModelRatio {
		if _, ok := newRatio[k]; !ok {
			newRatio[k] = v
		}
	}
	jsonBytes, err := json.Marshal(newRatio)
	if err != nil {
		logger.SysError("error marshalling new ratio: " + err.Error())
		return oldRatio
	}
	return string(jsonBytes)
}

func ModelRatio2JSONString() string {
	jsonBytes, err := json.Marshal(ModelRatio)
	if err != nil {
		logger.SysError("error marshalling model ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateModelRatioByJSONString(jsonStr string) error {
	modelRatioLock.Lock()
	defer modelRatioLock.Unlock()
	ModelRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &ModelRatio)
}

func GetModelRatio(name string, channelType int) float64 {
	modelRatioLock.RLock()
	defer modelRatioLock.RUnlock()
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	if strings.HasPrefix(name, "command-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	if ratio, ok := ModelRatio[model]; ok {
		return ratio
	}
	if ratio, ok := DefaultModelRatio[model]; ok {
		return ratio
	}
	if ratio, ok := ModelRatio[name]; ok {
		return ratio
	}
	if ratio, ok := DefaultModelRatio[name]; ok {
		return ratio
	}
	logger.SysError("model ratio not found: " + name)
	return 30
}

func CompletionRatio2JSONString() string {
	jsonBytes, err := json.Marshal(CompletionRatio)
	if err != nil {
		logger.SysError("error marshalling completion ratio: " + err.Error())
	}
	return string(jsonBytes)
}

func UpdateCompletionRatioByJSONString(jsonStr string) error {
	CompletionRatio = make(map[string]float64)
	return json.Unmarshal([]byte(jsonStr), &CompletionRatio)
}

func GetCompletionRatio(name string, channelType int) float64 {
	if strings.HasPrefix(name, "qwen-") && strings.HasSuffix(name, "-internet") {
		name = strings.TrimSuffix(name, "-internet")
	}
	model := fmt.Sprintf("%s(%d)", name, channelType)
	if ratio, ok := CompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[model]; ok {
		return ratio
	}
	if ratio, ok := CompletionRatio[name]; ok {
		return ratio
	}
	if ratio, ok := DefaultCompletionRatio[name]; ok {
		return ratio
	}
	if strings.HasPrefix(name, "gpt-3.5") {
		if name == "gpt-3.5-turbo" || strings.HasSuffix(name, "0125") {
			// https://openai.com/blog/new-embedding-models-and-api-updates
			// Updated GPT-3.5 Turbo model and lower pricing
			return 3
		}
		if strings.HasSuffix(name, "1106") {
			return 2
		}
		return 4.0 / 3.0
	}
	if strings.HasPrefix(name, "gpt-4") {
		if strings.HasPrefix(name, "gpt-4o") {
			if name == "gpt-4o-2024-05-13" {
				return 3
			}
			return 4
		}
		if strings.HasPrefix(name, "gpt-4-turbo") ||
			strings.HasSuffix(name, "preview") {
			return 3
		}
		return 2
	}
	// including o1, o1-preview, o1-mini
	if strings.HasPrefix(name, "o1") {
		return 4
	}
	if name == "chatgpt-4o-latest" {
		return 3
	}
	if strings.HasPrefix(name, "claude-3") {
		return 5
	}
	if strings.HasPrefix(name, "claude-") {
		return 3
	}
	if strings.HasPrefix(name, "mistral-") {
		return 3
	}
	if strings.HasPrefix(name, "gemini-") {
		return 3
	}
	if strings.HasPrefix(name, "deepseek-") {
		return 2
	}

	switch name {
	case "llama2-70b-4096":
		return 0.8 / 0.64
	case "llama3-8b-8192":
		return 2
	case "llama3-70b-8192":
		return 0.79 / 0.59
	case "command", "command-light", "command-nightly", "command-light-nightly":
		return 2
	case "command-r":
		return 3
	case "command-r-plus":
		return 5
	case "grok-beta":
		return 3
	// Replicate Models
	// https://replicate.com/pricing
	case "ibm-granite/granite-20b-code-instruct-8k":
		return 5
	case "ibm-granite/granite-3.0-2b-instruct":
		return 8.333333333333334
	case "ibm-granite/granite-3.0-8b-instruct",
		"ibm-granite/granite-8b-code-instruct-128k":
		return 5
	case "meta/llama-2-13b",
		"meta/llama-2-13b-chat",
		"meta/llama-2-7b",
		"meta/llama-2-7b-chat",
		"meta/meta-llama-3-8b",
		"meta/meta-llama-3-8b-instruct":
		return 5
	case "meta/llama-2-70b",
		"meta/llama-2-70b-chat",
		"meta/meta-llama-3-70b",
		"meta/meta-llama-3-70b-instruct":
		return 2.750 / 0.650 // ≈4.230769
	case "meta/meta-llama-3.1-405b-instruct":
		return 1
	case "mistralai/mistral-7b-instruct-v0.2",
		"mistralai/mistral-7b-v0.1":
		return 5
	case "mistralai/mixtral-8x7b-instruct-v0.1":
		return 1.000 / 0.300 // ≈3.333333
	}

	return 1
}
