package questdb

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"pharma-platform/internal/models"
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
) ([]models.Sample, error) {
	result, err := r.Query(
		ctx,
		`SELECT * FROM plc_samples LATEST ON timestamp PARTITION BY plc_id, tag_id`,
	)
	if err != nil {
		return nil, err
	}

	return decodeSamples(result.Dataset)
}

func (r *Reader) History(
	ctx context.Context,
	plcID string,
	tagID string,
	start time.Time,
	end time.Time,
) ([]models.Sample, error) {
	query := fmt.Sprintf(`
SELECT *
FROM plc_samples
WHERE plc_id = '%s'
  AND tag_id = '%s'
  AND timestamp BETWEEN '%s' AND '%s'
ORDER BY timestamp;
`,
		plcID,
		tagID,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return decodeSamples(result.Dataset)
}

func decodeSamples(
	dataset [][]any,
) ([]models.Sample, error) {
	samples := make(
		[]models.Sample,
		0,
		len(dataset),
	)

	for _, row := range dataset {
		if len(row) != 5 {
			return nil, fmt.Errorf(
				"unexpected QuestDB row length: %d",
				len(row),
			)
		}

		timestamp, err := time.Parse(
			time.RFC3339Nano,
			row[0].(string),
		)
		if err != nil {
			return nil, err
		}

		samples = append(
			samples,
			models.Sample{
				Timestamp: timestamp,
				PLCID:     row[1].(string),
				TagID:     row[2].(string),
				Value:     row[3].(float64),
				Quality: models.Quality(
					uint8(row[4].(float64)),
				),
			},
		)
	}

	return samples, nil
}
