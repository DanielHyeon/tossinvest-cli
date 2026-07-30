package releaseupdate

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	maxTUFResponse    = int64(16 << 20)
	tufRequestTimeout = 30 * time.Second
)

type tufFetcher struct {
	ctx      context.Context
	client   *http.Client
	allowURL func(*url.URL) bool
}

func newTUFFetcher(
	ctx context.Context,
	client *http.Client,
	allowURL func(*url.URL) bool,
) (*tufFetcher, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if client == nil {
		client = &http.Client{Timeout: tufRequestTimeout}
	}
	if allowURL == nil {
		return nil, errors.New("releaseupdate: TUF URL policy is missing")
	}
	return &tufFetcher{ctx: ctx, client: client, allowURL: allowURL}, nil
}

// DownloadFile implements go-tuf's fetcher interface while restoring the
// security properties its default synchronous fetcher does not provide:
// caller cancellation, an explicit timeout, no redirects, and a fixed URL
// policy. maxLength is TUF's metadata-derived bound; this fetcher additionally
// caps every response independently.
func (f *tufFetcher) DownloadFile(
	raw string,
	maxLength int64,
	_ time.Duration,
) ([]byte, error) {
	target, err := url.Parse(raw)
	if err != nil || !f.allowURL(target) {
		return nil, fmt.Errorf("releaseupdate: TUF URL is outside the fixed allowlist: %q", raw)
	}
	if maxLength <= 0 {
		return nil, fmt.Errorf("releaseupdate: invalid TUF response bound %d", maxLength)
	}
	limit := min(maxLength, maxTUFResponse)
	request, err := http.NewRequestWithContext(f.ctx, http.MethodGet, target.String(), nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", userAgent)

	client := *f.client
	if client.Timeout <= 0 || client.Timeout > tufRequestTimeout {
		client.Timeout = tufRequestTimeout
	}
	client.CheckRedirect = func(*http.Request, []*http.Request) error {
		return errors.New("releaseupdate: TUF redirects are forbidden")
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("releaseupdate: TUF HTTP %d", response.StatusCode)
	}
	if response.ContentLength > limit {
		return nil, fmt.Errorf(
			"releaseupdate: TUF Content-Length %d exceeds %d",
			response.ContentLength, limit)
	}
	body, err := io.ReadAll(io.LimitReader(response.Body, limit+1))
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, fmt.Errorf("releaseupdate: TUF response exceeds %d bytes", limit)
	}
	return body, nil
}

func productionTUFURLAllowed(u *url.URL) bool {
	return productionURLAllowed(u, requestTUF)
}
