package client

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetLTP(t *testing.T) {
	// Mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
{
  "error": [],
  "result": {
    "XXBTZEUR": {
      "a": [
        "93200.00000",
        "1",
        "1.000"
      ],
      "b": [
        "93199.90000",
        "6",
        "6.000"
      ],
      "c": [
        "93200.00000",
        "0.00010571"
      ],
      "v": [
        "93.41559906",
        "106.62849385"
      ],
      "p": [
        "93067.03469",
        "93029.51199"
      ],
      "t": [
        11280,
        12081
      ],
      "l": [
        "92755.90000",
        "92500.00000"
      ],
      "h": [
        "93409.60000",
        "93409.60000"
      ],
      "o": "92952.00000"
    }
  }
}
        `))
	}))
	defer server.Close()

	client := NewKrakenClient(server.URL)
	response, err := client.fetchTickerPrice("XXBTZEUR")
	require.NoError(t, err)

	assert.Equal(t, "93200.00000", response)
}

func TestFetchTickerPriceWithRetries(t *testing.T) {
	// Track number of attempts and timestamps
	attempts := 0
	var requestTimes []time.Time

	// Mock server that fails twice then succeeds
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestTimes = append(requestTimes, time.Now())
		attempts++

		if attempts <= 2 {
			// Simulate failure for first two attempts
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		// Succeed on third attempt
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`
{
  "error": [],
  "result": {
    "XXBTZEUR": {
      "c": ["93200.00000", "0.00010571"]
    }
  }
}
		`))
	}))
	defer server.Close()

	client := NewKrakenClient(server.URL)
	client.maxRetries = 3
	client.backoffBase = 100 * time.Millisecond // Shorter duration for testing

	// Make the request
	response, err := client.fetchTickerPrice("XXBTZEUR")
	require.NoError(t, err)
	assert.Equal(t, "93200.00000", response)

	// Verify number of attempts
	assert.Equal(t, 3, attempts, "Expected exactly 3 attempts")

	// Verify backoff timing
	require.Equal(t, 3, len(requestTimes), "Expected 3 request timestamps")

	// Check intervals between requests
	// First retry should be after ~100ms (backoffBase * 2^0)
	interval1 := requestTimes[1].Sub(requestTimes[0])
	assert.GreaterOrEqual(t, interval1, 100*time.Millisecond)
	assert.Less(t, interval1, 150*time.Millisecond) // Allow some buffer

	// Second retry should be after ~200ms (backoffBase * 2^1)
	interval2 := requestTimes[2].Sub(requestTimes[1])
	assert.GreaterOrEqual(t, interval2, 200*time.Millisecond)
	assert.Less(t, interval2, 250*time.Millisecond) // Allow some buffer
}

func TestFetchTickerPriceMaxRetriesExceeded(t *testing.T) {
	// Mock server that always fails
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewKrakenClient(server.URL)
	client.maxRetries = 2
	client.backoffBase = 100 * time.Millisecond // Shorter duration for testing

	// Make the request
	_, err := client.fetchTickerPrice("XXBTZEUR")

	// Verify error
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed after 2 attempts")
}
