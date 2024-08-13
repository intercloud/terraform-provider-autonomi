BUILD_DIR = "build"

.DEFAULT_GOAL :=
default: lint test

.PHONY: go-lint
go-lint: # Check Go code
	$(title)
	@golangci-lint run $(GOPACKAGE)
	@echo "$(GRE)OK$(END) Go code checked"

lint:  go-lint              # Check Go code	

.PHONY: go-test
go-test: # Run unit tests and generate coverage report
	$(title)
	@mkdir -p $(BUILD_DIR)
	@go test -coverprofile $(BUILD_DIR)/coverage-unit.out $(GOPACKAGE) || (echo "$(RED)ERROR$(END) unit tests failed"; exit 1)
	@go tool cover -html=$(BUILD_DIR)/coverage-unit.out -o $(BUILD_DIR)/coverage-unit.html
	@echo "Unit test coverage report in $$(pwd)/$(BUILD_DIR)/coverage-unit.html"
	@echo "$(GRE)OK$(END) unit tests passed"

test: go-test # Run unit tests