package handle_business_message

import "errors"

var (
	ErrResolveOwner   = errors.New("resolve owner failed")
	ErrWhitelistCheck = errors.New("whitelist check failed")
	ErrFloodDetect    = errors.New("flood detection failed")
	ErrLLMGenerate    = errors.New("llm generate failed")
	ErrSend           = errors.New("send reply failed")
)
