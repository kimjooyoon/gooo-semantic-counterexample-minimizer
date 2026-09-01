package minimizer

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
)

func ParseSemanticSource(path string) (SourceDecl, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return SourceDecl{}, nil, err
	}
	var source SourceDecl
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) == 0 {
			continue
		}
		switch fields[0] {
		case "gooo":
			if len(fields) != 3 || fields[1] != "semantic_counterexample_minimizer" {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid gooo header", lineNumber)
			}
			source.Version = fields[2]
			source.Schema = SourceSchema
		case "namespace":
			if len(fields) != 2 {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid namespace", lineNumber)
			}
			source.Namespace = fields[1]
		case "effect":
			if len(fields) != 2 {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid effect", lineNumber)
			}
			source.Effects = append(source.Effects, fields[1])
		case "activity":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Activities = append(source.Activities, Activity{Ordinal: len(source.Activities) + 1, ID: values["id"], Role: values["role"]})
		case "preservation_predicate":
			if len(fields) != 2 {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid preservation predicate", lineNumber)
			}
			source.PreservationPredicate = fields[1]
		case "precedence":
			if len(fields) != 2 {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid precedence", lineNumber)
			}
			source.Precedence = strings.Split(fields[1], ">")
		case "unknown_fields":
			if len(fields) != 2 {
				return SourceDecl{}, nil, fmt.Errorf("line %d: invalid unknown fields", lineNumber)
			}
			source.UnknownFields = strings.Split(fields[1], ",")
		case "authority":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.RepositoryWrites, err = parseInt(values, "repository_writes")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.CommitAuthority, err = parseInt(values, "commit_authority")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.MergeAuthority, err = parseInt(values, "merge_authority")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.ReleaseMutation, err = parseInt(values, "release_mutation")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Authority.LocalTestExecutions, err = parseInt(values, "local_test_executions")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "denominator":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.DenominatorID = values["id"]
			source.CellCount, err = parseInt(values, "cell_count")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
		case "case":
			values, err := parseKeyValues(fields[1:])
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			ordinal, err := parseInt(values, "ordinal")
			if err != nil {
				return SourceDecl{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			source.Scenarios = append(source.Scenarios, ScenarioDecl{Ordinal: ordinal, ID: values["id"], Expected: values["expected"], Rule: values["rule"], UnknownClass: values["unknown_class"]})
		default:
			return SourceDecl{}, nil, fmt.Errorf("line %d: unknown declaration %q", lineNumber, fields[0])
		}
	}
	if err := scanner.Err(); err != nil {
		return SourceDecl{}, nil, err
	}
	return source, data, nil
}

