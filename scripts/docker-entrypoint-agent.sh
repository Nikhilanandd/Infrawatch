#!/bin/sh
set -e

# Fix cert permissions when bind-mounted from host
if [ -d /etc/infrawatch/certs ]; then
    chmod -R a+r /etc/infrawatch/certs 2>/dev/null || true
fi

# Drop to infrawatch user and exec the agent
exec su-exec infrawatch "$@"
