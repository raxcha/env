#!/usr/bin/env bash
set -Eeuo pipefail

# Diretório exclusivo do deploy; nunca é a pasta de dados da API.
runtime_dir=${1:-/var/cache/prsnlspc}
source_dir="$runtime_dir/source"
export HOME="$runtime_dir"
export GOCACHE="$runtime_dir/go-build"
export GOPATH="$runtime_dir/go"
export GOTOOLCHAIN=auto
export CGO_ENABLED=0
export GIT_TERMINAL_PROMPT=0
export GIT_SSH_COMMAND='ssh -o BatchMode=yes -o ConnectTimeout=15'
trap 'echo "Falha na atualização; o serviço usará o binário anterior." >&2' ERR

mkdir -p "$runtime_dir"
if [[ ! -d "$source_dir/.git" ]]; then
    # Clone temporário evita deixar uma cópia incompleta após falha de rede.
    clone_dir=$(mktemp -d "$runtime_dir/clone.XXXXXX")
    trap 'rm -rf -- "$clone_dir"' EXIT
    git clone --single-branch --branch main https://github.com/raxcha/env.git "$clone_dir"
    mv -- "$clone_dir" "$source_dir"
    trap - EXIT
else
    git -C "$source_dir" pull --ff-only origin main
fi

cd -- "$source_dir"
echo "Compilando API do commit $(git rev-parse HEAD)"
build_dir=$(mktemp -d "$runtime_dir/build.XXXXXX")
trap 'rm -rf -- "$build_dir"' EXIT
go build -trimpath -o "$build_dir/prsnlspc-api" ./cmd/api
chmod 0755 "$build_dir/prsnlspc-api"
# Publica somente uma compilação concluída, por rename no mesmo filesystem.
mv -f -- "$build_dir/prsnlspc-api" "$runtime_dir/prsnlspc-api"
echo 'Atualização concluída; iniciando a API.'
