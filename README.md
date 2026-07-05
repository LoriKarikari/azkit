# azkit

[![Go Version](https://img.shields.io/github/go-mod/go-version/LoriKarikari/pimctl)](https://go.dev/dl)
[![CI](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml/badge.svg)](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml)
[![golangci-lint](https://github.com/LoriKarikari/pimctl/actions/workflows/lint.yml/badge.svg?branch=main)](https://github.com/LoriKarikari/pimctl/actions/workflows/lint.yml)
[![License](https://img.shields.io/github/license/LoriKarikari/pimctl)](./LICENSE)

`azkit` is an umbrella CLI for Azure operator workflows. It keeps Azure resource-role PIM, tenant contexts, and subscription switching in one small tool.

PIM support is focused on Azure resource roles. `azkit` does not manage Entra roles or PIM for Groups.

## Install

Install the latest release on macOS or Linux:

```bash
curl -fsSL https://raw.githubusercontent.com/LoriKarikari/pimctl/main/install.sh | sh
```

Install a specific version or directory:

```bash
curl -fsSL https://raw.githubusercontent.com/LoriKarikari/pimctl/main/install.sh \
  | AZKIT_INSTALL_VERSION=v0.3.0 AZKIT_INSTALL_DIR=/usr/local/bin sh
```

The installer downloads the matching release archive, verifies `checksums.txt`, and installs the `azkit` binary.

## Shell integration

Context and subscription switching change environment variables in your current shell, so enable shell integration first:

```bash
eval "$(azkit shell-init bash)"
```

Use `zsh`, `powershell`, or `pwsh` instead of `bash` if that is your shell.

To make it permanent, add the same line to your shell startup file.

## Bootstrap a tenant context

Create a named tenant context:

```bash
azkit ctx add prod --tenant 00000000-0000-0000-0000-000000000000
```

Switch to it:

```bash
azkit ctx prod
```

If the context has not been logged in yet, `azkit` prints the Azure login command to run. The switch sets `AZURE_TENANT_ID`, `ARM_TENANT_ID`, and an isolated `AZURE_CONFIG_DIR` for that context.

```bash
az login --tenant 00000000-0000-0000-0000-000000000000
```

Useful context commands:

```bash
azkit ctx -l
azkit ctx -l --json
azkit ctx current
azkit ctx current --json
azkit ctx -        # switch to the previous context
azkit ctx rm prod --force
```

## Select a subscription

List subscriptions for the active context:

```bash
azkit sub -l
```

Refresh the subscription cache:

```bash
azkit sub --refresh
```

Switch by alias, exact subscription ID, or exact subscription name:

```bash
azkit sub prod
azkit sub 11111111-1111-1111-1111-111111111111
azkit sub "Production"
```

Create and remove aliases:

```bash
azkit sub alias prod "Production"
azkit sub unalias prod
```

Check or switch back:

```bash
azkit sub current
azkit sub current --json
azkit sub -
```

Subscription switching sets `AZKIT_SUBSCRIPTION_ID`, `AZURE_SUBSCRIPTION_ID`, `ARM_SUBSCRIPTION_ID`, and `ARM_SUBSCRIPTION_NAME` in the current shell.

## PIM workflows

PIM commands live under `azkit pim`.

List eligible role assignments:

```bash
azkit pim list
```

Activate a role:

```bash
azkit pim activate \
  --subscription "Production" \
  --resource-group rg-app \
  --role Contributor \
  --reason "Investigating incident"
```

Check active assignments:

```bash
azkit pim status
```

Find the `assignment_id` for an active assignment:

```bash
azkit pim status --extended
azkit pim status --json
```

Default status output stays short. Use `--extended` or `--json` when you need the ID for a script or deactivation.

Deactivate an active assignment:

```bash
azkit pim deactivate <assignment-id>
```

Add a reason if you want one recorded with the request:

```bash
azkit pim deactivate <assignment-id> --reason "Incident resolved"
```

In a terminal, `azkit pim activate`, `azkit pim deactivate`, `azkit ctx`, and `azkit sub` can open pickers when required inputs are missing.

For scripts:

```bash
azkit pim list --json
azkit pim status --json
azkit pim deactivate <assignment-id> --json
azkit ctx -l --json
azkit sub -l --json
```

## Shell completion

Generate completion for your login shell:

```bash
azkit completion
```

For bash or zsh, load it in the current shell with:

```bash
eval "$(azkit completion)"
```

If you are testing a local build:

```bash
go build -o azkit ./cmd/azkit
alias azkit="$PWD/azkit"
eval "$(azkit completion)"
```

## Configuration

Config is optional. By default, `azkit` reads:

```text
~/.config/azkit/config.yaml
```

Example:

```yaml
pim:
  default_duration: 2h
  subscription_id: 00000000-0000-0000-0000-000000000000
```

Environment variables use the `AZKIT_` prefix:

```bash
export AZKIT_PIM_DEFAULT_DURATION=2h
export AZKIT_PIM_SUBSCRIPTION_ID=00000000-0000-0000-0000-000000000000
```

## License

MIT
