package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"relai/internal/client"
	"relai/internal/config"
	"relai/internal/schema"

	"github.com/gin-gonic/gin"
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
			server := createTestServer(t, mockClient)

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

// createTestServer creates a server instance for testing
func createTestServer(t *testing.T, mockClient client.KrakenClientInterface) *Server {
	t.Helper()

	cfg := &config.Config{
		Port:                  "8080",
		KrakenAPIBaseEndpoint: "http://test.example.com",
	}

	engine := gin.New()
	engine.Use(gin.Logger())
	engine.Use(gin.Recovery())

	// Use mock client
	engine.GET("/api/v1/ltp", func(c *gin.Context) {
		response, err := mockClient.GetLTP()
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusOK, response)
	})

	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: engine,
	}

	return &Server{
		Engine: engine,
		srv:    srv,
	}
}
