GOOS=darwin
GOARCH=amd64

.PHONY: build
b: build
build: test
	mkdir -p _bin
	env GOOS=${GOOS} GOARCH=${GOARCH} GOARM=6 go build -ldflags="-s -w" -o _bin/s3.${GOOS}.${GOARCH} github.com/razzo-lunare/s3/cmd/s3

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


.PHONY: clean
c: clean
clean:
	rm -rf _bin

.PHONY: test
t: test
test:
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
lint:
	"$$(go env GOPATH)/bin/golint" -set_exit_status ./...

.PHONY: test-all
ta: test-all
test-all:
	go test ./...
