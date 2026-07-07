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

func (w *Writer) flushLoop() {
	defer w.wg.Done()

	for buf := range w.flushBuf {
		w.flushBuffer(buf)

		select {
		case w.freeBuf <- buf[:0]:
		default:
		}
	}
}

func (w *Writer) flushBuffer(buf []models.Sample) {
	if len(buf) == 0 {
		return
	}

	data := encode(w.table, buf)

	if err := writeAll(w.client.conn, []byte(data)); err != nil {
		if err := w.client.reconnect(context.Background()); err != nil {
			log.Printf("questdb flush error: %v", err)
			return
		}
		if err := writeAll(w.client.conn, []byte(data)); err != nil {
			log.Printf("questdb flush error: %v", err)
			return
		}
	}

	writeCount.Add(uint64(len(buf)))
}
