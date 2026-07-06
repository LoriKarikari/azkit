# Deactivate accepts assignment IDs, not selectors

`azkit pim deactivate` takes an active assignment ID and submits the deactivation request. It does not mirror `azkit pim activate`'s `--role`, `--subscription`, and `--resource-group` flags. Active assignments are a small known set the user already holds. `azkit pim status` lists them, the interactive picker selects from them, and the ID flows through `--json` into scripts. Selectors would re-add the resolver and ambiguity errors to act on a known object, which is a worse UX than copy-paste.
