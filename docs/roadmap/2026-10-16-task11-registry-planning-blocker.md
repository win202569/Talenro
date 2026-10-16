# Task11 registry planning — stopped at request-authentication contract gap

Latest:C1 clarification was locally approved after247f1f41. Resume from the
[complete source index/C2 report](2026-10-16-task11-registry-reconciliation.md),
not the historical C1-open recommendations below. Correct object count is146;
earlier147 incorrectly included the table header. No protocol object removed.
Canonical-set registration/deployment gates remain open;no executable plan.

Follow-up: user approved preparing the [C1 clarification proposal](../superpowers/specs/2026-10-16-authority-v7-request-auth-clarification.md).
The proposal suggests a dedicated request role/common envelope with no bundle
evidence eligibility,not a retroactive claim about the old spec. Written approval
is pending;C1 remains OPEN and no implementation-ready plan exists.

Baseline:59bb7226277b04510aeb0068fe45e5be3573c2b2. User approved the
[written design](../superpowers/specs/2026-10-16-authority-v7-registry-design.md)
and authorized scoped plan preparation. This report is not an executable
implementation plan. No code/fixture/test/native/Go/SDK/Docker/DB operation
occurred;only read-only source/contract inspection and recovery documentation.

## C1: security-admin request policy is not an exact evidence signer mapping

The approved [v7 specification](../superpowers/specs/2026-08-24-nodecontrol-authority-v7-upgrade-design.md)
§4.1 registers PublishClaimV1CredentialPolicyRequestV1 as
publish-claim-v1-credential-policy-request.v1 (line133 at this baseline).
§4.3 line325 requires that request to be authenticated through the provider's
fixed security-admin request signer policy. The same paragraph separately
assigns signer_role=claim_v1_security_policy to the **response**
ClaimV1CredentialPolicyV1,not to the request.

The registry design requires each external signed object to have its exact
specified role,and non-external items to use null. It also explicitly forbids
equating authenticated requests with externally signed evidence or guessing
roles/evidence eligibility. The inspected contract does not unambiguously state
whether this request uses the general §4.1 envelope/metadata role,what that exact
request role is,or whether its authentication is a separate transport/protocol
mechanism that is not evidence-envelope eligibility. Assigning the response role,
inventing a security-admin role literal,or treating null as “authentication not
required” would each introduce policy not established by that paragraph.

Read-only exact-name/schema/policy searches across the four files in
c12-spec-set.v1.json locate the request mapping/definition in v7 §4.1/§4.3;
no additional exact request-role contract was found. Targeted source search
under internal/nodecontrol found no implementation of this request schema/type.
An unrelated Task9 node_security_admin fixture is not a role definition for it.
Search absence is not a proof about external provider deployments or other
noncanonical documents;neither was inspected or contacted.

## Reconciliation progress and limits

The147-name/object set remains the written design's verified mapping basis.
This increment inspected early mapping entries and traced the request above;
it did not complete147 role/evidence eligibility certifications. In particular,
§4.2 explicitly gives EnvironmentInventoryV1 the release_deployment_operator
role and DeploymentEnvironmentAttestationV1 the deployment_environment_attestor
role. Those examples do not authorize inferring roles for neighboring entries.
EnvironmentInventoryRecordV1 is an embedded complete body with its own domain
digest;neither its name nor digest alone establishes DB-immutable evidence status.
DatabaseAuthorityHeadAttestationV1 (§4.4) explicitly carries external signed
evidence,while its temporary embedded Head must not be disguised as
database_immutable_body. These distinctions demonstrate why a binary
“signed-or-database” default cannot safely finish the table.

Per written-design §5,one concrete unresolved role/eligibility row is sufficient
to stop producing an implementation-ready full registry. Do not present a plan
with fabricated147 rows,placeholder roles,guessed defaults or “resolve during
coding” instructions. No implementation plan is claimed complete or approved.
The old missing-digest-prefix conflict remains resolved in favor of the approved
talenro.c12. prefix;C1 is a separate issue,not a reason to change that decision.

## Recommended next decision

Prepare a narrow written clarification proposal that separates **request
authentication policy** from **bundle evidence-envelope signer role**. For this
specific request it must explicitly decide:

1. Is request authentication carried by the §4.1 signature envelope or a distinct
   authenticated-request mechanism? No assumption that an authenticated request
   is a reusable attestation.
2. If the common envelope is used,what is the exact request signer role and its
   trust-policy ownership? Do not borrow claim_v1_security_policy from the response.
3. What evidence eligibility,if any,does this request have? A null evidence signer
   field must never disable the independent mandatory request authentication.

Recommend that clarification route before continuing147-row reconciliation.
An alternative is a newly approved smaller name/object-only registry without
policy metadata;that changes the approved design's deliverable and cannot be
chosen silently. Neither recommendation authorizes protocol edits,trust-policy
changes,implementation or tests. Any clarification affecting the canonical
approved spec-set needs explicit scope/approval;this checkpoint does not alter
its files,digests or manifests.

## Recovery

Resume from latest codex/c12-b01-task9-coordinator tip at the clarification
decision,not at code execution or old Task11 fuzz commands. Preserve approved
design status and C1 in current-status/recovery. All historical validation,
physical/security/C12/main gates remain open;no new acceptance evidence.
