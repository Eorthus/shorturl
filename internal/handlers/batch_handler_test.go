package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleBatchShorten(t *testing.T) {
	r, _ := setupRouter(t)

	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name: "Valid batch",
			requestBody: `[
				{"correlation_id": "1", "original_url": "https://example.com"},
				{"correlation_id": "2", "original_url": "https://example.org"}
			]`,
			expectedStatus: http.StatusCreated,
		},
		{
			name:           "Empty batch",
			requestBody:    `[]`,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid URL in batch",
			requestBody: `[
				{"correlation_id": "1", "original_url": "not-a-url"}
			]`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, err := http.NewRequest("POST", "/api/shorten/batch", bytes.NewBufferString(tt.requestBody))
			require.NoError(t, err)

			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatus, rr.Code, "handler returned wrong status code")

			if tt.expectedStatus == http.StatusCreated {
				var response []BatchResponse
				err := json.Unmarshal(rr.Body.Bytes(), &response)
				require.NoError(t, err, "Failed to unmarshal response")

				assert.Len(t, response, 2, "Expected 2 items in response")
				for _, item := range response {
					assert.NotEmpty(t, item.CorrelationID, "CorrelationID should not be empty")
					assert.True(t, strings.HasPrefix(item.ShortURL, "http://localhost:8080/"),
						"ShortURL should start with base URL")
				}
			}
		})
	}
}
