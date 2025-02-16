# Bitcoin Last Traded Price (LTP) API

A Go-based REST API that retrieves the Last Traded Price (LTP) of Bitcoin for USD, CHF, and EUR currency pairs using the Kraken public API.

## Features

- **Real-time Bitcoin Prices**: Fetches latest BTC prices for USD, CHF, and EUR
- **Rate Limiting**: IP-based rate limiting with configurable requests per minute
- **Caching**: LRU cache for rate limiter state management
- **Configuration**: YAML-based configuration for easy deployment
- **Docker Support**: Containerized application for consistent deployment
- **Clean Architecture**: Separation of concerns with modular design

## API Endpoint

### Get Last Traded Prices

```
GET /api/v1/ltp
```

Example Response:
```json
{
    "ltp": [
        {
            "pair": "BTC/CHF",
            "amount": "49000.12"
        },
        {
            "pair": "BTC/EUR",
            "amount": "50000.12"
        },
        {
            "pair": "BTC/USD",
            "amount": "52000.12"
        }
    ]
}
```

## Configuration

The application is configured via `internal/config/config.yaml`:

```yaml
port: "8080"
kraken_api_base_endpoint: "https://api.kraken.com/0/public/Ticker"
log_level: "info"
rate_limit:
  requests_per_minute: 60
  burst_size: 5
  ip_tracking_ttl: "1h"
  cache_size: 10000
```

## Running the Application

### Using Docker

1. Build the image:
```bash
make docker-build
```

2. Run the container:
```bash
make docker-run
```

### Without Docker

1. Build the binary:
```bash
make build
```

2. Run the tests:
```bash
make test
```

## Testing the API

Once running, you can test the API using curl:

```bash
curl -X GET -H "Accept: application/json" http://localhost:8080/api/v1/ltp
```

## Rate Limiting

The API implements IP-based rate limiting:
- Default: 60 requests per minute
- Burst size: 5 requests
- Response headers include `X-RateLimit-Remaining`
- Returns 429 Too Many Requests when limit is exceeded

## Project Structure

```
.
├── Dockerfile              # Multi-stage Docker build
├── Makefile               # Build and run commands
├── internal/
│   ├── api/              # API handlers and middleware
│   ├── client/           # Kraken API client
│   ├── config/           # Configuration
│   └── schema/           # Data models
└── main.go               # Application entry point
```

## Future Improvements

TODO list for potential enhancements:

1. **API Enhancement**
   - [ ] Add API versioning strategy
   - [ ] Add OpenAPI/Swagger documentation
   - [ ] Add health check endpoint

2. **Performance**
   - [ ] Implement response caching
   - [ ] Add circuit breaker for Kraken API calls
   - [ ] Implement request timeout handling
   - [ ] Add metrics collection (Prometheus)

3. **Security**
   - [ ] Add CORS configuration
   - [ ] Implement API key authentication
   - [ ] Add request signing
   - [ ] Add SSL/TLS support

5. **CI/CD**
   - [ ] Set up GitHub Actions workflow
   - [ ] Add automated testing
   - [ ] Implement automated deployments
   - [ ] Add version tagging
