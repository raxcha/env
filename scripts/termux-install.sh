#!/usr/bin/env bash
set -euo pipefail

ENV_REPO="${ENV_REPO:-}"
ROOT_DIR="${ROOT_DIR:-$HOME}"
ENV_DIR="${ENV_DIR:-$ROOT_DIR/env}"
DATA_DIR="${ENV_ROOT:-$ROOT_DIR/prsnl.spc}"
BIN_DIR="${BIN_DIR:-$HOME/.local/bin}"
DEFAULT_ENV_ADDR="${ENV_ADDR:-https://prsnlspc.xyz}"
DEFAULT_ENV_PROFILE="${ENV_PROFILE:-phone}"

need_env_repo() {
	if [ -z "$ENV_REPO" ]; then
		echo "ENV_REPO is required."
		echo "Example:"
		echo "  ENV_REPO=https://github.com/you/env.git bash scripts/termux-install.sh"
		exit 1
	fi
}

install_termux_packages() {
	if command -v pkg >/dev/null 2>&1; then
		pkg update -y
		pkg install -y git golang
	fi
}

clone_or_update_env() {
	if [ -d "$ENV_DIR/.git" ]; then
		git -C "$ENV_DIR" pull --ff-only
		return
	fi

	if [ -e "$ENV_DIR" ]; then
		echo "$ENV_DIR exists but is not a git repository."
		exit 1
	fi

	git clone "$ENV_REPO" "$ENV_DIR"
}

write_phone_commands() {
	mkdir -p "$BIN_DIR" "$DATA_DIR"

	cat > "$BIN_DIR/env-update" <<EOF_UPDATE
#!/usr/bin/env bash
set -euo pipefail
cd "$ENV_DIR"
git pull --ff-only
go build -o "$BIN_DIR/env" .
echo "env updated: $BIN_DIR/env"
EOF_UPDATE

	cat > "$BIN_DIR/env-run" <<EOF_RUN
#!/usr/bin/env bash
set -euo pipefail
export ENV_ROOT="\${ENV_ROOT:-$DATA_DIR}"
export ENV_ADDR="\${ENV_ADDR:-$DEFAULT_ENV_ADDR}"
export ENV_PROFILE="\${ENV_PROFILE:-$DEFAULT_ENV_PROFILE}"
cd "$ENV_DIR"
exec "$BIN_DIR/env" "\$@"
EOF_RUN

	chmod +x "$BIN_DIR/env-update" "$BIN_DIR/env-run"
}

build_env() {
	mkdir -p "$BIN_DIR" "$DATA_DIR"
	cd "$ENV_DIR"
	go build -o "$BIN_DIR/env" .
}

print_next_steps() {
	echo
	echo "env installed:"
	echo "  $BIN_DIR/env"
	echo
	echo "Commands:"
	echo "  env-update        # pull latest git version and rebuild"
	echo "  env-run           # run with Termux-friendly defaults"
	echo
	echo "If the commands are not found, add this to your shell rc:"
	echo "  export PATH=\"$BIN_DIR:\$PATH\""
	echo
	echo "API token example:"
	echo "  ENV_TOKEN=your_token env-run"
}

main() {
	need_env_repo
	install_termux_packages
	clone_or_update_env
	build_env
	write_phone_commands
	print_next_steps
}

main "$@"
