#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built minimizer binary is required}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
work=$(mktemp -d "${RUNNER_TEMP:-/tmp}/gooo-semantic-counterexample-minimizer.XXXXXX")
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/first" "$work/replay"

before=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')

for source in "$root"/examples/counterexamples/*.gooo; do
  scenario=$(basename "$source" .gooo)
  mkdir -p "$work/first/$scenario" "$work/replay/$scenario"
  "$bin" run --meta "$root/.gooo/counterexample-minimizer.gooo" --contract "$root/contracts/denominator-v1.json" --source "$source" --out "$work/first/$scenario" >/dev/null
  "$bin" run --meta "$root/.gooo/counterexample-minimizer.gooo" --contract "$root/contracts/denominator-v1.json" --source "$source" --out "$work/replay/$scenario" >/dev/null
  for artifact in minimization-events.ndjson minimized-counterexample.gooo preservation-receipt.json human-report.md; do
    cmp -s "$work/first/$scenario/$artifact" "$work/replay/$scenario/$artifact"
  done
  test "$(find "$work/first/$scenario" -maxdepth 1 -type f | wc -l | tr -d ' ')" = 4
  jq -e --arg scenario "$scenario" '
    .schema == "gooo/semantic-counterexample-minimizer/preservation-receipt/v1" and
    .scenario == $scenario and
    (.metrics.input_bytes | type) == "number" and
    (.metrics.output_bytes | type) == "number" and
    (.metrics.nodes | type) == "number" and
    (.metrics.attempts | type) == "number" and
    (.metrics.accepted_reductions | type) == "number" and
    (.metrics.rejected_reductions | type) == "number" and
    (.metrics.oracle_calls | type) == "number" and
    (.preserved_failure_digest | type) == "string" and
    .precedence == ["REFUTED", "UNKNOWN", "CLOSED"] and
    .authority == {repository_writes:0,commit_authority:0,merge_authority:0,release_mutation:0,local_test_executions:0} and
    .repository_boundary == "CALLER_OWNED_OUTPUT_ONLY" and
    .inventory.root_readme_excluded == true and
    .inventory.git_excluded == true and
    .inventory.caller_output_excluded == true and
    .inventory.cache_excluded == true and
    .inventory.vendor_excluded == true and
    .inventory.toolchain_excluded == true
  ' "$work/first/$scenario/preservation-receipt.json" >/dev/null
  state=$(jq -r '.state' "$work/first/$scenario/preservation-receipt.json")
  case "$scenario:$state" in
    remove-unreachable-declaration:CLOSED|shrink-expression:CLOSED|reduce-effect-sequence:CLOSED|already-minimal:CLOSED|deterministic-replay:CLOSED|missing-oracle:UNKNOWN|nondeterministic-oracle:UNKNOWN|failure-lost-during-reduction:REFUTED|origin-provenance-drop:REFUTED) ;;
    *) echo "unexpected state for $scenario: $state" >&2; exit 1 ;;
  esac
  if [ "$state" = UNKNOWN ]; then
    jq -e '.claim.stage != "" and .claim.step != "" and .claim.reason != "" and .claim.unknown_class != "" and .claim.next_operation != "" and (.claim.blocked_by | length) > 0' "$work/first/$scenario/preservation-receipt.json" >/dev/null
  fi
done

after=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
test "$before" = "$after"

closed=$(find "$work/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.state' | grep -c '^CLOSED$')
unknown=$(find "$work/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.state' | grep -c '^UNKNOWN$')
refuted=$(find "$work/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.state' | grep -c '^REFUTED$')
test "$closed" = 5
test "$unknown" = 2
test "$refuted" = 2

forbidden="$root/.gooo-semantic-counterexample-minimizer-forbidden-output"
set +e
"$bin" run --meta "$root/.gooo/counterexample-minimizer.gooo" --contract "$root/contracts/denominator-v1.json" --source "$root/examples/counterexamples/already-minimal.gooo" --out "$forbidden" >/dev/null 2>&1
status=$?
set -e
test "$status" -ne 0
test ! -e "$forbidden"

echo "semantic counterexample minimization conformance: PASS (denominator=9 closed=5 unknown=2 refuted=2)"
