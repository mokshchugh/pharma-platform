package questdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Reader struct {
	client *Client
}

func NewReader(client *Client) *Reader {
	return &Reader{
		client: client,
	}
}

type queryResponse struct {
	Dataset [][]any `json:"dataset"`
}

func (r *Reader) Query(
	ctx context.Context,
	query string,
) (*queryResponse, error) {

	endpoint := fmt.Sprintf(
		"http://%s:%d/exec?query=%s",
		r.client.cfg.Host,
		9000,
		url.QueryEscape(query),
	)

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		endpoint,
		nil,
	)
	if err != nil {
		return nil, err
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var result queryResponse

	if err := json.Unmarshal(body, &result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (r *Reader) Latest(
	ctx context.Context,
) (*queryResponse, error) {

	return r.Query(
		ctx,
		`SELECT * FROM plc_samples LATEST ON timestamp PARTITION BY plc_id, tag_id`,
	)
}
