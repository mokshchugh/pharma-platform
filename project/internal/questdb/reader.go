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

type MachineTimestamp struct {
	MachineID string
	Timestamp time.Time
}

func (r *Reader) LatestTimestamps(ctx context.Context) ([]MachineTimestamp, error) {
	result, err := r.Query(ctx,
		`SELECT machine_id, max(timestamp) FROM plc_samples GROUP BY machine_id`,
	)
	if err != nil {
		return nil, err
	}

	rows := make([]MachineTimestamp, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 2 {
			continue
		}
		ts, err := time.Parse(time.RFC3339Nano, row[1].(string))
		if err != nil {
			continue
		}
		rows = append(rows, MachineTimestamp{
			MachineID: row[0].(string),
			Timestamp: ts,
		})
	}

	return rows, nil
}

func (r *Reader) Latest(
	ctx context.Context,
) ([]models.Sample, error) {
	result, err := r.Query(
		ctx,
		`SELECT * FROM plc_samples LATEST ON timestamp PARTITION BY machine_id, tag_name`,
	)
	if err != nil {
		return nil, err
	}

	return decodeSamples(result.Dataset)
}

func (r *Reader) LatestByPLC(
	ctx context.Context,
	machineID string,
) ([]models.Sample, error) {
	query := fmt.Sprintf(
		`SELECT * FROM plc_samples WHERE machine_id = '%s' LATEST ON timestamp PARTITION BY tag_name`,
		machineID,
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return decodeSamples(result.Dataset)
}

func (r *Reader) LatestByPLCAndTag(
	ctx context.Context,
	machineID string,
	tagName string,
) (*models.Sample, error) {
	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, value, quality
FROM plc_samples
WHERE machine_id = '%s' AND tag_name = '%s'
ORDER BY timestamp DESC
LIMIT 1`,
		machineID,
		tagName,
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
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
) ([]models.Sample, error) {
	query := fmt.Sprintf(`
SELECT *
FROM plc_samples
WHERE machine_id = '%s'
  AND tag_name = '%s'
  AND timestamp BETWEEN '%s' AND '%s'
ORDER BY timestamp;
`,
		machineID,
		tagName,
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

// Stream types for the data-stream page

type RawRow struct {
	Timestamp   string  `json:"timestamp"`
	MachineID   string  `json:"machine_id"`
	MachineName string  `json:"machine_name"`
	TagName     string  `json:"tag_name"`
	Value       float64 `json:"value"`
	Quality     int     `json:"quality"`
}

type AggregateRow struct {
	Timestamp   string  `json:"timestamp"`
	MachineID   string  `json:"machine_id"`
	MachineName string  `json:"machine_name"`
	TagName     string  `json:"tag_name"`
	MinValue    float64 `json:"min_value"`
	MaxValue    float64 `json:"max_value"`
	AvgValue    float64 `json:"avg_value"`
	SampleCount int64   `json:"sample_count"`
}

type StreamResponse struct {
	Data       any    `json:"data"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	PageSize   int    `json:"page_size"`
	Resolution string `json:"resolution"`
}

func (r *Reader) StreamRawAll(
	ctx context.Context,
	table string,
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
) ([]RawRow, error) {
	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, value, quality
FROM %s
WHERE 1=1`, table)

	if machineID != "" {
		query += fmt.Sprintf(" AND machine_id = '%s'", machineID)
	}
	if tagName != "" {
		query += fmt.Sprintf(" AND tag_name = '%s'", tagName)
	}
	query += fmt.Sprintf(" AND timestamp BETWEEN '%s' AND '%s'",
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)
	query += " ORDER BY timestamp DESC, machine_id ASC, tag_name ASC"

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rows := make([]RawRow, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 6 {
			continue
		}
		rows = append(rows, RawRow{
			Timestamp:   row[0].(string),
			MachineID:   row[1].(string),
			MachineName: row[2].(string),
			TagName:     row[3].(string),
			Value:       row[4].(float64),
			Quality:     int(row[5].(float64)),
		})
	}

	return rows, nil
}

func (r *Reader) StreamAggregateAll(
	ctx context.Context,
	table string,
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
) ([]AggregateRow, error) {
	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, min_value, max_value, avg_value, sample_count
FROM %s
WHERE 1=1`, table)

	if machineID != "" {
		query += fmt.Sprintf(" AND machine_id = '%s'", machineID)
	}
	if tagName != "" {
		query += fmt.Sprintf(" AND tag_name = '%s'", tagName)
	}
	query += fmt.Sprintf(" AND timestamp BETWEEN '%s' AND '%s'",
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)
	query += " ORDER BY timestamp DESC, machine_id ASC, tag_name ASC"

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rows := make([]AggregateRow, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 8 {
			continue
		}
		rows = append(rows, AggregateRow{
			Timestamp:   row[0].(string),
			MachineID:   row[1].(string),
			MachineName: row[2].(string),
			TagName:     row[3].(string),
			MinValue:    row[4].(float64),
			MaxValue:    row[5].(float64),
			AvgValue:    row[6].(float64),
			SampleCount: int64(row[7].(float64)),
		})
	}

	return rows, nil
}

