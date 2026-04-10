#!/usr/bin/env bash

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "${SCRIPT_DIR}/../.."

"${HALFPIPE}" dependabot > "${SCRIPT_DIR}/dependabot.actual.yml"

echo "Dependabot"
diff --ignore-blank-lines "${SCRIPT_DIR}/dependabot.actual.yml" "${SCRIPT_DIR}/dependabot.expected.yml"
