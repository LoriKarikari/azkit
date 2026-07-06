# Prompt segment and exec-under-context; no per-directory auto-switch

`azkit prompt` emits a single-line prompt segment — context:subscription, PROD marker, and a PIM expiry countdown when under the warning threshold — with a hard budget under 50ms: it reads only local state and never touches the network, tolerating stale PIM data. Color follows the TTY rules of the output contract. The README documents starship, PS1, and pwsh recipes.

`azkit exec [--ctx <name>] [--sub <selector>] [--temp] -- <cmd>` computes the context/subscription environment and runs one command without mutating the parent shell. `--temp` seeds an ephemeral `AZURE_CONFIG_DIR` copy and deletes it afterwards, which is the answer for CI jobs and parallel pipelines racing on shared az state.

Per-directory contexts (a `.azkit` file auto-switching on cd) are rejected: auto-switching credentials by directory is a security-sensitive footgun, and direnv already owns that trust prompt. We document a direnv recipe that calls azkit instead.
