package middleware

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

func TestURLDeleter_DeleteURLs(t *testing.T) {
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name         string
		shortIDs     []string
		userID       string
		mockBehavior func(*MockStorage)
		expectError  bool
	}{
		{
			name:     "Successful deletion",
			shortIDs: []string{"abc123", "def456", "ghi789"},
			userID:   "user1",
			mockBehavior: func(m *MockStorage) {
				m.On("MarkURLsAsDeleted", mock.Anything, []string{"abc123", "def456", "ghi789"}, "user1").Return(nil)
			},
			expectError: false,
		},
		{
			name:     "Partial failure",
			shortIDs: []string{"abc123", "def456", "ghi789", "jkl012"},
			userID:   "user2",
			mockBehavior: func(m *MockStorage) {
				m.On("MarkURLsAsDeleted", mock.Anything, []string{"abc123", "def456", "ghi789", "jkl012"}, "user2").Return(errors.New("failed to delete some URLs"))
			},
			expectError: true,
		},
		{
			name:     "Empty shortIDs",
			shortIDs: []string{},
			userID:   "user3",
			mockBehavior: func(m *MockStorage) {
				// Не ожидаем вызовов MarkURLsAsDeleted
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockStorage := new(MockStorage)
			tt.mockBehavior(mockStorage)

			deleter := NewURLDeleter(mockStorage, logger)

			err := deleter.DeleteURLs(context.Background(), tt.shortIDs, tt.userID)

			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			mockStorage.AssertExpectations(t)
		})
	}
}

func TestURLDeleter_generateBatches(t *testing.T) {
	deleter := &URLDeleter{}
	shortIDs := []string{"1", "2", "3", "4", "5", "6", "7"}
	batchSize := 3
	batches := make(chan []string)

	go deleter.generateBatches(shortIDs, batchSize, batches)

	var results [][]string
	for batch := range batches {
		results = append(results, batch)
	}

	expected := [][]string{
		{"1", "2", "3"},
		{"4", "5", "6"},
		{"7"},
	}

	assert.Equal(t, expected, results)
}

func TestURLDeleter_worker(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	mockStorage := new(MockStorage)
	deleter := NewURLDeleter(mockStorage, logger)

	ctx := context.Background()
	userID := "testUser"
	batches := make(chan []string, 2)
	results := make(chan error, 2)
	var wg sync.WaitGroup

	batches <- []string{"1", "2", "3"}
	batches <- []string{"4", "5", "6"}
	close(batches)

	mockStorage.On("MarkURLsAsDeleted", mock.Anything, []string{"1", "2", "3"}, userID).Return(nil)
	mockStorage.On("MarkURLsAsDeleted", mock.Anything, []string{"4", "5", "6"}, userID).Return(nil)

	wg.Add(1)
	go deleter.worker(ctx, userID, batches, results, &wg)

	wg.Wait()
	close(results)

	for err := range results {
		assert.NoError(t, err)
	}

	mockStorage.AssertExpectations(t)
}
