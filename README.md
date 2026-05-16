# pimctl

[![Go Version](https://img.shields.io/github/go-mod/go-version/LoriKarikari/pimctl)](https://go.dev/dl)
[![CI](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml/badge.svg)](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LoriKarikari/pimctl)](https://goreportcard.com/report/github.com/LoriKarikari/pimctl)
[![License](https://img.shields.io/github/license/LoriKarikari/pimctl)](./LICENSE)

`pimctl` manages Azure PIM resource roles. It lists eligible assignments, activates roles, deactivates active assignments, and shows what is active now.

## Scope

`pimctl` supports Azure resource role PIM. It does not manage Entra roles or PIM for Groups.

## Install

Install the latest release on macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/LoriKarikari/pimctl/main/install.sh | sh
```

Install a specific version or directory:

```bash
PIMCTL_INSTALL_VERSION=v0.3.0 PIMCTL_INSTALL_DIR=/usr/local/bin \
  curl -fsSL https://raw.githubusercontent.com/LoriKarikari/pimctl/main/install.sh | sh
```

## Quickstart

List roles you can activate:

```bash
pimctl list
```

Activate a role:

```bash
pimctl activate \
  --subscription "Production" \
  --resource-group rg-app \
  --role Contributor \
  --reason "Investigating incident"
```

Check active assignments:

```bash
pimctl status
```

Find the `assignment_id` for an active assignment:

```bash
pimctl status --verbose
pimctl status --json
```

Default status output stays short. Use `--verbose` or `--json` when you need the ID for a script or deactivation.

Deactivate an active assignment:

```bash
pimctl deactivate <assignment-id>
```

Add a reason if you want one recorded with the request:

```bash
pimctl deactivate <assignment-id> --reason "Incident resolved"
```

In a terminal, `pimctl deactivate` opens a picker.

Success means Azure accepted the deactivation request. The assignment may still appear in `pimctl status` for a short time.

For scripts:

```bash
pimctl list --json
pimctl status --json
pimctl deactivate <assignment-id> --json
```

## Shell completion

Generate completion for your login shell:

```bash
pimctl completion
```

For bash or zsh, load it in the current shell with:

```bash
eval "$(pimctl completion)"
```

Completion is registered for the command name `pimctl`. If you are testing a local build with `./pimctl`, add an alias first:

```bash
go build -o pimctl ./cmd/pimctl
alias pimctl="$PWD/pimctl"
eval "$(pimctl completion)"
```

## Configuration

Config is optional. By default, `pimctl` reads:

```text
~/.config/pimctl/config.yaml
```

Example:

```yaml
default_duration: 2h
subscription_id: 00000000-0000-0000-0000-000000000000
```

Environment variables use the `PIMCTL_` prefix:

```bash
export PIMCTL_DEFAULT_DURATION=2h
export PIMCTL_SUBSCRIPTION_ID=00000000-0000-0000-0000-000000000000
```

## License

MIT
