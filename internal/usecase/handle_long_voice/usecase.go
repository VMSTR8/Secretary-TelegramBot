package handle_long_voice

import (
	"context"
	"fmt"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
	"strings"
)

type Config struct {
	LongVoicePrompt string
}

type Usecase struct {
	cfg         Config
	downloader  repository.VoiceDownloader
	transcriber repository.Transcriber
	llmClient   repository.LLMClient
	sender      repository.BusinessSender
	log         *slog.Logger
}

func New(
	cfg Config,
	downloader repository.VoiceDownloader,
	transcriber repository.Transcriber,
	llmClient repository.LLMClient,
	sender repository.BusinessSender,
	log *slog.Logger,
) *Usecase {
	return &Usecase{
		cfg:         cfg,
		downloader:  downloader,
		transcriber: transcriber,
		llmClient:   llmClient,
		sender:      sender,
		log:         log.With("usecase", "handle_long_voice"),
	}
}

func (uc *Usecase) Execute(ctx context.Context, msg model.IncomingMessage) error {
	reader, err := uc.downloader.Download(ctx, msg.VoiceFileID)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrDownload, err)
	}

	text, err := uc.transcriber.Transcribe(ctx, reader)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrTranscribe, err)
	}

	if strings.TrimSpace(text) == "" {
		uc.log.InfoContext(ctx, "empty transcript, skip",
			slog.String("voice_file_id", msg.VoiceFileID),
		)

		return nil
	}

	rd := model.ReplyDraft{
		BusinessConnectionID: msg.BusinessConnectionID,
		GuestID:              msg.GuestID,
	}

	if shThrErr := uc.sender.ShowThinking(ctx, rd); shThrErr != nil {
		uc.log.WarnContext(ctx, "show thinking failed",
			slog.String("error", shThrErr.Error()),
		)
	}

	reply, err := uc.llmClient.Generate(ctx, uc.cfg.LongVoicePrompt, text)
	if err != nil {
		return fmt.Errorf("%w: %w", ErrLLMGenerate, err)
	}

	rd.Text = reply

	if sndErr := uc.sender.Send(ctx, rd); sndErr != nil {
		return fmt.Errorf("%w: %w", ErrSend, sndErr)
	}

	return nil
}
