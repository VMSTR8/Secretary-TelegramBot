package files

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type fileGetter interface {
	GetFile(ctx context.Context, params *bot.GetFileParams) (*models.File, error)
	FileDownloadLink(f *models.File) string
}
type Downloader struct {
	client     fileGetter
	httpClient *http.Client
}

func NewDownloader(b *bot.Bot, timeout time.Duration) *Downloader {
	return &Downloader{
		client: b,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (d *Downloader) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	f, err := d.client.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("tg downloader: get file %s: %w", fileID, err)
	}

	l := d.client.FileDownloadLink(f)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("tg downloader: new request: %w", err)
	}

	resp, err := d.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("tg downloader: do request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		_ = resp.Body.Close()

		return nil, fmt.Errorf(
			"tg downloader: %w, status: %s",
			ErrUnexpectedStatus,
			resp.Status,
		)
	}

	return resp.Body, nil
}
