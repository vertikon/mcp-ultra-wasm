.PHONY: lint test coverage-html mocks

lint:
	golangci-lint run --timeout=5m

test:
	go test ./... -count=1

coverage-html:
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage HTML em coverage.html"

mocks:
	bash scripts/regenerate_mocks.sh