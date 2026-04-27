.PHONY: test bench test-one test-failures

# Run all tests with better formatting using gotestsum
test:
	go run gotest.tools/gotestsum@latest --format=short-verbose ./...

# Run tests and show only failures
test-failures:
	go run gotest.tools/gotestsum@latest --format=short-verbose ./... | grep -E "FAIL|--- FAIL"

# Run a specific test by name: make test-one name=TestName
test-one:
	go run gotest.tools/gotestsum@latest --format=short-verbose -run $(name) ./...
