# Deactivate accepts assignment IDs, not selectors

`pimctl deactivate` takes an active assignment ID and submits the deactivation request. It does not mirror `pimctl activate`'s `--role`, `--subscription`, and `--resource-group` flags. Active assignments are a small known set the user already holds. `pimctl status` lists them, the interactive picker selects from them, and the ID flows through `--json` into scripts. Selectors would re-add the resolver and ambiguity errors to act on a known object, which is a worse UX than copy-paste.
