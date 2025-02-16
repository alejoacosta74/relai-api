package client

import (
	"net/http"
	"net/http/httptest"
	"testing"

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
