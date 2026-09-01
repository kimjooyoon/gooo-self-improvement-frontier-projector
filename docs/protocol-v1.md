# Frontier projection protocol v1

## Input graph

The input is a stable-ID claim/activity graph. A claim binds exactly one
activity. An activity has one of `CLOSED`, `UNKNOWN`, or `REFUTED`, an
`actionable` bit, a `historical` bit, and an explicit `blocked_by` list. The
`PRECEDES` edge is directed from predecessor to dependent. The `BLOCKED_BY`
edge is directed from blocked activity to blocker. Both are normalized into a
predecessor relation for projection.

The graph is finite only when `graph_bounded=true`. A missing, stale,
ambiguous, or unbounded graph evidence record is `UNKNOWN` and carries all six
unknown fields. Missing stable identity, contradictory binding, cycles,
non-append history, and false closure are `REFUTED`.

## Projection

The projector filters to activities that are live, actionable, non-historical,
and `UNKNOWN`. A candidate is in the canonical frontier only when every
predecessor is either complete `CLOSED` or historical. If another candidate
precedes it, it is moved to the blocked frontier. The result is sorted by
stable activity ID. No score, rank, percentage, estimated effort, textual
similarity, or LLM decision exists in the protocol.

Historical `REFUTED` items are evidence records. They are listed in
`historical_refutations_excluded`, retained in causal trace, and never become
automatic repair actions. A live refutation wins over all lower-precedence
states; `UNKNOWN` wins over `CLOSED` when the projector's graph evidence is
unresolved.

## Outputs

For one input the runtime writes only to the caller-owned output directory:

- `canonical-frontier.json`
- `blocked-frontier.json`
- `causal-trace.ndjson`
- `human-report.md`
- `semantic-ir.json`
- `provenance-graph.json`
- `receipt.json`

The conformance contract has exactly 12 cases with a `4/4/4` decision
denominator. Each case is evaluated twice and compared as a canonical output
bundle. The comparison ignores execution order because all input collections,
frontier IDs, blockers, and trace events are sorted by stable ID and sequence.

## Authority

The runtime's repository writes, source mutations, commit, merge, release,
local test executions, cross-project required gates, and acceptance-required
gate are all zero. Operator authoring is a separate boundary. A frontier is a
proposal and confers no product permission.
