package businessmsg

import "errors"

var (
	ErrResolveOwner   = errors.New("resolve owner failed")
	ErrWhitelistCheck = errors.New("whitelist check failed")
	ErrFloodDetect    = errors.New("flood detection failed")
	ErrVoiceWindow    = errors.New("voice reply window failed")
	ErrLLMGenerate    = errors.New("llm generate failed")
	ErrSend           = errors.New("send reply failed")
)
