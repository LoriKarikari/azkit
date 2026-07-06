# PIM presets and activation lifecycle

Activation presets are named bundles in config (`pim.presets.<name>`: role, subscription, resource group, duration, reason). `azkit pim activate <preset>` resolves the positional selector against presets first, and explicit flags override preset fields. A preset may bundle multiple role activations — presets are the bulk-activation mechanism; we are not adding a separate bulk flag.

`--wait` polls an approval-gated activation until it is active, bounded by `--wait-timeout` (default 30m), exits 0 only when active, and is interruptible. `azkit pim renew <assignment-id>` re-activates with the same parameters, and `azkit pim renew --last` re-submits the most recent activation from per-context local history. `azkit pim deactivate --all` deactivates everything active after confirmation (`--yes` for scripts).

`pim status` gains a time-remaining column with warning color under 15 minutes (TTY only), and the same local expiry data feeds the prompt segment. Interactive pickers adopt frecency ordering. The PIM approval workflow (approving others' requests) stays in the later backlog.
