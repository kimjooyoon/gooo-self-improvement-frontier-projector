# Gooo self-improvement frontier projector

This repository projects a deterministic, minimal actionable frontier from a
claim/activity graph. It is a proposal surface for the next operation; it has
zero authority to commit, merge, release, or mutate the product repository.

The semantic source is [`.gooo`](.gooo). Go is only the parser, generated
semantic-IR emitter, evaluator, and runtime for caller-owned output. The
projector uses stable IDs and graph edges. It does not rank work, emit scores
or percentages, infer natural-language intent, or make an LLM decision.

## Decision and frontier semantics

The fixed decision precedence is `REFUTED > UNKNOWN > CLOSED`.

- `CLOSED` means the projection completed with a structurally valid graph.
- `UNKNOWN` means graph evidence is missing, stale, ambiguous, or unbounded.
- `REFUTED` means the input contradicts its own immutable history, edges,
  closure evidence, or acyclicity requirements.

An actionable frontier item is live, non-historical, and `UNKNOWN`. A node is
minimal when no unresolved predecessor reaches it through `PRECEDES` or
`BLOCKED_BY`. The canonical frontier is sorted by stable activity ID and is an
antichain. Dependency-blocked items are kept in the blocked frontier. A
historical `REFUTED` record is preserved in the trace and explicitly excluded
from automatic action.

Every `UNKNOWN` record preserves exactly these six semantic fields:
`stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.

## Usage

The CI workflow runs the following operations with Go 1.27.x:

```text
go run ./cmd/gooo-frontier project \
  --source .gooo \
  --contract contracts/frontier-denominator-v1.json \
  --input fixtures/inputs/shared-ledger-v0480.json \
  --output output/project

go run ./cmd/gooo-frontier project \
  --source .gooo \
  --contract contracts/frontier-denominator-v1.json \
  --input fixtures/inputs/immutable-ledger-v0490.json \
  --output output/immutable-ledger

go run ./cmd/gooo-frontier conformance \
  --source .gooo \
  --contract contracts/frontier-denominator-v1.json \
  --fixtures fixtures/cases \
  --output output/conformance
```

The projector emits `canonical-frontier.json`, `blocked-frontier.json`,
`causal-trace.ndjson`, `human-report.md`, a generated `semantic-ir.json`, a
provenance graph, and a machine receipt. Conformance has exactly 12 cases:
four `CLOSED`, four `UNKNOWN`, and four `REFUTED`.

The v0.49 causal trace uses stable IDs and explicit classes: live unresolved
cells are `ACTIONABLE_FRONTIER`, dependency-held cells are
`BLOCKED_FRONTIER`, and immutable operational refutations are
`HISTORICAL_REFUTATION`. The five operational refutations do not become
repair actions. Missing, stale, and ambiguous adapter observations retain all
six UNKNOWN fields; schema or identity contradictions are REFUTED.

The legacy shared-ledger v0.48 graph remains accepted as a compatibility
fixture. The current input path is a generic immutable-ledger envelope: it
requires an explicit `ledger_version`, released asset metadata, release ID,
annotated tag object, peeled target commit, and SHA-256 digest. The adapter
parses the profile and stable-ID cells from that envelope; it never derives a
ledger version from a release title, tag wording, or natural-language text.

The live v0.49.0 fixture observes release `380810861`, tag object
`36f4fa271a72616a39a703c9658e905b670f5f64`, target
`036d2d1e25df72a5568aeb16f6ac0a077ce4471f`, asset `540115901`, and digest
`sha256:e680c234fee34e36bae27685a29c716208cf83bb67e9375a31a9ee5194ca5208`.
Its projector validity is `CLOSED` while its input status is `UNKNOWN`: the
external utility cell remains unresolved, the v0.49 parent-cache observation
is dependency-blocked, and five `OPERATIONAL_REFUTED` events are preserved as
historical evidence. These statuses are separate and are visible in the
causal trace.

## Verification boundary

All test, build, vet, format, conformance, and product-validation commands are
CI-only. The runtime counters preserve zero local test executions and zero
cross-project required gates. GitHub Actions uses only the workflow-scoped
`github.token` with read-only checkout and artifact permissions. Release
promotion is draft-first and must verify `immutable=true` through the GitHub
API before publication.
