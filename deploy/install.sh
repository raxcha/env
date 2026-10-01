#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo 'Execute: sudo bash deploy/install.sh' >&2
    exit 1
fi
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
for command in go caddy openssl curl; do
    command -v "$command" >/dev/null || {
        echo 'Instale: sudo apt install -y golang-go caddy openssl curl' >&2
        exit 1
    }
done

build_dir=$(mktemp -d)
trap 'rm -rf -- "$build_dir"' EXIT
# O Go do Ubuntu baixa automaticamente a versão exigida pelo go.mod.
GOTOOLCHAIN=auto CGO_ENABLED=0 go build -trimpath -o "$build_dir/prsnlspc-api" ./cmd/api
caddy validate --config "$PWD/deploy/Caddyfile" --adapter caddyfile

if ! id prsnlspc >/dev/null 2>&1; then
    useradd --system --user-group --home-dir /var/lib/prsnlspc --no-create-home \
        --shell /usr/sbin/nologin prsnlspc
fi
install -d -m 0700 /etc/prsnlspc
if [[ ! -e /etc/prsnlspc/api.env ]]; then
    (umask 077; printf 'API_USERNAME=prsnlspc\nAPI_PASSWORD=%s\n' "$(openssl rand -hex 24)" > /etc/prsnlspc/api.env)
fi
chmod 0600 /etc/prsnlspc/api.env

# Rename permite atualizar mesmo com o binário em execução.
install -m 0755 "$build_dir/prsnlspc-api" /usr/local/bin/prsnlspc-api.new
mv -f /usr/local/bin/prsnlspc-api.new /usr/local/bin/prsnlspc-api
install -m 0644 deploy/prsnlspc-api.service /etc/systemd/system/prsnlspc-api.service
systemctl daemon-reload
systemctl enable prsnlspc-api
systemctl restart prsnlspc-api

ready=false
for ((attempt=0; attempt<20; attempt++)); do
    code=$(curl --silent --output /dev/null --write-out '%{http_code}' \
        --max-time 2 -X POST http://127.0.0.1:8081/load || true)
    if [[ "$code" == 401 ]]; then
        ready=true
        break
    fi
    sleep 1
done
if [[ "$ready" != true ]]; then
    journalctl -u prsnlspc-api -n 30 --no-pager
    exit 1
fi

if [[ -f /etc/caddy/Caddyfile && ! -e /etc/caddy/Caddyfile.before-prsnlspc ]]; then
    cp -p /etc/caddy/Caddyfile /etc/caddy/Caddyfile.before-prsnlspc
fi
install -m 0644 deploy/Caddyfile /etc/caddy/Caddyfile
systemctl enable caddy
systemctl reload-or-restart caddy
echo 'API instalada. HTTPS depende do DNS e das portas 80/443 liberadas.'
echo 'Veja as credenciais com: sudo cat /etc/prsnlspc/api.env'
