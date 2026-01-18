package client

import (
	"context"
	"log"
	"net/http"

	"github.com/tmaxmax/go-sse"
)

type SSEClient struct {
	url        string
	httpClient *http.Client
}

func NewSSEClient(url string) *SSEClient {
	return &SSEClient{
		url:        url,
		httpClient: &http.Client{Timeout: 0},
	}
}

func (c *SSEClient) Connect(ctx context.Context, out chan<- SSEEvent, errs chan<- error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.url, nil)
	if err != nil {
		log.Printf("sse: build request failed: %v", err)
		errs <- err
		return
	}
	req.Header.Set("Accept", "text/event-stream")
	req.Header.Set("Cache-Control", "no-cache")

	log.Printf("sse: connecting url=%s", c.url)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Printf("sse: connect error url=%s err=%v", c.url, err)
		errs <- err
		return
	}
	defer resp.Body.Close()

	log.Printf("sse: connected url=%s status=%d", c.url, resp.StatusCode)

	if resp.StatusCode != http.StatusOK {
		log.Printf("sse: unexpected status url=%s status=%d", c.url, resp.StatusCode)
		errs <- &HTTPError{StatusCode: resp.StatusCode}
		return
	}

	for ev, err := range sse.Read(resp.Body, nil) {
		if err != nil {
			if ctx.Err() == nil {
				log.Printf("sse: read error: %v", err)
				errs <- err
			}
			return
		}

		data := ev.Data
		preview := data
		if len(preview) > 512 {
			preview = preview[:512] + "..."
		}
		log.Printf("sse: event=%q size=%d preview=%s", ev.Type, len(data), preview)

		out <- SSEEvent{
			Event: ev.Type,
			Data:  []byte(data),
		}
	}

	log.Printf("sse: disconnected url=%s", c.url)
}

type HTTPError struct {
	StatusCode int
}

func (e *HTTPError) Error() string {
	return "unexpected HTTP status"
}
