-include .env

test-controller:
	@go test github.com/BenFaruna/url-shortener/internal/controller -v

test-model:
	@go test github.com/BenFaruna/url-shortener/internal/model -v
test-api:
	@go test github.com/BenFaruna/url-shortener/internal/api -v
run:
	@go build && ./url-shortener