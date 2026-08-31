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
echo "Enabling the logging collector, so log_directory exists to be listed..."
su - postgres -c "echo 'logging_collector = on' >> /var/lib/postgresql/18/main/postgresql.conf"

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
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"CREATE TABLE part_parent (id int, ts date) PARTITION BY RANGE (ts);\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"CREATE TABLE part_child PARTITION OF part_parent FOR VALUES FROM ('2020-01-01') TO ('2030-01-01');\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"CREATE EXTENSION pg_statviz;\""
su - postgres -c "/usr/lib/postgresql/18/bin/psql -d testdb -c \"SELECT pgstatviz.snapshot();\""

echo ""
echo "Creating non-root system user..."
useradd -m -s /bin/bash radaruser || true
cp radar /home/radaruser/
chown radaruser:radaruser /home/radaruser/radar

echo ""
echo "Setting up a PgBouncer configuration fixture..."
mkdir -p /etc/pgbouncer
cat > /etc/pgbouncer/pgbouncer.ini << 'PGBINI'
; PgBouncer test fixture
[databases]
testdb = host=127.0.0.1 port=5432 dbname=testdb user=pooler password=ini-secret-value

[pgbouncer]
listen_addr = 127.0.0.1
listen_port = 6432
auth_type = scram-sha-256
auth_file = /etc/pgbouncer/userlist.txt
pool_mode = transaction
max_client_conn = 500
PGBINI
cat > /etc/pgbouncer/userlist.txt << 'PGBUSERS'
"pooler" "SCRAM-SHA-256$4096:userlist-secret-value"
PGBUSERS
# pgbouncer.ini is world-readable here so that every scenario exercises the
# redaction filter, which is what protects it. userlist.txt stays root-only:
# radar records that it is there and never reads its contents.
chmod 0644 /etc/pgbouncer/pgbouncer.ini
chmod 0600 /etc/pgbouncer/userlist.txt

# Verifies PgBouncer collection. pgbouncer.ini ships with the [pgbouncer] section
# intact and every other value removed, and userlist.txt is recorded as a
# directory entry only: its contents are credentials.
validate_pgbouncer() {
	local zip_file="$1"
	local scenario="$2"

	local ini
	ini=$(unzip -p "$zip_file" "pgbouncer/pgbouncer.ini" 2>/dev/null || true)
	if [ -z "$ini" ]; then
		echo -e "${RED}✗ $scenario FAILED: pgbouncer/pgbouncer.ini missing${NC}"
		return 1
	fi
	if ! echo "$ini" | grep -q "^max_client_conn = 500$"; then
		echo -e "${RED}✗ $scenario FAILED: pgbouncer.ini lost its [pgbouncer] values${NC}"
		return 1
	fi
	if ! echo "$ini" | grep -q "^\[databases\]$"; then
		echo -e "${RED}✗ $scenario FAILED: pgbouncer.ini lost its section headers${NC}"
		return 1
	fi
	# No comment reaches the archive, so the fixture's header comment is gone.
	if echo "$ini" | grep -qE "^[[:space:]]*[;#]"; then
		echo -e "${RED}✗ $scenario FAILED: pgbouncer.ini still contains comments${NC}"
		return 1
	fi

	if ! unzip -p "$zip_file" "pgbouncer/files.tsv" 2>/dev/null | grep -q "userlist.txt"; then
		echo -e "${RED}✗ $scenario FAILED: pgbouncer/files.tsv does not record userlist.txt${NC}"
		return 1
	fi

	# Neither secret may appear anywhere in the archive.
	if unzip -p "$zip_file" 2>/dev/null | grep -q "ini-secret-value"; then
		echo -e "${RED}✗ $scenario FAILED: a [databases] password reached the archive${NC}"
		return 1
	fi
	if unzip -p "$zip_file" 2>/dev/null | grep -q "userlist-secret-value"; then
		echo -e "${RED}✗ $scenario FAILED: userlist.txt contents reached the archive${NC}"
		return 1
	fi

	return 0
}

# Verifies the log directory listing, which comes from pg_ls_logdir(). That is
# executable by pg_monitor and needs no filesystem access, so every scenario
# gets it, including the non-root ones that cannot traverse $PGDATA. It carries
# names, sizes and timestamps only, never log contents.
validate_log_directory() {
	local zip_file="$1"
	local scenario="$2"

	local listing
	listing=$(unzip -p "$zip_file" "postgresql/log_directory.tsv" 2>/dev/null || true)

	if [ -z "$listing" ]; then
		echo -e "${RED}✗ $scenario FAILED: postgresql/log_directory.tsv missing${NC}"
		return 1
	fi

	if ! echo "$listing" | head -1 | grep -q "^name.size.modification$"; then
		echo -e "${RED}✗ $scenario FAILED: unexpected log_directory.tsv header${NC}"
		return 1
	fi

	if ! echo "$listing" | tail -n +2 | grep -q "\.log"; then
		echo -e "${RED}✗ $scenario FAILED: log_directory.tsv lists no log file${NC}"
		return 1
	fi

	# Names only: nothing from a log body may appear.
	if echo "$listing" | grep -qE "LOG:|FATAL:|database system is ready"; then
		echo -e "${RED}✗ $scenario FAILED: log_directory.tsv contains log file contents${NC}"
		return 1
	fi

	return 0
}

