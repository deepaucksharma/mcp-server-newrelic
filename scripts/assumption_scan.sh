#!/bin/bash
# Fail CI if production Go code uses forbidden field names outside discovery packages.
set -e
PATTERN='"(appName|duration|error|Transaction)"'
FILES=$(git ls-files '*.go')
if grep -R --line-number --exclude-dir=internal/discovery -E $PATTERN $FILES; then
  echo "Hard-coded discovery assumption detected." >&2
  exit 1
fi
exit 0
