package questdb

import (
	"context"
	"log"
	"net"
	"sync/atomic"
	"time"

	"pharma-platform/internal/models"
)

var writeCount atomic.Uint64

func writeAll(conn net.Conn, data []byte) error {
	for len(data) > 0 {
		n, err := conn.Write(data)
		if err != nil {
			return err
		}
		data = data[n:]
	}
	return nil
}

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

	go func() {
		metricsTick := time.NewTicker(time.Second)
		defer metricsTick.Stop()

		for {
			select {
			case <-ctx.Done():
				return

			case <-metricsTick.C:
				n := writeCount.Swap(0)
				log.Printf("QuestDB writes: %d samples/sec", n)
			}
		}
	}()

	flushTick := time.NewTicker(w.client.cfg.FlushInterval)
	defer flushTick.Stop()

	for {
		select {
		case <-ctx.Done():
			return w.flush(ctx)

		case sample, ok := <-w.samples:
			if !ok {
				return w.flush(ctx)
			}

			w.buffer = append(w.buffer, sample)

			if len(w.buffer) >= w.client.cfg.BatchSize {
				if err := w.flush(ctx); err != nil {
					log.Printf("questdb flush error: %v", err)
				}
			}

		case <-flushTick.C:
			if len(w.buffer) == 0 {
				continue
			}

			if err := w.flush(ctx); err != nil {
				log.Printf("questdb flush error: %v", err)
			}
		}
	}
}

func (w *Writer) Stop() error {
	if err := w.flush(context.Background()); err != nil {
		return err
	}

	return w.client.Close()
}

func (w *Writer) flush(ctx context.Context) error {
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

	if err := writeAll(w.client.conn, []byte(data)); err != nil {

		if err := w.client.reconnect(ctx); err != nil {
			return err
		}

		if err := writeAll(w.client.conn, []byte(data)); err != nil {
			return err
		}
	}

	writeCount.Add(uint64(len(w.buffer)))

	w.buffer = w.buffer[:0]

	return nil
}
