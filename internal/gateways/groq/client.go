package groq

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

type Config struct {
	BaseURL string
	APIKey  string
	Model   string
}

type Client struct {
	cfg  Config
	http *http.Client
}

func NewClient(cfg Config) *Client {
	return &Client{
		cfg: cfg,
		http: &http.Client{
			Timeout: http.DefaultClient.Timeout,
		},
	}
}

func (c *Client) Transcribe(ctx context.Context, reader io.Reader) (string, error) {
	var buf bytes.Buffer

	w := multipart.NewWriter(&buf)

	part, err := w.CreateFormFile("file", "audio.ogg")
	if err != nil {
		return "", fmt.Errorf("groq transcribe: create form file err: %w", err)
	}

	if _, copyErr := io.Copy(part, reader); copyErr != nil {
		return "", fmt.Errorf("groq transcribe: copy file err: %w", copyErr)
	}

	if wfErr := w.WriteField("model", c.cfg.Model); wfErr != nil {
		return "", fmt.Errorf("groq transcribe: write field err: %w", wfErr)
	}

	if clsErr := w.Close(); clsErr != nil {
		return "", fmt.Errorf("groq transcribe: close form file err: %w", clsErr)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.cfg.BaseURL,
		&buf,
	)
	if err != nil {
		return "", fmt.Errorf("groq transcribe: create request err: %w", err)
	}

	req.Header.Set("Content-Type", w.FormDataContentType())
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.cfg.APIKey))

	resp, err := c.http.Do(req)
	if err != nil {
		return "", fmt.Errorf("groq transcribe: http request err: %w", err)
	}

	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)

		return "", fmt.Errorf(
			"groq transcribe: %w, status: %d, body: %s",
			ErrUnexpectedStatus,
			resp.StatusCode,
			bodyBytes,
		)
	}

	type Response struct {
		Text string `json:"text"`
	}

	var response Response

	if ndErr := json.NewDecoder(resp.Body).Decode(&response); ndErr != nil {
		return "", fmt.Errorf("groq transcribe: decode response err: %w", ndErr)
	}

	return response.Text, nil
}
