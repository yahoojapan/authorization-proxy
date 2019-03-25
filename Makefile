GO_VERSION:=$(shell go version)

.PHONY: bench profile clean test

all: install

bench:
	go test -count=5 -run=NONE -bench . -benchmem

profile: clean
	mkdir pprof
	mkdir bench
	go test -count=10 -run=NONE -bench . -benchmem -o pprof/test.bin -cpuprofile pprof/cpu.out -memprofile pprof/mem.out
	go tool pprof --svg pprof/test.bin pprof/mem.out > bench/mem.svg
	go tool pprof --svg pprof/test.bin pprof/cpu.out > bench/cpu.svg

clean:
	rm -rf bench
	rm -rf pprof
	rm -rf ./*.svg
	rm -rf ./*.log

test:
	GOCACHE=off go test --race -coverprofile=cover.out ./...

docker-push:
	sudo docker build --pull=true --file=Dockerfile -t yahoojapan/authorization-proxy:latest .
	sudo docker push yahoojapan/authorization-proxy:latest

coverage:
	go test -v -race -covermode=atomic -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out
