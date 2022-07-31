all: test build

.PHONY: test
test:
	go test .

build: build-server build-client

.PHONY: build-server
build-server:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/gshd ./cmd/server

.PHONY: build-client
build-client:
	@mkdir -p bin
	CGO_ENABLED=0 go build -o bin/gsh ./cmd/client

.PHONY: install
install:
	install bin/* ~/.local/bin

.PHONY: gen
gen:
	protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative ./pb/service.proto

.PHONY: clean
clean:
	rm -rf bin
