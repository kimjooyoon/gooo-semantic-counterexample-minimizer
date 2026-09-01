package minimizer

const (
	SourceSchema       = "gooo/semantic-counterexample-minimizer/source/v1"
	ContractSchema     = "gooo/semantic-counterexample-minimizer/denominator/v1"
	ReceiptSchema      = "gooo/semantic-counterexample-minimizer/preservation-receipt/v1"
	ReportSchema       = "gooo/semantic-counterexample-minimizer/report/v1"
	FixedDenominator   = 9
	Toolchain          = "go1.27.0"
	Runner             = "github-actions-ubuntu-latest"
	RootReadmePolicy   = "EXCLUDED_FROM_SOURCE_INVENTORY"
	RepositoryBoundary = "CALLER_OWNED_OUTPUT_ONLY"
)

type Authority struct {
	RepositoryWrites    int `json:"repository_writes"`
	CommitAuthority     int `json:"commit_authority"`
	MergeAuthority      int `json:"merge_authority"`
	ReleaseMutation     int `json:"release_mutation"`
	LocalTestExecutions int `json:"local_test_executions"`
}

type Activity struct {
	Ordinal int    `json:"ordinal"`
	ID      string `json:"id"`
	Role    string `json:"role"`
}

type ScenarioDecl struct {
	Ordinal      int    `json:"ordinal"`
	ID           string `json:"id"`
	Expected     string `json:"expected"`
	Rule         string `json:"rule"`
	UnknownClass string `json:"unknown_class"`
}

type SourceDecl struct {
	Schema                string         `json:"schema"`
	Version               string         `json:"version"`
	Namespace             string         `json:"namespace"`
	Effects               []string       `json:"effects"`
	Activities            []Activity     `json:"activities"`
	PreservationPredicate string         `json:"preservation_predicate"`
	Precedence            []string       `json:"precedence"`
	UnknownFields         []string       `json:"unknown_fields"`
	Authority             Authority      `json:"authority"`
	DenominatorID         string         `json:"denominator_id"`
	CellCount             int            `json:"cell_count"`
	Scenarios             []ScenarioDecl `json:"scenarios"`
}

type Contract struct {
	Schema    string         `json:"schema"`
	ID        string         `json:"id"`
	Version   string         `json:"version"`
	CellCount int            `json:"cell_count"`
	Fixed     bool           `json:"fixed"`
	Scenarios []ScenarioDecl `json:"scenarios"`
}

type Node struct {
	ID    int    `json:"id"`
	Kind  string `json:"kind"`
	Role  string `json:"role"`
	Value string `json:"value"`
}

type Counterexample struct {
	Schema        string   `json:"schema"`
	Version       string   `json:"version"`
	Scenario      string   `json:"scenario"`
	FailureDigest string   `json:"failure_digest"`
	Origin        string   `json:"origin"`
	Provenance    string   `json:"provenance"`
	Expression    string   `json:"expression"`
	Effects       []string `json:"effects"`
	Nodes         []Node   `json:"nodes"`
}

type Claim struct {
	State         string   `json:"state"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func UnknownClaim(stage, step, reason, class, next string, blockedBy []string) Claim {
	return Claim{State: "UNKNOWN", Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: blockedBy}
}

type Metrics struct {
	InputBytes         int `json:"input_bytes"`
	OutputBytes        int `json:"output_bytes"`
	InputNodes         int `json:"input_nodes"`
	OutputNodes        int `json:"output_nodes"`
	Nodes              int `json:"nodes"`
	Attempts           int `json:"attempts"`
	AcceptedReductions int `json:"accepted_reductions"`
	RejectedReductions int `json:"rejected_reductions"`
	OracleCalls        int `json:"oracle_calls"`
}

type ReductionEvent struct {
	Ordinal         int    `json:"ordinal"`
	Operation       string `json:"operation"`
	Outcome         string `json:"outcome"`
	BeforeDigest    string `json:"before_digest"`
	CandidateDigest string `json:"candidate_digest"`
	AfterDigest     string `json:"after_digest"`
	OracleCalls     int    `json:"oracle_calls"`
	Reason          string `json:"reason"`
}

type Inventory struct {
	Files                int  `json:"files"`
	Directories          int  `json:"directories"`
	GoFiles              int  `json:"go_files"`
	GoooFiles            int  `json:"gooo_files"`
	PhysicalLines        int  `json:"physical_lines"`
	RootReadmeExcluded   bool `json:"root_readme_excluded"`
	GitExcluded          bool `json:"git_excluded"`
	CallerOutputExcluded bool `json:"caller_output_excluded"`
	CacheExcluded        bool `json:"cache_excluded"`
	VendorExcluded       bool `json:"vendor_excluded"`
	ToolchainExcluded    bool `json:"toolchain_excluded"`
}

type CaseResult struct {
	Scenario               string  `json:"scenario"`
	Expected               string  `json:"expected"`
	State                  string  `json:"state"`
	Rule                   string  `json:"rule"`
	Claim                  Claim   `json:"claim"`
	SourceDigest           string  `json:"source_digest"`
	CounterexampleDigest   string  `json:"counterexample_digest"`
	ContractDigest         string  `json:"contract_digest"`
	Toolchain              string  `json:"toolchain"`
	Runner                 string  `json:"runner"`
	Metrics                Metrics `json:"metrics"`
	PreservedFailureDigest string  `json:"preserved_failure_digest"`
	WitnessDigest          string  `json:"witness_digest"`
	EventsDigest           string  `json:"events_digest"`
	ReplayEqual            bool    `json:"replay_equal"`
}

type PreservationReceipt struct {
	Schema                 string    `json:"schema"`
	Scenario               string    `json:"scenario"`
	State                  string    `json:"state"`
	Expected               string    `json:"expected"`
	Precedence             []string  `json:"precedence"`
	SourceDigest           string    `json:"source_digest"`
	CounterexampleDigest   string    `json:"counterexample_digest"`
	ContractDigest         string    `json:"contract_digest"`
	Toolchain              string    `json:"toolchain"`
	Runner                 string    `json:"runner"`
	PreservationPredicate  string    `json:"preservation_predicate"`
	Claim                  Claim     `json:"claim"`
	Metrics                Metrics   `json:"metrics"`
	PreservedFailureDigest string    `json:"preserved_failure_digest"`
	WitnessDigest          string    `json:"witness_digest"`
	EventsDigest           string    `json:"events_digest"`
	Inventory              Inventory `json:"inventory"`
	Authority              Authority `json:"authority"`
	RepositoryBoundary     string    `json:"repository_boundary"`
	SemanticActivities     []string  `json:"semantic_activities"`
}

type Report struct {
	Schema               string            `json:"schema"`
	Decision             string            `json:"decision"`
	Scenario             string            `json:"scenario"`
	Expected             string            `json:"expected"`
	State                string            `json:"state"`
	Rule                 string            `json:"rule"`
	SourceDigest         string            `json:"source_digest"`
	CounterexampleDigest string            `json:"counterexample_digest"`
	ContractDigest       string            `json:"contract_digest"`
	Precedence           []string          `json:"precedence"`
	Metrics              Metrics           `json:"metrics"`
	Case                 CaseResult        `json:"case"`
	Improvement          Claim             `json:"improvement"`
	Inventory            Inventory         `json:"inventory"`
	Authority            Authority         `json:"authority"`
	SemanticActivities   []string          `json:"semantic_activities"`
	ArtifactDigests      map[string]string `json:"artifact_digests"`
}
