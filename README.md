# pimctl

`pimctl` is a high-quality Go CLI for activating and managing Azure Privileged Identity Management access for Azure resource roles.

The project is currently in planning. The first product slice will be `pimctl list`, showing eligible Azure resource role assignments with polished human output.

## Design direction

- Native Go client, not an Azure CLI wrapper.
- Azure resource role PIM first; Entra roles and PIM for Groups are out of scope for the initial product.
- Human-friendly terminal UX by default, stable `--json` output for scripts.
- Kong for command parsing, Koanf for layered configuration, and the Charm stack for interactive flows.
- Vertical slices only: no empty architecture folders.

## Documentation

- Domain language: [`CONTEXT.md`](./CONTEXT.md)
- Architecture decisions: [`docs/adr/`](./docs/adr/)
- Agent/project conventions: [`AGENTS.md`](./AGENTS.md)

## License

MIT
