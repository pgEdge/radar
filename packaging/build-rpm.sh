#!/bin/bash
set -euo pipefail

RHEL="$(rpm --eval %rhel)"
ARCH=$(uname -m)
if [ "$ARCH" = "x86_64" ]; then
  ARCH="amd64"
elif [ "$ARCH" = "aarch64" ]; then
  ARCH="arm64"
fi

# The workflow stages the GoReleaser binary + docs under release-artifacts/ so
# simulate_tag runs (no published release) still build. Prefer them; fall back
# to the GitHub release for the TAG (RADAR_BRANCH), never v${RADAR_VERSION}
# (which 404s on pre-release tags and carries a '~' suffix for DEB ordering).
CWD="$(pwd)"
ARTIFACT_DIR="${ARTIFACT_DIR:-${CWD}/release-artifacts}"

prepare() {
  setup_dnf_build_env

  echo "Copying packaging files..."
  cp ${COMPONENT_NAME}/rpm/radar.spec ~/rpmbuild/SPECS/

  # This function is for debugging purpose if you have your own keys. GH workflow does not need it.
  #import_gpg_keys

  echo "Staging radar binary and docs..."
  if [ -f "${ARTIFACT_DIR}/radar-linux-${ARCH}" ]; then
    echo "Using staged binary ${ARTIFACT_DIR}/radar-linux-${ARCH}"
    cp "${ARTIFACT_DIR}/radar-linux-${ARCH}" ~/rpmbuild/SOURCES/radar-linux-${ARCH}
  else
    echo "Downloading radar binary from release ${RADAR_BRANCH}"
    wget -q -O ~/rpmbuild/SOURCES/radar-linux-${ARCH} "https://github.com/pgEdge/radar/releases/download/${RADAR_BRANCH}/radar-linux-${ARCH}"
  fi
  if [ -f "${ARTIFACT_DIR}/LICENCE.md" ]; then
    cp "${ARTIFACT_DIR}/LICENCE.md" ~/rpmbuild/SOURCES/LICENCE.md
  else
    wget -q -O ~/rpmbuild/SOURCES/LICENCE.md "https://raw.githubusercontent.com/pgEdge/radar/${RADAR_BRANCH}/LICENCE.md"
  fi
  if [ -f "${ARTIFACT_DIR}/README.md" ]; then
    cp "${ARTIFACT_DIR}/README.md" ~/rpmbuild/SOURCES/README.md
  else
    wget -q -O ~/rpmbuild/SOURCES/README.md "https://raw.githubusercontent.com/pgEdge/radar/${RADAR_BRANCH}/README.md"
  fi

  echo "Installing RPM build dependencies..."
  dnf builddep -y \
    --define "radar_version ${RADAR_VERSION}" \
    --define "radar_buildnum ${RADAR_BUILDNUM}" \
    --define "arch ${ARCH}" \
    ~/rpmbuild/SPECS/radar.spec
}

build() {
  echo "Building RPM and SRPM..."
  QA_RPATHS=$(( 0xffff )) rpmbuild -ba ~/rpmbuild/SPECS/radar.spec \
    --define "radar_version ${RADAR_VERSION}" \
    --define "radar_buildnum ${RADAR_BUILDNUM}" \
    --define "arch ${ARCH}"
}

post_build() {
  echo "Copying built RPMs to /output..."
  mkdir -p /output
  cp -v ~/rpmbuild/RPMS/*/*.rpm /output/ || echo "No binary RPMs found"
  cp -v ~/rpmbuild/SRPMS/*.src.rpm /output/ || echo "No SRPM found"

  sign_rpms /output/*.rpm
  validate_signatures /output/*.rpm
}
