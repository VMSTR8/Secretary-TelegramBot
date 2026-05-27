package service

import (
	"noirbot/internal/domain/model"
	"strings"
	"unicode"
)

type GreetingDetector struct {
	tokens map[string]struct{}
}

func NewGreetingDetector(greetings []string) *GreetingDetector {
	tokens := make(map[string]struct{}, len(greetings))
	for _, g := range greetings {
		tokens[strings.ToLower(g)] = struct{}{}
	}

	return &GreetingDetector{tokens: tokens}
}

func (d *GreetingDetector) Detect(msg model.IncomingMessage) model.TriggerDecision {
	text := normalize(msg.Text)

	if _, ok := d.tokens[text]; ok {
		return model.TriggerDecision{
			Kind:   model.TriggerKindGreeting,
			Reason: "greeting match: " + text,
		}
	}

	return model.TriggerDecision{Kind: model.TriggerKindNone}
}

func normalize(s string) string {
	s = strings.ToLower(strings.TrimSpace(s))

	return strings.TrimFunc(s, unicode.IsPunct)
}
