#!/bin/bash
set -e

DOMAIN_ENTRY="127.0.0.1 appliance-store.local store.local"

if grep -q "appliance-store.local" /etc/hosts; then
    echo "==> Domain appliance-store.local already exists in /etc/hosts:"
    grep "appliance-store.local" /etc/hosts
else
    echo "==> Adding domain to /etc/hosts..."
    if [ "$EUID" -ne 0 ]; then
        echo "Please enter sudo password to append to /etc/hosts:"
        sudo sh -c "echo '$DOMAIN_ENTRY' >> /etc/hosts"
    else
        echo "$DOMAIN_ENTRY" >> /etc/hosts
    fi
    echo "==> Added '$DOMAIN_ENTRY' to /etc/hosts."
fi
