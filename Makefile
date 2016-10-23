export GORACE := halt_on_error=1

all: cover

race:
	go test -v -race

cover:
	go test -v -coverprofile=profile.cov

bench:
	go test -v -run=NONE -bench=. -benchmem -benchtime=10s
