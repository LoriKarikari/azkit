# pimctl

A CLI for Azure PIM resource role workflows — list eligible assignments, activate a role, check active status.

## Design

- Native Go client. No shelling out to `az`.
- Azure resource role PIM only. Entra roles and PIM for Groups are out of scope for now.
- Human output by default, `--json` for scripts.
- Kong for command parsing, Koanf for config, Charm stack for interactive flows.
- Vertical slices. No empty architecture folders.

## Docs

- Domain language: [`CONTEXT.md`](./CONTEXT.md)
- Architecture decisions: [`docs/adr/`](./docs/adr/)
- Agent/project conventions: [`AGENTS.md`](./AGENTS.md)

## License

MIT
