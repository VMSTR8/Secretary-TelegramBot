package handle_long_voice

import "errors"

var (
	ErrDownload    = errors.New("handle long voice: download failed")
	ErrTranscribe  = errors.New("handle long voice: transcribe failed")
	ErrLLMGenerate = errors.New("handle long voice: llm generate failed")
	ErrSend        = errors.New("handle long voice: send failed")
)
