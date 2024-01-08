ALL_PACKAGES=$(shell go list ./... | grep -v "vendor")

fmt:
	go fmt $(ALL_PACKAGES)

vet:
	go vet $(ALL_PACKAGES)

tidy:
	go mod tidy

serve: fmt vet
	go run cmd/*.go
