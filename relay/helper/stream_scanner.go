package helper

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"strings"
	"sync"
	"tea-api/common"
	"tea-api/constant"
	relaycommon "tea-api/relay/common"
	"tea-api/setting/operation_setting"
	"time"

	"github.com/bytedance/gopkg/util/gopool"

	"github.com/gin-gonic/gin"
)

const (
	InitialScannerBufferSize = 1 << 20  // 1MB (1*1024*1024)
	MaxScannerBufferSize     = 10 << 20 // 10MB (10*1024*1024)
	DefaultPingInterval      = 10 * time.Second
)

func StreamScannerHandler(c *gin.Context, resp *http.Response, info *relaycommon.RelayInfo, dataHandler func(data string) bool) {

	if resp == nil || dataHandler == nil {
		return
	}

	defer resp.Body.Close()

	streamingTimeout := time.Duration(constant.StreamingTimeout) * time.Second
	if strings.HasPrefix(info.UpstreamModelName, "o") {
		// twice timeout for thinking model
		streamingTimeout *= 2
	}

	var (
		stopChan   = make(chan bool, 2)
		scanner    = bufio.NewScanner(resp.Body)
		ticker     = time.NewTicker(streamingTimeout)
		pingTicker *time.Ticker
		writeMutex sync.Mutex // Mutex to protect concurrent writes
	)

	generalSettings := operation_setting.GetGeneralSetting()
	pingEnabled := generalSettings.PingIntervalEnabled
	pingInterval := time.Duration(generalSettings.PingIntervalSeconds) * time.Second
	if pingInterval <= 0 {
		pingInterval = DefaultPingInterval
	}

	if pingEnabled {
		pingTicker = time.NewTicker(pingInterval)
	}

	defer func() {
		ticker.Stop()
		if pingTicker != nil {
			pingTicker.Stop()
		}
		close(stopChan)
	}()
	scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
	scanner.Split(bufio.ScanLines)
	SetEventStreamHeaders(c)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ctx = context.WithValue(ctx, "stop_chan", stopChan)

	// Handle ping data sending
	if pingEnabled && pingTicker != nil {
		gopool.Go(func() {
			for {
				select {
				case <-pingTicker.C:
					writeMutex.Lock() // Lock before writing
					err := PingData(c)
					writeMutex.Unlock() // Unlock after writing
					if err != nil {
						common.LogError(c, "ping data error: "+err.Error())
						common.SafeSendBool(stopChan, true)
						return
					}
					if common.DebugEnabled {
						println("ping data sent")
					}
				case <-ctx.Done():
					if common.DebugEnabled {
						println("ping data goroutine stopped")
					}
					return
				}
			}
		})
	}

	// 使用单独的goroutine处理数据扫描
	dataDone := make(chan bool, 1)
	common.RelayCtxGo(ctx, func() {
		defer func() {
			dataDone <- true
		}()

		for scanner.Scan() {
			ticker.Reset(streamingTimeout)
			data := scanner.Text()
			if common.DebugEnabled {
				println(data)
			}

			if len(data) < 6 {
				continue
			}
			if data[:5] != "data:" && data[:6] != "[DONE]" {
				continue
			}
			data = data[5:]
			data = strings.TrimLeft(data, " ")
			data = strings.TrimSuffix(data, "\r")
			if !strings.HasPrefix(data, "[DONE]") {
				info.SetFirstResponseTime()
				writeMutex.Lock() // Lock before writing
				success := dataHandler(data)
				writeMutex.Unlock() // Unlock after writing
				// 确保每次写入后都刷新
				if flusher, ok := c.Writer.(http.Flusher); ok {
					flusher.Flush()
				}
				if !success {
					break
				}
			}
		}

		if err := scanner.Err(); err != nil {
			if err != io.EOF {
				common.LogError(c, "scanner error: "+err.Error())
			}
		}

		common.SafeSendBool(stopChan, true)
	})

	// 使用select监听多个通道
	select {
	case <-ticker.C:
		// 超时处理逻辑
		common.LogError(c, "streaming timeout")
		common.SafeSendBool(stopChan, true)
	case <-stopChan:
		// 正常结束
		common.LogInfo(c, "streaming finished")
	}

	// 确保数据处理goroutine完成
	select {
	case <-dataDone:
		// 数据处理完成
	case <-time.After(2 * time.Second):
		// 防止长时间等待
		common.LogError(c, "waiting for data processing goroutine timed out")
	}

	// 结束时发送最后的刷新信号
	if flusher, ok := c.Writer.(http.Flusher); ok {
		flusher.Flush()
	}
}
