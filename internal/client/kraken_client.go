package client

import (
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"relai/internal/schema"
	"time"

	"github.com/alejoacosta74/go-logger"
)

// KrakenClientInterface defines the interface for the Kraken client
type KrakenClientInterface interface {
	GetLTP() (schema.LTPResponse, error)
}

var _ KrakenClientInterface = (*KrakenClient)(nil)

type KrakenClient struct {
	baseEndpoint string
	httpClient   *http.Client
	maxRetries   int
	backoffBase  time.Duration
	cache        *CachedResponse
}

func NewKrakenClient(endpoint string) *KrakenClient {

	return &KrakenClient{
		baseEndpoint: endpoint,
		httpClient:   &http.Client{},
		maxRetries:   3,               // default to 3 retries
		backoffBase:  1 * time.Second, // start with 1 second backoff
		cache:        NewCachedResponse(),
	}
}

func (c *KrakenClient) GetLTP() (schema.LTPResponse, error) {

	// if there is a cached response, return it inmediately
	if data, err := c.cache.Get(); err == nil {
		return data, nil
	}

	// Kraken uses different symbols than what we want to display
	pairs := map[string]string{
		"BTC/USD": "XXBTZUSD",
		"BTC/EUR": "XXBTZEUR",
		"BTC/CHF": "XBTCHF",
	}

	response := schema.LTPResponse{
		LTP: make([]schema.LTPItem, 0, len(pairs)),
	}

	for displayPair, krakenPair := range pairs {
		price, err := c.fetchTickerPrice(krakenPair)
		if err != nil {
			return schema.LTPResponse{}, err
		}
		response.LTP = append(response.LTP, schema.LTPItem{Pair: displayPair, Amount: price})
	}

	// cache the response
	c.cache.Set(response)

	return response, nil
}

func (c *KrakenClient) fetchTickerPrice(pair string) (string, error) {
	var lastError error
	for attempt := 0; attempt < c.maxRetries; attempt++ {
		if attempt > 0 {
			backoffDuration := c.backoffBase * time.Duration(1<<(attempt-1))
			// Jitter the backoff duration to avoid thundering herd
			jitter := time.Duration(rand.Int63n(int64(backoffDuration) / 2)) // 50% jitter
			time.Sleep(backoffDuration + jitter)
			logger.Debugf("Retrying request for pair %s (attempt %d/%d)", pair, attempt, c.maxRetries)
		}

		price, err := c.dofetchTickerPrice(pair)
		if err == nil {
			return price, nil
		}
		logger.Errorf("Error fetching ticker price for pair %s (attempt %d/%d): %v", pair, attempt, c.maxRetries, err)
	}

	return "", fmt.Errorf("failed after %d attempts: %w", c.maxRetries, lastError)

}

func (c *KrakenClient) dofetchTickerPrice(pair string) (string, error) {
	url := fmt.Sprintf("%s?pair=%s", c.baseEndpoint, pair)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var krakenResponse schema.KrakenTickerResponse
	if err := json.Unmarshal(body, &krakenResponse); err != nil {
		return "", err
	}

	logger.Debugf("Kraken response: %+v\n", krakenResponse)

	if len(krakenResponse.Error) > 0 {
		return "", fmt.Errorf("kraken API error: %v", krakenResponse.Error)
	}

	lastTrade := krakenResponse.Result[pair].LastTrade
	if len(lastTrade) == 0 {
		return "", fmt.Errorf("no last trade found for pair: %s", pair)
	}

	return lastTrade[0], nil
}
