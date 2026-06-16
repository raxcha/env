#!/usr/bin/env bash
set -euo pipefail

ENV_REPO="${ENV_REPO:-}"
API_REPO="${API_REPO:-}"
ROOT_DIR="${ROOT_DIR:-$HOME}"
ENV_DIR="${ENV_DIR:-$ROOT_DIR/env}"
API_DIR="${API_DIR:-$ROOT_DIR/api}"
DATA_DIR="${ENV_ROOT:-$ROOT_DIR/prsnl.spc}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"

need_repo() {
	local name="$1"
	local value="$2"

	if [ -z "$value" ]; then
		echo "$name is required."
		echo "Example:"
		echo "  ENV_REPO=https://github.com/you/env.git API_REPO=https://github.com/you/api.git bash scripts/termux-install.sh"
		exit 1
	fi
}

install_termux_packages() {
	if command -v pkg >/dev/null 2>&1; then
		pkg update -y
		pkg install -y git golang
	fi
}

clone_or_update() {
	local repo="$1"
	local dir="$2"

	if [ -d "$dir/.git" ]; then
		git -C "$dir" pull --ff-only
		return
	fi

	if [ -e "$dir" ]; then
		echo "$dir exists but is not a git repository."
		exit 1
	fi

	git clone "$repo" "$dir"
}

build_env() {
	mkdir -p "$BIN_DIR" "$DATA_DIR"

	cd "$ENV_DIR"
	go build -o "$BIN_DIR/env" .
	go build -o "$BIN_DIR/env-api" ./cmd/apiserver
}

print_next_steps() {
	echo
	echo "env installed:"
	echo "  $BIN_DIR/env"
	echo "  $BIN_DIR/env-api"
	echo
	echo "Run:"
	echo "  ENV_ROOT=$DATA_DIR $BIN_DIR/env"
	echo
	echo "Run API:"
	echo "  API_ROOT=$API_DIR/prsnl.spc $BIN_DIR/env-api"
}

main() {
	need_repo "ENV_REPO" "$ENV_REPO"
	need_repo "API_REPO" "$API_REPO"
	install_termux_packages
	clone_or_update "$API_REPO" "$API_DIR"
	clone_or_update "$ENV_REPO" "$ENV_DIR"
	build_env
	print_next_steps
}

main "$@"
