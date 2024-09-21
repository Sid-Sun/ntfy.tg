ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

tidy:
	go mod tidy

serve: fmt vet
	go run cmd/*.go

build: fmt vet
	CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o bin/ntfy.tg ./cmd
