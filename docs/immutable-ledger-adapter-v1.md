# Generic immutable-ledger adapter v1

The adapter accepts a caller-owned JSON envelope with schema
`gooo/self-improvement-frontier/immutable-ledger-input/v1`. The envelope has
five independent identity inputs:

- `ledger_version`: an explicit profile schema. It is required and is never
  inferred from a release tag, release title, asset name, or prose.
- `release`: repository, release ID, immutable flag, tag, annotated tag object,
  peeled target commit, and the release asset list.
- `tag`: the annotated tag object and peeled target commit observed separately.
- `released_asset`: the selected asset ID, name, size, digest, and payload
  entry path.
- `profile` and `cells`: the profile identity, precedence, exact summary, and
  stable-ID cell records.

The v0.49 fixture adds `operational_events` as append-only evidence. A live
unknown event is adapted into the graph. An `OPERATIONAL_REFUTED` event is
adapted into immutable history and is emitted with the causal class
`HISTORICAL_REFUTATION`; it cannot become an automatic action.

The adapter performs no network fetch and grants no authority. Release
coordinates and asset digests are observations supplied by the caller. The
projector decision describes projector validity; `subject.input_status`
describes unresolved live input evidence. An input can therefore have
projector decision `CLOSED` and input status `UNKNOWN`.

## Failure classification

An explicit `failure.kind` of `MISSING`, `STALE`, or `AMBIGUOUS` produces a
six-field UNKNOWN tuple. Missing release identity or profile data follows the
same rule. Present but contradictory schema, stable identity, release, tag,
target, asset, digest, cell, or event data produces REFUTED. This preserves
`REFUTED > UNKNOWN > CLOSED` without coercing missing values to zero.

The projector does not emit scores, percentages, averages, or inferred
priority. The improvement contract remains exact-pair and per-indicator only;
external utility remains UNKNOWN without independent user evidence.
