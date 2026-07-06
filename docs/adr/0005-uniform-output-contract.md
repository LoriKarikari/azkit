# Uniform output contract with a global --output flag

All commands share one output contract: a global `-o/--output` flag with exactly two formats, `table` (default) and `json`. `--json` remains a documented alias until v1.0. We are deliberately not shipping yaml, tsv, or jsonpath — `jq` exists, and formats can be added additively if real demand shows.

Every command supports `-o json` except the shell-env switch paths (`ctx <name>`, `ctx -`, `sub <selector>`, `sub -`), whose stdout belongs to shell integration; they keep an explicit error pointing at `ctx current -o json` and `sub current -o json`. JSON output is one document per command on stdout with snake_case fields; schemas are additive-only before v1.0 and freeze at v1.0 except for additive fields. Errors keep the existing machine-readable JSON error shape.

Human table output is colored only when stdout is a TTY, honors `NO_COLOR`, and stays plain and byte-stable when piped, so golden tests and scripts never churn.
