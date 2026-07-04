# Rename pimctl to azkit umbrella CLI

We will rename the primary binary from `pimctl` to `azkit` and move existing PIM workflows under the `azkit pim` command group. The CLI is becoming an Azure operator toolkit rather than a single-purpose PIM tool. PIM remains the first supported workflow, but context and subscription switching will live beside it as first-class commands.

Root-level PIM aliases such as `azkit list` and `azkit status` will not exist. Users should run `azkit pim list`, `azkit pim status`, `azkit pim activate`, and `azkit pim deactivate`. Keeping PIM under an explicit command group leaves room for `azkit ctx` and `azkit sub` without overloading the root namespace.

The rename is a hard switch. The Go module, release binary, completion command name, configuration shape, and documentation should use azkit naming once each migration slice lands. Existing PIM behavior and JSON contracts should stay stable where the command path is the only thing changing.
