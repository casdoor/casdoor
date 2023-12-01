#!/bin/bash
set -e

atest run -p testsuite.yaml --report md --monitor-docker e2e-casdoor-1 \
    --report github \
    --report-file /var/data/report.json --report-github-repo casbin/casdoor \
    --report-github-pr ${PULL_REQUEST:-0}
