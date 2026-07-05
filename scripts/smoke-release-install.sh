#!/usr/bin/env sh
set -eu

version="${1:-}"
if [ -z "$version" ]; then
	echo "usage: smoke-release-install.sh <version>" >&2
	exit 2
fi
expected="${version#v}"
install_dir="$(mktemp -d "${TMPDIR:-/tmp}/azkit-release-install.XXXXXX")"

cleanup() {
	if [ -n "${install_dir:-}" ] && [ -d "$install_dir" ]; then
		rm -rf "$install_dir"
	fi
}
trap cleanup EXIT INT TERM

attempt=1
while :; do
	if AZKIT_INSTALL_VERSION="$version" AZKIT_INSTALL_DIR="$install_dir" sh install.sh; then
		break
	fi
	if [ "$attempt" -ge 6 ]; then
		echo "release install failed after ${attempt} attempts" >&2
		exit 1
	fi
	attempt=$((attempt + 1))
	sleep 10
done

"$install_dir/azkit" --version > "$install_dir/version.txt"
grep -q "^azkit ${expected}$" "$install_dir/version.txt"

"$install_dir/azkit" version > "$install_dir/version-command.txt"
grep -q "^azkit ${expected}$" "$install_dir/version-command.txt"

"$install_dir/azkit" pim --help > "$install_dir/pim-help.txt"
grep -q "activate" "$install_dir/pim-help.txt"

"$install_dir/azkit" ctx --help > "$install_dir/ctx-help.txt"
grep -q "ctx" "$install_dir/ctx-help.txt"

"$install_dir/azkit" sub --help > "$install_dir/sub-help.txt"
grep -q "sub" "$install_dir/sub-help.txt"

printf 'release install smoke ok: azkit %s\n' "$expected"
