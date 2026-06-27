package redis_test

import (
	"context"
	"testing"
	"time"

	redisstore "noirbot/internal/gateways/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newVoiceReplyStore(t *testing.T) (*redisstore.VoiceReplyWindowStore, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	t.Cleanup(func() { _ = client.Close() })

	return redisstore.NewVoiceReplyWindowStore(client), mr
}

func TestVoiceReplyWindowStore_TryEnter_OpensThenBlocks(t *testing.T) {
	store, _ := newVoiceReplyStore(t)
	ctx := context.Background()

	const (
		connID  = "conn-1"
		guestID = int64(42)
		ttl     = time.Minute
	)

	opened, err := store.TryEnter(ctx, connID, guestID, ttl)
	require.NoError(t, err)
	require.True(t, opened, "first call must open the window")

	blocked, err := store.TryEnter(ctx, connID, guestID, ttl)
	require.NoError(t, err)
	require.False(t, blocked, "second call must see the window already open")
}

func TestVoiceReplyWindowStore_TryEnter_TTLIsSet(t *testing.T) {
	store, mr := newVoiceReplyStore(t)
	ctx := context.Background()

	const ttl = 30 * time.Second

	_, err := store.TryEnter(ctx, "conn-1", 42, ttl)
	require.NoError(t, err)

	keys := mr.Keys()
	require.Len(t, keys, 1)
	// miniredis rounds TTL to nearest second — check ≈ ttl
	require.InDelta(t, ttl.Seconds(), mr.TTL(keys[0]).Seconds(), 1.0)
}

func TestVoiceReplyWindowStore_TryEnter_KeyExpires(t *testing.T) {
	store, mr := newVoiceReplyStore(t)
	ctx := context.Background()

	const ttl = 3 * time.Second

	opened, err := store.TryEnter(ctx, "conn-1", 42, ttl)
	require.NoError(t, err)
	require.True(t, opened)

	// fast-forward past TTL
	mr.FastForward(ttl + time.Second)

	// window must reopen after expiry
	openedAgain, err := store.TryEnter(ctx, "conn-1", 42, ttl)
	require.NoError(t, err)
	require.True(t, openedAgain, "after TTL expires, window must reopen")
}

func TestVoiceReplyWindowStore_TryEnter_ReleaseAllowsReentry(t *testing.T) {
	store, _ := newVoiceReplyStore(t)
	ctx := context.Background()

	const (
		connID  = "conn-1"
		guestID = int64(42)
		ttl     = time.Minute
	)

	opened, err := store.TryEnter(ctx, connID, guestID, ttl)
	require.NoError(t, err)
	require.True(t, opened)

	require.NoError(t, store.Release(ctx, connID, guestID))

	openedAgain, err := store.TryEnter(ctx, connID, guestID, ttl)
	require.NoError(t, err)
	require.True(t, openedAgain, "after Release window must reopen")
}

func TestVoiceReplyWindowStore_TryEnter_IsolatesPairs(t *testing.T) {
	store, mr := newVoiceReplyStore(t)
	ctx := context.Background()
	ttl := time.Minute

	// conn-1, guest 42 — open
	opened, err := store.TryEnter(ctx, "conn-1", 42, ttl)
	require.NoError(t, err)
	require.True(t, opened)

	// conn-1, guest 99 — different guest, must open
	opened, err = store.TryEnter(ctx, "conn-1", 99, ttl)
	require.NoError(t, err)
	require.True(t, opened, "different guest must have own window")

	// conn-2, guest 42 — different connection, must open
	opened, err = store.TryEnter(ctx, "conn-2", 42, ttl)
	require.NoError(t, err)
	require.True(t, opened, "different connection must have own window")

	// conn-1, guest 42 — still blocked
	blocked, err := store.TryEnter(ctx, "conn-1", 42, ttl)
	require.NoError(t, err)
	require.False(t, blocked)

	require.Len(t, mr.Keys(), 3)
}
