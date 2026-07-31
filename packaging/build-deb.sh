#!/usr/bin/env bash
set -euo pipefail

BUILD_DIR="/tmp/pg_deb_build"

CWD="$(pwd)"

export DEBIAN_FRONTEND=noninteractive
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

# RADAR_VERSION may carry a '~beta…' suffix for DEB ordering, so it must NOT
# be used to build the download URL — use the tag (RADAR_BRANCH) instead.
TAG_VERSION="${RADAR_BRANCH#v}"
SRC_DIR="${BUILD_DIR}/pgedge-radar-${TAG_VERSION}"
# The workflow stages the GoReleaser binary + docs under release-artifacts/ so
# simulate_tag runs (no published release) still build. Prefer them; fall back
# to the GitHub release for the tag (RADAR_BRANCH).
ARTIFACT_DIR="${ARTIFACT_DIR:-${CWD}/release-artifacts}"

prepare() {
  setup_apt_build_env

  # This function is for debugging purpose if you have your own keys. GH workflow does not need it.
  #import_gpg_keys

  echo "Creating source directory..."
  rm -rf "$SRC_DIR"
  mkdir -p "$SRC_DIR"

  echo "Staging radar binary..."
  if [ -f "${ARTIFACT_DIR}/radar-linux-${ARCH}" ]; then
    echo "Using staged binary ${ARTIFACT_DIR}/radar-linux-${ARCH}"
    cp "${ARTIFACT_DIR}/radar-linux-${ARCH}" "$SRC_DIR/radar"
  else
    echo "Downloading radar binary from release ${RADAR_BRANCH}"
    wget -q -O "$SRC_DIR/radar" "https://github.com/pgEdge/radar/releases/download/${RADAR_BRANCH}/radar-linux-${ARCH}"
  fi
  chmod +x "$SRC_DIR/radar"

  echo "Moving Debian packaging into source directory..."
  cp -rp "${CWD}/${COMPONENT_NAME}/deb/debian" "$SRC_DIR/"

  if [ -f "${ARTIFACT_DIR}/LICENCE.md" ]; then
    cp "${ARTIFACT_DIR}/LICENCE.md" "$SRC_DIR/LICENCE.md"
  else
    wget -q -O "$SRC_DIR/LICENCE.md" "https://raw.githubusercontent.com/pgEdge/radar/${RADAR_BRANCH}/LICENCE.md"
  fi
  if [ -f "${ARTIFACT_DIR}/README.md" ]; then
    cp "${ARTIFACT_DIR}/README.md" "$SRC_DIR/README.md"
  else
    wget -q -O "$SRC_DIR/README.md" "https://raw.githubusercontent.com/pgEdge/radar/${RADAR_BRANCH}/README.md"
  fi

  echo "Installing build dependencies..."
  cd "$SRC_DIR"
  sudo apt-get update
  sudo apt-get build-dep -y .
}

build() {
  cd "$SRC_DIR"
  echo "Building Debian package..."
  DISTRO=$(lsb_release -cs)
  rm -f debian/changelog
cat > debian/changelog <<EOF
pgedge-radar (${RADAR_VERSION}-${RADAR_BUILDNUM}.${DISTRO}) stable; urgency=medium

  * Initial pgedge-radar package.

 -- pgEdge Build Team <support@pgedge.com>  $(date -R)
EOF

  dpkg-buildpackage -us -uc -b
}

post_build() {
  echo "Copying .deb packages to output..."
  sudo mkdir -p "/output"
  rename_ddeb_packages $BUILD_DIR
  sudo cp "$BUILD_DIR"/*.deb "/output" || echo "No .deb packages found."
}
