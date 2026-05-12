# pimctl

`pimctl` is a high-quality Go CLI for activating and managing Azure Privileged Identity Management access for Azure resource roles.

## Shell completions

Generate a completion script for your shell, then save it where your shell loads completions.

Bash:

```bash
pimctl completion bash > ~/.local/share/bash-completion/completions/pimctl
```

Zsh:

```bash
mkdir -p ~/.zsh/completions
pimctl completion zsh > ~/.zsh/completions/_pimctl
```

Add `~/.zsh/completions` to `fpath` if it is not already there:

```zsh
fpath=(~/.zsh/completions $fpath)
autoload -Uz compinit && compinit
```

Fish:

```fish
pimctl completion fish > ~/.config/fish/completions/pimctl.fish
```

PowerShell:

```powershell
pimctl completion powershell | Out-String | Invoke-Expression
```

To load it every time, add that PowerShell command to your profile.

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
