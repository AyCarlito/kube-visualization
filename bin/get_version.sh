#!/bin/bash

# Version defined as major.minor.patch-build
# - major & minor numbers are taken from the last git tag
# - patch comes from the git describe "distance" from the last tag
# - build is either:
#      - $BUILD_ID (when main or release branches are built in CI)
#      - dev.$BRANCH_NAME.$BUILD_ID (when dev branches are built in CI)
#      - dev.branch.sha (when building outside CI; BRANCH_NAME obtained from git)
#
# Note that branch names are "mangled" to keep the overall version SemVer compatible.
#
# *** If the git repo is not tagged, it will default to 0.0.0-X ***
#

short_desc=$(git describe --abbrev=0 --tags 2>/dev/null) || short_desc="0.0"
long_desc=$(git describe --long --tags 2>/dev/null) || long_desc="0.0-0"
if [ $short_desc = "0.0" ]; then
    >&2 echo -e "\n*** WARNING: No tag found, using $short_desc ***\n"
fi

readonly major=$(echo ${short_desc} | cut --delimiter=. --fields=1)
readonly minor=$(echo ${short_desc} | cut --delimiter=. --fields=2)
readonly patch=$(echo ${long_desc}  | cut --delimiter=- --fields=2)

# Handle BRANCH_NAME being defined (building in CI), or undefined (local builds)
if [ -z "${BRANCH_NAME}" ]; then
    current_branch="$(git symbolic-ref --short HEAD 2>/dev/null)" ||
        current_branch="detached" # detached HEAD
else
    current_branch=${BRANCH_NAME}
fi

# Generate the "branch" part of the version
readonly branch_number_prefix="$(echo "${current_branch}" | sed -e 's#\(^[0-9]*\).*#\1#')"
if [ "${branch_number_prefix}" ]; then
    branch_prefix="${branch_number_prefix}"
else
    branch_prefix="$(echo "${current_branch}" | sed -e 's/^dev[^[:alnum:]]//g' -e 's/[^[:alnum:]]//g')"
fi

# Use BUILD_ID if it's defined (CI), otherwise use SHA1 (local builds).
if [ -z "${BUILD_ID}" ]; then
    current_sha="$(git rev-parse --short HEAD)"
    build="dev.${branch_prefix}.${current_sha}"
else
    if [ "${current_branch}" = "main" ] || [ "${current_branch:0:7}" = "release" ]; then
        build=${BUILD_ID}
    else
        build="dev.${branch_prefix}.${BUILD_ID}"
    fi
fi

echo "${major}.${minor}.${patch}-${build}"
