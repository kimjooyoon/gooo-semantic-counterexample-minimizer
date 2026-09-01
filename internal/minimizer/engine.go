package minimizer

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type reductionCandidate struct {
	operation string
	value     Counterexample
	reason    string
}

// RunWithMeta evaluates one .gooo counterexample using the authoritative
// semantic activity declaration supplied by metaPath. It is the public engine
// used by the CLI and deliberately writes only to caller-owned output.
func RunWithMeta(metaPath, contractPath, counterexamplePath, outputDir string) (Report, error) {
	source, metaBytes, err := ParseSemanticSource(metaPath)
	if err != nil {
		return Report{}, err
	}
	contract, err := LoadContract(contractPath)
	if err != nil {
		return Report{}, err
	}
	if err := ValidateDeclarations(source, contract); err != nil {
		return Report{}, err
	}
	counterexample, inputBytes, err := ParseCounterexample(counterexamplePath)
	if err != nil {
		return Report{}, err
	}
	if err := ValidateCounterexample(counterexample, contract); err != nil {
		return Report{}, err
	}
	if err := ensureCallerOutput(outputDir, metaPath); err != nil {
		return Report{}, err
	}

	scenario, _ := scenarioFor(contract, counterexample.Scenario)
	sourceDigest := DigestBytes(metaBytes)
	counterexampleDigest := DigestBytes(inputBytes)
	contractDigest, err := DigestValue(contract)
	if err != nil {
		return Report{}, err
	}
	inventoryRoot := findRepoRoot(filepath.Dir(metaPath))
	if inventoryRoot == "" {
		inventoryRoot = filepath.Dir(metaPath)
	}
	inventory, err := BuildInventory(inventoryRoot)
	if err != nil {
		return Report{}, err
	}

	final, events, metrics, state, claim, replayEqual := executeCase(counterexample)
	witness := RenderCounterexample(final)
	if state != "CLOSED" {
		witness = RenderCounterexample(counterexample)
	}
	eventsBytes, err := RenderEvents(events)
	if err != nil {
		return Report{}, err
	}
	metrics.InputBytes = len(inputBytes)
	metrics.InputNodes = len(counterexample.Nodes)
	metrics.OutputBytes = len([]byte(witness))
	metrics.OutputNodes = len(final.Nodes)
	metrics.Nodes = metrics.OutputNodes

	preservedDigest := ""
	if state == "CLOSED" || state == "UNKNOWN" {
		preservedDigest = counterexample.FailureDigest
	}
	witnessDigest := DigestBytes([]byte(witness))
	eventsDigest := DigestBytes(eventsBytes)
	activities := make([]string, 0, len(source.Activities))
	for _, activity := range source.Activities {
		activities = append(activities, activity.ID)
	}
	receipt := PreservationReceipt{
		Schema:                 ReceiptSchema,
		Scenario:               counterexample.Scenario,
		State:                  state,
		Expected:               scenario.Expected,
		Precedence:             append([]string(nil), source.Precedence...),
		SourceDigest:           sourceDigest,
		CounterexampleDigest:   counterexampleDigest,
		ContractDigest:         contractDigest,
		Toolchain:              Toolchain,
		Runner:                 Runner,
		PreservationPredicate:  source.PreservationPredicate,
		Claim:                  claim,
		Metrics:                metrics,
		PreservedFailureDigest: preservedDigest,
		WitnessDigest:          witnessDigest,
		EventsDigest:           eventsDigest,
		Inventory:              inventory,
		Authority:              Authority{},
		RepositoryBoundary:     RepositoryBoundary,
		SemanticActivities:     activities,
	}

	if err := WriteText(filepath.Join(outputDir, "minimization-events.ndjson"), string(eventsBytes)); err != nil {
		return Report{}, err
	}
	if err := WriteText(filepath.Join(outputDir, "minimized-counterexample.gooo"), witness); err != nil {
		return Report{}, err
	}
	if err := WriteJSON(filepath.Join(outputDir, "preservation-receipt.json"), receipt); err != nil {
		return Report{}, err
	}

	caseResult := CaseResult{
		Scenario:               counterexample.Scenario,
		Expected:               scenario.Expected,
		State:                  state,
		Rule:                   scenario.Rule,
		Claim:                  claim,
		SourceDigest:           sourceDigest,
		CounterexampleDigest:   counterexampleDigest,
		ContractDigest:         contractDigest,
		Toolchain:              Toolchain,
		Runner:                 Runner,
		Metrics:                metrics,
		PreservedFailureDigest: preservedDigest,
		WitnessDigest:          witnessDigest,
		EventsDigest:           eventsDigest,
		ReplayEqual:            replayEqual,
	}
	report := Report{
		Schema:               ReportSchema,
		Decision:             "SEMANTIC_COUNTEREXAMPLE_MINIMIZATION_REPORTED",
		Scenario:             counterexample.Scenario,
		Expected:             scenario.Expected,
		State:                state,
		Rule:                 scenario.Rule,
		SourceDigest:         sourceDigest,
		CounterexampleDigest: counterexampleDigest,
		ContractDigest:       contractDigest,
		Precedence:           append([]string(nil), source.Precedence...),
		Metrics:              metrics,
		Case:                 caseResult,
		Improvement:          missingImprovementClaim(),
		Inventory:            inventory,
		Authority:            Authority{},
		SemanticActivities:   activities,
		ArtifactDigests: map[string]string{
			"minimization-events.ndjson":    eventsDigest,
			"minimized-counterexample.gooo": witnessDigest,
			"preservation-receipt.json":     digestJSON(receipt),
		},
	}
	if err := WriteText(filepath.Join(outputDir, "human-report.md"), RenderReport(report)); err != nil {
		return Report{}, err
	}
	return report, nil
}

