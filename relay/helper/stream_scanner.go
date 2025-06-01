package helper

import (
	"bufio"
	"context"
	"io"
	"net/http"
	"tea-api/common"
	"tea-api/constant"
	relaycommon "tea-api/relay/common"
	"tea-api/setting/operation_setting"
	"strings"
	"sync"
	"time"

	"github.com/bytedance/gopkg/util/gopool"

	"github.com/gin-gonic/gin"
)

const (
	// 优化缓冲区大小以降低首字时延
	InitialScannerBufferSize = 4 << 10   // 4KB (4*1024) - 减小初始缓冲区
	MaxScannerBufferSize     = 1 << 20   // 1MB (1*1024*1024) - 减小最大缓冲区
	DefaultPingInterval      = 10 * time.Second

	// 首字响应优化相关常量
	FirstTokenBufferSize     = 1 << 10   // 1KB - 首字响应专用小缓冲区
	StreamFlushInterval      = 50 * time.Millisecond // 流式响应刷新间隔
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

	// 优化缓冲区配置以降低首字时延
	// 使用较小的初始缓冲区，首字响应后再扩展
	scanner.Buffer(make([]byte, FirstTokenBufferSize), MaxScannerBufferSize)
	scanner.Split(bufio.ScanLines)
	SetEventStreamHeaders(c)

	// 首字响应标志
	var firstTokenSent bool
	var bufferExpanded bool

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

	common.RelayCtxGo(ctx, func() {
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
				// 首字响应优化：记录首字时间并扩展缓冲区
				if !firstTokenSent {
					info.SetFirstResponseTime()
					firstTokenSent = true

					// 首字响应后扩展缓冲区以提高后续处理效率
					if !bufferExpanded {
						scanner.Buffer(make([]byte, InitialScannerBufferSize), MaxScannerBufferSize)
						bufferExpanded = true
					}
				}

				writeMutex.Lock() // Lock before writing
				success := dataHandler(data)
				writeMutex.Unlock() // Unlock after writing

				// 首字响应后立即刷新，减少延迟
				if firstTokenSent {
					if flusher, ok := c.Writer.(http.Flusher); ok {
						flusher.Flush()
					}
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

	select {
	case <-ticker.C:
		// 超时处理逻辑
		common.LogError(c, "streaming timeout")
		common.SafeSendBool(stopChan, true)
	case <-stopChan:
		// 正常结束
		common.LogInfo(c, "streaming finished")
	}
}
