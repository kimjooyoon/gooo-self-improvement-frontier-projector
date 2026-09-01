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
	"immutable_ledger_version",
	"released_identity",
	"input_status",
	"operational_history",
	"adapter_failure",
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
	"ParseImmutableLedgerInput",
	"VerifyImmutableReleaseIdentity",
	"ParseLedgerProfileCells",
	"PreserveLedgerUnknownTuple",
	"PreserveLedgerRefutationHistory",
	"PreserveOperationalEventHistory",
	"ProjectImmutableLedgerInput",
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

type ReleaseAsset struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	SizeBytes   int64  `json:"size_bytes"`
	Digest      string `json:"digest"`
	DownloadURL string `json:"download_url,omitempty"`
	EntryPath   string `json:"entry_path,omitempty"`
}

type ReleaseObservation struct {
	Repository      string        `json:"repository"`
	ReleaseID       int64         `json:"release_id"`
	Tag             string        `json:"tag"`
	Immutable       bool          `json:"immutable"`
	TagObjectSHA    string        `json:"tag_object_sha"`
	TargetCommitSHA string        `json:"target_commit_sha"`
	Assets          []ReleaseAsset `json:"assets"`
}

type TagObservation struct {
	Name            string `json:"name"`
	ObjectSHA       string `json:"object_sha"`
	TargetCommitSHA string `json:"target_commit_sha"`
}

type LedgerSummary struct {
	Total   int `json:"total"`
	Closed  int `json:"closed"`
	Unknown int `json:"unknown"`
	Refuted int `json:"refuted"`
}

type LedgerProfile struct {
	Schema       string        `json:"schema"`
	ProfileID    string        `json:"profile_id"`
	AssessmentID string        `json:"assessment_id"`
	SubjectSHA   string        `json:"subject_sha"`
	Decision     string        `json:"decision"`
	Precedence   []string      `json:"precedence"`
	Summary      LedgerSummary `json:"summary"`
}

type LedgerCell struct {
	Ordinal     int      `json:"ordinal"`
	ID          string   `json:"id"`
	Axis        string   `json:"axis"`
	Proof       string   `json:"proof"`
	Indicator   string   `json:"indicator"`
	Activity    string   `json:"activity"`
	State       string   `json:"state"`
	Numerator   int      `json:"numerator"`
	Denominator int      `json:"denominator"`
	Unknown     *Unknown `json:"unknown,omitempty"`
	Refutation  *Unknown `json:"refutation,omitempty"`
}

type OperationalEvent struct {
	ID            string   `json:"id"`
	State         string   `json:"state"`
	Historical    bool     `json:"historical"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
	Source        string   `json:"source,omitempty"`
}

type AdapterFailure struct {
	Kind          string   `json:"kind"`
	Stage         string   `json:"stage"`
	Step          string   `json:"step"`
	Reason        string   `json:"reason"`
	UnknownClass  string   `json:"unknown_class"`
	NextOperation string   `json:"next_operation"`
	BlockedBy     []string `json:"blocked_by"`
}

type ImmutableLedgerInput struct {
	Schema            string             `json:"schema"`
	LedgerVersion     string             `json:"ledger_version"`
	Release           ReleaseObservation `json:"release"`
	Tag               TagObservation     `json:"tag"`
	ReleasedAsset     ReleaseAsset       `json:"released_asset"`
	Profile           LedgerProfile      `json:"profile"`
	Cells             []LedgerCell       `json:"cells"`
	OperationalEvents []OperationalEvent `json:"operational_events"`
	Failure           *AdapterFailure    `json:"failure,omitempty"`
}

type ImmutableLedgerMetadata struct {
	EnvelopeSchema          string             `json:"envelope_schema"`
	EnvelopeDigest          string             `json:"envelope_digest"`
	LedgerVersion           string             `json:"ledger_version"`
	Profile                 LedgerProfile      `json:"profile"`
	Release                 ReleaseObservation `json:"release"`
	Tag                     TagObservation     `json:"tag"`
	ReleasedAsset           ReleaseAsset       `json:"released_asset"`
	InputStatus             string             `json:"input_status"`
	Failure                 *AdapterFailure    `json:"failure,omitempty"`
	CellCount               int                `json:"cell_count"`
	UnknownCellIDs          []string           `json:"unknown_cell_ids"`
	RefutedCellIDs          []string           `json:"refuted_cell_ids"`
	OperationalRefutedCount int                `json:"operational_refuted_count"`
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
	Schema            string                  `json:"schema"`
	Source            SourceBinding           `json:"source"`
	ImmutableHistory  bool                    `json:"immutable_history"`
	GraphBounded      bool                    `json:"graph_bounded"`
	GraphEvidence     *GraphEvidence          `json:"graph_evidence,omitempty"`
	Claims            []Claim                 `json:"claims"`
	Activities        []Activity              `json:"activities"`
	Edges             []Edge                  `json:"edges"`
	History           []HistoryEntry          `json:"history"`
	ImmutableLedger   *ImmutableLedgerMetadata `json:"immutable_ledger,omitempty"`
	OperationalEvents []OperationalEvent      `json:"operational_events,omitempty"`
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
	InputStatus            string `json:"input_status"`
	ProfileID              string `json:"profile_id,omitempty"`
	AssessmentID           string `json:"assessment_id,omitempty"`
	LedgerVersion          string `json:"ledger_version,omitempty"`
	ReleaseID              int64  `json:"release_id,omitempty"`
	TagObjectSHA           string `json:"tag_object_sha,omitempty"`
	TargetCommitSHA        string `json:"target_commit_sha,omitempty"`
	ReleasedAssetID        int64  `json:"released_asset_id,omitempty"`
	ReleasedAssetDigest    string `json:"released_asset_digest,omitempty"`
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
	Class       string   `json:"class,omitempty"`
	Origin      string   `json:"origin,omitempty"`
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
	InputStatus               string            `json:"input_status"`
	OperationalRefutedCount   int               `json:"operational_refuted_count"`
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
