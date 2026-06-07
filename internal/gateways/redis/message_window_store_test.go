// internal/gateways/redis/message_window_store_test.go
package redis_test

import (
	"context"
	"noirbot/internal/domain/model"
	"sync"
	"testing"
	"time"

	redisstore "noirbot/internal/gateways/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Redis EXPIRE has 1-second granularity; using sub-second TTL would be
// silently rounded up by go-redis and break exact-match assertions.
// floodTestWindow stays sub-second because it is only used as a domain
// duration (compared client-side as nanoseconds) and never reaches EXPIRE.
const (
	floodTestWindow   = 100 * time.Millisecond
	floodTestTTL      = 2 * time.Second
	floodTestPoolSize = 50
	concurrentAppends = 100

	testOwnerA int64 = 1
	testGuestA int64 = 2
	testOwnerB int64 = 3
	testGuestB int64 = 4
)

// since0 is a "definitely in the past" bound used as the lower edge for
// CountSince when we just want to count everything that ever was appended.
var since0 = time.Unix(0, 0)

// newWindowStore wires a miniredis-backed MessageWindowStore for tests.
// The underlying client is closed automatically when the test finishes.
func newWindowStore(t *testing.T) (*redisstore.MessageWindowStore, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{
		Addr:     mr.Addr(),
		PoolSize: floodTestPoolSize,
	})

	t.Cleanup(func() { _ = client.Close() })

	return redisstore.NewMessageWindowStore(client, floodTestWindow, floodTestTTL), mr
}

func TestMessageWindowStore_Append_Single(t *testing.T) {
	store, _ := newWindowStore(t)
	ctx := context.Background()

	require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))

	count, err := store.CountSince(ctx, testOwnerA, testGuestA, since0)
	require.NoError(t, err)
	assert.Equal(t, 1, count)
}

func TestMessageWindowStore_Append_Multiple(t *testing.T) {
	const writes = 3

	store, _ := newWindowStore(t)
	ctx := context.Background()

	for range writes {
		require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))
	}

	count, err := store.CountSince(ctx, testOwnerA, testGuestA, since0)
	require.NoError(t, err)
	assert.Equal(t, writes, count)
}

func TestMessageWindowStore_CountSince_Empty(t *testing.T) {
	store, _ := newWindowStore(t)

	count, err := store.CountSince(context.Background(), testOwnerA, testGuestA, since0)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

func TestMessageWindowStore_CountSince_FutureBound(t *testing.T) {
	store, _ := newWindowStore(t)
	ctx := context.Background()
	require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))

	future := time.Now().Add(time.Hour)
	count, err := store.CountSince(ctx, testOwnerA, testGuestA, future)

	require.NoError(t, err)
	assert.Equal(t, 0, count)
}

// TestMessageWindowStore_AppendPrunesStale verifies that Append removes
// entries that have aged out of the sliding window. We use a very short
// window, sleep past it, then Append again and assert the previous entry
// is gone.
func TestMessageWindowStore_AppendPrunesStale(t *testing.T) {
	store, mr := newWindowStore(t)
	ctx := context.Background()

	require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))

	time.Sleep(floodTestWindow + 50*time.Millisecond)

	require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))

	keys := mr.Keys()
	require.Len(t, keys, 1)

	members, err := mr.ZMembers(keys[0])
	require.NoError(t, err)
	assert.Len(t, members, 1, "stale entry must be pruned by Append")
}

func TestMessageWindowStore_AppliesTTL(t *testing.T) {
	store, mr := newWindowStore(t)

	require.NoError(t, store.Append(context.Background(), testOwnerA, testGuestA, model.IncomingMessage{}))

	keys := mr.Keys()
	require.Len(t, keys, 1)
	assert.Equal(t, floodTestTTL, mr.TTL(keys[0]))
}

func TestMessageWindowStore_KeyExpires(t *testing.T) {
	store, mr := newWindowStore(t)
	require.NoError(t, store.Append(context.Background(), testOwnerA, testGuestA, model.IncomingMessage{}))

	keys := mr.Keys()
	require.Len(t, keys, 1)

	mr.FastForward(floodTestTTL + time.Second)

	assert.False(t, mr.Exists(keys[0]), "key must expire after TTL")
}

func TestMessageWindowStore_IsolatesPairs(t *testing.T) {
	store, mr := newWindowStore(t)
	ctx := context.Background()

	require.NoError(t, store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{}))
	require.NoError(t, store.Append(ctx, testOwnerB, testGuestB, model.IncomingMessage{}))

	countA, err := store.CountSince(ctx, testOwnerA, testGuestA, since0)
	require.NoError(t, err)
	assert.Equal(t, 1, countA)

	countB, err := store.CountSince(ctx, testOwnerB, testGuestB, since0)
	require.NoError(t, err)
	assert.Equal(t, 1, countB)

	cross, err := store.CountSince(ctx, testOwnerA, testGuestB, since0)
	require.NoError(t, err)
	assert.Equal(t, 0, cross, "different (owner, guest) pairs must not share a window")

	assert.Len(t, mr.Keys(), 2)
}

// TestMessageWindowStore_Concurrent verifies that the atomic per-process
// counter produces unique ZSET members under load. If members collided,
// the second ZADD with the same member would silently overwrite the first
// and the count would drop below N.
func TestMessageWindowStore_Concurrent(t *testing.T) {
	store, _ := newWindowStore(t)
	ctx := context.Background()

	var wg sync.WaitGroup

	wg.Add(concurrentAppends)

	for range concurrentAppends {
		go func() {
			defer wg.Done()

			_ = store.Append(ctx, testOwnerA, testGuestA, model.IncomingMessage{})
		}()
	}

	wg.Wait()

	count, err := store.CountSince(ctx, testOwnerA, testGuestA, since0)
	require.NoError(t, err)
	assert.Equal(t, concurrentAppends, count)
}
