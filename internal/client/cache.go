package client

import (
	"fmt"
	"relai/internal/schema"
	"sync"
	"time"
)

const (
	cacheTTL = 5 * time.Second
)

type CachedResponse struct {
	data      schema.LTPResponse
	timestamp time.Time
	mu        sync.Mutex
}

func NewCachedResponse() *CachedResponse {
	return &CachedResponse{
		data:      schema.LTPResponse{},
		timestamp: time.Now(),
	}
}

func (c *CachedResponse) Get() (schema.LTPResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if time.Since(c.timestamp) > cacheTTL {
		return schema.LTPResponse{}, fmt.Errorf("cache expired")
	}
	return c.data, nil
}

func (c *CachedResponse) Set(data schema.LTPResponse) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.data = data
	c.timestamp = time.Now()
}
