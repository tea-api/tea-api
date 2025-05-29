package relay

import (
	commonconstant "tea-api/constant"
	"tea-api/relay/channel"
	"tea-api/relay/channel/ali"
	"tea-api/relay/channel/aws"
	"tea-api/relay/channel/baidu"
	"tea-api/relay/channel/baidu_v2"
	"tea-api/relay/channel/claude"
	"tea-api/relay/channel/cloudflare"
	"tea-api/relay/channel/cohere"
	"tea-api/relay/channel/coze"
	"tea-api/relay/channel/deepseek"
	"tea-api/relay/channel/dify"
	"tea-api/relay/channel/gemini"
	"tea-api/relay/channel/jina"
	"tea-api/relay/channel/mistral"
	"tea-api/relay/channel/mokaai"
	"tea-api/relay/channel/ollama"
	"tea-api/relay/channel/openai"
	"tea-api/relay/channel/palm"
	"tea-api/relay/channel/perplexity"
	"tea-api/relay/channel/siliconflow"
	"tea-api/relay/channel/task/suno"
	"tea-api/relay/channel/tencent"
	"tea-api/relay/channel/vertex"
	"tea-api/relay/channel/volcengine"
	"tea-api/relay/channel/xai"
	"tea-api/relay/channel/xunfei"
	"tea-api/relay/channel/zhipu"
	"tea-api/relay/channel/zhipu_4v"
	"tea-api/relay/constant"
)

func GetAdaptor(apiType int) channel.Adaptor {
	switch apiType {
	case constant.APITypeAli:
		return &ali.Adaptor{}
	case constant.APITypeAnthropic:
		return &claude.Adaptor{}
	case constant.APITypeBaidu:
		return &baidu.Adaptor{}
	case constant.APITypeGemini:
		return &gemini.Adaptor{}
	case constant.APITypeOpenAI:
		return &openai.Adaptor{}
	case constant.APITypePaLM:
		return &palm.Adaptor{}
	case constant.APITypeTencent:
		return &tencent.Adaptor{}
	case constant.APITypeXunfei:
		return &xunfei.Adaptor{}
	case constant.APITypeZhipu:
		return &zhipu.Adaptor{}
	case constant.APITypeZhipuV4:
		return &zhipu_4v.Adaptor{}
	case constant.APITypeOllama:
		return &ollama.Adaptor{}
	case constant.APITypePerplexity:
		return &perplexity.Adaptor{}
	case constant.APITypeAws:
		return &aws.Adaptor{}
	case constant.APITypeCohere:
		return &cohere.Adaptor{}
	case constant.APITypeDify:
		return &dify.Adaptor{}
	case constant.APITypeJina:
		return &jina.Adaptor{}
	case constant.APITypeCloudflare:
		return &cloudflare.Adaptor{}
	case constant.APITypeSiliconFlow:
		return &siliconflow.Adaptor{}
	case constant.APITypeVertexAi:
		return &vertex.Adaptor{}
	case constant.APITypeMistral:
		return &mistral.Adaptor{}
	case constant.APITypeDeepSeek:
		return &deepseek.Adaptor{}
	case constant.APITypeMokaAI:
		return &mokaai.Adaptor{}
	case constant.APITypeVolcEngine:
		return &volcengine.Adaptor{}
	case constant.APITypeBaiduV2:
		return &baidu_v2.Adaptor{}
	case constant.APITypeOpenRouter:
		return &openai.Adaptor{}
	case constant.APITypeXinference:
		return &openai.Adaptor{}
	case constant.APITypeXai:
		return &xai.Adaptor{}
	case constant.APITypeCoze:
		return &coze.Adaptor{}
	}
	return nil
}

func GetTaskAdaptor(platform commonconstant.TaskPlatform) channel.TaskAdaptor {
	switch platform {
	//case constant.APITypeAIProxyLibrary:
	//	return &aiproxy.Adaptor{}
	case commonconstant.TaskPlatformSuno:
		return &suno.TaskAdaptor{}
	}
	return nil
}
