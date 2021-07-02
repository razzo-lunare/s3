
APP_NAME="github.com/razzo-lunare/fortuna"

path_to_repo="$${go env GOPATH}/src/${APP_NAME}"

.PHONY: build
b: build
build: test
	mkdir -p _bin
	go build -o _bin/ github.com/razzo-lunare/s3/cmd/s3

.PHONY: build-l
build-l: test
	mkdir -p _bin
	env GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o _bin/ github.com/razzo-lunare/s3/cmd/s3

.PHONY: build-a
build-a: test
	mkdir -p _bin
	env GOOS=linux GOARCH=arm GOARM=6 go build -ldflags="-s -w" -o _bin/ github.com/razzo-lunare/s3/cmd/s3

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
