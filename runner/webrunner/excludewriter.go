package webrunner

import (
	"context"
	"log"
	"sync/atomic"

	"github.com/gosom/google-maps-scraper/gmaps"
	"github.com/gosom/scrapemate"
)

// excludeWriter drops entries whose place_id appears in exclude set.
type excludeWriter struct {
	inner   scrapemate.ResultWriter
	exclude map[string]struct{}
	written atomic.Int64
	skipped atomic.Int64
}

func newExcludeWriter(inner scrapemate.ResultWriter, exclude map[string]struct{}) *excludeWriter {
	if exclude == nil {
		exclude = map[string]struct{}{}
	}

	return &excludeWriter{
		inner:   inner,
		exclude: exclude,
	}
}

func (w *excludeWriter) Run(ctx context.Context, in <-chan scrapemate.Result) error {
	filtered := make(chan scrapemate.Result)

	go func() {
		defer close(filtered)

		for result := range in {
			entry, isEntry := result.Data.(*gmaps.Entry)
			if isEntry && entry.PlaceID != "" {
				if _, skip := w.exclude[entry.PlaceID]; skip {
					w.skipped.Add(1)

					continue
				}
			}

			w.written.Add(1)

			select {
			case <-ctx.Done():
				return
			case filtered <- result:
			}
		}
	}()

	err := w.inner.Run(ctx, filtered)
	log.Printf("exclude filter: wrote %d, skipped %d duplicates", w.written.Load(), w.skipped.Load())

	return err
}
