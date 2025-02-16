

# Run tests
test:
	go test -v ./...

# Build the image
build:
	go build -o relai-api ./main.go

# Build the image
docker-build:
	docker build -t relai-api .

# Run the container
docker-run:
	docker run -p 8080:8080 relai-api