import PropTypes from 'prop-types';
import {
  Dialog,
  DialogTitle,
  DialogContent,
  DialogActions,
  Button,
  Divider,
  Box,
  Typography,
  Chip,
  Table,
  TableBody,
  TableCell,
  TableContainer,
  TableRow,
  Paper,
  Accordion,
  AccordionSummary,
  AccordionDetails
} from '@mui/material';
import { IconCopy, IconCheck } from '@tabler/icons-react';
import { useState } from 'react';
import { timestamp2string, renderQuota } from 'utils/common';
import LogType from '../type/LogType';
import ExpandMoreIcon from '@mui/icons-material/ExpandMore';

function renderType(type) {
  const typeOption = LogType[type];
  if (typeOption) {
    return typeOption.text;
  } else {
    return '未知';
  }
}

// 解析 content JSON
const parseContent = (content) => {
  if (!content) return null;
  try {
    return JSON.parse(content);
  } catch (e) {
    return { raw: content };
  }
};

// 格式化 JSON 字符串，失败则返回原始文本
const formatJsonText = (text) => {
  if (text === undefined || text === null || text === '') return '无内容';
  try {
    return JSON.stringify(JSON.parse(text), null, 2);
  } catch (e) {
    return typeof text === 'string' ? text : JSON.stringify(text, null, 2);
  }
};

// 复制到剪贴板
const copyToClipboard = async (text, onCopy) => {
  try {
    await navigator.clipboard.writeText(text);
    onCopy();
  } catch (err) {
    console.error('Failed to copy:', err);
  }
};

// 渲染消息内容
const renderMessage = (msg, index) => {
  if (!msg) return null;

  const role = msg.role || 'unknown';
  const content = msg.content || '';

  return (
    <Accordion key={index} defaultExpanded={index < 2}>
      <AccordionSummary expandIcon={<ExpandMoreIcon />}>
        <Box sx={{ display: 'flex', alignItems: 'center', gap: 1, width: '100%' }}>
          <Chip
            label={role === 'system' ? '系统' : role === 'user' ? '用户' : role === 'assistant' ? '助手' : role}
            color={role === 'user' ? 'primary' : role === 'assistant' ? 'success' : 'default'}
            size="small"
          />
          {msg.name && <Typography variant="body2" color="text.secondary">{msg.name}</Typography>}
          <Typography variant="body2" color="text.secondary" sx={{ ml: 'auto' }}>
            {content.length > 50 ? content.substring(0, 50) + '...' : content}
          </Typography>
        </Box>
      </AccordionSummary>
      <AccordionDetails>
        <Typography
          variant="body2"
          sx={{
            whiteSpace: 'pre-wrap',
            wordBreak: 'break-word',
            fontFamily: 'monospace',
            backgroundColor: '#f5f5f5',
            p: 2,
            borderRadius: 1
          }}
        >
          {content}
        </Typography>
        {msg.tool_calls && msg.tool_calls.length > 0 && (
          <Box mt={2}>
            <Typography variant="subtitle2" gutterBottom>工具调用:</Typography>
            {msg.tool_calls.map((call, idx) => (
              <Box key={idx} sx={{ backgroundColor: '#e3f2fd', p: 1, mt: 1, borderRadius: 1 }}>
                <Typography variant="body2">
                  <strong>函数:</strong> {call.function?.name}
                </Typography>
                <Typography variant="body2" sx={{ mt: 1 }}>
                  <strong>参数:</strong>
                  <pre style={{ margin: 0, whiteSpace: 'pre-wrap' }}>{call.function?.arguments}</pre>
                </Typography>
              </Box>
            ))}
          </Box>
        )}
      </AccordionDetails>
    </Accordion>
  );
};

