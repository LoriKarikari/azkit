# pimctl

A command-line tool for activating and managing Azure Privileged Identity Management access with high-quality Go CLI ergonomics.

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

**Credential**:
The Azure identity used by the CLI to call Azure and Microsoft APIs.
_Avoid_: Session, login state

**Interactive activation**:
A guided terminal flow for choosing and activating an eligible assignment.
_Avoid_: Wizard, dashboard

**Configuration**:
User-controlled defaults for Azure tenant, subscription, color, and activation duration.
_Avoid_: Preferences, settings

**Human output**:
The default terminal presentation optimized for readability rather than Azure CLI compatibility.
_Avoid_: Table mode, Azure-style output

**Activation reason**:
The user-provided audit justification for an activation.
_Avoid_: Comment, message

## Relationships

- An **Activation** activates exactly one **Azure resource role** for one Azure resource scope.

## Example dialogue

> **Dev:** "Should `az-pim activate` support Global Administrator?"
> **Domain expert:** "No — this CLI initially supports **PIM** for **Azure resource roles**, not Entra directory roles."

## Flagged ambiguities

- "PIM" was used ambiguously across Azure resource roles, Entra ID roles, and PIM for Groups — resolved: **PIM** means Azure resource role PIM first.
