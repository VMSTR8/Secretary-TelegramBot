package handle_business_message

import (
	"context"
	"errors"
	"log/slog"
	"noirbot/internal/domain/model"
	"noirbot/internal/domain/repository"
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
	errRedisRefusedStub      = errors.New("redis connection refused")
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

	responseWindow = 60 * time.Second
)

type stubLongVoiceHandler struct {
	called bool
}

func (s *stubLongVoiceHandler) Execute(_ context.Context, _ model.IncomingMessage) error {
	s.called = true

	return nil
}

func expectShowThinking(ctx context.Context, sender *mock.MockBusinessSender, msg model.IncomingMessage) {
	sender.EXPECT().ShowThinking(ctx, model.ReplyDraft{
		BusinessConnectionID: msg.BusinessConnectionID,
		GuestID:              msg.GuestID,
	}).Return(nil)
}

// usecaseMocks bundles the five collaborator mocks every test case builds.
type usecaseMocks struct {
	whitelist     *mock.MockOwnerWhitelist
	connStore     *mock.MockBusinessConnectionStore
	accountReader *mock.MockBusinessAccountReader
	llm           *mock.MockLLMClient
	sender        *mock.MockBusinessSender
}

func newMocks(ctrl *gomock.Controller) usecaseMocks {
	return usecaseMocks{
		whitelist:     mock.NewMockOwnerWhitelist(ctrl),
		connStore:     mock.NewMockBusinessConnectionStore(ctrl),
		accountReader: mock.NewMockBusinessAccountReader(ctrl),
		llm:           mock.NewMockLLMClient(ctrl),
		sender:        mock.NewMockBusinessSender(ctrl),
	}
}

// expectAllowedOwner sets up the common cache-hit + whitelisted-owner path.
func (m usecaseMocks) expectAllowedOwner(ctx context.Context) {
	m.connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
	m.whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
}

// expectTextReply sets up the full allowed-owner → think → generate → send happy
// path for testMsg. sendErr is what Send returns (nil for success).
func (m usecaseMocks) expectTextReply(ctx context.Context, sendErr error) {
	m.expectAllowedOwner(ctx)
	expectShowThinking(ctx, m.sender, testMsg)
	m.llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
	m.sender.EXPECT().Send(ctx, gomock.Any()).Return(sendErr)
}

func (m usecaseMocks) usecase(
	t *testing.T,
	voiceCooldown repository.VoiceReplyWindowStore,
	longVoiceUC LongVoiceHandler,
) *Usecase {
	t.Helper()

	return newUsecase(t, m.whitelist, m.connStore, m.accountReader, m.llm, m.sender, voiceCooldown, longVoiceUC)
}

// mockVoiceCooldown returns a mock VoiceReplyWindowStore that expects
// TryEnter to NOT be called. Use for non-voice test cases.
func mockVoiceCooldown(ctrl *gomock.Controller) repository.VoiceReplyWindowStore {
	return mock.NewMockVoiceReplyWindowStore(ctrl)
}

// mockVoiceCooldownAcquired returns a mock that expects TryEnter → (true, nil).
func mockVoiceCooldownAcquired(ctrl *gomock.Controller) repository.VoiceReplyWindowStore {
	store := mock.NewMockVoiceReplyWindowStore(ctrl)
	store.EXPECT().
		TryEnter(gomock.Any(), "conn-1", int64(999), responseWindow).
		Return(true, nil)

	return store
}

// mockVoiceCooldownBlocked returns a mock that expects TryEnter → (false, nil).
func mockVoiceCooldownBlocked(ctrl *gomock.Controller) repository.VoiceReplyWindowStore {
	store := mock.NewMockVoiceReplyWindowStore(ctrl)
	store.EXPECT().
		TryEnter(gomock.Any(), "conn-1", int64(999), responseWindow).
		Return(false, nil)

	return store
}

// mockVoiceCooldownError returns a mock that expects TryEnter → error.
func mockVoiceCooldownError(ctrl *gomock.Controller) repository.VoiceReplyWindowStore {
	store := mock.NewMockVoiceReplyWindowStore(ctrl)
	store.EXPECT().
		TryEnter(gomock.Any(), "conn-1", int64(999), responseWindow).
		Return(false, errRedisRefusedStub)

	return store
}

