#!/usr/bin/env bash
# Fail if production code contains hard-coded field names that
# should be discovered dynamically. Used in lint stage.

grep -R --line-number --exclude-dir=internal/discovery -E '\"(appName|duration|error|Transaction)\"' $(git ls-files '*.go') && exit 1 || exit 0
