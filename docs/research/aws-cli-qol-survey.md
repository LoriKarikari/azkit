# AWS CLI quality-of-life tooling survey

Research brief for azkit (github.com/LoriKarikari/azkit), 2026-07-06.

The AWS ecosystem grew a layer of community tooling around credentials, profiles, role
assumption, and SSO that Azure never developed. This survey maps each major tool to the
pain it solves, its adoption evidence, and what the Azure analog would be inside azkit's
boundary: **mutating resources is out; identity, context, session, and credential
ergonomics are in.**

## Ecosystem context

Two facts frame everything below:

- **AWS credentials are exportable session objects.** `AssumeRole`/`GetSessionToken`
  return short-lived key/secret/token triples that any tool can mint, cache, scope, and
  hand to a subprocess. This is why an assume-role tooling ecosystem could exist at all.
- **Azure tokens are MSAL cache state.** `az login` writes `~/.azure/msal_token_cache.json`
  — in **plaintext on macOS and Linux** ([Microsoft docs](https://learn.microsoft.com/en-us/cli/azure/msal-based-azure-cli),
  [azure-cli#27176](https://github.com/Azure/azure-cli/issues/27176)) — and there is one
  global cache per `AZURE_CONFIG_DIR`. Azure's closest thing to assume-role is **PIM role
  activation**, which is portal-first with REST/PowerShell as the only automation paths
  ([activation docs](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-eligible-activate)).
  azkit's `pim activate` already fills the biggest gap; the question is which adjacent
  ergonomics to adopt.

## Tool-by-tool survey

### aws-vault (99designs)

- **Pain solved:** plaintext long-lived keys in `~/.aws/credentials`. aws-vault stores IAM
  credentials in the **OS keychain** (macOS Keychain, Windows Credential Manager, Secret
  Service/KWallet, pass, encrypted file), mints short-lived STS credentials from them, and
  exposes them to a subprocess via env vars or a local metadata/ECS-style credential
  server — the long-lived secret never touches the child process.
- **Adoption:** ~9k stars, 800+ forks; packaged in Homebrew, Chocolatey, Scoop, Arch,
  Nix. The 99designs repo is now **abandoned**, with active development continuing in the
  [ByteNess fork](https://github.com/ByteNess/aws-vault) — a sign the workflow is load-bearing
  enough that the community kept it alive. ([repo](https://github.com/99designs/aws-vault))
- **Azure analog:** encrypt the token cache. Concretely: keychain-backed storage for MSAL
  cache/refresh tokens in azkit-managed contexts (azkit already isolates `AZURE_CONFIG_DIR`
  per context — the natural place to fix what `az` leaves in plaintext), and an
  `azkit exec <ctx> -- <cmd>` that runs a command with a context's identity without
  mutating the parent shell. Directly addresses [azure-cli#27176](https://github.com/Azure/azure-cli/issues/27176).

### Granted / assume (Common Fate, now fwdcloudsec)

- **Pain solved:** slow role switching across many accounts, plus one-console-session-per-
  browser. `assume <profile>` fuzzy-picks a profile with **frecency ordering**; `assume -c`
  opens the AWS console via the federation endpoint in an isolated **browser container/profile**,
  so multiple accounts can be open simultaneously. It **encrypts cached SSO tokens** in the
  OS keyring (`pkg/securestorage`) explicitly to avoid plaintext tokens on disk, and syncs
  profile config from shared **profile registries** (git repos merged into `~/.aws/config`).
- **Adoption:** transferred to the Forward Cloud Security (fwdcloudsec) community org;
  Homebrew-packaged, active releases including a Chrome extension for Identity Center
  device-code confirmation. Frequently the top recommendation in r/aws profile-management
  threads. ([repo](https://github.com/common-fate/granted), [docs](https://docs.granted.dev/))
- **Azure analog:** azkit `ctx`/`sub` pickers already mirror `assume`; the gaps are
  (1) **frecency ordering** in pickers, (2) **`--console`/portal deep-link**: open the Azure
  portal scoped to the active context/subscription (or the PIM activation blade for a role)
  in a dedicated browser profile, and (3) a **shared context registry** so a team can
  distribute tenant/context definitions from a git repo.

### awsume (Trek10)

- **Pain solved:** the tedium of `sts assume-role` + exporting three env vars. `awsume
  <profile>` is a shell wrapper that follows `source_profile` chains, handles MFA, caches
  sessions, auto-refreshes, and sets credentials in the *current* shell. Supports custom
  role-session names and session policies; has a plugin system (e.g. console opening).
- **Adoption:** long-lived PyPI package (since ~2017) from a visible AWS consultancy;
  widely referenced in AWS onboarding docs and blog posts, though it has ceded mindshare
  to Granted in recent years. ([awsu.me](https://awsu.me/), [repo](https://github.com/trek10inc/awsume))
- **Azure analog:** azkit's `shell-init` + env-var mutation *is* the awsume pattern.
  Remaining ideas: session **auto-refresh awareness** (surface token/PIM expiry before a
  long operation fails) and awsume's `-` previous-profile flip, which azkit already has
  for `ctx`/`sub`.

### aws-sso-util (Ben Kehoe)

- **Pain solved:** IAM Identity Center gives users N accounts × M roles, but the AWS CLI
  makes you configure each profile by hand. `aws-sso-util configure populate` enumerates
  every account/role you can access and **generates all profiles** in `~/.aws/config`;
  `aws-sso-util login` logs in once per SSO instance rather than per profile; `credential-process`
  support backports SSO to tools that predate it.
- **Adoption:** qualitative but strong — repeatedly praised in r/aws, packaged in
  conda-forge and nixpkgs; parts of it were later absorbed into the AWS CLI itself (the
  `sso-session` config format), the classic "community tool becomes the native feature"
  arc. ([repo](https://github.com/benkehoe/aws-sso-util), [primer](https://github.com/benkehoe/aws-sso-util/blob/master/docs/primer.md))
- **Azure analog:** **discovery-driven config generation.** azkit can enumerate what you
  can reach — tenants, subscriptions, PIM-eligible roles — and generate contexts/aliases
  from it: `azkit ctx populate` (from `az account tenant list` / accessible tenants) and
  richer `azkit sub --refresh`. `azkit pim list` is already the eligibility half of this.

### Steampipe (Turbot)

- **Pain solved:** answering read-only questions across accounts/regions without writing
  SDK scripts — query cloud APIs with SQL (581 tables / 3,057 canned queries in the AWS
  plugin alone), heavily used for compliance and inventory.
- **Adoption:** thousands of GitHub stars across turbot repos; AWS itself published
  multiple blog posts promoting it; large plugin ecosystem including an **Azure plugin**.
  ([steampipe.io](https://steampipe.io/), [AWS plugin](https://hub.steampipe.io/plugins/turbot/aws),
  [AWS blog](https://aws.amazon.com/blogs/opensource/querying-aws-at-scale-across-apis-regions-and-accounts/))
- **Azure analog:** mostly **out of scope** — general resource querying is a different
  product, and Steampipe's Azure plugin already exists. The in-scope kernel is *identity
  introspection*: read-only "who am I, what can I do" queries (`azkit whoami`: identity,
  tenant, subscription, active + eligible roles, token expiry, `--json`). Don't compete
  with Steampipe; own the identity slice.

### AWS CLI native wins

- **Named profiles** (`--profile`, `AWS_PROFILE`): first-class multi-identity in the CLI
  since v1. Azure has nothing equivalent — `az` has one login state per config dir —
  which is precisely the hole `azkit ctx` fills via `AZURE_CONFIG_DIR` isolation.
- **`aws sso login` + `sso-session`**: one browser auth per SSO instance, cached and
  refreshed for all profiles that reference it. Azure analog: azkit detecting login state
  per context and (today) printing the right `az login --tenant …`; a deeper version
  drives the device-code/browser flow itself so `azkit ctx prod` is one step.
  ([docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html))
- **`credential_process`**: a documented extension point where any external program can
  supply credentials to the CLI/SDKs — the reason aws-vault, Granted, and aws-sso-util
  compose with everything. Azure's rough equivalents are `az account get-access-token`
  as a producer and Workload Identity/`AZURE_FEDERATED_TOKEN_FILE`-style plumbing as consumers.
  azkit analog: `azkit token` — emit an access token for the active context/subscription
  in formats tools expect, so azkit can sit behind Terraform, kubelogin-style flows, and
  scripts the way `credential_process` helpers sit behind the AWS SDK.
  ([docs](https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html))

## Cross-cutting lessons for azkit

1. **Secure token storage is the moral of aws-vault and Granted.** Both treat "no
   plaintext secrets on disk" as a headline feature. Azure CLI stores tokens in plaintext
   on macOS/Linux and the issue asking for Keychain support has been open since 2023.
   azkit's per-context config-dir isolation is the perfect insertion point.
2. **Role assumption won because it was one command.** aws-vault/awsume/Granted collapsed
   a multi-step STS dance into `assume prod-admin`. `azkit pim activate` is the same
   collapse for Azure's much clunkier REST flow — the survey confirms it is the right bet,
   and the ergonomic frontier is speed (frecency, defaults per context, re-activate last).
3. **Session isolation matters at two layers:** shell (subprocess with scoped creds —
   `exec`) and browser (per-account console sessions — Granted's container trick).
   azkit has the shell layer via env vars; `exec` and portal-launch are the missing pieces.
4. **Generate config from reality, don't make users write it.** aws-sso-util's populate
   pattern is the highest-leverage onboarding feature in the AWS ecosystem.

## Ranked shortlist for azkit

1. **Keychain-backed token cache for contexts** — aws-vault's core value prop, aimed at a
   documented, still-open Azure CLI security gap; no other Azure tool does it.
2. **`azkit exec <ctx> -- <cmd>`** — session isolation without shell mutation; makes azkit
   safe in scripts/CI and composes with Terraform the way aws-vault does.
3. **PIM activation ergonomics: frecency + `activate --last`/per-context defaults** —
   compounds the killer feature; Granted proved fast re-assumption is what retains users.
4. **`azkit ctx populate` (tenant/subscription discovery)** — aws-sso-util's populate
   pattern; kills the worst onboarding friction for multi-tenant operators.
5. **`azkit token` (credential_process-style emitter)** — makes azkit the credential
   broker other tools sit on top of, the composability lesson of the AWS ecosystem.
6. **Portal deep-link (`--console`) with browser-profile isolation** — Granted's
   most-loved feature translated to the Azure portal and PIM blades; higher effort,
   clear delight.
7. **Shared context registry (git-synced)** — Granted profile registries for teams;
   valuable but only after single-user ergonomics are done.
8. **`azkit whoami` identity introspection** — the in-scope sliver of Steampipe's value;
   cheap, useful, and reinforces the identity-tool positioning.

## Sources

- aws-vault: https://github.com/99designs/aws-vault (fork: https://github.com/ByteNess/aws-vault)
- Granted: https://github.com/common-fate/granted · https://granted.dev/ · https://docs.granted.dev/
- awsume: https://github.com/trek10inc/awsume · https://awsu.me/
- aws-sso-util: https://github.com/benkehoe/aws-sso-util
- Steampipe: https://github.com/turbot/steampipe · https://hub.steampipe.io/plugins/turbot/aws
- AWS CLI SSO config: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sso.html
- credential_process: https://docs.aws.amazon.com/cli/latest/userguide/cli-configure-sourcing-external.html
- Azure CLI plaintext token cache: https://learn.microsoft.com/en-us/cli/azure/msal-based-azure-cli · https://github.com/Azure/azure-cli/issues/27176
- PIM activation (portal/REST): https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-eligible-activate
