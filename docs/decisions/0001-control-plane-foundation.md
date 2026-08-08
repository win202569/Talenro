# ADR 0001: Control-plane foundation

- Status: Accepted
- Date: 2026-08-04

## Context

Talenro needs a reproducible control-plane foundation that can evolve without coupling proprietary orchestration code to tunnel implementations or treating transient infrastructure as authoritative state.

## Decision

Go is the implementation language for control-plane services and node agents. The first control-plane delivery is intentionally a modular monolith process. A service may be extracted only when measured scaling pressure or demonstrated failure-isolation needs justify the added operational boundary.

PostgreSQL is the authoritative system of record. Redis holds only short-lived, rebuildable state and may be cleared and reconstructed from authoritative data. NATS JetStream events use at-least-once delivery; consumers must deduplicate by `event_id` or an explicit business idempotency key and must not assume exactly-once delivery.

OpenAPI is the governing contract for client-facing HTTP APIs. Protobuf is the governing contract for service events. Their generated artifacts are committed and must reproduce without drift using the repository-pinned tools.

Redis 8.8.1 is used unmodified, through the selected AGPLv3 licensing path, only for local development. Production use is blocked until a written legal review approving that use or a recorded commercial Redis license is in place. The Redis server must not be silently modified or embedded into a proprietary binary.

sing-box and Xray remain separate operating-system process boundaries outside the proprietary control-plane binary. Neither tunnel core is linked into or vendored as part of that binary. Their distribution, source-offer, license-notice, modification, and application-store obligations remain independent release gates described in the dependency policy.

## Consequences

- Durable decisions survive loss of Redis, while cache and coordination state are reconstructible.
- Event handlers must tolerate redelivery and expose explicit idempotency behavior.
- HTTP and event schema changes begin in OpenAPI and Protobuf respectively, followed by deterministic regeneration.
- The initial deployment avoids premature distributed-service complexity while preserving evidence-based extraction criteria.
- Redis and tunnel-core production or distribution work cannot pass release review until the applicable license gates are recorded as satisfied.
