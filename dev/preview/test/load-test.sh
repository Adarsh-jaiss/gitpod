#!/usr/bin/env bash

set -euo pipefail

function log {
    echo "[$(date)] $*"
}

git checkout main

BASE_BRANCH_NAME="vm-load-test-base"
git branch -D "${BASE_BRANCH_NAME}" > /dev/null || true
git checkout -b "${BASE_BRANCH_NAME}"
git commit --allow-empty -m "Prepare load test" -m "/werft with-vm=true"

for number in $(seq 1 15); do
    branch="vm-load-test/${number}"

    log "Creating and pushing branch ${branch}"
    git checkout -b "${branch}"
    git push -u origin "${branch}"

    log "Back to base branch ${BASE_BRANCH_NAME}"
    git checkout "${BASE_BRANCH_NAME}"

    log "Sleeping 30 seconds"
    sleep 30
done
