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

echo "Building radar..."
CGO_ENABLED=0 go build -ldflags="-s -w" -o radar .

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

# Scenario 1: Root + superuser
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 1: Root + superuser${NC}"
echo "========================================"
./radar -h localhost -d testdb -U postgres -vv
ZIP1=$(ls -t radar-*.zip | head -1)
PG_COUNT1=$(unzip -l "$ZIP1" | grep -c "postgresql/" || true)
SYS_COUNT1=$(unzip -l "$ZIP1" | grep -c "system/" || true)
STATVIZ1=$(unzip -l "$ZIP1" | grep -c "pg_statviz/" || true)
echo "  PostgreSQL: $PG_COUNT1, System: $SYS_COUNT1, pg_statviz: $STATVIZ1"
if [ "$PG_COUNT1" -lt 25 ] || [ "$SYS_COUNT1" -lt 50 ] || [ "$STATVIZ1" -eq 0 ]; then
	echo -e "${RED}✗ Scenario 1 FAILED${NC}"
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
PG_COUNT2=$(unzip -l "$ZIP2" | grep -c "postgresql/" || true)
SYS_COUNT2=$(unzip -l "$ZIP2" | grep -c "system/" || true)
STATVIZ2=$(unzip -l "$ZIP2" | grep -c "pg_statviz/" || true)
echo "  PostgreSQL: $PG_COUNT2, System: $SYS_COUNT2, pg_statviz: $STATVIZ2"
if [ "$PG_COUNT2" -lt 20 ] || [ "$SYS_COUNT2" -lt 50 ] || [ "$STATVIZ2" -eq 0 ]; then
	echo -e "${RED}✗ Scenario 2 FAILED${NC}"
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
PG_COUNT3=$(unzip -l "$ZIP3" | grep -c "postgresql/" || true)
SYS_COUNT3=$(unzip -l "$ZIP3" | grep -c "system/" || true)
STATVIZ3=$(unzip -l "$ZIP3" | grep -c "pg_statviz/" || true)
echo "  PostgreSQL: $PG_COUNT3, System: $SYS_COUNT3, pg_statviz: $STATVIZ3"
if [ "$PG_COUNT3" -lt 25 ] || [ "$SYS_COUNT3" -eq 0 ] || [ "$STATVIZ3" -eq 0 ]; then
	echo -e "${RED}✗ Scenario 3 FAILED${NC}"
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
PG_COUNT4=$(unzip -l "$ZIP4" | grep -c "postgresql/" || true)
SYS_COUNT4=$(unzip -l "$ZIP4" | grep -c "system/" || true)
STATVIZ4=$(unzip -l "$ZIP4" | grep -c "pg_statviz/" || true)
echo "  PostgreSQL: $PG_COUNT4, System: $SYS_COUNT4, pg_statviz: $STATVIZ4"
if [ "$PG_COUNT4" -lt 20 ] || [ "$SYS_COUNT4" -eq 0 ] || [ "$STATVIZ4" -eq 0 ]; then
	echo -e "${RED}✗ Scenario 4 FAILED${NC}"
	exit 1
fi
echo -e "${GREEN}✓ Scenario 4 PASSED${NC}"
rm -f "$ZIP4"

echo ""
echo "Stopping PostgreSQL..."
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D /var/lib/postgresql/18/main stop"

echo ""
echo -e "${GREEN}All 4 scenarios passed!${NC}"
