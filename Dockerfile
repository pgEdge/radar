FROM debian:latest

# Install PostgreSQL 18 from official PostgreSQL repository
RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    ca-certificates \
    gnupg \
    lsb-release \
    && curl -fsSL https://www.postgresql.org/media/keys/ACCC4CF8.asc | gpg --dearmor -o /usr/share/keyrings/postgresql-keyring.gpg \
    && echo "deb [signed-by=/usr/share/keyrings/postgresql-keyring.gpg] http://apt.postgresql.org/pub/repos/apt $(lsb_release -cs)-pgdg main" > /etc/apt/sources.list.d/pgdg.list \
    && apt-get update && apt-get install -y --no-install-recommends \
    postgresql-18 \
    postgresql-18-statviz \
    util-linux \
    procps \
    net-tools \
    iproute2 \
    unzip \
    kmod \
    pciutils \
    nfs-common \
    sysstat \
    policycoreutils \
    dmsetup \
    && rm -rf /var/lib/apt/lists/*

# Set up working directory
WORKDIR /radar

# Copy pre-built binary and test script
COPY radar ./radar
COPY test-radar.sh ./test-radar.sh

# Default command
CMD ["./test-radar.sh"]
