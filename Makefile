test:
	go test -v -race `go list ./... | grep -v /vendor/`
bench:
	go test --benchmem --bench=.