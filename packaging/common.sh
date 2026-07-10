#!/usr/bin/env bash
# common.sh - Common environment variables

export RADAR_REPO="https://github.com/pgEdge/radar.git"
export RADAR_BRANCH="${COMPONENT_BRANCH:-v0.4.1}"
export RADAR_VERSION=${COMPONENT_VERSION:-0.4.1}
export RADAR_BUILDNUM=${COMPONENT_BUILDNUM:-1}

# DEB only: move a pre-release pretag (e.g. BUILDNUM='beta3_1') into the
# upstream VERSION with a leading '~' (0.4.1~beta3, BUILDNUM=1) so '~' sorts
# pre-releases BELOW stable in dpkg/reprepro. Downloads use the tag
# (RADAR_BRANCH), not VERSION, so this never affects the source URL.
if command -v apt-get &>/dev/null; then
    if [[ "$RADAR_BUILDNUM" == *_* ]]; then
        RADAR_PRETAG="${RADAR_BUILDNUM%%_*}"
        export RADAR_VERSION="${RADAR_VERSION}~${RADAR_PRETAG}"
        RADAR_BUILDNUM="${RADAR_BUILDNUM#*_}"
    fi
fi

export REPO_TYPE="${REPO_TYPE:-daily}"
