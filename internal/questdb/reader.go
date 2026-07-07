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

func (r *Reader) LatestByPLC(
	ctx context.Context,
	plcID string,
) ([]models.Sample, error) {
	query := fmt.Sprintf(
		`SELECT * FROM plc_samples WHERE plc_id = '%s' LATEST ON timestamp PARTITION BY tag_id`,
		plcID,
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return decodeSamples(result.Dataset)
}

func (r *Reader) LatestByPLCAndTag(
	ctx context.Context,
	plcID string,
	tagID string,
) (*models.Sample, error) {
	query := fmt.Sprintf(`
SELECT timestamp, plc_id, tag_id, value, quality
FROM plc_samples
WHERE plc_id = '%s' AND tag_id = '%s'
ORDER BY timestamp DESC
LIMIT 1`,
		plcID,
		tagID,
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(result.Dataset) == 0 {
		return nil, nil
	}

	samples, err := decodeSamples(result.Dataset)
	if err != nil {
		return nil, err
	}

	return &samples[0], nil
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

type AggregateSample struct {
	Timestamp time.Time `json:"timestamp"`
	Avg       float64   `json:"avg"`
	Min       float64   `json:"min"`
	Max       float64   `json:"max"`
}

func (r *Reader) Aggregate(
	ctx context.Context,
	plcID string,
	tagID string,
	interval string,
	start time.Time,
	end time.Time,
) ([]AggregateSample, error) {
	query := fmt.Sprintf(`
SELECT timestamp, avg(value), min(value), max(value)
FROM plc_samples
WHERE plc_id = '%s'
  AND tag_id = '%s'
  AND timestamp BETWEEN '%s' AND '%s'
SAMPLE BY %s`,
		plcID,
		tagID,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
		interval,
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return decodeAggregate(result.Dataset)
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

func decodeAggregate(
	dataset [][]any,
) ([]AggregateSample, error) {
	samples := make([]AggregateSample, 0, len(dataset))

	for _, row := range dataset {
		if len(row) != 4 {
			continue
		}

		timestamp, err := time.Parse(time.RFC3339Nano, row[0].(string))
		if err != nil {
			return nil, err
		}

		samples = append(samples, AggregateSample{
			Timestamp: timestamp,
			Avg:       row[1].(float64),
			Min:       row[2].(float64),
			Max:       row[3].(float64),
		})
	}

	return samples, nil
}
