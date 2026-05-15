# pimctl

[![Go Version](https://img.shields.io/github/go-mod/go-version/LoriKarikari/pimctl)](https://go.dev/dl)
[![CI](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml/badge.svg)](https://github.com/LoriKarikari/pimctl/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/LoriKarikari/pimctl)](https://goreportcard.com/report/github.com/LoriKarikari/pimctl)
[![License](https://img.shields.io/github/license/LoriKarikari/pimctl)](./LICENSE)
[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FLoriKarikari%2Fpimctl.svg?type=shield)](https://app.fossa.com/projects/git%2Bgithub.com%2FLoriKarikari%2Fpimctl?ref=badge_shield)

`pimctl` manages Azure PIM resource roles. It lists eligible assignments, activates roles, deactivates active assignments, and shows what is active now.

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


[![FOSSA Status](https://app.fossa.com/api/projects/git%2Bgithub.com%2FLoriKarikari%2Fpimctl.svg?type=large)](https://app.fossa.com/projects/git%2Bgithub.com%2FLoriKarikari%2Fpimctl?ref=badge_large)