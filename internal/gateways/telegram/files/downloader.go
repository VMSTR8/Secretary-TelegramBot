package files

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/go-telegram/bot"
)

type Downloader struct {
	bot *bot.Bot
}

func NewDownloader(b *bot.Bot) *Downloader {
	return &Downloader{
		bot: b,
	}
}

func (d *Downloader) Download(ctx context.Context, fileID string) (io.ReadCloser, error) {
	f, err := d.bot.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, fmt.Errorf("tg downloader: get file %s: %w", fileID, err)
	}

	l := d.bot.FileDownloadLink(f)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, l, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("tg downloader: new request: %w", err)
	}

	resp, err := http.DefaultClient.Do(req)
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
