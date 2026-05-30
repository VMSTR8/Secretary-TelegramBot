package handle_business_message

import (
	"context"
	"fmt"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"noirbot/internal/domain/service"
)

type Config struct {
	SystemPrompt     string
	ShortVoicePrompt string
}

type Usecase struct {
	cfg                Config
	whitelist          repository.OwnerWhitelist
	connStore          repository.BusinessConnectionStore
	accountReader      repository.BusinessAccountReader
	greetingDetector   *service.GreetingDetector
	floodDetector      *service.FloodDetector
	shortVoiceDetector *service.ShortVoiceDetector
	llm                repository.LLMClient
	sender             repository.BusinessSender
	log                *slog.Logger
}

type llmInput struct {
	SystemPrompt string
	UserText     string
}

func New(
	cfg Config,
	whitelist repository.OwnerWhitelist,
	connStore repository.BusinessConnectionStore,
	accountReader repository.BusinessAccountReader,
	greetingDetector *service.GreetingDetector,
	floodDetector *service.FloodDetector,
	shortVoiceDetector *service.ShortVoiceDetector,
	llm repository.LLMClient,
	sender repository.BusinessSender,
	log *slog.Logger,
) *Usecase {
	return &Usecase{
		cfg:                cfg,
		whitelist:          whitelist,
		connStore:          connStore,
		accountReader:      accountReader,
		greetingDetector:   greetingDetector,
		floodDetector:      floodDetector,
		shortVoiceDetector: shortVoiceDetector,
		llm:                llm,
		sender:             sender,
		log:                log.With("usecase", "handle_business_message"),
	}
}

func (uc *Usecase) Execute(ctx context.Context, msg model.IncomingMessage) error {
	owner, err := uc.resolveOwner(ctx, msg.BusinessConnectionID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrResolveOwner, err)
	}

	allowed, err := uc.whitelist.IsAllowed(ctx, owner.UserID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrWhitelistCheck, err)
	}

	if !allowed {
		uc.log.InfoContext(ctx, "owner not in whitelist, skip",
			slog.Int64("owner_id", owner.UserID),
		)

		return nil
	}

	msg.OwnerID = owner.UserID

	decision, err := uc.classify(ctx, msg)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrFloodDetect, err)
	}

	if !decision.ShouldReply() {
		return nil
	}

	uc.log.InfoContext(ctx, "trigger fired",
		slog.String("kind", string(decision.Kind)),
		slog.String("reason", decision.Reason),
	)

	replyTarget := model.ReplyDraft{
		BusinessConnectionID: msg.BusinessConnectionID,
		GuestID:              msg.GuestID,
	}

	if shThErr := uc.sender.ShowThinking(ctx, replyTarget); shThErr != nil {
		uc.log.WarnContext(ctx, "show thinking failed",
			slog.String("error", shThErr.Error()),
		)
	}

	in := uc.llmInputs(msg)

	reply, err := uc.llm.Generate(ctx, in.SystemPrompt, in.UserText)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLMGenerate, err)
	}

	replyTarget.Text = reply

	if sndErr := uc.sender.Send(ctx, replyTarget); sndErr != nil {
		return fmt.Errorf("%w: %w", ErrSend, sndErr)
	}

	return nil
}

func (uc *Usecase) resolveOwner(ctx context.Context, connectionID string) (model.Owner, error) {
	conn, ok, err := uc.connStore.Get(ctx, connectionID)
	if err != nil {
		return model.Owner{}, fmt.Errorf("store get: %w", err)
	}

	if ok {
		return conn.Owner, nil
	}

	conn, err = uc.accountReader.GetConnection(ctx, connectionID)
	if err != nil {
		return model.Owner{}, fmt.Errorf("account reader: %w", err)
	}

	if putErr := uc.connStore.Put(ctx, conn); putErr != nil {
		uc.log.WarnContext(ctx, "failed to cache business connection",
			slog.String("connection_id", connectionID),
			slog.String("error", putErr.Error()),
		)
	}

	return conn.Owner, nil
}

func (uc *Usecase) classify(ctx context.Context, msg model.IncomingMessage) (model.TriggerDecision, error) {
	switch msg.Kind {
	case model.MessageKindVoice:
		return uc.shortVoiceDetector.Detect(msg), nil
	case model.MessageKindText:
		if decision := uc.greetingDetector.Detect(msg); decision.ShouldReply() {
			return decision, nil
		}

		return uc.floodDetector.Detect(ctx, msg)
	default:
		return model.TriggerDecision{Kind: model.TriggerKindNone}, nil
	}
}

func (uc *Usecase) llmInputs(msg model.IncomingMessage) llmInput {
	if msg.Kind == model.MessageKindVoice {
		return llmInput{SystemPrompt: uc.cfg.ShortVoicePrompt, UserText: ""}
	}

	return llmInput{SystemPrompt: uc.cfg.SystemPrompt, UserText: msg.Text}
}
