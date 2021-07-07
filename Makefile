GOOS=darwin
GOARCH=amd64

.PHONY: build
b: build
build: test
	mkdir -p _bin
	env GOOS=${GOOS} GOARCH=${GOARCH} GOARM=6 go build -ldflags="-s -w" -o _bin/s3.${GOOS}.${GOARCH} github.com/razzo-lunare/s3/cmd/s3

.PHONY: build-l
build-l: GOOS=linux
build-l: GOARCH=amd64
build-l: build

.PHONY: build-a
build-a: GOOS=linux
build-a: GOARCH=arm
build-a: build

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
