package middleware

import (
	"context"
	"sync"

	"github.com/Eorthus/shorturl/internal/storage"
	"go.uber.org/zap"
)

// URLDeleter управляет асинхронным удалением URL
type URLDeleter struct {
	store  storage.Storage
	logger *zap.Logger
}

// NewURLDeleter создает новый экземпляр URLDeleter
func NewURLDeleter(store storage.Storage, logger *zap.Logger) *URLDeleter {
	return &URLDeleter{
		store:  store,
		logger: logger,
	}
}

// DeleteURLs асинхронно удаляет указанные URL
func (ud *URLDeleter) DeleteURLs(ctx context.Context, shortIDs []string, userID string) error {
	ud.logger.Info("Starting URL deletion", zap.Strings("shortIDs", shortIDs), zap.String("userID", userID))

	const batchSize = 100
	numWorkers := 5

	batches := make(chan []string)
	results := make(chan error)

	go ud.generateBatches(shortIDs, batchSize, batches)

	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go ud.worker(ctx, userID, batches, results, &wg)
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var lastErr error
	for err := range results {
		if err != nil {
			ud.logger.Error("Error marking URLs as deleted", zap.Error(err))
			lastErr = err
		}
	}

	if lastErr != nil {
		return lastErr
	}

	ud.logger.Info("All URLs processed successfully")
	return nil
}

var batchPool = sync.Pool{
	New: func() interface{} {
		return make([]string, 0, 100) // Предполагаемый размер batch
	},
}

func (ud *URLDeleter) generateBatches(shortIDs []string, batchSize int, batches chan<- []string) {
	for i := 0; i < len(shortIDs); i += batchSize {
		end := i + batchSize
		if end > len(shortIDs) {
			end = len(shortIDs)
		}

		batch := batchPool.Get().([]string)
		batch = batch[:0] // Сбрасываем длину слайса
		batch = append(batch, shortIDs[i:end]...)

		batches <- batch
	}
	close(batches)
}

func (ud *URLDeleter) worker(ctx context.Context, userID string, batches <-chan []string, results chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()
	for batch := range batches {
		select {
		case <-ctx.Done():
			results <- ctx.Err()
			return
		default:
			err := ud.store.MarkURLsAsDeleted(ctx, batch, userID)
			results <- err
			batchPool.Put(batch) // Возвращаем batch в пул
		}
	}
}
