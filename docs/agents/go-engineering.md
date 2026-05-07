# Go Engineering Standards

`pimctl` should feel like a small, sharp, high-quality infrastructure tool: explicit, testable, boring where possible, and polished at the edges.

## Architecture rules

- Work in vertical slices. Create a package only when a feature needs it and tests exercise it.
- Do not pre-create empty architecture folders.
- Every new package must be justified by a working slice: CLI behavior, application behavior, adapter behavior, or tests.
- Keep command handlers thin. CLI packages parse input, call application use cases, and render output; they do not contain business logic.
- Define interfaces at the consumer boundary, not in provider packages.
- Do not leak Azure SDK types outside `internal/azurepim`.
- Keep domain types in `internal/domain` small, explicit, and independent from transport/UI concerns.
- Thread `context.Context` through all I/O and API boundaries.
- Avoid package-level mutable state. Prefer constructors and explicit dependencies.
- Keep configuration loading separate from application logic.
- Avoid `util` or `common` dumping-ground packages; use specific package names.

## Error handling

- Use typed application errors with stable machine-readable codes.
- Render errors differently for human output and JSON output.
- Wrap lower-level errors with enough context for debugging, but do not expose raw SDK errors directly to users.
- Prefer actionable messages: say what failed, why it likely failed, and what to try next.

## Testing

- Prefer table-driven tests for domain and application logic.
- Use fake Azure clients for app tests; normal tests must not require live Azure access.
- Use golden tests for human CLI output.
- Use JSON contract tests for `--json` output and error shapes.
- Use fake clocks when behavior depends on time.
- Gate live integration tests behind an explicit environment variable such as `PIMCTL_LIVE_TESTS=1`.

## Go style

- Run `gofmt`, `go vet`, `staticcheck`, and `golangci-lint` in CI.
- Keep functions short enough to read without jumping around; split by responsibility, not by arbitrary line count.
- Use generics only when they remove real duplication without obscuring intent.
- Use concurrency only where it improves user-visible latency or correctness. Prefer `errgroup` for coordinated concurrent work.
- Prefer clear names over clever abstractions.
- Make zero values useful where practical, but do not contort APIs to achieve it.

## CLI quality bar

- Human output is the default and should be readable, concise, and beautiful.
- `--json` output is for scripts and must remain stable except for additive fields before v1.0.
- Interactive flows should be interruptible, deterministic, and safe.
- Non-interactive commands must fail fast when required inputs are missing or ambiguous.