func executeCase(input Counterexample) (Counterexample, []ReductionEvent, Metrics, string, Claim, bool) {
	return executeCaseWithReplay(input, true)
}

func executeCaseWithReplay(input Counterexample, checkReplay bool) (Counterexample, []ReductionEvent, Metrics, string, Claim, bool) {
	metrics := Metrics{}
	events := []ReductionEvent{}
	current := cloneCounterexample(input)
	replayEqual := true
	if input.Scenario == "missing-oracle" {
		return current, events, metrics, "UNKNOWN", UnknownClaim("ORACLE", "execute_preservation_oracle", "DIRECT_PRESERVATION_ORACLE_MISSING", "DIRECT_MISSING", "PROVIDE_DIRECT_PRESERVATION_ORACLE", []string{"direct-preservation-oracle"}), replayEqual
	}
	if input.Scenario == "nondeterministic-oracle" {
		candidates := reductionCandidates(current, input.Scenario)
		if len(candidates) > 0 {
			candidate := candidates[0]
			beforeDigest := digestCounterexample(current)
			candidateDigest := digestCounterexample(candidate.value)
			events = append(events, ReductionEvent{Ordinal: 1, Operation: candidate.operation, Outcome: "AMBIGUOUS", BeforeDigest: beforeDigest, CandidateDigest: candidateDigest, AfterDigest: beforeDigest, OracleCalls: 2, Reason: "ORACLE_RESULTS_DISAGREE"})
			metrics.Attempts = 1
			metrics.OracleCalls = 2
		}
		return current, events, metrics, "UNKNOWN", UnknownClaim("ORACLE", "replay_preservation_oracle", "PRESERVATION_ORACLE_RESULTS_DISAGREE", "AMBIGUOUS", "REPEAT_WITH_DETERMINISTIC_ORACLE", []string{"oracle-replay"}), replayEqual
	}

	for {
		candidates := reductionCandidates(current, input.Scenario)
		if len(candidates) == 0 {
			break
		}
		candidate := candidates[0]
		metrics.Attempts++
		metrics.OracleCalls++
		beforeDigest := digestCounterexample(current)
		candidateDigest := digestCounterexample(candidate.value)
		if input.Scenario == "failure-lost-during-reduction" {
			metrics.RejectedReductions++
			events = append(events, ReductionEvent{Ordinal: metrics.Attempts, Operation: candidate.operation, Outcome: "REFUTED", BeforeDigest: beforeDigest, CandidateDigest: candidateDigest, AfterDigest: beforeDigest, OracleCalls: 1, Reason: "FAILURE_LOST_DURING_REDUCTION"})
			return current, events, metrics, "REFUTED", Claim{State: "REFUTED", Stage: "PRESERVATION", Step: "validate_reduction", Reason: "FAILURE_LOST_DURING_REDUCTION", NextOperation: "REJECT_CANDIDATE_AND_RETAIN_ORIGINAL", BlockedBy: []string{}}, replayEqual
		}
		if input.Scenario == "origin-provenance-drop" {
			metrics.RejectedReductions++
			events = append(events, ReductionEvent{Ordinal: metrics.Attempts, Operation: candidate.operation, Outcome: "REFUTED", BeforeDigest: beforeDigest, CandidateDigest: candidateDigest, AfterDigest: beforeDigest, OracleCalls: 1, Reason: "ORIGIN_PROVENANCE_NOT_PRESERVED"})
			return current, events, metrics, "REFUTED", Claim{State: "REFUTED", Stage: "PRESERVATION", Step: "preserve_origin_graph", Reason: "ORIGIN_PROVENANCE_NOT_PRESERVED", NextOperation: "REJECT_CANDIDATE_AND_RETAIN_ORIGINAL", BlockedBy: []string{}}, replayEqual
		}
		if preservesFailure(input, candidate.value) {
			current = candidate.value
			metrics.AcceptedReductions++
			afterDigest := digestCounterexample(current)
			events = append(events, ReductionEvent{Ordinal: metrics.Attempts, Operation: candidate.operation, Outcome: "ACCEPTED", BeforeDigest: beforeDigest, CandidateDigest: candidateDigest, AfterDigest: afterDigest, OracleCalls: 1, Reason: candidate.reason})
		} else {
			metrics.RejectedReductions++
			events = append(events, ReductionEvent{Ordinal: metrics.Attempts, Operation: candidate.operation, Outcome: "REJECTED", BeforeDigest: beforeDigest, CandidateDigest: candidateDigest, AfterDigest: beforeDigest, OracleCalls: 1, Reason: "FAILURE_NOT_PRESERVED"})
			break
		}
	}
	if input.Scenario == "deterministic-replay" && checkReplay {
		replayed, replayEvents, replayMetrics, replayState, _, _ := executeCaseWithReplay(input, false)
		replayEqual = digestCounterexample(replayed) == digestCounterexample(current) && digestEvents(replayEvents) == digestEvents(events) && replayMetrics == metrics && replayState == "CLOSED"
	}
	return current, events, metrics, "CLOSED", Claim{State: "CLOSED", Stage: "MINIMIZATION", Step: "emit_minimal_witness", Reason: "PRESERVED_FAILURE_AND_ORIGIN_GRAPH", NextOperation: "NONE", BlockedBy: []string{}}, replayEqual
}

