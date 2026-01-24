.PHONY: test bench test-one bench-one

# Run all tests with verbose output
test:
	go test -v ./...

# Run all benchmarks with memory allocation stats
bench:
	go test -bench=. -benchmem

# Run a specific test by name: make test-one name=TestName
test-one:
	go test -v -run $(name)

# Run a specific benchmark by name: make bench-one name=BenchmarkName
bench-one:
	go test -bench=$(name) -benchmem
