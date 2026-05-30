package model

type TriggerKind string

const (
	TriggerKindNone       TriggerKind = "none"
	TriggerKindGreeting   TriggerKind = "greeting"
	TriggerKindFlood      TriggerKind = "flood"
	TriggerKindShortVoice TriggerKind = "short_voice"
)

type TriggerDecision struct {
	Kind   TriggerKind
	Reason string
}

func (d TriggerDecision) ShouldReply() bool {
	return d.Kind != TriggerKindNone
}
