# pimctl

[![Go Version](https://img.shields.io/github/go-mod/go-version/LoriKarikari/pimctl)](https://go.dev/dl)
[![CI](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml/badge.svg)](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LoriKarikari/pimctl)](https://goreportcard.com/report/github.com/LoriKarikari/pimctl)
[![License](https://img.shields.io/github/license/LoriKarikari/pimctl)](./LICENSE)

`pimctl` is a small CLI for Azure PIM resource role workflows.

It lists eligible assignments, activates roles, and shows current active assignments.

## Scope

`pimctl` supports Azure resource role PIM. It does not manage Entra roles or PIM for Groups.

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

Use JSON output for scripts:

```bash
pimctl list --json
pimctl status --json
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

## Docs

- Domain language: [`CONTEXT.md`](./CONTEXT.md)
- Architecture decisions: [`docs/adr/`](./docs/adr/)
- Agent/project conventions: [`AGENTS.md`](./AGENTS.md)

## License

MIT
