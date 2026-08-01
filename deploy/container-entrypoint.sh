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

copy_secret_if_present() {
    source_path=$1
    target_path=$2
    if [ -f "$source_path" ]; then
        copy_secret "$source_path" "$target_path"
    fi
}

if [ -d /run/secrets ]; then
    # Copy only secrets mounted for the selected service. Trusted-network
    # console and the no-token read API do not use the retired remote-token;
    # requiring it here made an otherwise valid HTTP API container fail before
    # its own private-boundary validation could run.
    copy_secret_if_present /run/secrets/broker-session "$runtime_dir/session.json"
    copy_secret_if_present /run/secrets/tls-cert "$runtime_dir/tls.crt"
    copy_secret_if_present /run/secrets/tls-key "$runtime_dir/tls.key"
    copy_secret_if_present /run/secrets/httpapi-capability-public-key "$runtime_dir/httpapi-capability-public.key"
fi

exec /usr/local/bin/tossctl "$@"
