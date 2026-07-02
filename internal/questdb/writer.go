package questdb

import (
	"context"
	"time"

	"pharma-platform/internal/models"
)

type Writer struct {
	client  *Client
	table   string
	samples <-chan models.Sample

	buffer []models.Sample
}

func NewWriter(
	client *Client,
	table string,
	samples <-chan models.Sample,
) *Writer {
	return &Writer{
		client:  client,
		table:   table,
		samples: samples,
	}
}

func (w *Writer) Start(ctx context.Context) error {
	if err := w.client.Connect(ctx); err != nil {
		return err
	}

	ticker := time.NewTicker(w.client.cfg.FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return w.flush()

		case sample, ok := <-w.samples:
			if !ok {
				return w.flush()
			}

			w.buffer = append(w.buffer, sample)

			if len(w.buffer) >= w.client.cfg.BatchSize {
				if err := w.flush(); err != nil {
					return err
				}
			}

		case <-ticker.C:
			if len(w.buffer) == 0 {
				continue
			}

			if err := w.flush(); err != nil {
				return err
			}
		}
	}
}

func (w *Writer) Stop() error {
	if err := w.flush(); err != nil {
		return err
	}

	return w.client.Close()
}

func (w *Writer) flush() error {
	if len(w.buffer) == 0 {
		return nil
	}

	if w.client.conn == nil {
		return ErrNotConnected
	}

	data := encode(
		w.table,
		w.buffer,
	)

	if _, err := w.client.conn.Write([]byte(data)); err != nil {
		return err
	}

	w.buffer = w.buffer[:0]

	return nil
}
