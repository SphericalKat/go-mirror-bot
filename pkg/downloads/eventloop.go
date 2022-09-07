package downloads

import (
	"context"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

func ConsumeStatusQueue(ctx context.Context, wg *sync.WaitGroup) {
	log.Info().Msg("Starting status queue consumer event loop")
	dlm := GetDownloadManager()
	stop := false
	for !stop {
		select {
		case <-ctx.Done():
			log.Info()
			stop = true
		default:
			consume, ok := dlm.StatusQueue.TryDequeue()
			if ok {
				consume.(func())()
			} else {
				time.Sleep(50 * time.Millisecond)
			}
		}
	}
	wg.Done()
	log.Info().Msg("Gracefully shutting down event loop")
}
