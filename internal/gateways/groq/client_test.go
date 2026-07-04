package groq_test

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"noirbot/internal/gateways/groq"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testModel   = "whisper-large-v3"
	testAPIKey  = "secret-key"
	testAudio   = "fake-audio-bytes"
	wantTranscr = "детектив, мне нужна помощь" // лол, тесты писала нейронка, кста
)

// closeSpyReader tracks whether Close was called.
type closeSpyReader struct {
	io.Reader
	closed bool
	mu     sync.Mutex
}

func (r *closeSpyReader) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.closed = true

	return nil
}

func (r *closeSpyReader) wasClosed() bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.closed
}

func newClient(url string) *groq.Client {
	return groq.NewClient(groq.Config{
		BaseURL: url,
		APIKey:  testAPIKey,
		Model:   testModel,
		Timeout: 5 * time.Second,
	})
}

func TestClient_Transcribe_HappyPath(t *testing.T) {
	var (
		gotAuth      string
		gotModel     string
		gotFileName  string
		gotFileBytes []byte
	)

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")

		if !assert.NoError(t, r.ParseMultipartForm(1<<20)) {
			return
		}

		gotModel = r.FormValue("model")

		file, hdr, err := r.FormFile("file")
		if !assert.NoError(t, err) {
			return
		}

		defer func() { _ = file.Close() }()

		gotFileName = hdr.Filename
		gotFileBytes, err = io.ReadAll(file)
		assert.NoError(t, err)

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"text":"` + wantTranscr + `"}`))
	}))
	defer srv.Close()

	reader := &closeSpyReader{Reader: strings.NewReader(testAudio)}

	text, err := newClient(srv.URL).Transcribe(context.Background(), reader)

	require.NoError(t, err)
	require.Equal(t, wantTranscr, text)
	require.Equal(t, "Bearer "+testAPIKey, gotAuth)
	require.Equal(t, testModel, gotModel)
	require.Equal(t, "audio.ogg", gotFileName)
	require.Equal(t, testAudio, string(gotFileBytes))
	require.True(t, reader.wasClosed(), "reader must be closed")
}

func TestClient_Transcribe_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`invalid api key`))
	}))
	defer srv.Close()

	reader := &closeSpyReader{Reader: strings.NewReader(testAudio)}

	text, err := newClient(srv.URL).Transcribe(context.Background(), reader)

	require.ErrorIs(t, err, groq.ErrUnexpectedStatus)
	require.Empty(t, text)
	require.Contains(t, err.Error(), "invalid api key")
	require.Contains(t, err.Error(), "401")
	require.True(t, reader.wasClosed(), "reader must be closed even on error")
}

func TestClient_Transcribe_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{not-json`))
	}))
	defer srv.Close()

	reader := &closeSpyReader{Reader: strings.NewReader(testAudio)}

	text, err := newClient(srv.URL).Transcribe(context.Background(), reader)

	require.Error(t, err)
	require.Empty(t, text)
	require.Contains(t, err.Error(), "decode")
}

func TestClient_Transcribe_RequestError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, _ *http.Request) {}))
	srv.Close() // server down → connection refused

	reader := &closeSpyReader{Reader: strings.NewReader(testAudio)}

	text, err := newClient(srv.URL).Transcribe(context.Background(), reader)

	require.Error(t, err)
	require.Empty(t, text)
	require.True(t, reader.wasClosed(), "reader must be closed even on transport error")
}
