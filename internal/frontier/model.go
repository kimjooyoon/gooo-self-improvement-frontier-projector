package frontier

const (
	DecisionRefuted = "REFUTED"
	DecisionUnknown = "UNKNOWN"
	DecisionClosed  = "CLOSED"
	Toolchain       = "go1.27.x"
	Runner          = "github-actions/ubuntu-latest"
)

var DecisionPrecedence = []string{DecisionRefuted, DecisionUnknown, DecisionClosed}

var RequiredUnknownFields = []string{
	"stage",
	"step",
	"reason",
	"unknown_class",
	"next_operation",
	"blocked_by",
}

var RequiredSourceRules = []string{
	"precedence",
	"partial_order",
	"blocked_by",
	"frontier",
	"historical_refutation",
	"unknown_fields",
	"acceptance_required_gate",
	"runtime_authority",
	"product_authority",
	"improvement",
	"missing_metric",
	"external_utility",
}

var RequiredActivityIDs = []string{
	"ParseImmutableClaimGraph",
	"PreserveDecisionPrecedence",
	"NormalizePartialOrder",
	"ProjectMinimalActionableAntichain",
	"PreserveUnknownTuple",
	"ExcludeHistoricalRefutation",
	"EmitCausalTrace",
	"VerifyReplayExactness",
	"ProjectSharedLedgerV048",
	"PreserveRuntimeAuthorityBoundary",
	"EmitHumanReport",
	"PreserveCounterexampleHistory",
}

