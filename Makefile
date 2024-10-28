-include .env

test-controller:
	@go test github.com/BenFaruna/url-shortener/internal/controller -v

test-database:
	@go test github.com/BenFaruna/url-shortener/internal/database -v
test-api:
	@go test github.com/BenFaruna/url-shortener/internal/api -v
run:
	@swag init && go build && ./url-shortener