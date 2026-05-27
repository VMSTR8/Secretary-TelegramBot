package deepseek

import "errors"

var (
	ErrUnexpectedStatus = errors.New("deepseek: unexpected status")
	ErrEmptyChoices     = errors.New("deepseek: empty choices in response")
)
