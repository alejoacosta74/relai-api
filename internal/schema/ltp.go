package schema

// LTPResponse represents the response structure for the /api/v1/ltp endpoint
type LTPResponse struct {
	LTP []LTPItem `json:"ltp"`
}

// LTPItem represents a single price item in the response
type LTPItem struct {
	Pair   string `json:"pair"`
	Amount string `json:"amount"`
}

// KrakenTickerResponse represents the response from Kraken API
type KrakenTickerResponse struct {
	Error  []string              `json:"error"`
	Result map[string]TickerInfo `json:"result"`
}

// TickerInfo represents the ticker information for a single pair
type TickerInfo struct {
	LastTrade []string `json:"c"` // c = last trade closed (price, lot volume)
}
