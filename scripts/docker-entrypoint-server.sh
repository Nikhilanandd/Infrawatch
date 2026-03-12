#!/bin/sh
set -e

# Fix cert permissions when bind-mounted from host
if [ -d /etc/infrawatch/certs ]; then
    chmod -R a+r /etc/infrawatch/certs 2>/dev/null || true
fi

# Fix data directory ownership
chown -R infrawatch:infrawatch /var/lib/infrawatch 2>/dev/null || true

# Drop to infrawatch user and exec the server
exec su-exec infrawatch "$@"
