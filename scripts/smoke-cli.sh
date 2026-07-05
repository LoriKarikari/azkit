#!/usr/bin/env sh
set -eu

version="${1:-dev}"
root_dir="$(CDPATH='' cd -- "$(dirname -- "$0")/.." && pwd)"
smoke_dir="$(mktemp -d "${TMPDIR:-/tmp}/azkit-smoke.XXXXXX")"

cleanup() {
	if [ -n "${smoke_dir:-}" ] && [ -d "$smoke_dir" ]; then
		rm -rf "$smoke_dir"
	fi
}
trap cleanup EXIT INT TERM

commit="$(git -C "$root_dir" rev-parse --short HEAD 2>/dev/null || printf unknown)"
date="$(date -u +%Y-%m-%dT%H:%M:%SZ)"

go build -trimpath \
	-ldflags "-s -w -X github.com/LoriKarikari/azkit/internal/cli.version=${version} -X github.com/LoriKarikari/azkit/internal/cli.commit=${commit} -X github.com/LoriKarikari/azkit/internal/cli.date=${date}" \
	-o "$smoke_dir/azkit" "$root_dir/cmd/azkit"

export HOME="$smoke_dir/home"
export XDG_CONFIG_HOME="$smoke_dir/config"
export SHELL=/bin/bash
mkdir -p "$HOME" "$XDG_CONFIG_HOME"

"$smoke_dir/azkit" --version > "$smoke_dir/version-flag.txt"
grep -q "^azkit ${version}$" "$smoke_dir/version-flag.txt"

"$smoke_dir/azkit" version > "$smoke_dir/version-command.txt"
grep -q "^azkit ${version}$" "$smoke_dir/version-command.txt"
grep -q "commit: ${commit}" "$smoke_dir/version-command.txt"

"$smoke_dir/azkit" --help > "$smoke_dir/root-help.txt"
grep -q "azkit" "$smoke_dir/root-help.txt"

"$smoke_dir/azkit" pim --help > "$smoke_dir/pim-help.txt"
grep -q "activate" "$smoke_dir/pim-help.txt"

"$smoke_dir/azkit" ctx --help > "$smoke_dir/ctx-help.txt"
grep -q "ctx" "$smoke_dir/ctx-help.txt"

"$smoke_dir/azkit" sub --help > "$smoke_dir/sub-help.txt"
grep -q "sub" "$smoke_dir/sub-help.txt"

"$smoke_dir/azkit" shell-init bash > "$smoke_dir/shell-bash.txt"
grep -q "azkit" "$smoke_dir/shell-bash.txt"

"$smoke_dir/azkit" shell-init zsh > "$smoke_dir/shell-zsh.txt"
grep -q "azkit" "$smoke_dir/shell-zsh.txt"

"$smoke_dir/azkit" shell-init pwsh > "$smoke_dir/shell-pwsh.txt"
grep -q "azkit" "$smoke_dir/shell-pwsh.txt"

"$smoke_dir/azkit" completion > "$smoke_dir/completion.txt"
grep -q "azkit" "$smoke_dir/completion.txt"

"$smoke_dir/azkit" ctx -l --json > "$smoke_dir/ctx-empty.json"
python3 -m json.tool "$smoke_dir/ctx-empty.json" > /dev/null

"$smoke_dir/azkit" ctx add demo --tenant 00000000-0000-0000-0000-000000000000 > "$smoke_dir/ctx-add.txt"
"$smoke_dir/azkit" ctx -l --json > "$smoke_dir/ctx-list.json"
python3 - "$smoke_dir/ctx-list.json" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as f:
    data = json.load(f)

assert isinstance(data, dict), data
assert any(c.get("context") == "demo" for c in data.get("contexts", [])), data
PY

"$smoke_dir/azkit" --shell-env ctx demo > "$smoke_dir/ctx-switch.sh"
grep -q "export AZKIT_CONTEXT='demo'" "$smoke_dir/ctx-switch.sh"
grep -q "export AZURE_TENANT_ID='00000000-0000-0000-0000-000000000000'" "$smoke_dir/ctx-switch.sh"
grep -q '^unset AZKIT_SUBSCRIPTION_ID$' "$smoke_dir/ctx-switch.sh"

AZKIT_CONTEXT=demo "$smoke_dir/azkit" ctx current --json > "$smoke_dir/ctx-current.json"
python3 - "$smoke_dir/ctx-current.json" <<'PY'
import json
import sys

with open(sys.argv[1], encoding="utf-8") as f:
    data = json.load(f)

assert data["active"] is True, data
assert data["context"] == "demo", data
PY

if "$smoke_dir/azkit" --shell-env ctx --json -l > "$smoke_dir/bad-shell-json.out" 2>&1; then
	echo "ctx shell-env json list unexpectedly succeeded" >&2
	exit 1
fi
grep -q "cannot run through shell integration" "$smoke_dir/bad-shell-json.out"

if "$smoke_dir/azkit" list --help > "$smoke_dir/root-list-alias.out" 2>&1; then
	echo "unexpected root list alias" >&2
	exit 1
fi

printf 'smoke ok: azkit %s (%s)\n' "$version" "$commit"
