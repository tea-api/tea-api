package relay

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"tea-api/common"
	"tea-api/dto"
	relaycommon "tea-api/relay/common"
	"tea-api/service"
	"tea-api/setting"
	"tea-api/setting/operation_setting"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func WssHelper(c *gin.Context, ws *websocket.Conn) (openaiErr *dto.OpenAIErrorWithStatusCode) {
	relayInfo := relaycommon.GenRelayInfoWs(c, ws)

	// 添加请求和响应超时
	ws.SetReadDeadline(time.Now().Add(60 * time.Second))
	ws.SetWriteDeadline(time.Now().Add(30 * time.Second))

	// 设置Ping处理
	ws.SetPingHandler(func(appData string) error {
		if common.DebugEnabled {
			common.LogInfo(c, "收到Ping请求，发送Pong响应")
		}
		return ws.WriteControl(websocket.PongMessage, []byte{}, time.Now().Add(10*time.Second))
	})

	// map model name
	modelMapping := c.GetString("model_mapping")
	if modelMapping != "" && modelMapping != "{}" {
		modelMap := make(map[string]string)
		err := json.Unmarshal([]byte(modelMapping), &modelMap)
		if err != nil {
			return service.OpenAIErrorWrapperLocal(err, "unmarshal_model_mapping_failed", http.StatusInternalServerError)
		}
		if modelMap[relayInfo.OriginModelName] != "" {
			relayInfo.UpstreamModelName = modelMap[relayInfo.OriginModelName]
		}
	}

	modelPrice, getModelPriceSuccess := operation_setting.GetModelPrice(relayInfo.UpstreamModelName, false)
	groupRatio := setting.GetGroupRatio(relayInfo.Group)

	var preConsumedQuota int
	var ratio float64
	var modelRatio float64

	if !getModelPriceSuccess {
		preConsumedTokens := common.PreConsumedQuota
		modelRatio, _ = operation_setting.GetModelRatio(relayInfo.UpstreamModelName)
		ratio = modelRatio * groupRatio
		preConsumedQuota = int(float64(preConsumedTokens) * ratio)
	} else {
		preConsumedQuota = int(modelPrice * common.QuotaPerUnit * groupRatio)
		relayInfo.UsePrice = true
	}

	// pre-consume quota 预消耗配额
	preConsumedQuota, userQuota, openaiErr := preConsumeQuota(c, preConsumedQuota, relayInfo)
	if openaiErr != nil {
		return openaiErr
	}

	defer func() {
		if openaiErr != nil {
			returnPreConsumedQuota(c, relayInfo, userQuota, preConsumedQuota)
		}
	}()

	adaptor := GetAdaptor(relayInfo.ApiType)
	if adaptor == nil {
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("invalid api type: %d", relayInfo.ApiType), "invalid_api_type", http.StatusBadRequest)
	}

	adaptor.Init(relayInfo)

	// 获取请求体以便发送到上游服务
	requestBody, err := common.GetRequestBody(c)
	if err != nil {
		common.LogError(c, fmt.Sprintf("获取请求体失败: %s", err.Error()))
		return service.OpenAIErrorWrapperLocal(err, "get_request_body_failed", http.StatusInternalServerError)
	}

	// 将请求体保存到请求上下文，以便DoWssRequest可以使用
	c.Set("first_wss_request", requestBody)
	c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))

	statusCodeMappingStr := c.GetString("status_code_mapping")
	resp, err := adaptor.DoRequest(c, relayInfo, bytes.NewBuffer(requestBody))
	if err != nil {
		common.LogError(c, fmt.Sprintf("执行WebSocket请求失败: %s", err.Error()))
		return service.OpenAIErrorWrapper(err, "do_request_failed", http.StatusInternalServerError)
	}

	if resp != nil {
		relayInfo.TargetWs = resp.(*websocket.Conn)
		// 确保在函数返回时关闭连接
		defer relayInfo.TargetWs.Close()

		// 设置上游连接超时
		relayInfo.TargetWs.SetReadDeadline(time.Now().Add(60 * time.Second))
		relayInfo.TargetWs.SetWriteDeadline(time.Now().Add(30 * time.Second))

		common.LogInfo(c, "成功建立与上游服务的WebSocket连接")
	} else {
		common.LogError(c, "无法获取上游WebSocket连接")
		return service.OpenAIErrorWrapperLocal(fmt.Errorf("target websocket connection is nil"), "websocket_connection_failed", http.StatusInternalServerError)
	}

	usage, openaiErr := adaptor.DoResponse(c, nil, relayInfo)
	if openaiErr != nil {
		common.LogError(c, fmt.Sprintf("处理WebSocket响应失败: %s", openaiErr.Error.Message))
		// reset status code 重置状态码
		service.ResetStatusCode(openaiErr, statusCodeMappingStr)
		return openaiErr
	}

	common.LogInfo(c, "WebSocket请求处理完成")
	service.PostWssConsumeQuota(c, relayInfo, relayInfo.UpstreamModelName, usage.(*dto.RealtimeUsage), preConsumedQuota,
		userQuota, modelRatio, groupRatio, modelPrice, getModelPriceSuccess, "")
	return nil
}
