package service

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"tea-api/common"
	"time"

	"golang.org/x/net/proxy"
)

var httpClient *http.Client
var impatientHTTPClient *http.Client

func init() {
	// 优化的传输层配置，降低首字时延
	transport := &http.Transport{
		// 连接池配置
		MaxIdleConns:        100,              // 最大空闲连接数
		MaxIdleConnsPerHost: 20,               // 每个主机最大空闲连接数
		MaxConnsPerHost:     50,               // 每个主机最大连接数
		IdleConnTimeout:     90 * time.Second, // 空闲连接超时

		// TCP连接优化
		DialContext: (&net.Dialer{
			Timeout:   10 * time.Second, // 连接超时
			KeepAlive: 30 * time.Second, // Keep-Alive间隔
		}).DialContext,

		// HTTP/2和Keep-Alive优化
		ForceAttemptHTTP2:     true,
		DisableKeepAlives:     false,
		DisableCompression:    false,
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second, // 响应头超时，降低首字时延
		ExpectContinueTimeout: 1 * time.Second,  // 100-continue超时
	}

	if common.RelayTimeout == 0 {
		httpClient = &http.Client{
			Transport: transport,
		}
	} else {
		httpClient = &http.Client{
			Timeout:   time.Duration(common.RelayTimeout) * time.Second,
			Transport: transport,
		}
	}

	// 快速客户端，用于健康检查等
	impatientHTTPClient = &http.Client{
		Timeout: 5 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:        10,
			MaxIdleConnsPerHost: 2,
			IdleConnTimeout:     30 * time.Second,
			DialContext: (&net.Dialer{
				Timeout:   3 * time.Second,
				KeepAlive: 10 * time.Second,
			}).DialContext,
			TLSHandshakeTimeout:   3 * time.Second,
			ResponseHeaderTimeout: 5 * time.Second,
		},
	}
}

func GetHttpClient() *http.Client {
	return httpClient
}

func GetImpatientHttpClient() *http.Client {
	return impatientHTTPClient
}

// NewProxyHttpClient 创建支持代理的 HTTP 客户端，优化首字时延
func NewProxyHttpClient(proxyURL string) (*http.Client, error) {
	if proxyURL == "" {
		return GetHttpClient(), nil // 使用优化过的默认客户端
	}

	parsedURL, err := url.Parse(proxyURL)
	if err != nil {
		return nil, err
	}

	// 基础传输层配置，优化连接性能
	baseTransport := &http.Transport{
		MaxIdleConns:        50,               // 代理连接池较小
		MaxIdleConnsPerHost: 10,               // 每个主机最大空闲连接数
		MaxConnsPerHost:     20,               // 每个主机最大连接数
		IdleConnTimeout:     60 * time.Second, // 空闲连接超时

		// 优化超时设置
		TLSHandshakeTimeout:   10 * time.Second,
		ResponseHeaderTimeout: 30 * time.Second, // 降低首字时延
		ExpectContinueTimeout: 1 * time.Second,

		// HTTP/2和压缩优化
		ForceAttemptHTTP2:  true,
		DisableKeepAlives:  false,
		DisableCompression: false,
	}

	switch parsedURL.Scheme {
	case "http", "https":
		baseTransport.Proxy = http.ProxyURL(parsedURL)
		baseTransport.DialContext = (&net.Dialer{
			Timeout:   10 * time.Second,
			KeepAlive: 30 * time.Second,
		}).DialContext

		return &http.Client{
			Transport: baseTransport,
			Timeout:   time.Duration(common.RelayTimeout) * time.Second,
		}, nil

	case "socks5", "socks5h":
		// 获取认证信息
		var auth *proxy.Auth
		if parsedURL.User != nil {
			auth = &proxy.Auth{
				User:     parsedURL.User.Username(),
				Password: "",
			}
			if password, ok := parsedURL.User.Password(); ok {
				auth.Password = password
			}
		}

		// 创建 SOCKS5 代理拨号器
		dialer, err := proxy.SOCKS5("tcp", parsedURL.Host, auth, proxy.Direct)
		if err != nil {
			return nil, err
		}

		baseTransport.DialContext = func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialer.Dial(network, addr)
		}

		return &http.Client{
			Transport: baseTransport,
			Timeout:   time.Duration(common.RelayTimeout) * time.Second,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported proxy scheme: %s", parsedURL.Scheme)
	}
}
