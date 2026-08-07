#!/bin/bash
# Local CI — mirrors .github/workflows/ci.yml step for step.
# If you change this file, change ci.yml to match (and vice versa).
set -e
set -o pipefail

# Tool versions — MUST match .github/workflows/ci.yml
GO_VERSION="1.25.12"             # ci.yml: actions/setup-go with go-version '1.25'
GOLANGCI_LINT_VERSION="v2.12.2"  # ci.yml: install.sh -b $(go env GOPATH)/bin v2.12.2
TOOLDIR=".ci-tools"

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

# Step 0: Pin the Go toolchain to the version CI uses
# GOTOOLCHAIN makes any Go >= 1.21 fetch and use exactly this toolchain, so the
# compiler, stdlib and gofmt all match CI regardless of the system Go version.
echo -e "${YELLOW}Step 0: Setting up Go ${GO_VERSION}${NC}" | tee -a "$LOGFILE"
export GOTOOLCHAIN="go${GO_VERSION}"
if ! go version 2>&1 | tee -a "$LOGFILE" | grep -q "go${GO_VERSION}"; then
    echo -e "${RED}✗ Could not select Go ${GO_VERSION} (GOTOOLCHAIN=$GOTOOLCHAIN)${NC}" | tee -a "$LOGFILE"
    exit 1
fi
# Use the pinned toolchain's gofmt, not whatever gofmt is on PATH
GOFMT="$(go env GOROOT)/bin/gofmt"
echo -e "${GREEN}✓ Using Go ${GO_VERSION}${NC}" | tee -a "$LOGFILE"
log ""

# Step 1: Format check
echo -e "${YELLOW}Step 1: Checking Go code formatting${NC}" | tee -a "$LOGFILE"
if [ -n "$("$GOFMT" -l .)" ]; then
    echo -e "${RED}✗ Go code is not formatted. Run 'gofmt -w .' to fix:${NC}" | tee -a "$LOGFILE"
    "$GOFMT" -d . | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Code formatting check passed${NC}" | tee -a "$LOGFILE"
log ""

# Step 2: Linting (required)
# Pinned to the same version CI installs. A different golangci-lint will not
# read this repo's v2-schema .golangci.yml, so never fall back to $PATH.
echo -e "${YELLOW}Step 2: Linting Go code${NC}" | tee -a "$LOGFILE"
LINT="$TOOLDIR/golangci-lint"
if ! [ -x "$LINT" ] || ! "$LINT" --version 2>/dev/null | grep -q "${GOLANGCI_LINT_VERSION#v}"; then
    echo "Installing golangci-lint ${GOLANGCI_LINT_VERSION} into ${TOOLDIR}/..." | tee -a "$LOGFILE"
    mkdir -p "$TOOLDIR"
    # Fetch the installer from the pinned tag, not master: the script itself is
    # then immutable, and it sha256-verifies the binary it downloads.
    if ! curl -sSfL "https://raw.githubusercontent.com/golangci/golangci-lint/${GOLANGCI_LINT_VERSION}/install.sh" \
        | sh -s -- -b "$TOOLDIR" "$GOLANGCI_LINT_VERSION" 2>&1 | tee -a "$LOGFILE"; then
        echo -e "${RED}✗ Failed to install golangci-lint ${GOLANGCI_LINT_VERSION}${NC}" | tee -a "$LOGFILE"
        exit 1
    fi
fi
if ! "$LINT" run --timeout=5m 2>&1 | tee -a "$LOGFILE"; then
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
# Same command as ci.yml; the binary is what the Dockerfile COPYs, mirroring
# CI's upload-artifact/download-artifact handoff into the integration jobs.
echo -e "${YELLOW}Step 4: Building radar binary${NC}" | tee -a "$LOGFILE"
if ! CGO_ENABLED=0 go build -ldflags="-s -w" -o radar . 2>&1 | tee -a "$LOGFILE"; then
    echo -e "${RED}✗ Build failed${NC}" | tee -a "$LOGFILE"
    exit 1
fi
echo -e "${GREEN}✓ Build successful${NC}" | tee -a "$LOGFILE"
log ""

# Step 5: Integration test with Docker
# Covers the same ground as ci.yml's integration-test, cert-auth-test and
# gssapi-auth-test jobs: test-radar.sh runs all 6 scenarios (4 permission
# combinations + certificate auth + GSSAPI/Kerberos).
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
log ""

# Report what CI covers that this run did not, so a local pass is not mistaken
# for a full CI pass.
echo -e "${YELLOW}Not covered locally (runs in GitHub CI only):${NC}" | tee -a "$LOGFILE"
if [ "$(uname -s)" != "Darwin" ]; then
    log "  - macos-test job: macOS build, pg_statviz install, collection and archive verification"
fi
log "  - CI runs the 4 permission scenarios in 4 isolated containers; this script"
log "    runs all 6 scenarios sequentially in one container, sharing a PostgreSQL instance"
