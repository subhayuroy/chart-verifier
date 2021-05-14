
default: bin

.PHONY: all
all:  gomod_tidy gofmt bin test

.PHONY: gomod_tidy
gomod_tidy:
	 go mod tidy

.PHONY: gofmt
gofmt:
	go fmt -x ./...

.PHONY: bin
bin:
	 hack/build.sh

.PHONY: test
test:
	 go test -v ./...


.PHONY: build-image
build-image:
	hack/build-image.sh