func (r *Reader) StreamRaw(
	ctx context.Context,
	table string,
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
	page int,
	pageSize int,
) (*StreamResponse, error) {
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, value, quality
FROM %s
WHERE 1=1`, table)

	if machineID != "" {
		query += fmt.Sprintf(" AND machine_id = '%s'", machineID)
	}
	if tagName != "" {
		query += fmt.Sprintf(" AND tag_name = '%s'", tagName)
	}
	query += fmt.Sprintf(" AND timestamp BETWEEN '%s' AND '%s'",
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)
	query += " ORDER BY timestamp DESC, machine_id ASC, tag_name ASC"
	query += fmt.Sprintf(" LIMIT %d, %d", offset, pageSize)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rows := make([]RawRow, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 6 {
			continue
		}
		rows = append(rows, RawRow{
			Timestamp:   row[0].(string),
			MachineID:   row[1].(string),
			MachineName: row[2].(string),
			TagName:     row[3].(string),
			Value:       row[4].(float64),
			Quality:     int(row[5].(float64)),
		})
	}

	total, err := r.countStream(ctx, table, machineID, tagName, start, end)
	if err != nil {
		return nil, err
	}

	return &StreamResponse{
		Data:       rows,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		Resolution: "raw",
	}, nil
}

func (r *Reader) StreamAggregate(
	ctx context.Context,
	table string,
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
	page int,
	pageSize int,
	resolution string,
) (*StreamResponse, error) {
	offset := (page - 1) * pageSize

	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, min_value, max_value, avg_value, sample_count
FROM %s
WHERE 1=1`, table)

	if machineID != "" {
		query += fmt.Sprintf(" AND machine_id = '%s'", machineID)
	}
	if tagName != "" {
		query += fmt.Sprintf(" AND tag_name = '%s'", tagName)
	}
	query += fmt.Sprintf(" AND timestamp BETWEEN '%s' AND '%s'",
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)
	query += " ORDER BY timestamp DESC, machine_id ASC, tag_name ASC"
	query += fmt.Sprintf(" LIMIT %d, %d", offset, pageSize)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rows := make([]AggregateRow, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 8 {
			continue
		}
		rows = append(rows, AggregateRow{
			Timestamp:   row[0].(string),
			MachineID:   row[1].(string),
			MachineName: row[2].(string),
			TagName:     row[3].(string),
			MinValue:    row[4].(float64),
			MaxValue:    row[5].(float64),
			AvgValue:    row[6].(float64),
			SampleCount: int64(row[7].(float64)),
		})
	}

	total, err := r.countStream(ctx, table, machineID, tagName, start, end)
	if err != nil {
		return nil, err
	}

	return &StreamResponse{
		Data:       rows,
		Total:      total,
		Page:       page,
		PageSize:   pageSize,
		Resolution: resolution,
	}, nil
}

