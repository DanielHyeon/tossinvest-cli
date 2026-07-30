#!/bin/sh
set -eu

runtime_dir=/run/tossos
umask 077

if [ "$(id -u)" -eq 0 ]; then
    echo "TossOS container: refusing to run as root" >&2
    exit 1
fi

copy_secret() {
    source_path=$1
    target_path=$2
    if [ ! -f "$source_path" ]; then
        echo "TossOS container: required secret is missing: $source_path" >&2
        exit 1
    fi
    cp "$source_path" "$target_path"
    chmod 600 "$target_path"
}

if [ -d /run/secrets ]; then
    copy_secret /run/secrets/broker-session "$runtime_dir/session.json"
    copy_secret /run/secrets/remote-token "$runtime_dir/remote-token"
    copy_secret /run/secrets/tls-cert "$runtime_dir/tls.crt"
    copy_secret /run/secrets/tls-key "$runtime_dir/tls.key"
fi

exec /usr/local/bin/tossctl "$@"
