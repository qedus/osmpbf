export GORACE := halt_on_error=1

all: cover

race:
	go test -v -race

cover:
	go test -v -coverprofile=profile.cov

bench:
	env OSMPBF_BENCHMARK_BUFFER=1048576  go test -v -run=NONE -bench=. -benchmem -benchtime=10s | tee 01.txt
	env OSMPBF_BENCHMARK_BUFFER=33554432 go test -v -run=NONE -bench=. -benchmem -benchtime=10s | tee 32.txt
	benchcmp 01.txt 32.txt