const LogDetailModal = ({ open, logData, onClose }) => {
  const [copiedId, setCopiedId] = useState(null);
  const [copiedSection, setCopiedSection] = useState(null);

  // 调试日志 - 组件渲染时打印
  console.group('=== LogDetailModal Debug ===');
  console.log('open:', open);
  console.log('logData:', logData);
  
  if (!logData) {
    console.log('❌ logData is null, returning null');
    console.groupEnd();
    return null;
  }

  const parsedContent = parseContent(logData.content);
  const messages = parsedContent?.messages || [];
  const requestText = logData.request_text ?? logData.requestText ?? '';
  const responseText = logData.response_text ?? logData.responseText ?? '';
  const formattedRequest = formatJsonText(requestText);
  const formattedResponse = formatJsonText(responseText);

  console.log('request_text:', logData.request_text);
  console.log('requestText:', logData.requestText);
  console.log('response_text:', logData.response_text);
  console.log('responseText:', logData.responseText);
  console.log('formattedRequest:', formattedRequest);
  console.log('formattedResponse:', formattedResponse);
  console.log('Will show section?', formattedRequest !== '无内容' || formattedResponse !== '无内容');
  console.groupEnd();

  const handleCopy = (id) => {
    copyToClipboard(JSON.stringify(parsedContent, null, 2), () => {
      setCopiedId(id);
      setTimeout(() => setCopiedId(null), 2000);
    });
  };

  const handleCopySection = (section, text) => {
    copyToClipboard(text || '无内容', () => {
      setCopiedSection(section);
      setTimeout(() => setCopiedSection(null), 2000);
    });
  };

  return (
    <Dialog open={open} onClose={onClose} fullWidth maxWidth="lg">
      <DialogTitle
        sx={{
          margin: '0px',
          fontWeight: 700,
          lineHeight: '1.55556',
          padding: '24px',
          fontSize: '1.125rem'
        }}
      >
        日志详情
      </DialogTitle>
      <Divider />
      <DialogContent>
        {/* 基本信息 */}
        <Box sx={{ mb: 3 }}>
          <Typography variant="h6" gutterBottom>
            基本信息
          </Typography>
          <TableContainer component={Paper} variant="outlined">
            <Table size="small">
              <TableBody>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold', width: '15%' }}>时间</TableCell>
                  <TableCell>{timestamp2string(logData.created_at)}</TableCell>
                </TableRow>
                {logData.username && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>用户</TableCell>
                    <TableCell>{logData.username}</TableCell>
                  </TableRow>
                )}
                {logData.channel && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>渠道</TableCell>
                    <TableCell>{logData.channel}</TableCell>
                  </TableRow>
                )}
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>令牌</TableCell>
                  <TableCell>{logData.token_name || '-'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>类型</TableCell>
                  <TableCell>{renderType(logData.type)}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>模型</TableCell>
                  <TableCell>{logData.model_name || '-'}</TableCell>
                </TableRow>
                <TableRow>
                  <TableCell sx={{ fontWeight: 'bold' }}>Tokens</TableCell>
                  <TableCell>
                    {logData.prompt_tokens ?? 0} → {logData.completion_tokens ?? 0}
                  </TableCell>
                </TableRow>
                {logData.first_token_time !== undefined && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>首字耗时</TableCell>
                    <TableCell>{logData.first_token_time || 0} ms</TableCell>
                  </TableRow>
                )}
                {logData.elapsed_time !== undefined && logData.elapsed_time > 0 && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>生成速率</TableCell>
                    <TableCell>
                      {logData.completion_tokens > 0
                        ? `${Math.round(((logData.completion_tokens / logData.elapsed_time) * 1000) * 10) / 10} T/s`
                        : '-'}
                    </TableCell>
                  </TableRow>
                )}
                {logData.quota !== undefined && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>花费</TableCell>
                    <TableCell>{logData.quota ? renderQuota(logData.quota, 6) : '-'}</TableCell>
                  </TableRow>
                )}
               
                {logData.is_stream !== undefined && (
                  <TableRow>
                    <TableCell sx={{ fontWeight: 'bold' }}>流式</TableCell>
                    <TableCell>{logData.is_stream ? '是' : '否'}</TableCell>
                  </TableRow>
                )}
              </TableBody>
            </Table>
          </TableContainer>
        </Box>

        {/* 对话内容 */}
        <Box sx={{ mb: 2 }}>
          <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', mb: 2 }}>
            <Typography variant="h6">对话内容</Typography>
            <Button
              size="small"
              startIcon={copiedId === logData.id ? <IconCheck /> : <IconCopy />}
              onClick={() => handleCopy(logData.id)}
              variant="outlined"
            >
              {copiedId === logData.id ? '已复制' : '复制JSON'}
            </Button>
          </Box>

          {messages.length > 0 ? (
            <Box>
              {messages.map((msg, index) => renderMessage(msg, index))}
            </Box>
          ) : parsedContent && typeof parsedContent === 'object' ? (
            <Box>
              <Typography variant="subtitle2" gutterBottom>完整内容 (JSON):</Typography>
              <Typography
                component="pre"
                sx={{
                  whiteSpace: 'pre-wrap',
                  wordBreak: 'break-word',
                  fontFamily: 'monospace',
                  backgroundColor: '#f5f5f5',
                  p: 2,
                  borderRadius: 1,
                  fontSize: '0.875rem'
                }}
              >
                {JSON.stringify(parsedContent, null, 2)}
              </Typography>
            </Box>
          ) : (
            <Typography
              variant="body2"
              sx={{
                whiteSpace: 'pre-wrap',
                wordBreak: 'break-word',
                fontFamily: 'monospace',
                backgroundColor: '#f5f5f5',
                p: 2,
                borderRadius: 1
              }}
            >
              {logData.content || '无内容'}
            </Typography>
          )}
        </Box>

        {/* 请求与响应 */}
        {(formattedRequest !== '无内容' || formattedResponse !== '无内容') && (
          <Box sx={{ mb: 2 }}>
            <Typography variant="h6" gutterBottom>
              请求与响应
            </Typography>
            
            {/* Request */}
            {formattedRequest !== '无内容' && (
              <>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Typography variant="subtitle2" gutterBottom>request_text</Typography>
                  <Button
                    size="small"
                    startIcon={copiedSection === 'request' ? <IconCheck /> : <IconCopy />}
                    onClick={() => handleCopySection('request', formattedRequest)}
                    variant="outlined"
                  >
                    {copiedSection === 'request' ? '已复制' : '复制'}
                  </Button>
                </Box>
                <Typography
                  component="pre"
                  sx={{
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    fontFamily: 'monospace',
                    backgroundColor: '#f5f5f5',
                    p: 2,
                    borderRadius: 1,
                    fontSize: '0.875rem',
                    mb: 2,
                    maxHeight: '400px',
                    overflow: 'auto'
                  }}
                >
                  {formattedRequest}
                </Typography>
              </>
            )}

            {/* Response */}
            {formattedResponse !== '无内容' && (
              <>
                <Box sx={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                  <Typography variant="subtitle2" gutterBottom>response_text</Typography>
                  <Button
                    size="small"
                    startIcon={copiedSection === 'response' ? <IconCheck /> : <IconCopy />}
                    onClick={() => handleCopySection('response', formattedResponse)}
                    variant="outlined"
                  >
                    {copiedSection === 'response' ? '已复制' : '复制'}
                  </Button>
                </Box>
                <Typography
                  component="pre"
                  sx={{
                    whiteSpace: 'pre-wrap',
                    wordBreak: 'break-word',
                    fontFamily: 'monospace',
                    backgroundColor: '#f5f5f5',
                    p: 2,
                    borderRadius: 1,
                    fontSize: '0.875rem',
                    maxHeight: '400px',
                    overflow: 'auto'
                  }}
                >
                  {formattedResponse}
                </Typography>
              </>
            )}
            
            {formattedRequest === '无内容' && formattedResponse === '无内容' && (
              <Typography variant="body2" color="text.secondary">
                暂无请求和响应数据
              </Typography>
            )}
          </Box>
        )}
      </DialogContent>
      <DialogActions>
        <Button onClick={onClose}>关闭</Button>
      </DialogActions>
    </Dialog>
  );
};

export default LogDetailModal;

LogDetailModal.propTypes = {
  open: PropTypes.bool,
  logData: PropTypes.object,
  onClose: PropTypes.func
};
