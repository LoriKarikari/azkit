# Azure community pain points survey

Research brief for azkit (github.com/LoriKarikari/azkit), July 2026. Question: what have Azure
operators wanted for years that Microsoft has not shipped in the `az` CLI or portal tooling?
Sources: Azure/azure-cli GitHub issues, Microsoft Learn docs (as evidence of official-workflow
gaps), r/AZURE and Hacker News threads, Stack Overflow, and community workaround tools.

Each pain is classified as:

- **(a)** already solved by azkit (tenant/subscription contexts with per-context
  `AZURE_CONFIG_DIR` isolation, plus PIM role activation from the CLI)
- **(b)** inside azkit's boundary but unsolved — candidate feature
- **(c)** outside azkit's boundary (generic resource CRUD, IaC, portal UX)

## 1. Global mutable `az account set` state — (a)

The single most structural complaint. Azure CLI persists the "current" subscription in a shared
config directory (`~/.azure`), so `az account set` in one terminal silently changes context for
every other terminal, script, and CI job on the machine. Each `az` invocation is a separate
process reading persisted state, so there is no per-shell context. The community-standard
workaround is exactly what azkit productizes: point each shell at its own `AZURE_CONFIG_DIR`
([Stack Overflow](https://stackoverflow.com/questions/33137145/can-i-access-multiple-azure-accounts-with-azure-cli-from-the-same-machine-at-sam),
[PowerShell parallel isolation Q&A](https://stackoverflow.com/questions/71923820/powershell-foreach-object-parallel-with-az-cli-profile-isolation)).
The same shared state causes race conditions in parallel CI pipelines, where Microsoft's own
guidance is to set a per-job `AZURE_CONFIG_DIR` and pass `--subscription` on every command
([config docs](https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration)).

Demand is also visible in the ecosystem of kubectx-inspired switchers people built because `az`
lacks one: [whiteducksoftware/azctx](https://github.com/whiteducksoftware/azctx),
[azurectx](https://pypi.org/project/azurectx/). r/AZURE Kubernetes threads explicitly praise
`kubectx`/`kubens` ergonomics ([thread](https://www.reddit.com/r/AZURE/comments/1ro5kjm/kubernetes_admins_what_kubectl_commands_do_you/)).
azkit's isolated contexts go further than these tools, which only flip the shared default.

## 2. PIM activation is effectively portal-only — (a) core, (b) around the edges

Microsoft's activation docs for eligible Azure RBAC roles describe only portal and mobile-app
flows ([role-assignments-eligible-activate](https://learn.microsoft.com/en-us/azure/role-based-access-control/role-assignments-eligible-activate),
[pim-how-to-activate-role](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-how-to-activate-role)).
There is no native `az` command to self-activate an eligible role; operators script raw
`roleAssignmentScheduleRequests` REST calls with `"requestType": "SelfActivate"` instead
([community gist](https://gist.github.com/ryanotella/161cbeea9fceb679b73f49f7bc95b3d1),
[second gist](https://gist.github.com/trietsch/a10ff159b5fe85db56e8f5b2764e5ca5)). The same gap
exists for Entra roles — a feature request landed in a third-party CLI rather than Microsoft's
([cli-microsoft365 #5766](https://github.com/pnp/cli-microsoft365/issues/5766)). Because org
policies cap activation at a few hours ([role settings docs](https://learn.microsoft.com/en-us/entra/id-governance/privileged-identity-management/pim-resource-roles-configure-role-settings)),
operators repeat this portal dance several times a day. azkit's CLI activation solves the core;
the recurring-friction edges (expiry visibility, re-activation, group/Entra scope) remain — see
candidates below.

## 3. Login and token UX — (b)

Recurring complaints: re-running `az login` whenever refresh tokens are invalidated by
Conditional Access ([azure-cli #17714](https://github.com/Azure/azure-cli/issues/17714),
[Stack Overflow](https://stackoverflow.com/questions/65892743/how-to-stay-logged-in-with-azure-cli));
no first-class way to be logged in as multiple accounts simultaneously — one login, then flip
the shared default ([manage-subscriptions docs](https://learn.microsoft.com/en-us/cli/azure/manage-azure-subscriptions-azure-cli));
`az login --scope` requiring ARM access and `--allow-no-subscriptions` confusion
([azure-cli #30769](https://github.com/Azure/azure-cli/issues/30769)); "found multiple accounts
with the same username" errors ([Server Fault](https://serverfault.com/questions/1143038/azure-cli-az-login-error-found-multiple-accounts-with-the-same-username)).
azkit's per-context isolation already enables side-by-side identities, but the surrounding UX
(token expiry visibility, guided re-auth, per-context login status) is unbuilt.

## 4. Cross-tenant operations break subtly — (a) partially, (b) remainder

`--subscription` is acknowledged by maintainers as broken in cross-tenant scenarios (for
example the Key Vault `AKV10032: Invalid issuer` failure when targeting a non-default tenant,
[azure-cli #11871](https://github.com/Azure/azure-cli/issues/11871)); `MissingSubscription`
errors persist even with explicit flags
([azure-cli #28372](https://github.com/Azure/azure-cli/issues/28372)). azkit's tenant-scoped
contexts remove the ambient-tenant footgun for its own operations; diagnosing or surfacing
wrong-tenant-token conditions before a command fails is a candidate.

## 5. Slow CLI startup and general sluggishness — (c) mostly, (b) as positioning

Well-documented: `az` commands hanging or extremely slow
([Microsoft Q&A](https://learn.microsoft.com/en-us/answers/questions/2146588/az-commands-hang-or-are-extremely-slow),
[Stack Overflow](https://stackoverflow.com/questions/79112841/az-cli-runs-very-slow-after-a-while)),
Python packaging pain on Windows acknowledged by Microsoft
([Azure Tools blog](https://techcommunity.microsoft.com/blog/azuretoolsblog/azure-cli-windows-msi-upgrade-issue-root-cause-mitigation-and-performance-improv/4491691),
[broken 2.77.0 report](https://hungyi.net/Tech/Azure-CLI-2.77.0-Broken-on-Windows)), and AWS
migrants on r/AZURE calling the CLI "very slow in comparison to AWS"
([thread](https://www.reddit.com/r/AZURE/comments/1npasya)). azkit cannot fix `az`, but being a
fast single-binary Go tool for the hot paths (context switch, PIM activate) is exactly the
contrast operators are asking for — a positioning point, not a feature.

## 6. Output, formatting, and query ergonomics — (c) mostly

JSON-by-default plus JMESPath `--query` is a steady source of friction: table output drops
nested objects ([format docs](https://learn.microsoft.com/en-us/cli/azure/format-output-azure-cli)),
queried columns render inconsistently ([Stack Overflow](https://stackoverflow.com/questions/66507220/az-cli-query-not-displaying-all-columns-properly)),
and AWS users miss idioms like `--cli-input-json`
([r/AZURE](https://www.reddit.com/r/AZURE/comments/1tbz43m/is_there_an_az_cli_equivalent_to_aws_cliinputjson)).
Fixing this generally is out of scope; for azkit's own commands, human-first tables with
`--output json` for scripting is table stakes.

## 7. Shell integration gaps — (b)

PowerShell tab completion only arrived in az 2.49 after years of complaints
([PowerShell tips](https://learn.microsoft.com/en-us/cli/azure/use-azure-cli-successfully-powershell)),
and docs remain bash-first ([troubleshooting doc](https://learn.microsoft.com/en-us/cli/azure/use-azure-cli-successfully-troubleshooting)).
For a context-switching tool, prompt visibility of the active context (the `kube-ps1` pattern)
and completions across bash/zsh/pwsh are the expected ecosystem features. The repo's recent
"clarify completion on Windows" fix suggests completion exists; prompt integration is the gap.

## 8. Portal/CLI/PowerShell fragmentation, general Azure UX — (c)

HN threads ([My Poor Experience With Azure](https://news.ycombinator.com/item?id=32139672)) and
r/AZURE ([Why does the Azure Portal suck?](https://www.reddit.com/r/AZURE/comments/16w0ush),
[biggest pain with AWS/GCP/Azure](https://www.reddit.com/r/AZURE/comments/1oxlon1/whats_your_biggest_pain_with_awsgcpazure/))
complain about fragmented experiences across portal, `az`, and Az PowerShell, inconsistent
extensions, and doc quality. Out of azkit's boundary except where it can be the one coherent
tool for identity-context workflows.

## Ranked candidate features — category (b)

1. **PIM status and expiry visibility** (`azkit pim status` / prompt segment) — the daily pain
   is not just activating but knowing what is active, in which context, and when it expires;
   nothing surfaces this outside the portal.
2. **Re-activation ergonomics** (`azkit pim renew`, expiry warnings, optional
   notify-before-expiry) — org policies force activation every 4–8 hours; removing the
   repeated form-filling is the highest-frequency win after activation itself.
3. **Shell prompt integration** (kube-ps1-style active tenant/subscription/PIM indicator for
   bash/zsh/pwsh) — mutable context is only safe when it is visible; this is the proven
   kubectx-ecosystem companion pattern.
4. **Per-context login/token health** (`azkit ctx doctor` / status showing token expiry,
   Conditional Access re-auth needed, wrong-tenant token) — turns the #3 and #4 complaint
   clusters from silent failures into one diagnostic command.
5. **PIM for Entra roles and PIM groups** — the same portal-only gap exists beyond Azure RBAC
   (cli-microsoft365 #5766 shows demand); natural extension of azkit's PIM surface.
6. **Ephemeral/CI context mode** (`azkit ctx exec --temp -- <cmd>` with a throwaway config dir)
   — packages Microsoft's own per-job `AZURE_CONFIG_DIR` CI guidance into one command and
   kills the parallel-pipeline race class.
7. **Approval workflow support** (list/approve pending PIM requests from the CLI) — approvers
   have the same portal-only friction as requesters; lower frequency than 1–2 but same boundary.
