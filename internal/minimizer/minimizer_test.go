package minimizer

import (
	"path/filepath"
	"testing"
)

func TestFixedDenominatorHasNineScenarios(t *testing.T) {
	contract, err := LoadContract(filepath.Join("..", "..", "contracts", "denominator-v1.json"))
	if err != nil {
		t.Fatal(err)
	}
	if contract.CellCount != FixedDenominator || len(contract.Scenarios) != FixedDenominator || !contract.Fixed {
		t.Fatalf("unexpected fixed denominator: %+v", contract)
	}
}

func TestSuccessfulReductionPreservesFailureAndOrigin(t *testing.T) {
	input := Counterexample{Schema: "gooo/semantic-counterexample/v1", Version: "v1", Scenario: "remove-unreachable-declaration", FailureDigest: "sha256:failure", Origin: "origin", Provenance: "provenance", Expression: "fail", Effects: []string{"READ_INPUT"}, Nodes: []Node{{ID: 1, Kind: "declaration", Role: "unreachable", Value: "dead"}, {ID: 2, Kind: "declaration", Role: "failure-anchor", Value: "live"}}}
	final, events, metrics, state, claim, _ := executeCase(input)
	if state != "CLOSED" || claim.State != "CLOSED" || len(final.Nodes) != 1 || len(events) != 1 || metrics.AcceptedReductions != 1 || metrics.OracleCalls != 1 {
		t.Fatalf("unexpected reduction: final=%+v events=%+v metrics=%+v claim=%+v", final, events, metrics, claim)
	}
	if final.FailureDigest != input.FailureDigest || final.Origin != input.Origin || final.Provenance != input.Provenance {
		t.Fatal("reduction did not preserve failure identity and origin graph")
	}
}

func TestMissingAndAmbiguousOracleRemainUnknown(t *testing.T) {
	for _, scenario := range []string{"missing-oracle", "nondeterministic-oracle"} {
		input := Counterexample{Schema: "gooo/semantic-counterexample/v1", Version: "v1", Scenario: scenario, FailureDigest: "sha256:failure", Origin: "origin", Provenance: "provenance", Expression: "fail", Nodes: []Node{{ID: 1, Kind: "expression", Role: "failure-anchor", Value: "x"}}}
		_, _, _, state, claim, _ := executeCase(input)
		if state != "UNKNOWN" || claim.State != "UNKNOWN" || claim.Stage == "" || claim.Step == "" || claim.Reason == "" || claim.UnknownClass == "" || claim.NextOperation == "" || len(claim.BlockedBy) == 0 {
			t.Fatalf("incomplete unknown for %s: state=%s claim=%+v", scenario, state, claim)
		}
	}
}
