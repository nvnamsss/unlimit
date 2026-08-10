#!/bin/bash
RED=`tput setaf 1`
GRN=`tput setaf 2`
YLW=`tput setaf 3`
BLU=`tput setaf 4`
PUR=`tput setaf 13`
RESET=`tput sgr0`

# Define path to secrets file
SECRETS_FILE="$PWD/.secrets"

# Function to get a secret from the secrets file or environment
get_secret() {
    local secret_name=$1
    local secret_value=""
    
    # Check if secret is set in environment
    if [ -n "${!secret_name}" ]; then
        secret_value="${!secret_name}"
        echo "Using $secret_name from environment variable."
        return 0
    fi
    
    # If not in environment, try to load from secret file
    if [ -f "$SECRETS_FILE" ]; then
        # Extract the value for the specified key from the secrets file
        secret_value=$(grep "^$secret_name=" "$SECRETS_FILE" | cut -d '=' -f 2- | tr -d '[:space:]')
        if [ -n "$secret_value" ]; then
            # Export the secret to environment for child processes
            export "$secret_name"="$secret_value"
            echo "Using $secret_name from secret file."
            return 0
        fi
    fi
    
    echo "Error: $secret_name is not set in environment or secret file."
    echo "Please set the environment variable with: export $secret_name=your_value"
    echo "Or add it to the secrets file at: $SECRETS_FILE"
    return 1
}

# Get any secrets needed for testing if required
# Example: get_secret "TEST_API_KEY" || true

# Setup directories
echo "${BLU}Setting up test directories...${RESET}"
TEST_RESULT_DIR=./test-results
mkdir -p ${TEST_RESULT_DIR}

# Determine the root directory of the Go module
ROOT_DIR=$(go list -m -f "{{.Dir}}")
if [ -z "$ROOT_DIR" ]; then
    ROOT_DIR="."
fi
cd "$ROOT_DIR"

# Find all directories containing Go files
echo "${BLU}Finding Go modules...${RESET}"
GO_DIRS=$(find . -type f -name "*.go" | grep -v "/vendor/" | grep -v "/.git/" | xargs -n1 dirname | sort -u)


# Run tests with coverage on all modules
echo "${BLU}Running tests with coverage...${RESET}"
echo "----------------"
go test -v -race -covermode=atomic -coverprofile=${TEST_RESULT_DIR}/coverage.out ./... | tee ${TEST_RESULT_DIR}/test.log; 
TEST_EXIT_CODE=${PIPESTATUS[0]}
echo $TEST_EXIT_CODE > ${TEST_RESULT_DIR}/test.out

# Generate coverage report
echo "${BLU}Generating coverage report...${RESET}"
if [ -f "${TEST_RESULT_DIR}/coverage.out" ]; then
    go tool cover -func=${TEST_RESULT_DIR}/coverage.out | tee ${TEST_RESULT_DIR}/coverage_summary.txt
    TOTAL_COVERAGE=$(go tool cover -func=${TEST_RESULT_DIR}/coverage.out | grep total | awk '{print $3}')
else
    echo "${RED}No coverage data was generated${RESET}"
    TOTAL_COVERAGE="0.0%"
fi

# Generate HTML coverage report
if [ -f "${TEST_RESULT_DIR}/coverage.out" ]; then
    echo "${BLU}Generating HTML coverage report...${RESET}"
    go tool cover -html=${TEST_RESULT_DIR}/coverage.out -o ${TEST_RESULT_DIR}/coverage.html
    echo "HTML coverage report generated at: ${TEST_RESULT_DIR}/coverage.html"
fi

# Generate JUnit-style XML report for CI systems
echo "${BLU}Generating JUnit XML report...${RESET}"
cat ${TEST_RESULT_DIR}/test.log | go-junit-report > ${TEST_RESULT_DIR}/report.xml

echo "----------------"
echo "${GRN}Test Results:${RESET}"
echo "${BLU}Total Coverage:${RESET} ${YLW}${TOTAL_COVERAGE}${RESET}"

# Check if any tests failed and provide a summary
if [ $TEST_EXIT_CODE -ne 0 ]; then
    echo "${RED}⨯ Tests failed with exit code: ${TEST_EXIT_CODE}${RESET}"
    # Extract failures from the test output
    echo "${RED}Failed tests:${RESET}"
    grep "FAIL:" ${TEST_RESULT_DIR}/test.log || echo "${GRN}No test failures detected in log output${RESET}"
else
    echo "${GRN}✓ All tests passed${RESET}"
fi

# Print coverage thresholds
COVERAGE_PCT=$(echo $TOTAL_COVERAGE | tr -d '%')
if (( $(echo "$COVERAGE_PCT >= 80" | bc -l) )); then
    echo "${GRN}✓ Coverage threshold met (>= 80%)${RESET}"
elif (( $(echo "$COVERAGE_PCT >= 50" | bc -l) )); then
    echo "${YLW}⚠ Coverage below recommended threshold (>= 80%)${RESET}"
else
    echo "${RED}⨯ Coverage critically low (< 50%)${RESET}"
fi

# Exit with the test exit code
exit $TEST_EXIT_CODE
