# Keychain-backed token cache: feasibility

Research brief for azkit (Go CLI managing Azure tenant contexts via per-context `AZURE_CONFIG_DIR`,
subscription switching, and PIM role activation). Question: can azkit encrypt the per-context MSAL
token cache with the OS keychain, and in what design shape?

## 1. How az / MSAL read and write the token cache

Azure CLI persists tokens through [msal-extensions for Python](https://github.com/AzureAD/microsoft-authentication-extensions-for-python),
which offers four persistence backends:

| Backend | Platform | On-disk artifact |
|---|---|---|
| `FilePersistence` | all | plaintext `msal_token_cache.json` |
| `FilePersistenceWithDataProtection` | Windows | DPAPI-encrypted `msal_token_cache.bin` |
| `KeychainPersistence` | macOS | cache blob lives **in a Keychain item**; the file path is only a change-detection probe |
| `LibsecretPersistence` | Linux | blob in Secret Service (GNOME Keyring/KWallet); requires a D-Bus session |

What az actually enables ([`should_encrypt_token_cache`](https://github.com/Azure/azure-cli/issues/27176),
[MSAL-based Azure CLI docs](https://learn.microsoft.com/en-us/cli/azure/msal-based-azure-cli)):

- **Windows**: encryption on by default (`fallback = sys.platform.startswith('win32')`) — DPAPI-protected `msal_token_cache.bin`.
- **macOS / Linux**: plaintext `msal_token_cache.json` by default. There is an undocumented, comment-marked-EXPERIMENTAL
  config knob `core.encrypt_token_cache=true` that switches macOS to `KeychainPersistence` — but as
  [azure-cli#27176](https://github.com/Azure/azure-cli/issues/27176) documents, it stores the item under the
  hardcoded service/account names `"my_service_name"` / `"my_account_name"`, it is undocumented in the
  [official config reference](https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration), and Microsoft has
  left it off "for now". Known failure modes for keychain/libsecret persistence: locked keychain over SSH,
  headless sessions, and no Secret Service on minimal Linux hosts (msal-extensions exposes
  `fallback_to_plaintext` for exactly this reason).

**The decisive compatibility fact:** the cache format is not pluggable from the outside. az deserializes the
persisted blob directly as MSAL's JSON schema (or DPAPI-wrapped bytes on Windows). If azkit wrote its own
encrypted bytes into `msal_token_cache.json` inside a context's `AZURE_CONFIG_DIR`, az would fail to parse the
cache and treat the user as logged out (or error). There is **no** az/MSAL knob that says "decrypt with key X" —
the only knob is `core.encrypt_token_cache`, which merely selects among msal-extensions' *own* backends and is
only default-on and production-quality on Windows.

Two nuances worth knowing:

- az reads its config from `$AZURE_CONFIG_DIR/config`, so `core.encrypt_token_cache` **is settable per azkit
  context**. On Windows that is a no-op (already on). On macOS it flips az to the shared, generically-named
  Keychain item — which **breaks per-context isolation**: all contexts using `KeychainPersistence` would
  read/write the *same* keychain item regardless of `AZURE_CONFIG_DIR`, since the service/account names are
  hardcoded. Not usable as-is.
- The Go SDK has [`azidentity/cache`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache),
  which gives azkit's *own* Go code MSAL-compatible persistent caching encrypted with DPAPI/Keychain/libsecret on
  all three platforms — but it uses its own storage location and is not shared with az's cache.

## 2. Architecture options if az cannot read an azkit-encrypted cache

### (a) azkit-only protection — recommended
azkit stores the secrets *it* owns (tokens it acquires for PIM/ARM calls, plus any context metadata secrets) in
the OS keychain, and never touches az's cache format. az's cache stays exactly as az manages it.

- **Security value:** real but scoped — azkit's tokens get at-rest encryption + OS access control; az's cache on
  macOS/Linux remains plaintext, which is Microsoft's problem to fix, not azkit's to work around.
- **UX cost:** low. One keychain-unlock prompt per session at most; granted/aws-vault users already accept this.
- **Compatibility risk:** none for az. Headless/CI needs a fallback (encrypted file or plain env-based auth).

### (b) Broker pattern: decrypt-to-plaintext just-in-time for az
azkit keeps the canonical cache encrypted (keychain or encrypted file) and materializes a plaintext
`msal_token_cache.json` only while an az invocation runs, then re-encrypts and shreds it.

- **Security value:** mostly theater. The plaintext window exists on every az call; az writes back refreshed
  tokens azkit must re-capture; crashes leave plaintext behind; any local attacker who can read files can also
  wait for the window or read process memory. Race-prone with concurrent az invocations and background refresh.
- **UX cost:** every az call must go through `azkit exec`-style wrapping; direct `az` use silently sees a stale or
  missing cache.
- **Compatibility risk:** high — az's cache write-back, file locking (`.lock` files msal-extensions creates), and
  extensions that touch the cache all fight the broker.

### (c) Rely on OS-level protections; drop the feature
File permissions (0600), FileVault/BitLocker/LUKS full-disk encryption, and documenting
`core.encrypt_token_cache=true` for Windows-parity awareness.

- **Security value:** the honest baseline; identical to what az itself ships on macOS/Linux.
- **UX cost:** zero. **Compatibility risk:** zero.
- Weakness: no defense against same-user malware/exfiltration of the cache file — but neither does (b), really,
  and (a) only defends azkit's own tokens.

**Hybrid worth noting:** (a) + set `core.encrypt_token_cache=true` in per-context config **on Windows only**
(harmless, already default) and surface a doc note pointing at azure-cli#27176 for macOS/Linux users who want to
experiment. Do not auto-enable it on macOS because of the shared-hardcoded-item isolation break above.

## 3. Go secure-storage libraries

| Library | Backends | Assessment |
|---|---|---|
| [99designs/keyring](https://github.com/99designs/keyring) | macOS Keychain, Windows Credential Manager, Secret Service, KWallet, `pass`, keyctl, **encrypted-file fallback** | The de-facto standard (aws-vault, granted). Not archived; LFX Insights rates it "Stable" but with a small contributor base. Parent project aws-vault is explicitly abandoned (README points to the active [ByteNess/aws-vault](https://github.com/ByteNess/aws-vault) fork), so expect slow upstream movement — vendoring or pinning is fine, it's mature and low-churn. |
| [zalando/go-keyring](https://github.com/zalando/go-keyring) | macOS (`/usr/bin/security` exec), Windows wincred, Secret Service | Actively maintained, tiny API (`Set/Get/Delete`), zero cgo. No file fallback, no backend selection, no KWallet/pass. Good for "one small secret", weak for structured multi-item storage. |
| [keybase/go-keychain](https://github.com/keybase/go-keychain) | macOS/iOS native Keychain (cgo), Linux Secret Service | Best low-level Apple Keychain control (access groups, trusted applications). No Windows. Use only if you need Keychain-specific ACLs. |
| granted's `pkg/securestorage` | wraps **99designs/keyring** | See §4. |

Pain points to design around:

- **Item size limits:** Windows Credential Manager caps a credential blob at ~2560 bytes (`CRED_MAX_CREDENTIAL_BLOB_SIZE`).
  A full MSAL cache (multiple accounts × refresh/access/ID tokens) is routinely tens of KB — storing *the whole
  cache* as one keyring item is not portable. Store small, per-purpose secrets (a refresh token, or an encryption
  key that wraps a file) instead. This alone rules out "put msal_token_cache.json in the keychain" designs.
- **Headless/CI:** Secret Service needs D-Bus + an unlocked collection; macOS Keychain needs an unlocked login
  keychain (SSH sessions fail). 99designs/keyring's encrypted-file backend with a passphrase (env-suppliable, cf.
  granted's `CF_KEYRING_FILE_PASSWORD`) is the standard escape hatch; also honor an explicit backend override.
- **Prompt fatigue (macOS):** each new binary signature triggers "wants to access" prompts;
  `KeychainTrustApplication: true` and stable code signing mitigate.

## 4. Prior art

**aws-vault** ([99designs/aws-vault](https://github.com/99designs/aws-vault), maintained fork
[ByteNess/aws-vault](https://github.com/ByteNess/aws-vault)): a `vault` layer over 99designs/keyring. Patterns
worth copying: user-selectable backend via `--backend`/env with a sane platform default; small discrete items
(one item per profile's credentials/session, never one giant blob); JWE-encrypted flat-file backend as the
universal fallback; credentials handed to child processes via env or a local metadata-endpoint server — plaintext
never hits disk.

**granted** ([fwdcloudsec/granted `pkg/securestorage`](https://github.com/fwdcloudsec/granted/tree/main/pkg/securestorage)):
thin `SecureStorage{StorageSuffix}` wrapper over 99designs/keyring. Patterns worth copying: one storage namespace
per secret class (`granted-aws-sso-tokens`, `granted-session-credentials`, …) so items stay small and separately
access-controllable; every keyring `Config` field (backend allow-list, keychain name, file dir) overridable from
user config; `FilePasswordFunc` reads an env var so the file fallback works in CI; keyring errors degrade to a
warning + re-auth instead of hard failure; token structs carry refresh material (client id/secret + refresh
token) so `GetValidSSOToken` can silently refresh and re-store. Notably, granted does **not** encrypt the AWS
CLI's own `~/.aws/sso/cache` — it protects only the tokens it acquires itself. That is exactly option (a).

## Verdict: **feasible-with-constraints**

Encrypting `msal_token_cache.json` in place is **infeasible** — az cannot read a foreign format, the only native
knob is Windows-only-by-default and broken for per-context isolation on macOS, and Windows Credential Manager
size limits kill whole-cache-in-keychain designs. But keychain-backed protection of **azkit's own tokens** is
feasible and is what the mature prior art (granted, aws-vault) ships.

**Recommended design — option (a), granted-shaped:**

1. Add a `securestorage`-style package wrapping **99designs/keyring** (pin the version; it is stable but
   low-activity) with: platform-default backend, user-overridable backend allow-list, encrypted-file fallback
   whose passphrase can come from an env var, and per-secret-class service names namespaced per azkit context
   (e.g. `azkit-<context>-tokens`) to preserve context isolation.
2. Route azkit's own PIM/ARM token acquisition through it — or evaluate
   [`azidentity/cache`](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache), which gives
   Microsoft-maintained encrypted MSAL caching for Go on all three platforms for free.
3. Leave each context's az cache untouched; document its plaintext status on macOS/Linux, link azure-cli#27176,
   and rely on `AZURE_CONFIG_DIR` permissions + full-disk encryption as the az-side baseline.
4. Degrade gracefully: keyring unavailable → warn once, fall back to encrypted file or re-auth; never block a
   command on a keychain prompt in non-interactive mode.

**Open questions for the decision ticket:**

- Does azkit acquire tokens itself (azidentity) or purely shell out to az today? If purely shelling out, option
  (a) protects nothing until azkit owns at least the PIM token acquisition path — is that in scope?
- `99designs/keyring` vs `azidentity/cache` for azkit's own tokens: full backend control + file fallback vs
  Microsoft-maintained MSAL compatibility with less configurability. Pick one.
- Headless/CI policy: is an env-passphrase encrypted file acceptable, or should azkit require interactive
  keychains and refuse to persist otherwise?
- Should azkit set `core.encrypt_token_cache=true` in per-context az config on Windows (explicit, but redundant)
  and/or expose an opt-in flag for the experimental macOS path despite the shared-item isolation break?
- One keychain item per context vs one per (context, token-type): prompt frequency vs blast-radius trade-off.

## Sources

- [azure-cli#27176 — token cache encryption on macOS](https://github.com/Azure/azure-cli/issues/27176)
- [MSAL-based Azure CLI (cache encrypted on Windows, plaintext on Linux/macOS)](https://learn.microsoft.com/en-us/cli/azure/msal-based-azure-cli)
- [msal-extensions for Python (DPAPI / Keychain / libsecret backends, `fallback_to_plaintext`)](https://github.com/AzureAD/microsoft-authentication-extensions-for-python)
- [Azure CLI configuration reference (no `encrypt_token_cache` documented)](https://learn.microsoft.com/en-us/cli/azure/azure-cli-configuration)
- [azidentity/cache (Go persistent MSAL cache with platform encryption)](https://pkg.go.dev/github.com/Azure/azure-sdk-for-go/sdk/azidentity/cache)
- [99designs/keyring](https://github.com/99designs/keyring) · [99designs/aws-vault (abandoned)](https://github.com/99designs/aws-vault) · [ByteNess/aws-vault fork](https://github.com/ByteNess/aws-vault)
- [zalando/go-keyring](https://github.com/zalando/go-keyring) · [keybase/go-keychain](https://github.com/keybase/go-keychain)
- [granted securestorage source](https://github.com/fwdcloudsec/granted/tree/main/pkg/securestorage) · [granted security docs](https://docs.granted.dev/security/)