func newUsecase(
	t *testing.T,
	whitelist *mock.MockOwnerWhitelist,
	connStore *mock.MockBusinessConnectionStore,
	accountReader *mock.MockBusinessAccountReader,
	llm *mock.MockLLMClient,
	sender *mock.MockBusinessSender,
	voiceCooldown repository.VoiceReplyWindowStore,
	longVoiceUC LongVoiceHandler,
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

	longVoice := service.NewLongVoiceDetector(service.LongVoiceDetectorConfig{
		MinDuration: 10 * time.Second,
		MaxDuration: 600 * time.Second,
	})

	if longVoiceUC == nil {
		longVoiceUC = &stubLongVoiceHandler{}
	}

	return New(
		Config{
			SystemPrompt:             systemPrompt,
			ShortVoicePrompt:         shortVoicePrompt,
			ShortVoiceResponseWindow: responseWindow,
		},
		whitelist,
		connStore,
		accountReader,
		greeting,
		flood,
		shortVoice,
		longVoice,
		voiceCooldown,
		llm,
		sender,
		longVoiceUC,
		slog.Default(),
	)
}

func runUsecaseTests(t *testing.T, tests []struct {
	name    string
	setup   func(ctrl *gomock.Controller) *Usecase
	msg     model.IncomingMessage
	wantErr error
},
) {
	t.Helper()

	ctx := context.Background()

	for i := range tests {
		tt := &tests[i]
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

func TestUsecase_TextMessages(t *testing.T) {
	ctx := context.Background()

	runUsecaseTests(t, []struct {
		name    string
		setup   func(ctrl *gomock.Controller) *Usecase
		msg     model.IncomingMessage
		wantErr error
	}{
		{
			name: "owner not in whitelist — LLM и sender не вызываются",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.connStore.EXPECT().Get(ctx, testConn.ID).Return(testConn, true, nil)
				m.whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(false, nil)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "greeting match — LLM вызван, ответ отправлен",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)
				expectShowThinking(ctx, m.sender, testMsg)
				m.llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, model.ReplyDraft{
					BusinessConnectionID: testMsg.BusinessConnectionID,
					GuestID:              testMsg.GuestID,
					Text:                 testReply,
				}).Return(nil)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "длинное сообщение без приветствия — бот молчит",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
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
	})
}

func TestUsecase_ErrorPropagation(t *testing.T) {
	ctx := context.Background()

	runUsecaseTests(t, []struct {
		name    string
		setup   func(ctrl *gomock.Controller) *Usecase
		msg     model.IncomingMessage
		wantErr error
	}{
		{
			name: "LLM вернул ошибку — возвращаем ErrLLMGenerate",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)
				expectShowThinking(ctx, m.sender, testMsg)
				m.llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).
					Return("", errDeepseekTimeoutStub)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: ErrLLMGenerate,
		},
		{
			name: "sender вернул ошибку — возвращаем ErrSend",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectTextReply(ctx, errTelegramRateLimitStub)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: ErrSend,
		},
		{
			name: "show thinking failed — LLM и Send всё равно вызываются",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)
				m.sender.EXPECT().ShowThinking(ctx, model.ReplyDraft{
					BusinessConnectionID: testMsg.BusinessConnectionID,
					GuestID:              testMsg.GuestID,
				}).Return(errTelegramDraftStub)
				m.llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "voice window store error → ErrVoiceWindow",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)

				return m.usecase(t, mockVoiceCooldownError(ctrl), nil)
			},
			msg:     testVoiceMsg,
			wantErr: ErrVoiceWindow,
		},
	})
}

func TestUsecase_EdgeCases(t *testing.T) {
	ctx := context.Background()

	runUsecaseTests(t, []struct {
		name    string
		setup   func(ctrl *gomock.Controller) *Usecase
		msg     model.IncomingMessage
		wantErr error
	}{
		{
			name: "cache miss — идём в accountReader, кешируем",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.connStore.EXPECT().Get(ctx, testConn.ID).Return(model.BusinessConnection{}, false, nil)
				m.accountReader.EXPECT().GetConnection(ctx, testConn.ID).Return(testConn, nil)
				m.connStore.EXPECT().Put(ctx, testConn).Return(nil)
				m.whitelist.EXPECT().IsAllowed(ctx, testConn.Owner.UserID).Return(true, nil)
				expectShowThinking(ctx, m.sender, testMsg)
				m.llm.EXPECT().Generate(ctx, systemPrompt, testMsg.Text).Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, gomock.Any()).Return(nil)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: nil,
		},
		{
			name: "пустой whitelist (permissive) — любой owner проходит",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectTextReply(ctx, nil)

				return m.usecase(t, mockVoiceCooldown(ctrl), nil)
			},
			msg:     testMsg,
			wantErr: nil,
		},
	})
}

