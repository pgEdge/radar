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
for i in $(seq 1 30); do
    if su - postgres -c "/usr/lib/postgresql/18/bin/pg_isready -q" 2>/dev/null; then break; fi
    sleep 1
done

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

# Scenario 5: Certificate authentication
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 5: Certificate authentication${NC}"
echo "========================================"

PGDATA=/var/lib/postgresql/18/main
CERTDIR=/tmp/certs
mkdir -p "$CERTDIR"

# CA key and cert
openssl genpkey -algorithm RSA -out "$CERTDIR/ca.key" 2>/dev/null
openssl req -new -x509 -key "$CERTDIR/ca.key" -out "$CERTDIR/ca.crt" -days 1 -subj "/CN=TestCA" 2>/dev/null

# Server key and cert (SAN=localhost — Go requires SANs, not just CN)
openssl genpkey -algorithm RSA -out "$CERTDIR/server.key" 2>/dev/null
openssl req -new -key "$CERTDIR/server.key" -out "$CERTDIR/server.csr" -subj "/CN=localhost" \
    -addext "subjectAltName=DNS:localhost" 2>/dev/null
openssl x509 -req -in "$CERTDIR/server.csr" -CA "$CERTDIR/ca.crt" -CAkey "$CERTDIR/ca.key" \
    -CAcreateserial -out "$CERTDIR/server.crt" -days 1 -copy_extensions copyall 2>/dev/null

# Client key and cert (CN=testuser — must match PG username)
openssl genpkey -algorithm RSA -out "$CERTDIR/client.key" 2>/dev/null
openssl req -new -key "$CERTDIR/client.key" -out "$CERTDIR/client.csr" -subj "/CN=testuser" 2>/dev/null
openssl x509 -req -in "$CERTDIR/client.csr" -CA "$CERTDIR/ca.crt" -CAkey "$CERTDIR/ca.key" \
    -CAcreateserial -out "$CERTDIR/client.crt" -days 1 2>/dev/null

# Set permissions
cp "$CERTDIR/server.crt" "$CERTDIR/server.key" "$CERTDIR/ca.crt" "$PGDATA/"
chown postgres:postgres "$PGDATA/server.crt" "$PGDATA/server.key" "$PGDATA/ca.crt"
chmod 600 "$PGDATA/server.key"
chmod 600 "$CERTDIR/client.key"

# Configure PostgreSQL for SSL + cert auth
su - postgres -c "cat >> $PGDATA/postgresql.conf << 'SSLCONF'
ssl = on
ssl_cert_file = 'server.crt'
ssl_key_file = 'server.key'
ssl_ca_file = 'ca.crt'
SSLCONF"

# Replace pg_hba.conf: cert auth for testuser, trust for postgres (scenarios 1-4 already passed)
su - postgres -c "cat > $PGDATA/pg_hba.conf << 'HBA'
local   all             all                                     trust
hostssl testdb          testuser        127.0.0.1/32            cert
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
HBA"

# Restart PostgreSQL
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D $PGDATA restart -l /var/lib/postgresql/18/logfile"
for i in $(seq 1 30); do
    if su - postgres -c "/usr/lib/postgresql/18/bin/pg_isready -q" 2>/dev/null; then break; fi
    sleep 1
done

# Run radar with cert auth
./radar -h localhost -d testdb -U testuser \
    -sslmode verify-full -sslcert "$CERTDIR/client.crt" -sslkey "$CERTDIR/client.key" -sslrootcert "$CERTDIR/ca.crt" -vv
ZIP5=$(ls -t radar-*.zip | head -1)
if ! validate_zip "$ZIP5" "Scenario 5" "yes"; then
    exit 1
fi
echo -e "${GREEN}✓ Scenario 5 PASSED${NC}"
rm -f "$ZIP5"

# Scenario 6: GSSAPI/Kerberos authentication
echo ""
echo "========================================"
echo -e "${YELLOW}Scenario 6: GSSAPI/Kerberos authentication${NC}"
echo "========================================"

PGDATA=/var/lib/postgresql/18/main
KRB_REALM="RADAR.TEST"

# Initialize Kerberos KDC
mkdir -p /etc/krb5kdc
cat > /etc/krb5.conf << KRBCONF
[libdefaults]
    default_realm = $KRB_REALM
    dns_lookup_realm = false
    dns_lookup_kdc = false

[realms]
    $KRB_REALM = {
        kdc = localhost
        admin_server = localhost
    }
KRBCONF

cat > /etc/krb5kdc/kdc.conf << KDCCONF
[kdcdefaults]
    kdc_ports = 88

[realms]
    $KRB_REALM = {
        database_name = /var/lib/krb5kdc/principal
        key_stash_file = /etc/krb5kdc/stash
        max_life = 1h
    }
KDCCONF

# Create KDC database
kdb5_util create -s -r "$KRB_REALM" -P masterpass 2>/dev/null

# Create principals
kadmin.local -q "addprinc -pw testpass testuser@$KRB_REALM" 2>/dev/null
kadmin.local -q "addprinc -randkey postgres/localhost@$KRB_REALM" 2>/dev/null

# Export keytab for PostgreSQL
kadmin.local -q "ktadd -k $PGDATA/server.keytab postgres/localhost@$KRB_REALM" 2>/dev/null
chown postgres:postgres "$PGDATA/server.keytab"
chmod 600 "$PGDATA/server.keytab"

# Start KDC
krb5kdc

# Configure PostgreSQL for GSSAPI
su - postgres -c "/usr/lib/postgresql/18/bin/psql -c \"ALTER SYSTEM SET krb_server_keyfile = '$PGDATA/server.keytab';\""

# Add GSSAPI auth and ident map
cat > "$PGDATA/pg_hba.conf" << HBA
local   all             all                                     trust
hostgssenc testdb       testuser        127.0.0.1/32            gss include_realm=0
hostssl testdb          testuser        127.0.0.1/32            cert
host    all             all             127.0.0.1/32            trust
host    all             all             ::1/128                 trust
HBA

# Restart PostgreSQL
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D $PGDATA restart -l /var/lib/postgresql/18/logfile"
for i in $(seq 1 30); do
    if su - postgres -c "/usr/lib/postgresql/18/bin/pg_isready -q" 2>/dev/null; then break; fi
    sleep 1
done

# Obtain Kerberos ticket
echo "testpass" | kinit testuser@$KRB_REALM 2>/dev/null

# Run radar with GSSAPI
./radar -h localhost -d testdb -U testuser -sslmode disable -vv
ZIP6=$(ls -t radar-*.zip | head -1)
if ! validate_zip "$ZIP6" "Scenario 6" "yes"; then
    exit 1
fi
echo -e "${GREEN}✓ Scenario 6 PASSED${NC}"
rm -f "$ZIP6"

echo ""
echo "Stopping PostgreSQL..."
su - postgres -c "/usr/lib/postgresql/18/bin/pg_ctl -D /var/lib/postgresql/18/main stop"

echo ""
echo -e "${GREEN}All 6 scenarios passed!${NC}"