func ParseCounterexample(path string) (Counterexample, []byte, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Counterexample{}, nil, err
	}
	var counterexample Counterexample
	counterexample.Schema = "gooo/semantic-counterexample/v1"
	scanner := bufio.NewScanner(strings.NewReader(string(data)))
	lineNumber := 0
	for scanner.Scan() {
		lineNumber++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if strings.HasPrefix(line, "counterexample ") {
			values, err := parseKeyValues(strings.Fields(strings.TrimPrefix(line, "counterexample ")))
			if err != nil {
				return Counterexample{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			counterexample.Version = values["version"]
			counterexample.Scenario = values["scenario"]
			continue
		}
		switch {
		case strings.HasPrefix(line, "failure_digest="):
			counterexample.FailureDigest = strings.TrimPrefix(line, "failure_digest=")
		case strings.HasPrefix(line, "origin="):
			counterexample.Origin = strings.TrimPrefix(line, "origin=")
		case strings.HasPrefix(line, "provenance="):
			counterexample.Provenance = strings.TrimPrefix(line, "provenance=")
		case strings.HasPrefix(line, "expression="):
			counterexample.Expression = strings.TrimPrefix(line, "expression=")
		case strings.HasPrefix(line, "effect="):
			counterexample.Effects = append(counterexample.Effects, strings.TrimPrefix(line, "effect="))
		case strings.HasPrefix(line, "node "):
			values, err := parseKeyValues(strings.Fields(strings.TrimPrefix(line, "node ")))
			if err != nil {
				return Counterexample{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			id, err := parseInt(values, "id")
			if err != nil {
				return Counterexample{}, nil, fmt.Errorf("line %d: %w", lineNumber, err)
			}
			counterexample.Nodes = append(counterexample.Nodes, Node{ID: id, Kind: values["kind"], Role: values["role"], Value: values["value"]})
		default:
			return Counterexample{}, nil, fmt.Errorf("line %d: unknown counterexample declaration", lineNumber)
		}
	}
	if err := scanner.Err(); err != nil {
		return Counterexample{}, nil, err
	}
	return counterexample, data, nil
}

func LoadContract(path string) (Contract, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Contract{}, err
	}
	var contract Contract
	if err := json.Unmarshal(data, &contract); err != nil {
		return Contract{}, fmt.Errorf("decode contract: %w", err)
	}
	return contract, nil
}

func ValidateDeclarations(source SourceDecl, contract Contract) error {
	if source.Schema != SourceSchema || source.Version != "v1" || source.Namespace != "gooo://semantic-counterexample-minimizer/v1" {
		return fmt.Errorf("invalid semantic .gooo source declaration")
	}
	if contract.Schema != ContractSchema || contract.Version != "v1" || !contract.Fixed || contract.ID != source.DenominatorID || source.CellCount != FixedDenominator || contract.CellCount != FixedDenominator {
		return fmt.Errorf("fixed denominator declaration mismatch")
	}
	if len(source.Effects) != 6 || len(source.Activities) != 6 || len(source.Scenarios) != FixedDenominator || len(contract.Scenarios) != FixedDenominator {
		return fmt.Errorf("semantic source must declare six activities and nine fixed scenarios")
	}
	wantActivities := []string{"BindFailurePredicate", "EnumerateSemanticReductions", "ExecutePreservationOracle", "SelectDeterministicReduction", "PreserveOriginGraph", "EmitMinimalWitness"}
	for i, activity := range source.Activities {
		if activity.Ordinal != i+1 || activity.ID != wantActivities[i] || activity.Role == "" {
			return fmt.Errorf("activity %d is not owned by the declared semantic vocabulary", i+1)
		}
	}
	if source.PreservationPredicate != "failure_digest_and_origin_graph" || strings.Join(source.Precedence, ">") != "REFUTED>UNKNOWN>CLOSED" || strings.Join(source.UnknownFields, ",") != "stage,step,reason,unknown_class,next_operation,blocked_by" {
		return fmt.Errorf("preservation or resolution declaration mismatch")
	}
	if source.Authority != (Authority{}) {
		return fmt.Errorf("source authority must remain zero")
	}
	for i := range source.Scenarios {
		if source.Scenarios[i] != contract.Scenarios[i] || source.Scenarios[i].Ordinal != i+1 {
			return fmt.Errorf("scenario %d does not match the fixed denominator", i+1)
		}
	}
	return nil
}

func ValidateCounterexample(counterexample Counterexample, contract Contract) error {
	if counterexample.Schema != "gooo/semantic-counterexample/v1" || counterexample.Version != "v1" || counterexample.FailureDigest == "" || counterexample.Origin == "" || counterexample.Provenance == "" || counterexample.Expression == "" || len(counterexample.Nodes) == 0 {
		return fmt.Errorf("counterexample is missing semantic identity or witness nodes")
	}
	if _, ok := scenarioFor(contract, counterexample.Scenario); !ok {
		return fmt.Errorf("scenario %q is not in the fixed denominator", counterexample.Scenario)
	}
	seen := make(map[int]bool, len(counterexample.Nodes))
	for _, node := range counterexample.Nodes {
		if node.ID <= 0 || node.Kind == "" || node.Role == "" || node.Value == "" || seen[node.ID] {
			return fmt.Errorf("counterexample node identity is invalid")
		}
		seen[node.ID] = true
	}
	return nil
}

func parseKeyValues(fields []string) (map[string]string, error) {
	values := make(map[string]string, len(fields))
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
			return nil, fmt.Errorf("invalid key/value %q", field)
		}
		values[parts[0]] = strings.Trim(parts[1], "\"")
	}
	return values, nil
}

func parseInt(values map[string]string, key string) (int, error) {
	value, ok := values[key]
	if !ok {
		return 0, fmt.Errorf("missing %s", key)
	}
	n, err := strconv.Atoi(value)
	if err != nil {
		return 0, fmt.Errorf("invalid %s %q", key, value)
	}
	return n, nil
}

func scenarioFor(contract Contract, id string) (ScenarioDecl, bool) {
	for _, scenario := range contract.Scenarios {
		if scenario.ID == id {
			return scenario, true
		}
	}
	return ScenarioDecl{}, false
}

func sameStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
