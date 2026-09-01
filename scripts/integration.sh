#!/usr/bin/env bash
set -euo pipefail

bin=${1:?path to the built minimizer binary is required}
root=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
result=${INTEGRATION_RESULT_OUT:?INTEGRATION_RESULT_OUT is required}
work=$(mktemp -d "${RUNNER_TEMP:-/tmp}/gooo-semantic-counterexample-minimizer-integration.XXXXXX")
trap 'rm -rf "$work"' EXIT
mkdir -p "$work/first" "$work/replay"

source="$root/examples/counterexamples/shrink-expression.gooo"
meta="$root/.gooo/counterexample-minimizer.gooo"
contract="$root/contracts/denominator-v1.json"
before=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
"$bin" run --meta "$meta" --contract "$contract" --source "$source" --out "$work/first" >/dev/null
"$bin" run --meta "$meta" --contract "$contract" --source "$source" --out "$work/replay" >/dev/null

for artifact in minimization-events.ndjson minimized-counterexample.gooo preservation-receipt.json human-report.md; do
  test -f "$work/first/$artifact"
  test -f "$work/replay/$artifact"
  cmp -s "$work/first/$artifact" "$work/replay/$artifact"
done
test "$(find "$work/first" -maxdepth 1 -type f | wc -l | tr -d ' ')" = 4
jq -e '
  .schema == "gooo/semantic-counterexample-minimizer/preservation-receipt/v1" and
  .scenario == "shrink-expression" and .state == "CLOSED" and
  .preserved_failure_digest == "sha256:2222222222222222222222222222222222222222222222222222222222222222" and
  .metrics.input_bytes == 408 and .metrics.output_bytes == 370 and
  .metrics.attempts == 7 and .metrics.accepted_reductions == 7 and
  .metrics.rejected_reductions == 0 and .metrics.oracle_calls == 7
' "$work/first/preservation-receipt.json" >/dev/null
test "sha256:$(sha256sum "$work/first/minimized-counterexample.gooo" | awk '{print $1}')" = "$(jq -r '.witness_digest' "$work/first/preservation-receipt.json")"
test "sha256:$(sha256sum "$work/first/minimization-events.ndjson" | awk '{print $1}')" = "$(jq -r '.events_digest' "$work/first/preservation-receipt.json")"
after=$(git -C "$root" status --porcelain=v1 -z --untracked-files=all | sha256sum | awk '{print $1}')
test "$before" = "$after"

jq -n \
  --arg scenario shrink-expression \
  --arg state CLOSED \
  '{schema:"gooo/semantic-counterexample-minimizer/integration/v1",scenario:$scenario,state:$state,caller_owned_output:true,artifact_count:4,preservation_receipt_verified:true,replay_equal:true,digests_verified:true,repository_writes:0}' > "$result"
