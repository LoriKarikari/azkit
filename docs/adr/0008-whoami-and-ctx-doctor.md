# whoami snapshots, ctx doctor diagnoses

`azkit whoami` is a root command answering "who am I, where, until when" in one shot: identity from ARM token claims (upn/oid/tid) via the existing credential, active context, tenant, subscription, auth source, and token expiry — no Graph dependency and no extra network by default. `--full` adds active and eligible PIM role summaries. Not logged in yields an actionable error carrying the right `az login` command; the command works without shell integration.

`azkit ctx doctor [name]` (default: active context; `--all` for every context) is the health check: config dir existence and permissions, az login state, silent token acquisition, token expiry, token-tenant-matches-context-tenant, subscription cache staleness, and shell integration. Output is one row per check with a remediation hint; `-o json` emits an array of {check, status, detail, remediation}; the exit code is non-zero if any check fails, so it works in CI. Doctor only guides — it never mutates or drives re-auth itself.

The boundary is strict: whoami snapshots, doctor diagnoses. Neither grows the other's job.
