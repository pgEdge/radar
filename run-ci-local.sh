#!/bin/bash
set -e
set -o pipefail

# Create timestamped log file
LOGFILE="ci-$(date +%Y%m%d-%H%M%S).log"

# Function to log to both stdout and file
log() {
    echo "$@" | tee -a "$LOGFILE"
}

log "=== Local CI/CD Test ==="
log "Log file: $LOGFILE"
log ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Format check
echo -e "${YELLOW}Step 1: Checking Go code formatting${NC}" | tee -a "$LOGFILE"
if [ -n "$(gofmt -l .)" ]; then
    echo -e "${RED}✗ Go code is not formatted. Run 'gofmt -w .' to fix:${NC}" | tee -a "$LOGFILE"
    gofmt -d . | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Code formatting check passed${NC}" | tee -a "$LOGFILE"
log ""

# Step 2: Linting (required)
echo -e "${YELLOW}Step 2: Linting Go code${NC}" | tee -a "$LOGFILE"
# Add ~/go/bin to PATH for golangci-lint
export PATH="$HOME/go/bin:$PATH"
if ! command -v golangci-lint &> /dev/null; then
    echo -e "${RED}✗ golangci-lint not found${NC}" | tee -a "$LOGFILE"
    echo "  Install with: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest" | tee -a "$LOGFILE"
    exit 1
fi
if ! golangci-lint run --timeout=5m 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Linting failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Linting passed${NC}" | tee -a "$LOGFILE"
log ""

# Step 3: Unit tests
echo -e "${YELLOW}Step 3: Running unit tests${NC}" | tee -a "$LOGFILE"
if ! go test -v ./... 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Unit tests failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Unit tests passed${NC}" | tee -a "$LOGFILE"
log ""

# Step 4: Build
echo -e "${YELLOW}Step 4: Building radar binary${NC}" | tee -a "$LOGFILE"
if ! CGO_ENABLED=0 go build -ldflags="-s -w" -o radar . 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Build failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}" | tee -a "$LOGFILE"
log ""

# Step 5: Integration test with Docker
echo -e "${YELLOW}Step 5: Running integration test with Docker${NC}" | tee -a "$LOGFILE"
echo "Building Docker image..." | tee -a "$LOGFILE"

if ! docker build -t radar-local-test . 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Docker build failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi

echo "Running integration test container..." | tee -a "$LOGFILE"
if ! docker run --rm radar-local-test 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Integration test failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi

echo -e "${GREEN}✓ Integration test passed${NC}" | tee -a "$LOGFILE"
log ""

# Cleanup
echo -e "${YELLOW}Cleaning up Docker images${NC}" | tee -a "$LOGFILE"
docker rmi radar-local-test 2>&1 | tee -a "$LOGFILE" || true

log ""
echo -e "${GREEN}=== All CI/CD tests passed! ===${NC}" | tee -a "$LOGFILE"
