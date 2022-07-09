help:				## display help information
	@awk 'BEGIN {FS = ":.*##"; printf "Usage: make \033[36m<target>\033[0m\n"} /^[a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-12s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

check: lint test	## lint + test

lint:	hooks			## fmt vet staticcheck
	go fmt ./...
	go vet ./...
	staticcheck ./...

test: hooks 		## execute tests
	go test ./... -race

hooks: .git/hooks/pre-commit

.git/hooks/pre-commit:
	cp -r .githooks/* .git/hooks/

.PHONY: help check lint test hooks
