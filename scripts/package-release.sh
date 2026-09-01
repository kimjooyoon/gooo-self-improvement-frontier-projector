#!/usr/bin/env bash
set -euo pipefail

release_root="${RUNNER_TEMP}/gooo-frontier-projector"
asset_root="${RUNNER_TEMP}/gooo-frontier-release-assets"
mkdir -p "$asset_root"

tar -C "$release_root" -czf "$asset_root/frontier-projector-evidence.tar.gz" conformance project inventory.json
git archive --format=tar.gz --prefix=gooo-self-improvement-frontier-projector/ HEAD > "$asset_root/frontier-projector-source.tar.gz"

evidence_digest="sha256:$(sha256sum "$asset_root/frontier-projector-evidence.tar.gz" | awk '{print $1}')"
evidence_size="$(wc -c < "$asset_root/frontier-projector-evidence.tar.gz" | tr -d ' ')"
source_digest="sha256:$(sha256sum "$asset_root/frontier-projector-source.tar.gz" | awk '{print $1}')"
source_size="$(wc -c < "$asset_root/frontier-projector-source.tar.gz" | tr -d ' ')"

jq -n \
  --arg commit "$GITHUB_SHA" \
  --arg evidence_digest "$evidence_digest" \
  --arg source_digest "$source_digest" \
  --argjson evidence_size "$evidence_size" \
  --argjson source_size "$source_size" \
  '{schema:"gooo/self-improvement-frontier/release-manifest/v1",commit:$commit,assets:[{name:"frontier-projector-evidence.tar.gz",size:$evidence_size,sha256:$evidence_digest},{name:"frontier-projector-source.tar.gz",size:$source_size,sha256:$source_digest}]}' \
  > "$asset_root/release-manifest-v0.1.0.json"

printf '%s\n' "Release assets prepared in $asset_root"
