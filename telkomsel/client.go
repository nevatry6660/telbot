package telkomsel

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/sardanioss/httpcloak/client"

	"telkomsel-bot/config"
	"telkomsel-bot/model"
	"telkomsel-bot/util"
)

type Client struct {
	http *client.Client
}

func NewClient() *Client {
	c := client.NewClient(config.ChromePreset)
	return &Client{http: c}
}

func (c *Client) Close() {
	c.http.Close()
}

type apiResponse struct {
	Status        string          `json:"status"`
	Message       string          `json:"message"`
	TransactionID string          `json:"transaction_id"`
	Data          json.RawMessage `json:"data"`
}

func (c *Client) doRequest(ctx context.Context, method, endpoint string, session *model.Session, body interface{}) (*apiResponse, error) {
	url := config.BaseURL + endpoint

	for attempt := 0; attempt < config.MaxRetries; attempt++ {
		headers := util.BuildHeaders(
			session.AccessAuth,
			session.Authorization,
			session.FullPhone,
			session.XDevice,
			session.WebAppVersion,
		)

		var resp *client.Response
		var err error

		switch method {
		case "POST":
			bodyBytes, marshalErr := json.Marshal(body)
			if marshalErr != nil {
				return nil, fmt.Errorf("marshal body for %s: %w", endpoint, marshalErr)
			}
			resp, err = c.http.Post(ctx, url, bytes.NewReader(bodyBytes), headers)
		case "GET":
			resp, err = c.http.Get(ctx, url, headers)
		default:
			return nil, fmt.Errorf("unsupported method: %s", method)
		}

		if err != nil {
			if attempt < config.MaxRetries-1 {
				time.Sleep(backoff(attempt, 3*time.Second))
				continue
			}
			return nil, fmt.Errorf("%s %s failed: %w", method, endpoint, err)
		}

		respBody, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			return nil, fmt.Errorf("read response from %s: %w", endpoint, readErr)
		}

		if resp.StatusCode == 401 {
			return nil, ErrUnauthorized
		}

		if resp.StatusCode == 429 {
			if attempt < config.MaxRetries-1 {
				time.Sleep(backoff(attempt, 5*time.Second))
				continue
			}
			return nil, fmt.Errorf("rate limited (429) on %s", endpoint)
		}

		if resp.StatusCode != 200 {
			if attempt < config.MaxRetries-1 {
				time.Sleep(backoff(attempt, 3*time.Second))
				continue
			}
			return nil, fmt.Errorf("HTTP %d from %s: %s", resp.StatusCode, endpoint, string(respBody))
		}

		var apiResp apiResponse
		if unmarshalErr := json.Unmarshal(respBody, &apiResp); unmarshalErr != nil {
			if attempt < config.MaxRetries-1 {
				time.Sleep(3 * time.Second)
				continue
			}
			return nil, fmt.Errorf("invalid JSON from %s: %w", endpoint, unmarshalErr)
		}

		return &apiResp, nil
	}

	return nil, fmt.Errorf("max retries exceeded for %s %s", method, endpoint)
}

func (c *Client) doGet(ctx context.Context, endpoint string, session *model.Session) (*apiResponse, error) {
	return c.doRequest(ctx, "GET", endpoint, session, nil)
}

func (c *Client) doPost(ctx context.Context, endpoint string, session *model.Session, body interface{}) (*apiResponse, error) {
	return c.doRequest(ctx, "POST", endpoint, session, body)
}

func backoff(attempt int, base time.Duration) time.Duration {
	return time.Duration(1<<uint(attempt)) * base
}

var ErrUnauthorized = fmt.Errorf("unauthorized: token expired")
