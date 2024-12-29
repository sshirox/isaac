.PHONY: build
build:
	cd cmd/agent && go build -o agent main.go
	cd cmd/server && go build -o server main.go

.PHONY: tidy
tidy:
	go mod tidy

.PHONY: fmt
fmt:
	goimports -local "github.com/sshirox/isaac" -d -w $$(find . -type f -name '*.go' -not -path "*_mock.go")