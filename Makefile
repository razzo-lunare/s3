GOPATH := $(shell go env GOPATH)

GOOS=darwin
GOARCH=amd64

.PHONY: build
b: build
build: test
	mkdir -p _bin
	env GOOS=${GOOS} GOARCH=${GOARCH} GOARM=6 go build -ldflags="-s -w" -o _bin/s3.${GOOS}.${GOARCH} github.com/razzo-lunare/s3/cmd/s3

.PHONY: install
install: test
	mkdir -p _bin
	env GOOS=${GOOS} GOARCH=${GOARCH} GOARM=6 go install -ldflags="-s -w" github.com/razzo-lunare/s3/cmd/s3

.PHONY: build-linux
build-linux: GOOS=linux
build-linux: GOARCH=amd64
build-linux: build

.PHONY: build-arm
build-arm: GOOS=linux
build-arm: GOARCH=arm
build-arm: build

.PHONY: build-darwin
build-darwin: GOOS=darwin
build-darwin: GOARCH=amd64
build-darwin: build

.PHONY: build-windows
build-windows: GOOS=windows
build-windows: GOARCH=amd64
build-windows: build

.PHONY: clean
c: clean
clean:
	rm -rf _bin

.PHONY: test
t: test
test: fmt vet lint
	go test ./...

.PHONY: fmt
f: fmt
fmt:
	go fmt ./...

.PHONY: vet
v: vet
vet:
	go vet ./...

.PHONY: lint
lint: $(GOPATH)/bin/golint
	$(GOPATH)/bin/golint -set_exit_status ./...

.PHONY: test-all
ta: test-all
test-all:
	go test ./...

# Generic function to install a go package
# go_install,path
define go_install
	go install $(1)
endef

$(GOPATH)/bin/golint:
	$(call go_install,golang.org/x/lint/golint@latest)
