package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"relai/internal/config"
	"relai/internal/schema"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKrakenClient is a mock implementation of the Kraken client
type mockKrakenClient struct {
	response schema.LTPResponse
	err      error
}

func (m *mockKrakenClient) GetLTP() (schema.LTPResponse, error) {
	return m.response, m.err
}

func TestLTPEndpoint(t *testing.T) {
	tests := []struct {
		name           string
		mockResponse   schema.LTPResponse
		mockError      error
		expectedStatus int
		expectedBody   schema.LTPResponse
		expectError    bool
	}{
		{
			name: "successful response with all pairs",
			mockResponse: schema.LTPResponse{
				LTP: []schema.LTPItem{
					{Pair: "BTC/USD", Amount: "52000.12"},
					{Pair: "BTC/EUR", Amount: "50000.12"},
					{Pair: "BTC/CHF", Amount: "49000.12"},
				},
			},
			expectedStatus: http.StatusOK,
			expectedBody: schema.LTPResponse{
				LTP: []schema.LTPItem{
					{Pair: "BTC/USD", Amount: "52000.12"},
					{Pair: "BTC/EUR", Amount: "50000.12"},
					{Pair: "BTC/CHF", Amount: "49000.12"},
				},
			},
		},
		{
			name:           "kraken client error",
			mockError:      fmt.Errorf("kraken API error"),
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a new mock client
			mockClient := &mockKrakenClient{
				response: tt.mockResponse,
				err:      tt.mockError,
			}

			// Create a test server with the mock client
			cfg, err := config.LoadConfig("")
			require.NoError(t, err)
			server := NewServer(cfg)
			server.client = mockClient

			// Create a test request
			w := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp", nil)

			// Serve the request
			server.Engine.ServeHTTP(w, req)

			// Assert status code
			assert.Equal(t, tt.expectedStatus, w.Code)

			// Parse and verify response
			if tt.expectError {
				var errorResponse struct {
					Error string `json:"error"`
				}
				err := json.Unmarshal(w.Body.Bytes(), &errorResponse)
				require.NoError(t, err)
				assert.NotEmpty(t, errorResponse.Error)
			} else {
				var response schema.LTPResponse
				err := json.Unmarshal(w.Body.Bytes(), &response)
				require.NoError(t, err)
				assert.Equal(t, tt.expectedBody, response)

				assert.Len(t, response.LTP, 3)
				for _, item := range response.LTP {
					assert.Contains(t, []string{"BTC/USD", "BTC/EUR", "BTC/CHF"}, item.Pair)
					assert.Regexp(t, `^\d+\.\d{2}$`, item.Amount)
				}
			}
		})
	}
}

func TestServerRateLimiting(t *testing.T) {
	tests := []struct {
		name              string
		requestsPerMinute int
		burstSize         int
		cacheSize         int
		requestSequence   []struct {
			ip       string
			wait     time.Duration
			expected int
		}
	}{
		{
			name:              "basic rate limiting",
			requestsPerMinute: 60,
			burstSize:         2,
			cacheSize:         1000,
			requestSequence: []struct {
				ip       string
				wait     time.Duration
				expected int
			}{
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.1", wait: 0, expected: http.StatusTooManyRequests},
				{ip: "192.168.1.1", wait: time.Second, expected: http.StatusOK},
			},
		},
		{
			name:              "multiple IPs",
			requestsPerMinute: 60,
			burstSize:         2,
			cacheSize:         1000,
			requestSequence: []struct {
				ip       string
				wait     time.Duration
				expected int
			}{
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.2", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.2", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.1", wait: 0, expected: http.StatusTooManyRequests},
				{ip: "192.168.1.2", wait: 0, expected: http.StatusTooManyRequests},
			},
		},
		{
			name:              "cache eviction",
			requestsPerMinute: 60,
			burstSize:         2,
			cacheSize:         2, // Only store 2 IPs
			requestSequence: []struct {
				ip       string
				wait     time.Duration
				expected int
			}{
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.2", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.3", wait: 0, expected: http.StatusOK}, // Should evict 192.168.1.1
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK}, // Should work as IP was evicted
			},
		},
		{
			name:              "rate reset after wait",
			requestsPerMinute: 60,
			burstSize:         1,
			cacheSize:         1000,
			requestSequence: []struct {
				ip       string
				wait     time.Duration
				expected int
			}{
				{ip: "192.168.1.1", wait: 0, expected: http.StatusOK},
				{ip: "192.168.1.1", wait: 0, expected: http.StatusTooManyRequests},
				{ip: "192.168.1.1", wait: time.Second, expected: http.StatusOK},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup config
			cfg := &config.Config{
				Port:                  "8080",
				KrakenAPIBaseEndpoint: "http://test.example.com",
				RateLimit: config.RateLimit{
					RequestsPerMinute: tt.requestsPerMinute,
					BurstSize:         tt.burstSize,
					IPTrackingTTL:     time.Hour,
					CacheSize:         tt.cacheSize,
				},
			}

			// Setup mock client
			mockClient := &mockKrakenClient{
				response: schema.LTPResponse{
					LTP: []schema.LTPItem{
						{Pair: "BTC/USD", Amount: "52000.12"},
					},
				},
			}

			server := NewServer(cfg)
			server.client = mockClient

			// Helper function to make requests
			makeRequest := func(ip string) *httptest.ResponseRecorder {
				w := httptest.NewRecorder()
				req := httptest.NewRequest(http.MethodGet, "/api/v1/ltp", nil)
				req.Header.Set("X-Real-IP", ip)
				server.Engine.ServeHTTP(w, req)
				return w
			}

			// Execute request sequence
			for i, req := range tt.requestSequence {
				if req.wait > 0 {
					time.Sleep(req.wait)
				}

				w := makeRequest(req.ip)
				assert.Equal(t, req.expected, w.Code,
					"Request %d failed: expected status %d, got %d",
					i, req.expected, w.Code)

				// Verify headers for successful requests
				if w.Code == http.StatusOK {
					remaining := w.Header().Get("X-RateLimit-Remaining")
					assert.NotEmpty(t, remaining,
						"Missing rate limit header for request %d", i)
				}

				// Verify error response for rate limited requests
				if w.Code == http.StatusTooManyRequests {
					var response struct {
						Error      string        `json:"error"`
						RetryAfter time.Duration `json:"retry_after"`
					}
					err := json.NewDecoder(w.Body).Decode(&response)
					assert.NoError(t, err)
					assert.NotEmpty(t, response.Error)
					assert.Greater(t, response.RetryAfter, time.Duration(0))
				}
			}
		})
	}
}
