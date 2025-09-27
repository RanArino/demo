#!/bin/bash

# Integration test script for ms_canvas Go-Python communication
# This script sets up the environment and runs integration tests

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GO_APP_DIR="$PROJECT_ROOT/go_app"
PYTHON_APP_DIR="$PROJECT_ROOT/python_app"

echo -e "${YELLOW}Starting ms_canvas Integration Tests${NC}"
echo "Project root: $PROJECT_ROOT"

# Function to check if Python dependencies are installed
check_python_deps() {
    echo -e "${YELLOW}Checking Python dependencies...${NC}"
    cd "$PYTHON_APP_DIR"

    if [ ! -f "pyproject.toml" ] && [ ! -f "requirements.txt" ]; then
        echo -e "${RED}Error: Neither pyproject.toml nor requirements.txt found in $PYTHON_APP_DIR${NC}"
        exit 1
    fi

    # Check if virtual environment exists
    if [ ! -d "venv" ] && [ ! -d ".venv" ]; then
        echo -e "${YELLOW}Creating Python virtual environment...${NC}"
        python3 -m venv venv
    fi

    # Activate virtual environment
    if [ -d "venv" ]; then
        source venv/bin/activate
    elif [ -d ".venv" ]; then
        source .venv/bin/activate
    fi

    # Install dependencies
    echo -e "${YELLOW}Installing Python dependencies...${NC}"
    if [ -f "pyproject.toml" ]; then
        pip install -e . > /dev/null 2>&1 || {
            echo -e "${RED}Failed to install Python dependencies from pyproject.toml${NC}"
            exit 1
        }
    elif [ -f "requirements.txt" ]; then
        pip install -r requirements.txt > /dev/null 2>&1 || {
            echo -e "${RED}Failed to install Python dependencies from requirements.txt${NC}"
            exit 1
        }
    fi

    echo -e "${GREEN}Python dependencies OK${NC}"
}

# Function to check if Go dependencies are ready
check_go_deps() {
    echo -e "${YELLOW}Checking Go dependencies...${NC}"
    cd "$GO_APP_DIR"

    if [ ! -f "go.mod" ]; then
        echo -e "${RED}Error: go.mod not found in $GO_APP_DIR${NC}"
        exit 1
    fi

    # Ensure dependencies are downloaded
    go mod download > /dev/null 2>&1 || {
        echo -e "${RED}Failed to download Go dependencies${NC}"
        exit 1
    }

    echo -e "${GREEN}Go dependencies OK${NC}"
}

# Function to run the integration tests
run_integration_tests() {
    echo -e "${YELLOW}Running Integration Tests...${NC}"
    cd "$GO_APP_DIR"

    # Set environment variables for integration testing
    export INTEGRATION_TESTS=1
    export CANVAS_PY_HOST=localhost
    export CANVAS_PY_PORT=50054
    export CANVAS_CHUNK_TARGET_TOKENS=300
    export CANVAS_CHUNK_OVERLAP_PERCENT=10
    export CANVAS_TOKENIZER=tiktoken:cl100k_base
    export CANVAS_BATCH_SIZE=50

    # Run the integration tests
    echo -e "${YELLOW}Running: go test ./internal/gateway/python/... -run TestPythonServiceIntegration -v${NC}"
    go test ./internal/gateway/python/... -run TestPythonServiceIntegration -v -timeout 60s

    local test_result=$?

    if [ $test_result -eq 0 ]; then
        echo -e "${GREEN}✅ Integration tests PASSED${NC}"
    else
        echo -e "${RED}❌ Integration tests FAILED${NC}"
        exit 1
    fi
}

# Function to run unit tests as well
run_unit_tests() {
    echo -e "${YELLOW}Running Unit Tests...${NC}"
    cd "$GO_APP_DIR"

    echo -e "${YELLOW}Running: go test ./internal/gateway/python/... -v -short${NC}"
    go test ./internal/gateway/python/... -v -short

    local test_result=$?

    if [ $test_result -eq 0 ]; then
        echo -e "${GREEN}✅ Unit tests PASSED${NC}"
    else
        echo -e "${RED}❌ Unit tests FAILED${NC}"
        exit 1
    fi
}

# Function to show usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Options:"
    echo "  --unit-only    Run only unit tests (with mocks)"
    echo "  --integration-only  Run only integration tests (with real Python service)"
    echo "  --help         Show this help message"
    echo ""
    echo "Default: Run both unit and integration tests"
}

# Parse command line arguments
UNIT_ONLY=false
INTEGRATION_ONLY=false

while [[ $# -gt 0 ]]; do
    case $1 in
        --unit-only)
            UNIT_ONLY=true
            shift
            ;;
        --integration-only)
            INTEGRATION_ONLY=true
            shift
            ;;
        --help)
            show_usage
            exit 0
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            show_usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo -e "${YELLOW}=== ms_canvas Go-Python Communication Tests ===${NC}"

    # Check dependencies
    check_go_deps
    check_python_deps

    # Run tests based on options
    if [ "$UNIT_ONLY" = true ]; then
        run_unit_tests
    elif [ "$INTEGRATION_ONLY" = true ]; then
        run_integration_tests
    else
        # Run both
        run_unit_tests
        echo ""
        run_integration_tests
    fi

    echo ""
    echo -e "${GREEN}🎉 All tests completed successfully!${NC}"
    echo -e "${GREEN}✅ Go app can successfully communicate with Python app${NC}"
}

# Run main function
main "$@"