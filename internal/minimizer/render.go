package minimizer

import (
	"fmt"
	"strings"
)

func RenderReport(report Report) string {
	var builder strings.Builder
	builder.WriteString("# Semantic counterexample minimization\n\n")
	fmt.Fprintf(&builder, "- scenario: `%s`\n- expected: `%s`\n- state: `%s`\n- rule: `%s`\n", report.Scenario, report.Expected, report.State, report.Rule)
	builder.WriteString("- precedence: `REFUTED > UNKNOWN > CLOSED`\n")
	builder.WriteString("- improvement: `UNKNOWN` (exact matched before/after integer pair not provided)\n")
	builder.WriteString("- repository_writes: 0\n- commit_authority: 0\n- merge_authority: 0\n- release_mutation: 0\n- local_test_executions: 0\n\n")
	builder.WriteString("## Exact metrics\n\n")
	builder.WriteString("| metric | value |\n|---|---:|\n")
	fmt.Fprintf(&builder, "| input bytes | %d |\n| output bytes | %d |\n| input nodes | %d |\n| output nodes | %d |\n| nodes | %d |\n| attempts | %d |\n| accepted reductions | %d |\n| rejected reductions | %d |\n| oracle calls | %d |\n", report.Metrics.InputBytes, report.Metrics.OutputBytes, report.Metrics.InputNodes, report.Metrics.OutputNodes, report.Metrics.Nodes, report.Metrics.Attempts, report.Metrics.AcceptedReductions, report.Metrics.RejectedReductions, report.Metrics.OracleCalls)
	builder.WriteString("\n## Preservation evidence\n\n")
	fmt.Fprintf(&builder, "- preserved failure digest: `%s`\n- semantic source digest: `%s`\n- counterexample source digest: `%s`\n- contract digest: `%s`\n- toolchain: `%s`\n- runner: `%s`\n- witness digest: `%s`\n- events digest: `%s`\n\n", report.Case.PreservedFailureDigest, report.SourceDigest, report.Case.CounterexampleDigest, report.ContractDigest, report.Case.Toolchain, report.Case.Runner, report.Case.WitnessDigest, report.Case.EventsDigest)
	builder.WriteString("## Resolution claim\n\n")
	fmt.Fprintf(&builder, "- state: `%s`\n- stage: `%s`\n- step: `%s`\n- reason: `%s`\n- unknown_class: `%s`\n- next_operation: `%s`\n- blocked_by: `%s`\n\n", report.Case.Claim.State, report.Case.Claim.Stage, report.Case.Claim.Step, report.Case.Claim.Reason, report.Case.Claim.UnknownClass, report.Case.Claim.NextOperation, strings.Join(report.Case.Claim.BlockedBy, ","))
	builder.WriteString("## Source inventory\n\n")
	fmt.Fprintf(&builder, "- files: %d\n- directories: %d\n- Go files: %d\n- Gooo files: %d\n- physical lines: %d\n- root README excluded: %t\n- `.git`, caller-owned output, cache, vendor, and toolchain internals excluded: %t\n", report.Inventory.Files, report.Inventory.Directories, report.Inventory.GoFiles, report.Inventory.GoooFiles, report.Inventory.PhysicalLines, report.Inventory.RootReadmeExcluded, report.Inventory.GitExcluded && report.Inventory.CallerOutputExcluded && report.Inventory.CacheExcluded && report.Inventory.VendorExcluded && report.Inventory.ToolchainExcluded)
	return builder.String()
}
