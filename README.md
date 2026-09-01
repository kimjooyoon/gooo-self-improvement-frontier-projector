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

The optional shared-ledger v0.48 input is immutable input evidence only. Its
acceptance-required gate is fixed at zero, and external utility remains
`UNKNOWN`; neither fact grants runtime or product authority.

## Verification boundary

All test, build, vet, format, conformance, and product-validation commands are
CI-only. The runtime counters preserve zero local test executions and zero
cross-project required gates. GitHub Actions uses only the workflow-scoped
`github.token` with read-only checkout and artifact permissions. Release
promotion is draft-first and must verify `immutable=true` through the GitHub
API before publication.
