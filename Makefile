# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFORMAT=gofmt -s -w .
GOVET=$(GOCMD) vet

# Main package path
MAIN_PACKAGE=github.com/upamune/zstate

.PHONY: all test test-with-update lint coverage clean format help

all: test

lint: ## Run go vet
	$(GOVET) ./...

test: ## Run tests
	$(GOTEST) -count=1 -race -shuffle=on -v ./...

test-with-update: ## Run tests with update testdata
	$(GOTEST) -count=1 -race -shuffle=on -v -update ./...

coverage: ## Run tests with coverage
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

clean: ## Clean build files
	$(GOCLEAN)
	rm -f coverage.out

format: ## Format the code
	$(GOFORMAT)

help: ## Display this help screen
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'