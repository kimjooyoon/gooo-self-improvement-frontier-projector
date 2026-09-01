#!/usr/bin/env bash
set -euo pipefail

out_root="${RUNNER_TEMP}/gooo-frontier-projector"
mkdir -p "$out_root"

go test ./...
go build ./...
go vet ./...

go run ./cmd/gooo-frontier conformance \
  --source .gooo \
  --contract contracts/frontier-denominator-v1.json \
  --fixtures fixtures/cases \
  --output "$out_root/conformance"

go run ./cmd/gooo-frontier project \
	--source .gooo \
	--contract contracts/frontier-denominator-v1.json \
	--input fixtures/inputs/shared-ledger-v0480.json \
	--output "$out_root/project"

go run ./cmd/gooo-frontier project \
	--source .gooo \
	--contract contracts/frontier-denominator-v1.json \
	--input fixtures/inputs/immutable-ledger-v0490.json \
	--output "$out_root/immutable-ledger"

go run ./cmd/gooo-frontier project \
	--source .gooo \
	--contract contracts/frontier-denominator-v1.json \
	--input fixtures/inputs/immutable-ledger-v0490.json \
	--output "$out_root/immutable-ledger-replay"

go run ./cmd/gooo-frontier project \
  --source .gooo \
  --contract contracts/frontier-denominator-v1.json \
  --input fixtures/inputs/shared-ledger-v0480.json \
  --output "$out_root/project-replay"

cmp "$out_root/project/canonical-frontier.json" "$out_root/project-replay/canonical-frontier.json"
cmp "$out_root/project/blocked-frontier.json" "$out_root/project-replay/blocked-frontier.json"
cmp "$out_root/project/causal-trace.ndjson" "$out_root/project-replay/causal-trace.ndjson"
cmp "$out_root/project/human-report.md" "$out_root/project-replay/human-report.md"
cmp "$out_root/project/semantic-ir.json" "$out_root/project-replay/semantic-ir.json"
cmp "$out_root/project/provenance-graph.json" "$out_root/project-replay/provenance-graph.json"
cmp "$out_root/project/receipt.json" "$out_root/project-replay/receipt.json"
cmp "$out_root/immutable-ledger/canonical-frontier.json" "$out_root/immutable-ledger-replay/canonical-frontier.json"
cmp "$out_root/immutable-ledger/blocked-frontier.json" "$out_root/immutable-ledger-replay/blocked-frontier.json"
cmp "$out_root/immutable-ledger/causal-trace.ndjson" "$out_root/immutable-ledger-replay/causal-trace.ndjson"
cmp "$out_root/immutable-ledger/human-report.md" "$out_root/immutable-ledger-replay/human-report.md"
cmp "$out_root/immutable-ledger/semantic-ir.json" "$out_root/immutable-ledger-replay/semantic-ir.json"
cmp "$out_root/immutable-ledger/provenance-graph.json" "$out_root/immutable-ledger-replay/provenance-graph.json"
cmp "$out_root/immutable-ledger/receipt.json" "$out_root/immutable-ledger-replay/receipt.json"

go run ./cmd/gooo-frontier inventory \
  --root . \
  --output "$out_root/inventory.json"

printf '%s\n' "CI verification complete: output=$out_root"
