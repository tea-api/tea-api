package helper

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"tea-api/common"
	"tea-api/dto"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func SetEventStreamHeaders(c *gin.Context) {
	// 检查是否已经设置过头部
	if _, exists := c.Get("event_stream_headers_set"); exists {
		return
	}

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	c.Writer.Header().Set("X-Accel-Buffering", "no")

	// 设置标志，表示头部已经设置过
	c.Set("event_stream_headers_set", true)
}

func ClaudeData(c *gin.Context, resp dto.ClaudeResponse) error {
	jsonData, err := json.Marshal(resp)
	if err != nil {
		common.SysError("error marshalling stream response: " + err.Error())
	} else {
		c.Render(-1, common.CustomEvent{Data: fmt.Sprintf("event: %s\n", resp.Type)})
		c.Render(-1, common.CustomEvent{Data: "data: " + string(jsonData)})
	}
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return errors.New("streaming error: flusher not found")
	}
	return nil
}

func ClaudeChunkData(c *gin.Context, resp dto.ClaudeResponse, data string) {
	c.Render(-1, common.CustomEvent{Data: fmt.Sprintf("event: %s\n", resp.Type)})
	c.Render(-1, common.CustomEvent{Data: fmt.Sprintf("data: %s\n", data)})
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func ResponseChunkData(c *gin.Context, resp dto.ResponsesStreamResponse, data string) {
	c.Render(-1, common.CustomEvent{Data: fmt.Sprintf("event: %s\n", resp.Type)})
	c.Render(-1, common.CustomEvent{Data: fmt.Sprintf("data: %s", data)})
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}

func StringData(c *gin.Context, str string) error {
	//str = strings.TrimPrefix(str, "data: ")
	//str = strings.TrimSuffix(str, "\r")
	c.Render(-1, common.CustomEvent{Data: "data: " + str})
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return errors.New("streaming error: flusher not found")
	}
	return nil
}

func PingData(c *gin.Context) error {
	c.Writer.Write([]byte(": PING\n\n"))
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	} else {
		return errors.New("streaming error: flusher not found")
	}
	return nil
}

func ObjectData(c *gin.Context, object interface{}) error {
	if object == nil {
		return errors.New("object is nil")
	}
	jsonData, err := json.Marshal(object)
	if err != nil {
		return fmt.Errorf("error marshalling object: %w", err)
	}
	return StringData(c, string(jsonData))
}

func Done(c *gin.Context) {
	_ = StringData(c, "[DONE]")
}

func WssString(c *gin.Context, ws *websocket.Conn, str string) error {
	if ws == nil {
		common.LogError(c, "websocket connection is nil")
		return errors.New("websocket connection is nil")
	}

	// 更新写入超时
	ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// 使用文本消息类型(1)发送
	err := ws.WriteMessage(1, []byte(str))
	if err != nil {
		common.LogError(c, fmt.Sprintf("发送WebSocket消息失败: %s", err.Error()))
		return fmt.Errorf("发送WebSocket消息失败: %w", err)
	}

	if common.DebugEnabled {
		common.LogInfo(c, fmt.Sprintf("WebSocket消息发送成功, 长度: %d字节", len(str)))
	}

	return nil
}

func WssObject(c *gin.Context, ws *websocket.Conn, object interface{}) error {
	jsonData, err := json.Marshal(object)
	if err != nil {
		common.LogError(c, fmt.Sprintf("序列化对象失败: %s", err.Error()))
		return fmt.Errorf("error marshalling object: %w", err)
	}

	if ws == nil {
		common.LogError(c, "websocket connection is nil")
		return errors.New("websocket connection is nil")
	}

	// 更新写入超时
	ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// 使用文本消息类型(1)发送
	err = ws.WriteMessage(1, jsonData)
	if err != nil {
		common.LogError(c, fmt.Sprintf("发送WebSocket对象失败: %s", err.Error()))
		return fmt.Errorf("发送WebSocket对象失败: %w", err)
	}

	if common.DebugEnabled {
		common.LogInfo(c, fmt.Sprintf("WebSocket对象发送成功, 类型: %T, 长度: %d字节", object, len(jsonData)))
	}

	return nil
}

func WssError(c *gin.Context, ws *websocket.Conn, openaiError dto.OpenAIError) {
	errorObj := &dto.RealtimeEvent{
		Type:    "error",
		EventId: GetLocalRealtimeID(c),
		Error:   &openaiError,
	}
	_ = WssObject(c, ws, errorObj)
}

func GetResponseID(c *gin.Context) string {
	logID := c.GetString(common.RequestIdKey)
	return fmt.Sprintf("chatcmpl-%s", logID)
}

func GetLocalRealtimeID(c *gin.Context) string {
	logID := c.GetString(common.RequestIdKey)
	return fmt.Sprintf("evt_%s", logID)
}

func GenerateStopResponse(id string, createAt int64, model string, finishReason string) *dto.ChatCompletionsStreamResponse {
	return &dto.ChatCompletionsStreamResponse{
		Id:                id,
		Object:            "chat.completion.chunk",
		Created:           createAt,
		Model:             model,
		SystemFingerprint: nil,
		Choices: []dto.ChatCompletionsStreamResponseChoice{
			{
				FinishReason: &finishReason,
			},
		},
	}
}

func GenerateFinalUsageResponse(id string, createAt int64, model string, usage dto.Usage) *dto.ChatCompletionsStreamResponse {
	return &dto.ChatCompletionsStreamResponse{
		Id:                id,
		Object:            "chat.completion.chunk",
		Created:           createAt,
		Model:             model,
		SystemFingerprint: nil,
		Choices:           make([]dto.ChatCompletionsStreamResponseChoice, 0),
		Usage:             &usage,
	}
}
