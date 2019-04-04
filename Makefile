APP = galaxy
BUILD_DIR = build

.PHONY: bootstrap build

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

test: FORCE
	go test -cover -v pkg/$(APP)/*

integration:
	go test -v $(E2E_TEST_DIR)/*

FORCE: ;