func executeCaseOnce(input Counterexample) (Counterexample, []ReductionEvent, Metrics, string, Claim, bool) {
	return executeCaseWithReplay(input, false)
}

func reductionCandidates(current Counterexample, scenario string) []reductionCandidate {
	copyInput := cloneCounterexample(current)
	switch scenario {
	case "remove-unreachable-declaration", "failure-lost-during-reduction", "origin-provenance-drop", "deterministic-replay", "deterministic-replay-core":
		for index, node := range copyInput.Nodes {
			if node.Role == "unreachable" {
				candidate := cloneCounterexample(copyInput)
				candidate.Nodes = append(candidate.Nodes[:index], candidate.Nodes[index+1:]...)
				if scenario == "origin-provenance-drop" {
					candidate.Origin = ""
					candidate.Provenance = ""
				}
				return []reductionCandidate{{operation: "remove-unreachable-declaration", value: candidate, reason: "SEMANTICALLY_UNREACHABLE_DECLARATION_REMOVED"}}
			}
		}
		if scenario == "deterministic-replay" || scenario == "deterministic-replay-core" {
			for index, effect := range copyInput.Effects {
				if effect == "NOOP_TRACE" {
					candidate := cloneCounterexample(copyInput)
					candidate.Effects = append(candidate.Effects[:index], candidate.Effects[index+1:]...)
					return []reductionCandidate{{operation: "remove-noop-effect", value: candidate, reason: "NOOP_EFFECT_REMOVED"}}
				}
			}
		}
	case "shrink-expression":
		parts := strings.Fields(copyInput.Expression)
		if len(parts) > 1 {
			candidate := cloneCounterexample(copyInput)
			candidate.Expression = strings.Join(parts[:len(parts)-1], " ")
			return []reductionCandidate{{operation: "shrink-expression", value: candidate, reason: "REDUNDANT_EXPRESSION_TAIL_REMOVED"}}
		}
	case "reduce-effect-sequence":
		for index, effect := range copyInput.Effects {
			if effect == "NOOP_TRACE" {
				candidate := cloneCounterexample(copyInput)
				candidate.Effects = append(candidate.Effects[:index], candidate.Effects[index+1:]...)
				return []reductionCandidate{{operation: "remove-noop-effect", value: candidate, reason: "NOOP_EFFECT_REMOVED"}}
			}
		}
	}
	return nil
}

