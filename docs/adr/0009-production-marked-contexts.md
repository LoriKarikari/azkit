# Production-marked contexts

A context can be marked production (`production` field in the context store) via `--production` on `ctx add` or the new `azkit ctx edit <name> [--production|--no-production] [--tenant <id>]`. Marking is context-level only for now; subscription-level marking is deferred.

When a context is marked: switching to it prints a loud colored banner (TTY), `pim activate` requires interactive confirmation, and the prompt segment, whoami, and status output carry a distinct PROD marker. Confirmations never appear in `-o json` or non-TTY paths — scripts pass `--yes` or fail fast, per the CLI quality bar.
