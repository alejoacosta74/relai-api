package client

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"relai/internal/schema"

	"github.com/alejoacosta74/go-logger"
)

type KrakenClient struct {
	baseEndpoint string
	httpClient   *http.Client
}

func NewKrakenClient(endpoint string) *KrakenClient {

	return &KrakenClient{
		baseEndpoint: endpoint,
		httpClient:   &http.Client{},
	}
}

func (c *KrakenClient) GetLTP() (schema.LTPResponse, error) {
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

	return response, nil
}

func (c *KrakenClient) fetchTickerPrice(pair string) (string, error) {
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
