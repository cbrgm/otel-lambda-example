GOOS            := linux
GOARCH          := arm64
BUILD_DIR       := bin

GO_PREFIX       := CGO_ENABLED=0 GOFLAGS=-mod=mod
GO              := $(GO_PREFIX) go
PACKAGES        := $(shell $(GO) list ./... | grep -v node_modules)

.PHONY: build
build:
	env GOOS=$(GOOS) GOARCH=$(GOARCH) $(GO) build -ldflags="-s -w" -o bin/bootstrap

.PHONY: build-bootstrap
build-bootstrap: clean build
	zip --quiet --junk-paths bin/lambda.zip bin/bootstrap collector.yaml

.PHONY: vendor
vendor:
	rm -rf vendor/
	go mod vendor

.PHONY: diff
diff:
	cdk diff

.PHONY: synth
synth: build
	cdk synth

.PHONY: deploy
deploy: clean build-bootstrap
	cdk deploy --all

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)/

