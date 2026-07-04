package longvoice

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository/mock"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	errDownloadStub   = errors.New("download failed")
	errTranscribeStub = errors.New("transcribe failed")
	errLLMStub        = errors.New("llm failed")
	errSendStub       = errors.New("send failed")
	errThinkingStub   = errors.New("thinking failed")
)

const (
	longVoicePrompt = "Ответь на расшифровку голосового нуарно"
	voiceFileID     = "voice-file-123"
	transcript      = "детектив, мне нужна помощь"
	testReply       = "Говори, что случилось."
)

var testMsg = model.IncomingMessage{
	BusinessConnectionID: "conn-1",
	GuestID:              999,
	Kind:                 model.MessageKindVoice,
	VoiceFileID:          voiceFileID,
}

type usecaseMocks struct {
	downloader  *mock.MockVoiceDownloader
	transcriber *mock.MockTranscriber
	llm         *mock.MockLLMClient
	sender      *mock.MockBusinessSender
}

func newMocks(ctrl *gomock.Controller) usecaseMocks {
	return usecaseMocks{
		downloader:  mock.NewMockVoiceDownloader(ctrl),
		transcriber: mock.NewMockTranscriber(ctrl),
		llm:         mock.NewMockLLMClient(ctrl),
		sender:      mock.NewMockBusinessSender(ctrl),
	}
}

func (m usecaseMocks) usecase(t *testing.T) *Usecase {
	t.Helper()

	return New(
		Config{LongVoicePrompt: longVoicePrompt},
		m.downloader,
		m.transcriber,
		m.llm,
		m.sender,
		slog.Default(),
	)
}

func expectHappyPath(
	ctx context.Context,
	m usecaseMocks,
	reader io.ReadCloser,
	text string,
) {
	m.downloader.EXPECT().Download(ctx, voiceFileID).Return(reader, nil)
	m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).Return(text, nil)
	m.sender.EXPECT().ShowThinking(ctx, model.ReplyDraft{
		BusinessConnectionID: testMsg.BusinessConnectionID,
		GuestID:              testMsg.GuestID,
	}).Return(nil)
	m.llm.EXPECT().Generate(ctx, longVoicePrompt, text).Return(testReply, nil)
	m.sender.EXPECT().Send(ctx, model.ReplyDraft{
		BusinessConnectionID: testMsg.BusinessConnectionID,
		GuestID:              testMsg.GuestID,
		Text:                 testReply,
	}).Return(nil)
}

func TestUsecase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(m usecaseMocks)
		wantErr error
	}{
		{
			name: "happy path — download, transcribe, llm, send",
			setup: func(m usecaseMocks) {
				expectHappyPath(
					ctx,
					m,
					io.NopCloser(strings.NewReader("audio")),
					transcript,
				)
			},
		},
		{
			name: "пустая транскрипция — skip без llm и send",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(io.NopCloser(strings.NewReader("audio")), nil)
				m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).Return("   ", nil)
			},
		},
		{
			name: "download error",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(nil, errDownloadStub)
			},
			wantErr: ErrDownload,
		},
		{
			name: "transcribe error",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(io.NopCloser(strings.NewReader("audio")), nil)
				m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).
					Return("", errTranscribeStub)
			},
			wantErr: ErrTranscribe,
		},
		{
			name: "llm error",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(io.NopCloser(strings.NewReader("audio")), nil)
				m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).Return(transcript, nil)
				m.sender.EXPECT().ShowThinking(ctx, gomock.Any()).Return(nil)
				m.llm.EXPECT().Generate(ctx, longVoicePrompt, transcript).
					Return("", errLLMStub)
			},
			wantErr: ErrLLMGenerate,
		},
		{
			name: "send error",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(io.NopCloser(strings.NewReader("audio")), nil)
				m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).Return(transcript, nil)
				m.sender.EXPECT().ShowThinking(ctx, gomock.Any()).Return(nil)
				m.llm.EXPECT().Generate(ctx, longVoicePrompt, transcript).Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, gomock.Any()).Return(errSendStub)
			},
			wantErr: ErrSend,
		},
		{
			name: "show thinking error — пайплайн продолжается",
			setup: func(m usecaseMocks) {
				m.downloader.EXPECT().Download(ctx, voiceFileID).
					Return(io.NopCloser(strings.NewReader("audio")), nil)
				m.transcriber.EXPECT().Transcribe(ctx, gomock.Any()).Return(transcript, nil)
				m.sender.EXPECT().ShowThinking(ctx, gomock.Any()).Return(errThinkingStub)
				m.llm.EXPECT().Generate(ctx, longVoicePrompt, transcript).Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			m := newMocks(ctrl)
			tt.setup(m)

			err := m.usecase(t).Execute(ctx, testMsg)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
