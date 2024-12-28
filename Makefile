.PHONY: build
build:
	cd cmd/agent && go build -o agent main.go
	cd cmd/server && go build -o server main.go

.PHONY: tidy
tidy:
	go mod tidy
