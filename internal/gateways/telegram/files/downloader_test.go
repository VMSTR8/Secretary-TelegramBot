package files

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
	"github.com/stretchr/testify/require"
)

var errGetFileStub = errors.New("get file boom")

// mockFileGetter подменяет *bot.Bot: GetFile + FileDownloadLink.
type mockFileGetter struct {
	file   *models.File
	getErr error
	link   string
}

func (m mockFileGetter) GetFile(_ context.Context, _ *bot.GetFileParams) (*models.File, error) {
	return m.file, m.getErr
}

func (m mockFileGetter) FileDownloadLink(_ *models.File) string {
	return m.link
}

func newDownloader(getter fileGetter, timeout time.Duration) *Downloader {
	return &Downloader{
		client:     getter,
		httpClient: &http.Client{Timeout: timeout},
	}
}

func TestDownloader_Download_HappyPath(t *testing.T) {
	const body = "audio-bytes"

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(body))
	}))
	defer srv.Close()

	getter := mockFileGetter{file: &models.File{FileID: "f1"}, link: srv.URL}

	rc, err := newDownloader(getter, 5*time.Second).Download(context.Background(), "f1")

	require.NoError(t, err)
	require.NotNil(t, rc)

	defer func() { _ = rc.Close() }()

	got, err := io.ReadAll(rc)
	require.NoError(t, err)
	require.Equal(t, body, string(got))
}

func TestDownloader_Download_GetFileError(t *testing.T) {
	getter := mockFileGetter{getErr: errGetFileStub}

	rc, err := newDownloader(getter, 5*time.Second).Download(context.Background(), "f1")

	require.ErrorIs(t, err, errGetFileStub)
	require.Nil(t, rc)
}

func TestDownloader_Download_NonOKStatus(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	getter := mockFileGetter{file: &models.File{FileID: "f1"}, link: srv.URL}

	rc, err := newDownloader(getter, 5*time.Second).Download(context.Background(), "f1")

	require.ErrorIs(t, err, ErrUnexpectedStatus)
	require.Nil(t, rc)
}

func TestDownloader_Download_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		time.Sleep(200 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	getter := mockFileGetter{file: &models.File{FileID: "f1"}, link: srv.URL}

	rc, err := newDownloader(getter, 50*time.Millisecond).Download(context.Background(), "f1")

	require.Error(t, err)
	require.Nil(t, rc)
}
