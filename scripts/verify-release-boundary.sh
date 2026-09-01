#!/usr/bin/env bash
set -euo pipefail

tag=${1:?annotated release tag is required}
repo=${GITHUB_REPOSITORY:-kimjooyoon/gooo-semantic-counterexample-minimizer}
api_version='X-GitHub-Api-Version: 2026-03-10'

ref=$(gh api --header "$api_version" "repos/$repo/git/ref/tags/$tag")
test "$(jq -er '.object.type' <<<"$ref")" = tag
tag_object=$(jq -er '.object.sha' <<<"$ref")
tag_data=$(gh api --header "$api_version" "repos/$repo/git/tags/$tag_object")
test "$(jq -er '.object.type' <<<"$tag_data")" = commit
test "$(jq -er '.object.sha' <<<"$tag_data")" = "${EXPECTED_COMMIT:?EXPECTED_COMMIT is required}"
release=$(gh api --header "$api_version" "repos/$repo/releases/tags/$tag")
test "$(jq -er '.immutable' <<<"$release")" = true
test "$(jq -er '.assets | length' <<<"$release")" -gt 0
jq -e 'all(.assets[]; (.digest | type == "string" and startswith("sha256:")))' <<<"$release" >/dev/null
printf 'immutable public release boundary: closed\n'
