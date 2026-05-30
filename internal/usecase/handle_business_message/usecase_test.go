package handle_business_message

import (
	"context"
	"errors"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository/mock"
	"noirbot/internal/domain/service"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

var (
	errDeepseekTimeoutStub   = errors.New("deepseek timeout")
	errTelegramRateLimitStub = errors.New("telegram 429")
	errTelegramDraftStub     = errors.New("telegram draft unavailable")
)

var (
	testConn = model.BusinessConnection{
		ID:        "conn-1",
		Owner:     model.Owner{UserID: 111},
		IsEnabled: true,
		CanReply:  true,
	}
	testMsg = model.IncomingMessage{
		BusinessConnectionID: "conn-1",
		GuestID:              999,
		Kind:                 model.MessageKindText,
		Text:                 "привет",
		ReceivedAt:           time.Now(),
	}
	testVoiceMsg = model.IncomingMessage{
		BusinessConnectionID: "conn-1",
		GuestID:              999,
		Kind:                 model.MessageKindVoice,
		VoiceDuration:        5 * time.Second,
		ReceivedAt:           time.Now(),
	}
	testReply        = "Ну какой привет, пиши сразу, что тебе надо!"
	systemPrompt     = "Отвечай как нуарный детектив, повидавший некоторое дерьмо"
	shortVoicePrompt = "Тебе пришло голосовое — отреагируй нуарно"
)

func expectShowThinking(ctx context.Context, sender *mock.MockBusinessSender, msg model.IncomingMessage) {
	sender.EXPECT().ShowThinking(ctx, model.ReplyDraft{
		BusinessConnectionID: msg.BusinessConnectionID,
		GuestID:              msg.GuestID,
	}).Return(nil)
}

func newUsecase(
	t *testing.T,
	whitelist *mock.MockOwnerWhitelist,
	connStore *mock.MockBusinessConnectionStore,
	accountReader *mock.MockBusinessAccountReader,
	llm *mock.MockLLMClient,
	sender *mock.MockBusinessSender,
) *Usecase {
	t.Helper()

	greeting := service.NewGreetingDetector([]string{"привет", "здоров"})

	ctrl := gomock.NewController(t)
	windowStore := mock.NewMockMessageWindowStore(ctrl)
	windowStore.EXPECT().Append(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().Return(nil)
	windowStore.EXPECT().CountSince(gomock.Any(), gomock.Any(), gomock.Any(), gomock.Any()).
		AnyTimes().Return(0, nil)

	flood := service.NewFloodDetector(service.FloodDetectorConfig{
		WindowDuration: time.Minute,
		MaxLen:         20,
		Threshold:      5,
	}, windowStore)

	shortVoice := service.NewShortVoiceDetector(service.ShortVoiceDetectorConfig{
		MaxDuration: 10 * time.Second,
	})

	return New(
		Config{
			SystemPrompt:     systemPrompt,
			ShortVoicePrompt: shortVoicePrompt,
		},
		whitelist,
		connStore,
		accountReader,
		greeting,
		flood,
		shortVoice,
		llm,
		sender,
		slog.Default(),
	)
}

func TestUsecase_Execute(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name    string
		setup   func(ctrl *gomock.Controller) *Usecase
		msg     model.IncomingMessage
		wantErr error
	}{
		{
			name: "owner not in whitelist — LLM и sender не вызываются",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(false, nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "greeting match — LLM вызван, ответ отправлен",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testMsg)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				sender.EXPECT().Send(ctx, model.ReplyDraft{
					BusinessConnectionID: testMsg.BusinessConnectionID,
					GuestID:              testMsg.GuestID,
					Text:                 testReply,
				}).Return(nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "длинное сообщение без приветствия — бот молчит",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg: model.IncomingMessage{
				BusinessConnectionID: "conn-1",
				GuestID:              999,
				Kind:                 model.MessageKindText,
				Text:                 "это очень длинное сообщение которое точно больше двадцати символов",
				ReceivedAt:           time.Now(),
			},
			wantErr: nil,
		},
		{
			name: "LLM вернул ошибку — возвращаем ErrLLMGenerate",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testMsg)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).
					Return("", errDeepseekTimeoutStub)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: ErrLLMGenerate,
		},
		{
			name: "sender вернул ошибку — возвращаем ErrSend",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testMsg)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				sender.EXPECT().Send(ctx, gomock.Any()).Return(errTelegramRateLimitStub)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: ErrSend,
		},
		{
			name: "cache miss — идём в accountReader, кешируем",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(model.BusinessConnection{}, false, nil)
				accountReader.EXPECT().GetConnection(ctx, testConn.ID).Return(testConn, nil)
				connStore.EXPECT().Put(ctx, testConn).Return(nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testMsg)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "show thinking failed — LLM и Send всё равно вызываются",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				sender.EXPECT().ShowThinking(ctx, model.ReplyDraft{
					BusinessConnectionID: testMsg.BusinessConnectionID,
					GuestID:              testMsg.GuestID,
				}).Return(errTelegramDraftStub)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "пустой whitelist (permissive) — любой owner проходит",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testMsg)
				llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "short voice ≤ порога — LLM вызван с short voice prompt и пустым userText",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, sender, testVoiceMsg)
				llm.EXPECT().Generate(ctx, shortVoicePrompt, "").Return(testReply, nil)
				sender.EXPECT().Send(ctx, model.ReplyDraft{
					BusinessConnectionID: testVoiceMsg.BusinessConnectionID,
					GuestID:              testVoiceMsg.GuestID,
					Text:                 testReply,
				}).Return(nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg:     testVoiceMsg,
			wantErr: nil,
		},
		{
			name: "long voice > порога — бот молчит, LLM не вызывается",
			setup: func(ctrl *gomock.Controller) *Usecase {
				whitelist := mock.NewMockOwnerWhitelist(ctrl)
				connStore := mock.NewMockBusinessConnectionStore(ctrl)
				accountReader := mock.NewMockBusinessAccountReader(ctrl)
				llm := mock.NewMockLLMClient(ctrl)
				sender := mock.NewMockBusinessSender(ctrl)

				connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)

				return newUsecase(t, whitelist, connStore, accountReader, llm, sender)
			},
			msg: model.IncomingMessage{
				BusinessConnectionID: "conn-1",
				GuestID:              999,
				Kind:                 model.MessageKindVoice,
				VoiceDuration:        30 * time.Second,
				ReceivedAt:           time.Now(),
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			uc := tt.setup(ctrl)

			err := uc.Execute(ctx, tt.msg)

			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
			} else {
				require.NoError(t, err)
			}
		})
	}
}