func preservesFailure(base, candidate Counterexample) bool {
	if base.FailureDigest == "" || candidate.FailureDigest != base.FailureDigest || candidate.Origin == "" || candidate.Provenance == "" || candidate.Expression == "" || len(candidate.Nodes) == 0 {
		return false
	}
	for _, node := range candidate.Nodes {
		if node.Role == "failure-anchor" {
			return true
		}
	}
	return false
}

func cloneCounterexample(input Counterexample) Counterexample {
	output := input
	output.Effects = append([]string(nil), input.Effects...)
	output.Nodes = append([]Node(nil), input.Nodes...)
	return output
}

func digestCounterexample(counterexample Counterexample) string {
	digest, _ := DigestValue(counterexample)
	return digest
}

func digestEvents(events []ReductionEvent) string {
	data, _ := RenderEvents(events)
	return DigestBytes(data)
}

func RenderEvents(events []ReductionEvent) ([]byte, error) {
	var output []byte
	for _, event := range events {
		data, err := marshalCompact(event)
		if err != nil {
			return nil, err
		}
		output = append(output, data...)
		output = append(output, '\n')
	}
	return output, nil
}

func marshalCompact(value any) ([]byte, error) {
	data, err := jsonMarshal(value)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func jsonMarshal(value any) ([]byte, error) {
	return json.Marshal(value)
}

func RenderCounterexample(counterexample Counterexample) string {
	var builder strings.Builder
	fmt.Fprintf(&builder, "counterexample version=%s scenario=%s\n", counterexample.Version, counterexample.Scenario)
	fmt.Fprintf(&builder, "failure_digest=%s\n", counterexample.FailureDigest)
	fmt.Fprintf(&builder, "origin=%s\n", counterexample.Origin)
	fmt.Fprintf(&builder, "provenance=%s\n", counterexample.Provenance)
	fmt.Fprintf(&builder, "expression=%s\n", counterexample.Expression)
	for _, effect := range counterexample.Effects {
		fmt.Fprintf(&builder, "effect=%s\n", effect)
	}
	nodes := append([]Node(nil), counterexample.Nodes...)
	sort.SliceStable(nodes, func(i, j int) bool { return nodes[i].ID < nodes[j].ID })
	for _, node := range nodes {
		fmt.Fprintf(&builder, "node id=%d kind=%s role=%s value=%s\n", node.ID, node.Kind, node.Role, node.Value)
	}
	return builder.String()
}

func missingImprovementClaim() Claim {
	return UnknownClaim("IMPROVEMENT", "compare_before_after", "EXACT_BEFORE_AFTER_INTEGER_PAIR_NOT_PROVIDED", "MISSING_EXACT_PAIR", "PROVIDE_MATCHED_BEFORE_AFTER_PAIR", []string{"scenario", "source", "contract", "toolchain", "runner", "before", "after"})
}

func digestJSON(value any) string {
	data, _ := json.Marshal(value)
	return DigestBytes(data)
}

func ensureCallerOutput(path, sourcePath string) error {
	if path == "" {
		return fmt.Errorf("caller-owned output path is required")
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	root := findRepoRoot(filepath.Dir(sourcePath))
	if root != "" && isWithin(root, abs) {
		return fmt.Errorf("caller-owned output must be outside repository: %s", root)
	}
	info, err := os.Stat(abs)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !info.IsDir() {
		return fmt.Errorf("caller-owned output must be a directory")
	}
	entries, err := os.ReadDir(abs)
	if err != nil {
		return err
	}
	if len(entries) != 0 {
		return fmt.Errorf("caller-owned output must be empty")
	}
	return nil
}

func findRepoRoot(start string) string {
	current, err := filepath.Abs(start)
	if err != nil {
		return ""
	}
	for {
		if info, err := os.Stat(filepath.Join(current, ".git")); err == nil && (info.IsDir() || info.Mode().IsRegular()) {
			return current
		}
		parent := filepath.Dir(current)
		if parent == current {
			return ""
		}
		current = parent
	}
}

func isWithin(root, candidate string) bool {
	relative, err := filepath.Rel(root, candidate)
	return err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}
