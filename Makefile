.PHONY: run test race vet fmt fmt-check build docker check

# Token chỉ dùng khi chạy local; môi trường thật đặt AUTH_TOKEN qua secret store.
run:
	AUTH_TOKEN=$${AUTH_TOKEN:-dev-only-token-change-me} go run .

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

fmt:
	gofmt -w .

fmt-check:
	@out=$$(gofmt -l .); if [ -n "$$out" ]; then echo "chưa gofmt:"; echo "$$out"; exit 1; fi

build:
	CGO_ENABLED=0 go build -trimpath -o bin/app .

docker:
	docker build -t go-modular-template:local .

# Cổng trước khi push: format + vet + test có race.
check: fmt-check vet race
