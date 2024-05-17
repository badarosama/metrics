#!/usr/bin/env bash

# STEP 1: Determinate the required values

PACKAGE="metrics"
VERSION="$(git describe --tags --always --abbrev=0 --match='v[0-9]*.[0-9]*.[0-9]*' 2> /dev/null | sed 's/^.//')"
COMMIT_HASH="$(git rev-parse --short HEAD)"
BUILD_TIMESTAMP=$(date '+%Y-%m-%dT%H:%M:%S')

# Print the values
echo "Version: ${VERSION}"
echo "Commit Hash: ${COMMIT_HASH}"
echo "Build Timestamp: ${BUILD_TIMESTAMP}"

# STEP 2: Build the ldflags

LDFLAGS=(
  "-X '${PACKAGE}/server/version.Version=${VERSION}'"
  "-X '${PACKAGE}/server/version.CommitHash=${COMMIT_HASH}'"
  "-X '${PACKAGE}/server/version.BuildTimestamp=${BUILD_TIMESTAMP}'"
)


# STEP 3: Actual Go build process

go build -ldflags="${LDFLAGS[*]}"