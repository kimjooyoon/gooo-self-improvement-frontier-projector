#!/usr/bin/env bash
set -euo pipefail

: "${GITHUB_REPOSITORY:?GITHUB_REPOSITORY is required}"
: "${RELEASE_ID:?RELEASE_ID is required}"
: "${GITHUB_TOKEN:?GITHUB_TOKEN is required}"

audit_dir="${RUNNER_TEMP}/gooo-frontier-release-audit"
audit_output="${AUDIT_OUTPUT:-${audit_dir}/release-audit.json}"
mkdir -p "$audit_dir"

release_json="$(gh api "repos/${GITHUB_REPOSITORY}/releases/${RELEASE_ID}")"
release_id="$(jq -r '.id' <<<"$release_json")"
draft="$(jq -r '.draft' <<<"$release_json")"
immutable="$(jq -r '.immutable' <<<"$release_json")"
tag_name="$(jq -r '.tag_name' <<<"$release_json")"
target="$(jq -r '.target_commitish' <<<"$release_json")"

test "$release_id" = "$RELEASE_ID"
test "$draft" = "false"
test "$immutable" = "true"
test -n "$tag_name"
test -n "$target"

assets_json='[]'
while IFS=$'\t' read -r asset_id asset_name download_url; do
  asset_path="$audit_dir/$asset_name"
  gh api -H 'Accept: application/octet-stream' "$download_url" > "$asset_path"
  asset_digest="sha256:$(sha256sum "$asset_path" | awk '{print $1}')"
  asset_size="$(wc -c < "$asset_path" | tr -d ' ')"
  assets_json="$(jq --argjson id "$asset_id" --arg name "$asset_name" --arg digest "$asset_digest" --argjson size "$asset_size" '. + [{id:$id,name:$name,size:$size,digest:$digest}]' <<<"$assets_json")"
done < <(jq -r '.assets[] | [.id,.name,.browser_download_url] | @tsv' <<<"$release_json")

jq -n \
  --argjson id "$release_id" \
  --arg tag "$tag_name" \
  --arg target "$target" \
  --argjson immutable true \
  --argjson assets "$assets_json" \
  --argjson asset_count "$(jq 'length' <<<"$assets_json")" \
  '{schema:"gooo/self-improvement-frontier/release-audit/v1",release_id:$id,tag_name:$tag,target_commitish:$target,immutable:$immutable,assets:$assets,asset_count:$asset_count}' \
  | tee "$audit_output"
