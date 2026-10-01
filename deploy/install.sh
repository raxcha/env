#!/usr/bin/env bash
set -euo pipefail

if [[ $EUID -ne 0 ]]; then
    echo 'Execute: sudo bash deploy/install.sh' >&2
    exit 1
fi
cd -- "$(dirname -- "${BASH_SOURCE[0]}")/.."
for command in go git timeout caddy openssl curl; do
    command -v "$command" >/dev/null || {
        echo 'Instale: sudo apt install -y git golang-go caddy openssl curl' >&2
        exit 1
    }
done

build_dir=$(mktemp -d)
trap 'rm -rf -- "$build_dir"' EXIT
# O Go do Ubuntu baixa automaticamente a versão exigida pelo go.mod.
GOTOOLCHAIN=auto CGO_ENABLED=0 go build -trimpath -o "$build_dir/prsnlspc-api" ./cmd/api
caddy validate --config "$PWD/deploy/Caddyfile" --adapter caddyfile

if ! id nausea >/dev/null 2>&1; then
    useradd --user-group --create-home --home-dir /home/nausea --shell /bin/bash nausea
fi
if [[ $(getent passwd nausea | cut -d: -f6) != /home/nausea ]]; then
    echo 'O usuário nausea existente precisa ter home /home/nausea.' >&2
    exit 1
fi
nausea_group=$(id -gn nausea)
# Não mistura silenciosamente dois conjuntos de dados.
if [[ -d /var/lib/prsnlspc ]] && [[ -n $(find /var/lib/prsnlspc -mindepth 1 -print -quit) ]]; then
    if [[ -d /home/nausea/prsnlspc ]] && [[ -n $(find /home/nausea/prsnlspc -mindepth 1 -print -quit) ]]; then
        if [[ ! -f /etc/prsnlspc/home-migration-complete ]]; then
            echo 'Há dados no caminho antigo e no novo; faça a conciliação antes de instalar.' >&2
            exit 1
        fi
    else
        systemctl stop prsnlspc-api || true
        install -d -o nausea -g "$nausea_group" -m 0700 /home/nausea/prsnlspc
        cp -a /var/lib/prsnlspc/. /home/nausea/prsnlspc/
        chown -R nausea:"$nausea_group" /home/nausea/prsnlspc
        install -d -m 0700 /etc/prsnlspc
        touch /etc/prsnlspc/home-migration-complete
    fi
fi
install -d -o nausea -g "$nausea_group" -m 0700 /home/nausea/prsnlspc
install -d -m 0700 /etc/prsnlspc
if [[ ! -e /etc/prsnlspc/api.env ]]; then
    (umask 077; printf 'API_USERNAME=prsnlspc\nAPI_PASSWORD=%s\n' "$(openssl rand -hex 24)" > /etc/prsnlspc/api.env)
fi
chmod 0600 /etc/prsnlspc/api.env

# Binário inicial permite iniciar mesmo sem GitHub disponível.
systemctl stop prsnlspc-api || true
install -d -o nausea -g "$nausea_group" -m 0700 /var/cache/prsnlspc
chown -R nausea:"$nausea_group" /var/cache/prsnlspc
install -o nausea -g "$nausea_group" -m 0755 "$build_dir/prsnlspc-api" /var/cache/prsnlspc/prsnlspc-api.new
mv -f /var/cache/prsnlspc/prsnlspc-api.new /var/cache/prsnlspc/prsnlspc-api
install -d -m 0755 /usr/local/libexec
install -m 0755 deploy/update.sh /usr/local/libexec/prsnlspc-update
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
echo 'Escolha a senha do usuário Linux com: sudo passwd nausea'
