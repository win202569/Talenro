# Dependency and license policy

This policy is an engineering release gate, not legal advice. Dependency versions remain pinned in the repository; changing a version, source, license selection, linking model, or distribution method requires renewed license and security review.

## Runtime infrastructure

| Dependency | Use and boundary | Release gate |
| --- | --- | --- |
| PostgreSQL 18.4 | Authoritative database, run as a separately distributed service. | Preserve its PostgreSQL License notices and review the exact image and bundled packages before distribution. |
| Redis 8.8.1 | Unmodified official server image, selected under the AGPLv3 path for local development only. Redis state is non-authoritative and rebuildable. | **Production use is blocked** until written legal approval of the planned AGPLv3 use or a recorded commercial Redis license exists. Any server modification, embedding, redistribution, hosted-service exposure, or license-path change requires a new review; no one may silently modify the server. |
| NATS Server 2.14.3 | At-least-once event transport, run as a separate service. | Preserve Apache-2.0 notices and review the exact image and bundled packages before distribution. |

Container tags must be immutable project pins rather than `latest`. A version update is a deliberate change accompanied by dependency provenance, vulnerability, and license review.

## Tunnel cores

sing-box is GPLv3+ and Xray-core is MPL-2.0. Both must remain separate operating-system processes outside the proprietary Go control-plane binary; they may communicate only across an explicit process boundary. They must not be linked into or vendored inside that proprietary binary.

Before distributing sing-box or a work based on it, release owners must record the applicable GPLv3+ obligations, including license text, notices, complete corresponding source, build material, and the treatment of modifications. iOS and other application-store distribution requires a separate written review of license and store-term compatibility.

Before distributing Xray-core or modifications to it, release owners must record MPL-2.0 notices, source availability for covered modified files, and a review confirming that proprietary control-plane files remain outside the covered-file boundary.

Changes to either tunnel core, its packaging, IPC interface, privilege model, or distribution channel reopen legal, security, and application-store review. Merely preserving a process boundary does not waive upstream license obligations.

## Go modules and build tools

### Identity and trust foundations

The following reviewed direct modules are pinned for the C1.1 identity and trust foundation. Distribution artifacts must retain the applicable BSD-3-Clause or Apache-2.0 license notices and required attribution from the upstream distribution.

| Dependency | Version | Upstream | License notice |
| --- | --- | --- | --- |
| `github.com/go-webauthn/webauthn` | `v0.17.4` | https://github.com/go-webauthn/webauthn | BSD-3-Clause |
| `github.com/gowebpki/jcs` | `v1.0.1` | https://github.com/gowebpki/jcs | Apache-2.0 |
| `github.com/google/uuid` | `v1.6.0` | https://github.com/google/uuid | BSD-3-Clause |
| `github.com/oapi-codegen/runtime` | `v1.6.0` | https://github.com/oapi-codegen/runtime | Apache-2.0 |
| `github.com/pquerna/otp` | `v1.5.0` | https://github.com/pquerna/otp | Apache-2.0 |
| `golang.org/x/crypto` | `v0.54.0` | https://go.googlesource.com/crypto | BSD-3-Clause |

The following reviewed indirect modules are the exact transitive closure used by the pinned WebAuthn implementation. They must remain indirect; adding test-only dependencies from upstream modules is not permitted.

| Dependency | Version | Upstream | License notice |
| --- | --- | --- | --- |
| `github.com/fxamacker/cbor/v2` | `v2.9.2` | https://github.com/fxamacker/cbor | MIT |
| `github.com/go-webauthn/x` | `v0.2.6` | https://github.com/go-webauthn/x | BSD-3-Clause |
| `github.com/golang-jwt/jwt/v5` | `v5.3.1` | https://github.com/golang-jwt/jwt | MIT |
| `github.com/google/go-tpm` | `v0.9.8` | https://github.com/google/go-tpm | Apache-2.0 |
| `github.com/philhofer/fwd` | `v1.2.0` | https://github.com/philhofer/fwd | MIT |
| `github.com/tinylib/msgp` | `v1.6.4` | https://github.com/tinylib/msgp | MIT |
| `github.com/x448/float16` | `v0.8.4` | https://github.com/x448/float16 | MIT |

The pinned TOTP implementation also compiles with the following reviewed indirect module. It remains an implementation detail of `github.com/pquerna/otp`; its upstream test-only dependencies are excluded.

| Dependency | Version | Upstream | License notice |
| --- | --- | --- | --- |
| `github.com/boombuler/barcode` | `v1.0.1-0.20190219062509-6c824513bacc` | https://github.com/boombuler/barcode | MIT |

`github.com/oapi-codegen/runtime` is linked into the control API because the pinned generated server uses its required header and path parameter binders, JSON merge helpers for closed `oneOf` models, and OpenAPI email/UUID types. The reviewed `v1.6.0` module carries an Apache-2.0 license; distributions must preserve its license and notice obligations. Changing the generator/runtime version pairing or replacing these helpers requires renewed compatibility, provenance, security, and license review.

Go modules and code-generation, migration, formatting, lint, and test tools are pinned by `go.mod` and `go.sum`. Generated code retains any notices required by its source or generator. Before a production release, the dependency inventory must be regenerated from the locked module graph and reviewed for:

- provenance and checksum integrity;
- vulnerability and maintenance status;
- license compatibility, notices, attribution, and source-disclosure duties;
- copyleft, network-copyleft, noncommercial, source-available, or unknown terms;
- transitive dependencies and generated or vendored material.

No dependency may be upgraded, replaced, vendored, relicensed, or fetched from an unapproved source merely to bypass a failed gate. Exceptions require a documented owner, scope, legal approval where applicable, and an expiration or re-review condition.
