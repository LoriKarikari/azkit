# Kubernetes CLI ecosystem survey

**Question:** beyond kubectx/kubens, what did the Kubernetes CLI ecosystem build that made daily
operator life dramatically better — and which of those patterns translate to azkit?

**azkit's hard boundary:** mutating Azure resources is OUT; understanding your session, identity,
and world is IN. azkit today does tenant context, subscription switching, and Azure PIM role
activation. Every analog below is filtered through that boundary.

**Baseline for scale:** [kubectx/kubens](https://github.com/ahmetb/kubectx) sits at ~19.9k GitHub
stars — the pattern azkit already borrows. The tools below show what came next.

---

## 1. k9s — the read-only-first TUI

- **Pain:** `kubectl get` round-trips are slow for exploration. Answering "what is going on in
  this cluster right now?" takes a dozen commands. k9s gives a keyboard-driven terminal UI that
  watches resources live, so navigation replaces command composition.
- **Adoption:** ~33.8k stars ([derailed/k9s](https://github.com/derailed/k9s)); a fixture of every
  "must-have K8s tools" thread (e.g. [r/kubernetes](https://www.reddit.com/r/kubernetes/comments/1nnooxu/your_must_have_tools/)).
  Notably, most k9s usage is *observation* — people live in it to look, not to mutate.
- **Azure analog:** `azkit tui` (or `azkit dash`) — an interactive view of the session world:
  tenants, subscriptions, your eligible and active PIM roles with expiry countdowns, current
  identity. Navigation and switching/activation are the only "writes", and both are already in
  azkit's IN column. This is a natural umbrella over azkit's existing commands rather than new scope.

## 2. stern — tail the world, not one object

- **Pain:** `kubectl logs` targets one pod; workloads are many ephemeral pods. stern tails logs
  across pods/containers matched by regex, colored by source, surviving pod churn — "`tail -f`
  for dynamic multi-pod workloads."
- **Adoption:** ~4.8k stars ([stern/stern](https://github.com/stern/stern)); standard issue in
  debugging-tools roundups ([example](https://kubezilla.io/top-10-kubernetes-debugging-tools-every-devops-engineer-needs-in-2026/)).
- **Azure analog:** tailing *your own audit trail* — a follow-mode view of Azure Activity Log /
  PIM audit events / sign-in logs scoped to you or your subscriptions ("what did my session just
  do, and did my role activation land?"). Purely read-only, so inside the boundary, but it pulls
  azkit toward log plumbing — a real feature area, not a UX polish item.

## 3. krew — a plugin manager that grew an ecosystem

- **Pain:** kubectl can't ship every niche workflow. krew made `kubectl foo` extensions
  discoverable and installable, so the community filled the gaps.
- **Adoption:** ~7k stars ([kubernetes-sigs/krew](https://github.com/kubernetes-sigs/krew));
  200+ plugins in the [krew index](https://krew.sigs.k8s.io/docs/user-guide/discovering-plugins/).
- **Azure analog:** two readings. (a) Build a plugin system for azkit — premature; krew worked
  because kubectl already had massive install base. (b) *Be the plugin*: krew's real lesson is
  that focused single-purpose tools thrive when they compose with the main CLI. azkit should stay
  sharply scoped and interoperate cleanly with `az` (shared auth, `AZURE_CONFIG_DIR`, no config
  forking) rather than grow a marketplace.

## 4. kube-ps1 — context in the prompt

- **Pain:** the most expensive Kubernetes mistake is running a command against the wrong cluster.
  kube-ps1 puts current context + namespace in `$PS1` so the answer to "where am I?" is always on
  screen ([jonmosco/kube-ps1](https://github.com/jonmosco/kube-ps1), ~3.8k stars).
- **Azure analog:** `azkit prompt` — a fast, cache-backed command emitting current tenant,
  subscription, and (crucially, with no K8s equivalent) *active PIM roles and time-to-expiry* for
  embedding in zsh/fish/starship prompts. "You are Owner on prod for 47 more minutes" in the
  prompt is a genuinely novel safety win Kubernetes never needed. Squarely session-awareness: IN.

## 5. kubie — per-shell context isolation

- **Pain:** `kubectx` mutates a *global* kubeconfig, so switching context in one terminal silently
  retargets every other terminal — a classic wrong-cluster incident source. kubie spawns a
  sub-shell with an isolated context instead ([kubie-org/kubie](https://github.com/kubie-org/kubie),
  ~2.6k stars; explicitly positioned as the safe alternative to kubectx).
- **Azure analog:** the `az` CLI has the *identical* flaw — `az account set` writes global
  `~/.azure` state shared by all shells. `azkit shell <subscription>` could spawn a sub-shell with
  a scoped `AZURE_CONFIG_DIR` (or copied profile), so subscription/tenant context is per-terminal.
  Pure session management: IN, and arguably the highest-leverage safety feature available.

## 6. kubecolor — output humans can scan

- **Pain:** kubectl output is a monochrome wall; triage means re-reading. kubecolor colorizes it
  (errors red, headers distinct) with zero behavior change
  ([kubecolor/kubecolor](https://github.com/kubecolor/kubecolor), ~1.4k stars).
- **Azure analog:** not a separate tool — a design standard for azkit itself. `az` JSON dumps are
  the monochrome wall of the Azure world. azkit output should default to colored, aligned,
  human-first tables (expiring roles highlighted, current subscription marked) with color disabled
  when piped.

## 7. kubectl's own UX wins

- **fzf-backed completion / interactive selection** — kubectl completion plus fzf turned resource
  names from "copy-paste from a previous command" into fuzzy search. **Analog:** every azkit
  argument (tenant, subscription, role) should have shell completion *and* an interactive fuzzy
  picker when the argument is omitted — subscription GUIDs are even less typeable than pod names.
- **`-o` output formats** (`json`, `yaml`, `jsonpath`, `custom-columns`, `wide`) — one consistent
  contract made kubectl scriptable everywhere. **Analog:** a uniform `--output table|json|yaml`
  across all azkit commands, with stable JSON schemas so azkit slots into scripts and CI.
- **`kubectl auth whoami`** (stable in K8s 1.28, alpha 1.26;
  [docs](https://kubernetes.io/docs/reference/kubectl/generated/kubectl_auth/kubectl_auth_whoami/)) —
  kubectl went *years* without a way to ask "who does the server think I am?", painful exactly in
  dynamic-auth setups. **Analog:** `azkit whoami` — identity, tenant, subscription, token claims,
  auth method, active + eligible PIM roles, token expiry, in one command. This is the purest
  possible expression of azkit's mission statement.
- **`kubectl explain`** — schema documentation at the CLI, no browser. **Analog:** `azkit explain
  <role|scope>` — what does this RBAC role actually grant, what does this PIM eligibility mean,
  rendered from role definitions. Read-only reference lookup: IN.

---

## Ranked shortlist for azkit

| # | Feature | Rationale (one line) |
|---|---------|----------------------|
| 1 | `azkit whoami` | Cheapest-to-build, purest fit: one command answering "who am I, where, with which roles, until when." |
| 2 | Prompt integration (`azkit prompt`) | kube-ps1's proven wrong-target prevention, plus a PIM-expiry countdown Kubernetes never had. |
| 3 | Per-shell context isolation (`azkit shell`) | Fixes `az account set`'s global-state footgun the way kubie fixed kubectx's — biggest safety win. |
| 4 | Fuzzy pickers + completion everywhere | Subscription GUIDs and role names are hostile to typing; fzf-style selection is table stakes post-kubectx. |
| 5 | Uniform `--output` formats | Stable JSON/table contract makes azkit composable in scripts and CI, following kubectl's most durable win. |
| 6 | Read-only session TUI | k9s-scale demand proves operators live in observation dashboards; unifies items 1–4 in one screen. |
| 7 | `azkit explain` for roles/scopes | Low-effort read-only reference that demystifies what a PIM activation actually grants. |
| 8 | Activity/PIM audit log tailing | stern-style follow mode for your own audit trail; valuable but a heavier, later-stage investment. |

Plugin manager (krew analog) is deliberately unranked: azkit should *be* the sharp single-purpose
tool krew's ecosystem rewards, not host one.

## Sources

- k9s — https://github.com/derailed/k9s · https://k9scli.io/
- stern — https://github.com/stern/stern
- krew — https://github.com/kubernetes-sigs/krew · https://krew.sigs.k8s.io/docs/user-guide/discovering-plugins/
- kubectx/kubens — https://github.com/ahmetb/kubectx
- kube-ps1 — https://github.com/jonmosco/kube-ps1
- kubie — https://github.com/kubie-org/kubie
- kubecolor — https://github.com/kubecolor/kubecolor
- kubectl auth whoami — https://kubernetes.io/docs/reference/kubectl/generated/kubectl_auth/kubectl_auth_whoami/
- kubectl output formats & explain — https://kubernetes.io/docs/reference/kubectl/
- Community testimonials — https://www.reddit.com/r/kubernetes/comments/1nnooxu/your_must_have_tools/
