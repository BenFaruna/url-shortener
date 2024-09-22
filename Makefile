-include .env

test-controller:
	@go test github.com/BenFaruna/url-shortener/internal/controller -v

test-model:
	@go test github.com/BenFaruna/url-shortener/internal/model -v