func TestUsecase_VoiceMessages(t *testing.T) {
	ctx := context.Background()

	runUsecaseTests(t, []struct {
		name    string
		setup   func(ctrl *gomock.Controller) *Usecase
		msg     model.IncomingMessage
		wantErr error
	}{
		{
			name: "short voice ≤ порога + окно свободно — LLM вызван с short voice prompt",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)
				expectShowThinking(ctx, m.sender, testVoiceMsg)
				m.llm.EXPECT().Generate(ctx, shortVoicePrompt, "").Return(testReply, nil)
				m.sender.EXPECT().Send(ctx, model.ReplyDraft{
					BusinessConnectionID: testVoiceMsg.BusinessConnectionID,
					GuestID:              testVoiceMsg.GuestID,
					Text:                 testReply,
				}).Return(nil)

				return m.usecase(t, mockVoiceCooldownAcquired(ctrl), nil)
			},
			msg:     testVoiceMsg,
			wantErr: nil,
		},
		{
			name: "short voice + LLM error — Release вызывается, окно снимается",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				voiceWindow := mock.NewMockVoiceReplyWindowStore(ctrl)

				m.expectAllowedOwner(ctx)
				voiceWindow.EXPECT().
					TryEnter(gomock.Any(), "conn-1", int64(999), responseWindow).
					Return(true, nil)
				expectShowThinking(ctx, m.sender, testVoiceMsg)
				m.llm.EXPECT().Generate(ctx, shortVoicePrompt, "").
					Return("", errDeepseekTimeoutStub)
				voiceWindow.EXPECT().
					Release(gomock.Any(), "conn-1", int64(999)).
					Return(nil)

				return m.usecase(t, voiceWindow, nil)
			},
			msg:     testVoiceMsg,
			wantErr: ErrLLMGenerate,
		},
		{
			name: "short voice ≤ порога + окно занято — бот молчит",
			setup: func(ctrl *gomock.Controller) *Usecase {
				m := newMocks(ctrl)
				m.expectAllowedOwner(ctx)

				return m.usecase(t, mockVoiceCooldownBlocked(ctrl), nil)
			},
			msg:     testVoiceMsg,
			wantErr: nil,
		},
	})
}

func TestUsecase_LongVoice(t *testing.T) {
	ctx := context.Background()

	t.Run("в диапазоне — делегирует в longVoiceUC, LLM не вызывается", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		m := newMocks(ctrl)
		m.expectAllowedOwner(ctx)

		longVoiceUC := &stubLongVoiceHandler{}
		uc := m.usecase(t, mockVoiceCooldown(ctrl), longVoiceUC)

		msg := model.IncomingMessage{
			BusinessConnectionID: "conn-1",
			GuestID:              999,
			Kind:                 model.MessageKindVoice,
			VoiceDuration:        30 * time.Second,
			ReceivedAt:           time.Now(),
		}

		err := uc.Execute(ctx, msg)

		require.NoError(t, err)
		require.True(t, longVoiceUC.called)
	})

	t.Run("длиннее max — молчит, longVoiceUC не вызывается", func(t *testing.T) {
		ctrl := gomock.NewController(t)
		m := newMocks(ctrl)
		m.expectAllowedOwner(ctx)

		longVoiceUC := &stubLongVoiceHandler{}
		uc := m.usecase(t, mockVoiceCooldown(ctrl), longVoiceUC)

		msg := model.IncomingMessage{
			BusinessConnectionID: "conn-1",
			GuestID:              999,
			Kind:                 model.MessageKindVoice,
			VoiceDuration:        700 * time.Second,
			ReceivedAt:           time.Now(),
		}

		err := uc.Execute(ctx, msg)

		require.NoError(t, err)
		require.False(t, longVoiceUC.called)
	})
}
