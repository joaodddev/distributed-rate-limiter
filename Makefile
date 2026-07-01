.PHONY: run test bench up down

up:
	docker-compose up -d

down:
	docker-compose down

test:
	go test ./... -v

bench:
	go test ./bench/... -bench=. -benchmem -cpuprofile=cpu.prof -memprofile=mem.prof

run:
	go run ./cmd/server

test-integration:
	docker-compose up -d
	go test ./... -tags=integration -v
	docker-compose down
