// internal/gateways/redis/business_connection_store_test.go
package redis_test

import (
	"context"
	"noirbot/internal/domain/model"
	"testing"
	"time"

	redisstore "noirbot/internal/gateways/redis"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const bcTestTTL = time.Hour

// newBCStore wires a miniredis-backed BusinessConnectionStore for tests.
// The underlying client is closed automatically when the test finishes.
func newBCStore(t *testing.T) (*redisstore.BusinessConnectionStore, *miniredis.Miniredis) {
	t.Helper()

	mr := miniredis.RunT(t)
	client := redis.NewClient(&redis.Options{Addr: mr.Addr()})

	t.Cleanup(func() { _ = client.Close() })

	return redisstore.NewBusinessConnectionStore(client, bcTestTTL), mr
}

// sampleConnection returns a fully-populated business connection.
// UserChatID is included on purpose to guard against the regression where
// the field was silently dropped in the DTO mapper.
func sampleConnection() model.BusinessConnection {
	return model.BusinessConnection{
		ID:          "conn-42",
		Owner:       model.Owner{UserID: 111},
		UserChatID:  222,
		IsEnabled:   true,
		CanReply:    true,
		ConnectedAt: time.Date(2025, 1, 2, 3, 4, 5, 0, time.UTC),
	}
}

func TestBusinessConnectionStore_Get_Missing(t *testing.T) {
	store, _ := newBCStore(t)

	conn, ok, err := store.Get(context.Background(), "nope")

	require.NoError(t, err)
	assert.False(t, ok)
	assert.Equal(t, model.BusinessConnection{}, conn)
}

func TestBusinessConnectionStore_Get_InvalidJSON(t *testing.T) {
	store, mr := newBCStore(t)
	require.NoError(t, mr.Set("bc:bad", "{not json"))

	_, _, err := store.Get(context.Background(), "bad")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "redis unmarshal business_connection bad")
}

func TestBusinessConnectionStore_Put_EmptyID(t *testing.T) {
	store, _ := newBCStore(t)

	conn := sampleConnection()
	conn.ID = ""

	err := store.Put(context.Background(), conn)

	require.ErrorIs(t, err, redisstore.ErrEmptyConnectionID)
}

func TestBusinessConnectionStore_Put_AppliesTTL(t *testing.T) {
	store, mr := newBCStore(t)
	conn := sampleConnection()

	require.NoError(t, store.Put(context.Background(), conn))

	key := "bc:" + conn.ID
	assert.True(t, mr.Exists(key))
	assert.Equal(t, bcTestTTL, mr.TTL(key))
}

func TestBusinessConnectionStore_Delete_Missing(t *testing.T) {
	store, _ := newBCStore(t)

	err := store.Delete(context.Background(), "ghost")

	assert.NoError(t, err)
}

func TestBusinessConnectionStore_Delete_Existing(t *testing.T) {
	store, mr := newBCStore(t)
	conn := sampleConnection()
	require.NoError(t, store.Put(context.Background(), conn))

	key := "bc:" + conn.ID
	require.True(t, mr.Exists(key))

	err := store.Delete(context.Background(), conn.ID)

	require.NoError(t, err)
	assert.False(t, mr.Exists(key))
}

// TestBusinessConnectionStore_RoundTrip verifies that every domain field
// survives JSON marshal/unmarshal through the DTO mapper. This is the
// regression guard for the UserChatID-drop bug.
func TestBusinessConnectionStore_RoundTrip(t *testing.T) {
	store, _ := newBCStore(t)
	original := sampleConnection()
	require.NoError(t, store.Put(context.Background(), original))

	got, ok, err := store.Get(context.Background(), original.ID)

	require.NoError(t, err)
	require.True(t, ok)
	assert.Equal(t, original, got)
}