# Verifies the dedicated freeze-age listing. It is ranked by age rather than
# size, so it must reach relations tables.tsv excludes: pg_catalog and pg_toast.
validate_freeze_age_file() {
	local zip_file="$1"
	local scenario="$2"

	local tsv
	tsv=$(unzip -p "$zip_file" "databases/testdb/table_freeze_age.tsv" 2>/dev/null || true)

	if [ -z "$tsv" ]; then
		echo -e "${RED}✗ $scenario FAILED: databases/testdb/table_freeze_age.tsv missing${NC}"
		return 1
	fi

	if ! echo "$tsv" | head -1 | grep -q "^schemaname.tablename.relkind.relpages.relfrozenxid_age.relminmxid_age$"; then
		echo -e "${RED}✗ $scenario FAILED: unexpected table_freeze_age.tsv header${NC}"
		return 1
	fi

	# Every row carries a numeric age, and the partitioned parent is excluded
	# because its relkind is 'p'.
	if ! echo "$tsv" | tail -n +2 | awk -F'\t' '$5 ~ /^[0-9]+$/ {n++} END {exit !(n>0)}'; then
		echo -e "${RED}✗ $scenario FAILED: table_freeze_age.tsv has no numeric ages${NC}"
		return 1
	fi
	if echo "$tsv" | tail -n +2 | awk -F'\t' '$2 == "part_parent" {found=1} END {exit !found}'; then
		echo -e "${RED}✗ $scenario FAILED: partitioned parent should be excluded (relkind 'p')${NC}"
		return 1
	fi

	# An invalid relfrozenxid or relminmxid comes back as 2147483647, the INT_MAX
	# both age() and mxid_age() return for one, so no age may read that.
	if echo "$tsv" | tail -n +2 | awk -F'\t' '$5 == 2147483647 || $6 == 2147483647 {found=1} END {exit !found}'; then
		echo -e "${RED}✗ $scenario FAILED: table_freeze_age.tsv reports an INT_MAX age${NC}"
		return 1
	fi

	return 0
}

# Verifies the per-table freeze-age columns in tables.tsv. A partitioned table
# carries relfrozenxid = 0, and age('0'::xid) returns a huge meaningless number
# that reads as an imminent wraparound, so the partitioned parent must come back
# empty while its partition reports a real age.
validate_freeze_age() {
	local zip_file="$1"
	local scenario="$2"

	if ! unzip -p "$zip_file" "databases/testdb/tables.tsv" | awk -F'\t' '
		NR == 1 {
			for (i = 1; i <= NF; i++) col[$i] = i
			if (!("relfrozenxid_age" in col) || !("relminmxid_age" in col)) {
				print "tables.tsv has no freeze-age columns" > "/dev/stderr"
				exit 1
			}
			next
		}
		$col["tablename"] == "part_parent" {
			parent = 1
			if ($col["relfrozenxid_age"] != "" || $col["relminmxid_age"] != "") {
				print "partitioned table part_parent reported a freeze age" > "/dev/stderr"
				exit 1
			}
		}
		$col["tablename"] == "part_child" {
			child = 1
			if ($col["relfrozenxid_age"] !~ /^[0-9]+$/ || $col["relminmxid_age"] !~ /^[0-9]+$/) {
				print "partition part_child freeze age is not a number" > "/dev/stderr"
				exit 1
			}
		}
		END {
			if (!parent) { print "part_parent row missing from tables.tsv" > "/dev/stderr"; exit 1 }
			if (!child)  { print "part_child row missing from tables.tsv" > "/dev/stderr"; exit 1 }
		}'; then
		echo -e "${RED}✗ $scenario FAILED: per-table freeze age${NC}"
		return 1
	fi

	return 0
}

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

	if ! validate_freeze_age "$zip_file" "$scenario"; then
		return 1
	fi

	if ! validate_freeze_age_file "$zip_file" "$scenario"; then
		return 1
	fi

	if ! validate_log_directory "$zip_file" "$scenario"; then
		return 1
	fi

	if ! validate_pgbouncer "$zip_file" "$scenario"; then
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
