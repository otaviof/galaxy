# application name
APP = galaxy
# build directory
BUILD_DIR ?= build
# directory containing end-to-end tests
E2E_TEST_DIR ?= test/e2e
# project version, used as docker tag
VERSION ?= $(shell cat ./version)

.PHONY: bootstrap build test

default: build

dep:
	go get -u github.com/golang/dep/cmd/dep

bootstrap:
	dep ensure -v -vendor-only

build: clean
	go build -v -o $(BUILD_DIR)/$(APP) cmd/$(APP)/*

clean:
	rm -rf $(BUILD_DIR) > /dev/null

clean-vendor:
	rm -rf ./vendor > /dev/null

test:
	go test -race -coverprofile=coverage.txt -covermode=atomic -cover -v pkg/$(APP)/*

integration:
	go test -v $(E2E_TEST_DIR)/*

codecov:
	mkdir .ci || true
	curl -s -o .ci/codecov.sh https://codecov.io/bash
	bash .ci/codecov.sh -t $(CODECOV_TOKEN)

snapshot:
	goreleaser --rm-dist --snapshot

release:
	git tag $(VERSION)
	git push origin $(VERSION)
	goreleaser --rm-dist