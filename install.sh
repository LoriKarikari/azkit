#!/usr/bin/env sh
set -eu

repo="LoriKarikari/pimctl"
bin_name="pimctl"
version="${PIMCTL_INSTALL_VERSION:-latest}"
bin_dir="${PIMCTL_INSTALL_DIR:-${HOME}/.local/bin}"
tmp_dir=""

usage() {
	cat <<EOF
Install pimctl from GitHub releases.

Usage:
  curl -fsSL https://raw.githubusercontent.com/${repo}/main/install.sh | sh

Environment:
  PIMCTL_INSTALL_VERSION  Version to install, for example v0.3.0. Defaults to latest.
  PIMCTL_INSTALL_DIR      Install directory. Defaults to ~/.local/bin.
  GITHUB_TOKEN            Optional token for GitHub API rate limits.
EOF
}

log() {
	printf '%s\n' "$*" >&2
}

fail() {
	log "error: $*"
	exit 1
}

need() {
	command -v "$1" >/dev/null 2>&1 || fail "missing required command: $1"
}

cleanup() {
	if [ -n "$tmp_dir" ] && [ -d "$tmp_dir" ]; then
		rm -rf "$tmp_dir"
	fi
}
trap cleanup EXIT INT TERM

github_curl() {
	if [ -n "${GITHUB_TOKEN:-}" ]; then
		curl -fsSL \
			-H "Accept: application/vnd.github+json" \
			-H "Authorization: Bearer ${GITHUB_TOKEN}" \
			"$1"
		return
	fi
	curl -fsSL -H "Accept: application/vnd.github+json" "$1"
}

download() {
	curl -fsSL "$1" -o "$2"
}

os() {
	case "$(uname -s)" in
		Linux) printf linux ;;
		Darwin) printf darwin ;;
		*) fail "unsupported OS: $(uname -s)" ;;
	esac
}

arch() {
	case "$(uname -m)" in
		x86_64 | amd64) printf amd64 ;;
		arm64 | aarch64) printf arm64 ;;
		*) fail "unsupported architecture: $(uname -m)" ;;
	esac
}

sha256_verify() {
	file="$1"
	checksums="$2"
	checksum_line="$3"

	printf '%s\n' "$checksum_line" >"${checksums}.one"
	if command -v sha256sum >/dev/null 2>&1; then
		(cd "$(dirname "$file")" && sha256sum -c "${checksums}.one" >/dev/null)
	elif command -v shasum >/dev/null 2>&1; then
		(cd "$(dirname "$file")" && shasum -a 256 -c "${checksums}.one" >/dev/null)
	else
		fail "missing sha256sum or shasum for checksum verification"
	fi
}

if [ "${1:-}" = "--help" ] || [ "${1:-}" = "-h" ]; then
	usage
	exit 0
fi

need basename
need curl
need find
need grep
need head
need install
need mktemp
need sed
need tar

case "$version" in
	latest | v*) ;;
	*) version="v${version}" ;;
esac

os_name="$(os)"
arch_name="$(arch)"
api_url="https://api.github.com/repos/${repo}/releases"
if [ "$version" = "latest" ]; then
	release_url="${api_url}/latest"
else
	release_url="${api_url}/tags/${version}"
fi

log "Finding ${bin_name} ${version} for ${os_name}/${arch_name}..."
release_json="$(github_curl "$release_url")"
asset_url="$(
	printf '%s\n' "$release_json" |
		sed -n 's/.*"browser_download_url": "\([^"]*\.tar\.gz\)".*/\1/p' |
		grep -i "${os_name}" |
		grep -i "${arch_name}" |
		head -n 1
)"

if [ -z "$asset_url" ]; then
	fail "could not find a release archive for ${os_name}/${arch_name}"
fi

checksums_url="$(
	printf '%s\n' "$release_json" |
		sed -n 's/.*"browser_download_url": "\([^"]*checksums\.txt\)".*/\1/p' |
		head -n 1
)"
if [ -z "$checksums_url" ]; then
	fail "release is missing checksums.txt"
fi

archive="$(basename "$asset_url")"
tmp_dir="$(mktemp -d "${TMPDIR:-/tmp}/pimctl.XXXXXX")"
archive_path="${tmp_dir}/${archive}"
checksums_path="${tmp_dir}/checksums.txt"

log "Downloading ${archive}..."
download "$asset_url" "$archive_path"

log "Verifying checksum..."
download "$checksums_url" "$checksums_path"
checksum_line="$(grep "[[:space:]]${archive}$" "$checksums_path" || true)"
if [ -z "$checksum_line" ]; then
	fail "could not find ${archive} in checksums.txt"
fi
sha256_verify "$archive_path" "$checksums_path" "$checksum_line"

log "Installing to ${bin_dir}/${bin_name}..."
mkdir -p "$bin_dir"
tar -xzf "$archive_path" -C "$tmp_dir"
bin_path="$(find "$tmp_dir" -type f -name "$bin_name" -perm -u+x | head -n 1)"
if [ -z "$bin_path" ]; then
	bin_path="$(find "$tmp_dir" -type f -name "$bin_name" | head -n 1)"
fi
if [ -z "$bin_path" ]; then
	fail "archive did not contain ${bin_name}"
fi

install -m 0755 "$bin_path" "${bin_dir}/${bin_name}"
log "Installed: $(${bin_dir}/${bin_name} version 2>/dev/null || printf '%s' "${bin_dir}/${bin_name}")"

case ":${PATH}:" in
	*:"${bin_dir}":*) ;;
	*) log "Note: ${bin_dir} is not on PATH." ;;
esac
