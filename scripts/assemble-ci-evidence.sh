#!/usr/bin/env bash
set -euo pipefail

stage_root=${1:?stage measurement directory is required}
test_events=${2:?go test json event file is required}
conformance_counts=${3:?conformance counts file is required}
integration_result=${4:?integration result file is required}
output=${5:?machine receipt output path is required}

for stage in compile build test conformance integration; do
  test -f "$stage_root/$stage.json"
  jq -e --arg stage "$stage" '.stage == $stage and (.wall_ms | type) == "number" and (.wall_ms | floor) == .wall_ms and (.wall_ms >= 0) and (.peak_rss_kib | type) == "number" and (.peak_rss_kib | floor) == .peak_rss_kib and (.peak_rss_kib >= 0)' "$stage_root/$stage.json" >/dev/null
done

go_test_total=$(jq -s '[.[] | select(.Action == "run" and (.Test // "") != "")] | length' "$test_events")
go_test_selected=$go_test_total
go_test_executed=$(jq -s '[.[] | select(.Action == "pass" and (.Test // "") != "")] | length' "$test_events")
go_test_reused=$(jq -s '[.[] | select((.Test // "") != "" and .Cached == true)] | length' "$test_events")
go_test_failed=$(jq -s '[.[] | select(.Action == "fail" and (.Test // "") != "")] | length' "$test_events")
go_test_unknown=0

source_digest=$(find "${CONFORMANCE_WORK_ROOT:?}/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.source_digest' | sort -u)
contract_digest=$(find "${CONFORMANCE_WORK_ROOT:?}/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.contract_digest' | sort -u)
toolchain=$(find "${CONFORMANCE_WORK_ROOT:?}/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.toolchain' | sort -u)
runner=$(find "${CONFORMANCE_WORK_ROOT:?}/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -r '.runner' | sort -u)
inventory=$(find "${CONFORMANCE_WORK_ROOT:?}/first" -mindepth 2 -name preservation-receipt.json -print0 | xargs -0 jq -c '.inventory' | sort -u)
test "$(printf '%s\n' "$source_digest" | wc -l | tr -d ' ')" = 1
test "$(printf '%s\n' "$contract_digest" | wc -l | tr -d ' ')" = 1
test "$(printf '%s\n' "$toolchain" | wc -l | tr -d ' ')" = 1
test "$(printf '%s\n' "$runner" | wc -l | tr -d ' ')" = 1
test "$(printf '%s\n' "$inventory" | wc -l | tr -d ' ')" = 1

jq -e '.artifact_count == 4 and .caller_owned_output == true and .preservation_receipt_verified == true and .replay_equal == true and .digests_verified == true and .repository_writes == 0' "$integration_result" >/dev/null
jq -e '.total == 9 and .selected == 9 and .executed == 9 and .reused == 0 and .closed == 5 and .unknown == 2 and .refuted == 2' "$conformance_counts" >/dev/null

stage_field_count=$(jq '["compile.wall_ms","compile.peak_rss_kib","build.wall_ms","build.peak_rss_kib","test.wall_ms","test.peak_rss_kib","conformance.wall_ms","conformance.peak_rss_kib","integration.wall_ms","integration.peak_rss_kib"] | length')

jq -n \
  --arg schema "gooo/semantic-counterexample-minimizer/ci-evidence/v1" \
  --arg commit "${GITHUB_SHA:-unknown}" \
  --arg source_digest "$source_digest" \
  --arg contract_digest "$contract_digest" \
  --arg toolchain "$toolchain" \
  --arg runner "$runner" \
  --argjson stage_field_count "$stage_field_count" \
  --argjson go_test_total "$go_test_total" \
  --argjson go_test_selected "$go_test_selected" \
  --argjson go_test_executed "$go_test_executed" \
  --argjson go_test_reused "$go_test_reused" \
  --argjson go_test_failed "$go_test_failed" \
  --argjson go_test_unknown "$go_test_unknown" \
  --argjson inventory "$inventory" \
  --slurpfile compile "$stage_root/compile.json" \
  --slurpfile build "$stage_root/build.json" \
  --slurpfile test "$stage_root/test.json" \
  --slurpfile conformance "$stage_root/conformance.json" \
  --slurpfile integration "$stage_root/integration.json" \
  --slurpfile conformance_counts "$conformance_counts" \
  --slurpfile integration_result "$integration_result" \
  '{schema:$schema,commit:$commit,source_digest:$source_digest,contract_digest:$contract_digest,toolchain:$toolchain,runner:$runner,stage_measurement_fields_required:10,stage_measurement_fields_recorded:$stage_field_count,stage_measurements:{compile:($compile[0]|{wall_ms,peak_rss_kib}),build:($build[0]|{wall_ms,peak_rss_kib}),test:($test[0]|{wall_ms,peak_rss_kib}),conformance:($conformance[0]|{wall_ms,peak_rss_kib}),integration:($integration[0]|{wall_ms,peak_rss_kib})},tests:{total:$go_test_total,selected:$go_test_selected,executed:$go_test_executed,reused:$go_test_reused,failed:$go_test_failed,unknown:$go_test_unknown},conformance:$conformance_counts[0],integration:$integration_result[0],inventory:$inventory,authority:{repository_writes:0,commit_authority:0,push_authority:0,merge_authority:0,release_mutation:0,local_test_executions:0},improvement:{state:"UNKNOWN",stage:"IMPROVEMENT",step:"compare_before_after",reason:"EXACT_BEFORE_AFTER_INTEGER_PAIR_NOT_PROVIDED",unknown_class:"MISSING_EXACT_PAIR",next_operation:"PROVIDE_MATCHED_BEFORE_AFTER_PAIR",blocked_by:["scenario","source","contract","toolchain","runner","before","after"]},instrumentation_coverage:{state:"CLOSED",scenario:"ci-stage-measurement",source:$source_digest,contract:$contract_digest,toolchain:$toolchain,runner:$runner,before:0,after:10,unit:"required_stage_measurement_fields"}}' > "$output"
