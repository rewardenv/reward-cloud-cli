.DEFAULT_GOAL 	= help

SHELL         	= bash
project       	= reward
GIT_AUTHOR    	= janosmiko

help: ## Outputs this help screen
	@grep -E '(^[\/a-zA-Z0-9_-]+:.*?##.*$$)|(^##)' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}{printf "\033[32m%-30s\033[0m %s\n", $$1, $$2}' | sed -e 's/\[32m##/[33m/'

# If the first argument is "gen"...
ifeq (gen,$(firstword $(MAKECMDGOALS)))
  # use the rest as arguments for "run"
  GEN_ARGS := $(wordlist 2,$(words $(MAKECMDGOALS)),$(MAKECMDGOALS))
  # ...and turn them into do-nothing targets
  $(eval $(GEN_ARGS):;@:)
endif

## —— Commands —————————————————————————————————————————————————————————————————
build: ## Build the command to ./dist
	go mod download
	go generate ./...
	CGO_ENABLED=0 go build -ldflags="-s -w" -o dist/reward-cloud ./cmd/cloud/main.go

package: ## Build the binaries and packages using goreleaser (without releasing it)
	goreleaser --clean --snapshot --skip-publish

build-local: ## Build the binaries only using goreleaser (without releasing it)
	goreleaser --clean --snapshot --skip-publish --config .local.goreleaser.yml

## —— Go Commands ——————————————————————————————————————————————————————————————
gomod: ## Update Go Dependencies
	go mod tidy

lint: ## Lint Go Code
	golangci-lint run ./...

lint-fix: ## Lint Go Code
	golangci-lint run --fix ./...

test: ## Run Go tests
	go test -race -v ./...

## —— API Generation Commands ——————————————————————————————————————————————————
api-remove-jsonld: ## Remove JSON-LD files
	docker run --rm -v "$${PWD}:/app" -w /app python:3 /bin/bash -c 'pip3 install -r hack/requirements.txt && python3 hack/main.py --input pkg/api/reward-cloud-api.yaml --output pkg/api/reward-cloud-api-no-jsonld.yaml'

.PHONY: generate
api-generate: ## Generate Go code from OpenAPI spec
	#docker run --rm -v "$${PWD}/pkg/api/:/app" -w /app openapitools/openapi-generator-cli:v7.4.0 generate -i reward-cloud-api.yaml -c config.yaml -g go --global-property apiTests=false,apiDocs=false,modelTests=false,modelDocs=false -o rewardcloud
	docker run --rm -v "$${PWD}/pkg/api/:/app" -w /app openapitools/openapi-generator-cli:v7.0.1 generate -i reward-cloud-api-no-jsonld.yaml -c config.yaml -g go --global-property apiTests=false,apiDocs=false,modelTests=false,modelDocs=false -o rewardcloud
	rm -f pkg/api/rewardcloud/go.mod && rm -f pkg/api/rewardcloud/go.sum
	go fmt ./pkg/api/...

.PHONY: all
api: ## Generate Go code from OpenAPI spec
	make api-remove-jsonld
	make api-generate