type Unknown struct {
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

func (u *Unknown) Valid() bool {
	return u != nil && u.Stage != "" && u.Step != "" && u.Reason != "" &&
		u.UnknownClass != "" && u.NextOperation != "" && u.BlockedBy != nil
}

type Evidence struct {
	Complete bool `json:"complete"`
}

type SourceBinding struct {
	Kind                   string `json:"kind"`
	Release                string `json:"release"`
	Immutable              bool   `json:"immutable"`
	AcceptanceRequiredGate int    `json:"acceptance_required_gate"`
	ExternalUtilityState   string `json:"external_utility_state"`
}

type Claim struct {
	ID         string `json:"id"`
	ActivityID string `json:"activity_id"`
	State      string `json:"state"`
	Historical bool   `json:"historical"`
	Immutable  bool   `json:"immutable"`
}

type Activity struct {
	ID         string    `json:"id"`
	ClaimID    string    `json:"claim_id"`
	State      string    `json:"state"`
	Actionable bool      `json:"actionable"`
	Historical bool      `json:"historical"`
	BlockedBy  []string  `json:"blocked_by"`
	Unknown    *Unknown  `json:"unknown,omitempty"`
	Evidence   *Evidence `json:"evidence,omitempty"`
}

type Edge struct {
	ID         string `json:"id"`
	From       string `json:"from"`
	To         string `json:"to"`
	Relation   string `json:"relation"`
	Historical bool   `json:"historical"`
	Immutable  bool   `json:"immutable"`
}

type HistoryEntry struct {
	ID         string `json:"id"`
	ActivityID string `json:"activity_id"`
	State      string `json:"state"`
	Historical bool   `json:"historical"`
	Immutable  bool   `json:"immutable"`
	Mutation   string `json:"mutation"`
}

type GraphEvidence struct {
	State   string   `json:"state"`
	Unknown *Unknown `json:"unknown"`
}

type Input struct {
	Schema          string         `json:"schema"`
	Source          SourceBinding  `json:"source"`
	ImmutableHistory bool          `json:"immutable_history"`
	GraphBounded    bool           `json:"graph_bounded"`
	GraphEvidence   *GraphEvidence `json:"graph_evidence,omitempty"`
	Claims          []Claim        `json:"claims"`
	Activities      []Activity     `json:"activities"`
	Edges           []Edge         `json:"edges"`
	History         []HistoryEntry `json:"history"`
}

type SemanticRule struct {
	ID    string `json:"id"`
	Value string `json:"value"`
}

type SourceSpec struct {
	Language   string         `json:"language"`
	Namespace  string         `json:"namespace"`
	Activities []string       `json:"activities"`
	Rules      []SemanticRule `json:"rules"`
}

type ImprovementPolicy struct {
	Mode           string   `json:"mode"`
	IdentityFields []string `json:"identity_fields"`
	Aggregate      string   `json:"aggregate"`
	MissingMetric  string   `json:"missing_metric"`
}

type Contract struct {
	Schema             string              `json:"schema"`
	DenominatorID      string              `json:"denominator_id"`
	Total              int                 `json:"total"`
	DecisionPrecedence []string            `json:"decision_precedence"`
	ClassCounts        ClassCounts         `json:"class_counts"`
	UnknownFields      []string            `json:"unknown_fields"`
	Authority          Authority            `json:"authority"`
	Improvement        ImprovementPolicy   `json:"improvement"`
	Cases              []ContractCase      `json:"cases"`
}

type ClassCounts struct {
	Closed  int `json:"CLOSED"`
	Unknown int `json:"UNKNOWN"`
	Refuted int `json:"REFUTED"`
}

type Authority struct {
	RepositoryWrites          int `json:"repository_writes"`
	LocalTestExecutions       int `json:"local_test_executions"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
	AcceptanceRequiredGate   int `json:"acceptance_required_gate"`
	ProductCommitMergeRelease int `json:"product_commit_merge_release"`
}

type ContractCase struct {
	Ordinal          int    `json:"ordinal"`
	ID               string `json:"id"`
	Class            string `json:"class"`
	Fixture          string `json:"fixture"`
	ExpectedDecision string `json:"expected_decision"`
	Coverage         string `json:"coverage"`
}

type CaseFixture struct {
	CaseID           string `json:"case_id"`
	Class            string `json:"class"`
	ExpectedDecision string `json:"expected_decision"`
	Description      string `json:"description"`
	Input            Input  `json:"input"`
}

type FrontierItem struct {
	ActivityID string  `json:"activity_id"`
	ClaimID    string  `json:"claim_id"`
	Next       string  `json:"next_operation"`
	Unknown    Unknown `json:"unknown"`
}

type BlockedItem struct {
	ActivityID     string   `json:"activity_id"`
	ClaimID        string   `json:"claim_id"`
	Unknown        Unknown  `json:"unknown"`
	BlockingIDs    []string `json:"blocking_ids"`
	BlockingStates []string `json:"blocking_states"`
}

type Subject struct {
	InputDigest            string `json:"input_digest"`
	SourceKind             string `json:"source_kind"`
	SourceRelease          string `json:"source_release"`
	SourceImmutable        bool   `json:"source_immutable"`
	ExternalUtilityState   string `json:"external_utility_state"`
	AcceptanceRequiredGate int `json:"acceptance_required_gate"`
	Toolchain              string `json:"toolchain"`
	Runner                 string `json:"runner"`
}

type CanonicalFrontier struct {
	Schema                         string         `json:"schema"`
	Subject                        Subject        `json:"subject"`
	Decision                      string         `json:"decision"`
	Precedence                    []string       `json:"precedence"`
	Antichain                     bool           `json:"antichain"`
	Frontier                      []FrontierItem `json:"frontier"`
	HistoricalRefutationsExcluded []string       `json:"historical_refutations_excluded"`
}

type BlockedFrontier struct {
	Schema       string        `json:"schema"`
	Subject      Subject       `json:"subject"`
	Decision     string        `json:"decision"`
	Precedence   []string      `json:"precedence"`
	Blocked      []BlockedItem `json:"blocked"`
	DirectCount  int           `json:"direct_count"`
	BlockedCount int           `json:"blocked_count"`
}

type TraceEvent struct {
	Sequence    int      `json:"sequence"`
	Kind        string   `json:"kind"`
	SubjectID   string   `json:"subject_id"`
	State       string   `json:"state,omitempty"`
	Historical  bool     `json:"historical,omitempty"`
	Relation    string   `json:"relation,omitempty"`
	BlockingIDs []string `json:"blocking_ids,omitempty"`
	Unknown     *Unknown `json:"unknown,omitempty"`
	Decision    string   `json:"decision,omitempty"`
	Reason      string   `json:"reason,omitempty"`
}

type AuthorityReport struct {
	RuntimeAuthority RuntimeAuthority `json:"runtime_authority"`
	OperatorBoundary OperatorBoundary `json:"operator_boundary"`
}

type RuntimeAuthority struct {
	RepositoryWrites          int `json:"repository_writes"`
	SourceMutations           int `json:"source_mutations"`
	Commit                    int `json:"commit"`
	Merge                     int `json:"merge"`
	Release                   int `json:"release"`
	LocalTestExecutions       int `json:"local_test_executions"`
	CrossProjectRequiredGates int `json:"cross_project_required_gates"`
	AcceptanceRequiredGate   int `json:"acceptance_required_gate"`
}

type OperatorBoundary struct {
	SeparateFromRuntime bool `json:"separate_from_runtime"`
	RepositoryCreate   int  `json:"repository_create"`
	PRCreate           int  `json:"pr_create"`
	Merge              int  `json:"merge"`
	AnnotatedTag       int  `json:"annotated_tag"`
	ReleaseCreate      int  `json:"release_create"`
}

type Inventory struct {
	DescendantDirectories int  `json:"descendant_directories"`
	RegularFiles          int  `json:"regular_files"`
	GoFiles               int  `json:"go_files"`
	GoPhysicalLines       int  `json:"go_physical_lines"`
	GoooFiles             int  `json:"gooo_files"`
	GoooPhysicalLines     int  `json:"gooo_physical_lines"`
	RootREADMEExcluded    bool `json:"root_readme_excluded"`
}

type Receipt struct {
	Schema                    string            `json:"schema"`
	Subject                   Subject           `json:"subject"`
	Decision                  string            `json:"decision"`
	FrontierCount             int               `json:"frontier_count"`
	BlockedCount              int               `json:"blocked_count"`
	HistoricalRefutationCount int               `json:"historical_refutation_count"`
	UnknownFields             []string          `json:"unknown_fields"`
	Improvement               ImprovementPolicy `json:"improvement"`
	Authority                 AuthorityReport   `json:"authority"`
	LocalExecutionCount       LocalExecution    `json:"local_execution_count"`
	Artifacts                 []string          `json:"artifacts"`
	ReplayExact               bool              `json:"replay_exact"`
	SemanticIRDigest          string            `json:"semantic_ir_digest"`
	GraphDigest               string            `json:"graph_digest"`
}

type LocalExecution struct {
	GoTest      int `json:"go_test"`
	GoBuild     int `json:"go_build"`
	GoVet       int `json:"go_vet"`
	Conformance int `json:"conformance"`
	Product     int `json:"product_validation"`
}

type SemanticIR struct {
	Schema       string         `json:"schema"`
	SourceDigest string         `json:"source_digest"`
	Toolchain    string         `json:"toolchain"`
	Rules        []SemanticRule `json:"rules"`
	Activities   []string       `json:"activities"`
}

type GraphNode struct {
	ID         string `json:"id"`
	Kind       string `json:"kind"`
	State      string `json:"state"`
	Historical bool   `json:"historical"`
}

type GraphEdge struct {
	ID       string `json:"id"`
	From     string `json:"from"`
	To       string `json:"to"`
	Relation string `json:"relation"`
}

type ProvenanceGraph struct {
	Schema       string      `json:"schema"`
	SourceDigest string      `json:"source_digest"`
	Nodes        []GraphNode `json:"nodes"`
	Edges        []GraphEdge `json:"edges"`
}

type Projection struct {
	Canonical  CanonicalFrontier
	Blocked    BlockedFrontier
	Trace      []TraceEvent
	Receipt    Receipt
	Report     string
	SemanticIR SemanticIR
	Graph      ProvenanceGraph
}

func validDecision(value string) bool {
	return value == DecisionRefuted || value == DecisionUnknown || value == DecisionClosed
}
