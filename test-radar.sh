#!/bin/bash
# Test script for radar in a Debian container with PostgreSQL 18
# Tests all 4 permission scenarios:
# 1. Root + superuser
# 2. Root + pg_monitor
# 3. Non-root + superuser
# 4. Non-root + pg_monitor

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Binary is pre-built by run-ci-local.sh and copied into container by Dockerfile
echo "Using pre-built radar binary..."
./radar --help > /dev/null 2>&1 || { echo -e "${RED}✗ radar binary not found or not executable${NC}"; exit 1; }

echo ""
echo "Initializing PostgreSQL 18..."
rm -rf /var/lib/postgresql/18/main
mkdir -p /var/lib/postgresql/18/main
chown -R postgres:postgres /var/lib/postgresql
su - postgres -c "/usr/lib/postgresql/18/bin/initdb -D /var/lib/postgresql/18/main"

echo ""
echo "Starting PostgreSQL 18..."
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D /var/lib/postgresql/18/main -l /var/lib/postgresql/18/logfile start"

echo ""
echo "Waiting for PostgreSQL to start..."
sleep 3

echo ""
echo "Creating test database and users..."
su - postgres -c "/usr/lib/postgresql/18/bin/createdb testdb"
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"CREATE USER testuser WITH PASSWORD 'testpass';\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"GRANT CONNECT ON DATABASE testdb TO testuser;\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"GRANT pg_monitor TO testuser;\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"CREATE EXTENSION pg_statviz;\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"SELECT pgstatviz.snapshot();\""

echo ""
echo "Creating non-root system user..."
useradd -m -s /bin/bash radaruser || true
cp radar /home/radaruser/
chown radaruser:radaruser /home/radaruser/radar

# Helper function to validate ZIP contents
validate_zip() {
	local zip_file="$1"
	local scenario="$2"
	local require_system="$3"  # "yes" or "no"

	local pg_count=$(unzip -l "$zip_file" | grep -c "postgresql/" || true)
	local sys_count=$(unzip -l "$zip_file" | grep -c "system/" || true)
	local statviz_count=$(unzip -l "$zip_file" | grep -c "pg_statviz/" || true)
	local empty_count=$(unzip -l "$zip_file" | awk '$1 == 0 {count++} END {print count+0}')

	echo "  PostgreSQL: $pg_count, System: $sys_count, pg_statviz: $statviz_count, Empty files: $empty_count"

	# Check for empty files (should be 0)
	if [ "$empty_count" -gt 0 ]; then
		echo -e "${RED}✗ $scenario FAILED: Found $empty_count empty files in archive${NC}"
		return 1
	fi

	# Must have PostgreSQL data
	if [ "$pg_count" -eq 0 ]; then
		echo -e "${RED}✗ $scenario FAILED: No PostgreSQL data collected${NC}"
		return 1
	fi

	# System data check (optional for non-root scenarios)
	if [ "$require_system" = "yes" ] && [ "$sys_count" -eq 0 ]; then
		echo -e "${RED}✗ $scenario FAILED: No system data collected${NC}"
		return 1
	fi

	# Must have pg_statviz data
	if [ "$statviz_count" -eq 0 ]; then
		echo -e "${RED}✗ $scenario FAILED: No pg_statviz data collected${NC}"
		return 1
	fi

	return 0
}

# Scenario 1: Root + superuser
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 1: Root + superuser${NC}"
echo "========================================"
./radar -h localhost -d testdb -U postgres -vv
ZIP1=$(ls -t radar-*.zip | head -1)
if ! validate_zip "$ZIP1" "Scenario 1" "yes"; then
	exit 1
fi
echo -e "${GREEN}✓ Scenario 1 PASSED${NC}"
rm -f "$ZIP1"

# Scenario 2: Root + pg_monitor
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 2: Root + pg_monitor${NC}"
echo "========================================"
PGPASSWORD=testpass ./radar -h localhost -d testdb -U testuser -vv
ZIP2=$(ls -t radar-*.zip | head -1)
if ! validate_zip "$ZIP2" "Scenario 2" "yes"; then
	exit 1
fi
echo -e "${GREEN}✓ Scenario 2 PASSED${NC}"
rm -f "$ZIP2"

# Scenario 3: Non-root + superuser
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 3: Non-root + superuser${NC}"
echo "========================================"
su - radaruser -c "cd /home/radaruser && ./radar -h localhost -d testdb -U postgres -vv"
ZIP3=$(su - radaruser -c "ls -t /home/radaruser/radar-*.zip | head -1")
if ! validate_zip "$ZIP3" "Scenario 3" "no"; then
	exit 1
fi
echo -e "${GREEN}✓ Scenario 3 PASSED${NC}"
rm -f "$ZIP3"

# Scenario 4: Non-root + pg_monitor
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 4: Non-root + pg_monitor${NC}"
echo "========================================"
su - radaruser -c "cd /home/radaruser && PGPASSWORD=testpass ./radar -h localhost -d testdb -U testuser -vv"
ZIP4=$(su - radaruser -c "ls -t /home/radaruser/radar-*.zip | head -1")
if ! validate_zip "$ZIP4" "Scenario 4" "no"; then
	exit 1
fi
echo -e "${GREEN}✓ Scenario 4 PASSED${NC}"
rm -f "$ZIP4"

echo ""
echo "Stopping PostgreSQL..."
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D /var/lib/postgresql/18/main stop"

echo ""
echo -e "${GREEN}All 4 scenarios passed!${NC}"
