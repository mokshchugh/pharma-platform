package questdb

import (
	"context"
	"log"
	"net"
	"sync"
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

	flushBuf chan []models.Sample
	freeBuf  chan []models.Sample

	wg sync.WaitGroup
}

func NewWriter(
	client *Client,
	table string,
	samples <-chan models.Sample,
) *Writer {
	w := &Writer{
		client:   client,
		table:    table,
		samples:  samples,
		flushBuf: make(chan []models.Sample, 2),
		freeBuf:  make(chan []models.Sample, 3),
	}
	for i := 0; i < 3; i++ {
		w.freeBuf <- make([]models.Sample, 0, client.cfg.BatchSize)
	}
	return w
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

	w.wg.Add(2)
	go w.accumulate(ctx)
	go w.flushLoop()

	return nil
}

func (w *Writer) Stop() error {
	w.wg.Wait()
	return w.client.Close()
}

func (w *Writer) accumulate(ctx context.Context) {
	defer func() {
		close(w.flushBuf)
		w.wg.Done()
	}()

	buffer := <-w.freeBuf

	flushTick := time.NewTicker(w.client.cfg.FlushInterval)
	defer flushTick.Stop()

	for {
		select {
		case <-ctx.Done():
			if len(buffer) > 0 {
				w.flushBuf <- buffer
			}
			return

		case sample, ok := <-w.samples:
			if !ok {
				if len(buffer) > 0 {
					w.flushBuf <- buffer
				}
				return
			}

			buffer = append(buffer, sample)

			if len(buffer) >= w.client.cfg.BatchSize {
				w.flushBuf <- buffer

				select {
				case buffer = <-w.freeBuf:
				default:
					buffer = make([]models.Sample, 0, w.client.cfg.BatchSize)
				}
			}

		case <-flushTick.C:
			if len(buffer) == 0 {
				continue
			}
			w.flushBuf <- buffer

			select {
			case buffer = <-w.freeBuf:
			default:
				buffer = make([]models.Sample, 0, w.client.cfg.BatchSize)
			}
		}
	}
}

// maxPendingBatches bounds how many failed batches flushLoop will hold
// onto and keep retrying during a sustained QuestDB outage, so a long
// outage degrades to bounded memory use (and clearly-logged data loss)
// instead of growing without limit.
const maxPendingBatches = 5

func (w *Writer) flushLoop() {
	defer w.wg.Done()

	var pending [][]models.Sample

	for buf := range w.flushBuf {
		pending = append(pending, buf)

		remaining := pending[:0]
		for _, b := range pending {
			if err := w.writeBatch(b); err != nil {
				remaining = append(remaining, b)
				continue
			}

			select {
			case w.freeBuf <- b[:0]:
			default:
			}
		}
		pending = remaining

		if len(pending) > 0 {
			log.Printf("questdb: %d batch(es) pending retry after write failure", len(pending))
		}

		for len(pending) > maxPendingBatches {
			dropped := pending[0]
			pending = pending[1:]
			log.Printf("questdb: dropping %d samples after sustained write failure (retry backlog exceeded %d batches)", len(dropped), maxPendingBatches)
		}
	}
}

// writeBatch encodes and writes one batch, retrying once via a fresh
// connection on failure. It returns an error (rather than swallowing it)
// so flushLoop can requeue the batch for another attempt instead of
// silently dropping it on the first sustained failure.
func (w *Writer) writeBatch(buf []models.Sample) error {
	if len(buf) == 0 {
		return nil
	}

	data := encode(w.table, buf)

	if err := writeAll(w.client.conn, []byte(data)); err != nil {
		if rerr := w.client.reconnect(context.Background()); rerr != nil {
			return rerr
		}
		if err := writeAll(w.client.conn, []byte(data)); err != nil {
			return err
		}
	}

	writeCount.Add(uint64(len(buf)))
	return nil
}
