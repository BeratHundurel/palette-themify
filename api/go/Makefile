.PHONY: test bench test-one

# Run all tests with verbose output
test:
	go test -v ./...

# Run a specific test by name: make test-one name=TestName
test-one:
	go test -v -run $(name)