func (r *Reader) countStream(
	ctx context.Context,
	table string,
	machineID string,
	tagName string,
	start time.Time,
	end time.Time,
) (int64, error) {
	query := fmt.Sprintf("SELECT count() FROM %s WHERE 1=1", table)

	if machineID != "" {
		query += fmt.Sprintf(" AND machine_id = '%s'", machineID)
	}
	if tagName != "" {
		query += fmt.Sprintf(" AND tag_name = '%s'", tagName)
	}
	query += fmt.Sprintf(" AND timestamp BETWEEN '%s' AND '%s'",
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return 0, err
	}

	if len(result.Dataset) == 0 || len(result.Dataset[0]) == 0 {
		return 0, nil
	}

	return int64(result.Dataset[0][0].(float64)), nil
}

func (r *Reader) Aggregate(
	ctx context.Context,
	machineID string,
	tagName string,
	interval string,
	start time.Time,
	end time.Time,
) ([]AggregateSample, error) {
	table := "plc_samples_" + interval
	query := fmt.Sprintf(`
SELECT timestamp, avg_value, min_value, max_value
FROM %s
WHERE machine_id = '%s'
  AND tag_name = '%s'
  AND timestamp BETWEEN '%s' AND '%s'`,
		table,
		machineID,
		tagName,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
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
		if len(row) != 6 {
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
				Timestamp:   timestamp,
				MachineID:   row[1].(string),
				MachineName: row[2].(string),
				TagName:     row[3].(string),
				Value:       row[4].(float64),
				Quality: models.Quality(
					uint8(row[5].(float64)),
				),
			},
		)
	}

	return samples, nil
}

func (r *Reader) LatestFromView(
	ctx context.Context,
	machineID string,
	resolution string,
	tagName string,
) (*AggregateRow, error) {
	table := "plc_samples_" + resolution
	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, min_value, max_value, avg_value, sample_count
FROM %s
WHERE machine_id = '%s' AND tag_name = '%s'
ORDER BY timestamp DESC
LIMIT 1`,
		table, machineID, tagName,
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	if len(result.Dataset) == 0 {
		return nil, nil
	}

	row := result.Dataset[0]
	if len(row) != 8 {
		return nil, nil
	}

	return &AggregateRow{
		Timestamp:   row[0].(string),
		MachineID:   row[1].(string),
		MachineName: row[2].(string),
		TagName:     row[3].(string),
		MinValue:    row[4].(float64),
		MaxValue:    row[5].(float64),
		AvgValue:    row[6].(float64),
		SampleCount: int64(row[7].(float64)),
	}, nil
}

func (r *Reader) SeriesFromView(
	ctx context.Context,
	machineID string,
	resolution string,
	tagName string,
	start time.Time,
	end time.Time,
) ([]AggregateRow, error) {
	table := "plc_samples_" + resolution
	query := fmt.Sprintf(`
SELECT timestamp, machine_id, machine_name, tag_name, min_value, max_value, avg_value, sample_count
FROM %s
WHERE machine_id = '%s' AND tag_name = '%s'
  AND timestamp BETWEEN '%s' AND '%s'
ORDER BY timestamp ASC`,
		table, machineID, tagName,
		start.UTC().Format(time.RFC3339Nano),
		end.UTC().Format(time.RFC3339Nano),
	)

	result, err := r.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	rows := make([]AggregateRow, 0, len(result.Dataset))
	for _, row := range result.Dataset {
		if len(row) != 8 {
			continue
		}
		rows = append(rows, AggregateRow{
			Timestamp:   row[0].(string),
			MachineID:   row[1].(string),
			MachineName: row[2].(string),
			TagName:     row[3].(string),
			MinValue:    row[4].(float64),
			MaxValue:    row[5].(float64),
			AvgValue:    row[6].(float64),
			SampleCount: int64(row[7].(float64)),
		})
	}

	return rows, nil
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
