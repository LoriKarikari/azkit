# Switcher parity: rename, unset, import, and dynamic completion

We close the remaining kubectx/kubens parity gaps with subcommand-style surfaces, consistent with the existing `add`/`rm`/`current` shape rather than kubectx's `new=old` and `-u` flag syntax. `azkit ctx rename <old> <new>` atomically updates the stored context, active/previous references, and any name-keyed state. `azkit ctx unset` clears the context environment (including `AZURE_CONFIG_DIR`) and implies subscription unset; `azkit sub unset` clears the subscription variables. Both require shell integration, the same rule as switching.

`azkit ctx import` reads az's `azureProfile.json` from the default `AZURE_CONFIG_DIR` and proposes one context per distinct tenant, with interactive multi-select in a TTY and `--all` for non-interactive use. It never modifies az state.

Shell completion completes context names, subscription aliases, and subscription names from local stores only — the completion path never touches the network. Windows completion status is unchanged.
