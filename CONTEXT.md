# azkit

An umbrella CLI for Azure operator workflows, starting with Azure resource-role PIM, tenant contexts, and subscription switching.

## Language

**PIM**:
Azure Privileged Identity Management for eligible Azure resource roles.
_Avoid_: Entra role activation, PIM for Groups

**Azure resource role**:
An Azure RBAC role assigned at a resource scope, such as subscription or resource group.
_Avoid_: Directory role, Entra role

**Resource scope**:
The canonical Azure resource ID where an Azure resource role applies, such as a subscription or resource group ID.
_Avoid_: Target, location, environment

**Subscription selector**:
A subscription ID or exact subscription name used to derive a subscription-level resource scope; fuzzy matching is only allowed in interactive flows.
_Avoid_: Subscription scope, account

**Resource group selector**:
A resource group name within a selected subscription used to derive a resource-group-level resource scope.
_Avoid_: Resource group scope, target group

**Eligible assignment**:
A configured permission to activate an Azure resource role when needed.
_Avoid_: Available role, inactive role

**Eligibility expiry**:
The time when an eligible assignment stops being available for activation.
_Avoid_: Max duration, activation duration

**Activation**:
A time-bound elevation from eligible assignment to active Azure resource role assignment.
_Avoid_: Login, assume role

**Activation expiry**:
The time when an active Azure resource role assignment ends and the elevation drops.
_Avoid_: Logout, session timeout

**Deactivation**:
A self-initiated request to end an active assignment before its activation expiry.
_Avoid_: Logout, cancellation, expiry

**Deactivation request**:
The Azure request submitted for a deactivation. A successful request means Azure accepted the request; it does not prove the active assignment has disappeared from status yet.
_Avoid_: Deactivated assignment, confirmed deactivation

**Credential**:
The Azure identity used by the CLI to call Azure and Microsoft APIs.
_Avoid_: Session, login state

**Tenant context**:
A user-named binding to an Azure tenant. A tenant context owns a local credential cache directory for that tenant.
_Avoid_: Profile, environment

**Context catalog**:
The local list of saved tenant contexts. Listing the catalog is a local operation and does not call Azure.
_Avoid_: Account list, tenant registry

**Context credential cache**:
The local directory used as the Azure credential cache for one tenant context.
_Avoid_: Config directory, token store

**Active context**:
The tenant context selected for the current shell.
_Avoid_: Default tenant, current account

**Previous context**:
The tenant context that was active before the last context switch in the current shell.
_Avoid_: Backup context, fallback tenant

**Subscription cache**:
The per-context local cache of Azure subscriptions. It avoids repeated Azure calls and can be refreshed explicitly.
_Avoid_: Subscription catalog, account cache

**Interactive activation**:
A guided terminal flow for choosing and activating an eligible assignment.
_Avoid_: Wizard, dashboard

**Configuration**:
User-controlled defaults for activation duration and subscription selection.
_Avoid_: Preferences, settings

**Human output**:
The default terminal presentation optimized for readability rather than Azure CLI compatibility.
_Avoid_: Table mode, Azure-style output

**Activation reason**:
The user-provided audit justification for an activation.
_Avoid_: Comment, message

## Relationships

- An **Activation** activates exactly one **Azure resource role** for one Azure resource scope.
- A **Deactivation** ends one **Activation** before its **Activation expiry**.

## Example dialogue

> **Dev:** "Should `azkit pim activate` support Global Administrator?"
> **Domain expert:** "No — this CLI initially supports **PIM** for **Azure resource roles**, not Entra directory roles."

## Flagged ambiguities

- "PIM" was used ambiguously across Azure resource roles, Entra ID roles, and PIM for Groups — resolved: **PIM** means Azure resource role PIM first.
