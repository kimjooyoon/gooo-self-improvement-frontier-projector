package frontier

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
)

const (
	inputSchema            = "gooo/self-improvement-frontier/input/v1"
	frontierSchema         = "gooo/self-improvement-frontier/canonical-frontier/v1"
	blockedSchema          = "gooo/self-improvement-frontier/blocked-frontier/v1"
	traceSchema            = "gooo/self-improvement-frontier/causal-trace/v1"
	receiptSchema          = "gooo/self-improvement-frontier/receipt/v1"
	semanticIRSchema       = "gooo/self-improvement-frontier/semantic-ir/v1"
	provenanceGraphSchema  = "gooo/self-improvement-frontier/provenance-graph/v1"
)

type graphIssue struct {
	Reason     string
	SubjectID  string
	BlockingID string
}

func DigestBytes(raw []byte) string {
	digest := sha256.Sum256(raw)
	return "sha256:" + hex.EncodeToString(digest[:])
}

func CanonicalJSON(value any) ([]byte, error) {
	return json.Marshal(value)
}

func CanonicalDigest(value any) (string, error) {
	raw, err := CanonicalJSON(value)
	if err != nil {
		return "", err
	}
	return DigestBytes(raw), nil
}

func BuildSemanticIR(spec SourceSpec, sourceDigest string) SemanticIR {
	activities := append([]string(nil), spec.Activities...)
	rules := append([]SemanticRule(nil), spec.Rules...)
	return SemanticIR{
		Schema:       semanticIRSchema,
		SourceDigest: sourceDigest,
		Toolchain:    Toolchain,
		Rules:        rules,
		Activities:   activities,
	}
}

func BuildProvenanceGraph(in Input, sourceDigest string) ProvenanceGraph {
	activities := append([]Activity(nil), in.Activities...)
	claims := append([]Claim(nil), in.Claims...)
	edges := append([]Edge(nil), in.Edges...)
	sort.Slice(activities, func(i, j int) bool { return activities[i].ID < activities[j].ID })
	sort.Slice(claims, func(i, j int) bool { return claims[i].ID < claims[j].ID })
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].ID != edges[j].ID {
			return edges[i].ID < edges[j].ID
		}
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	nodes := make([]GraphNode, 0, len(claims)+len(activities))
	for _, claim := range claims {
		nodes = append(nodes, GraphNode{ID: claim.ID, Kind: "claim", State: claim.State, Historical: claim.Historical})
	}
	for _, activity := range activities {
		nodes = append(nodes, GraphNode{ID: activity.ID, Kind: "activity", State: activity.State, Historical: activity.Historical})
	}
	graphEdges := make([]GraphEdge, 0, len(edges)+len(claims))
	for _, claim := range claims {
		graphEdges = append(graphEdges, GraphEdge{ID: "claim-binds-" + claim.ID, From: claim.ID, To: claim.ActivityID, Relation: "CLAIM_BINDS_ACTIVITY"})
	}
	for _, edge := range edges {
		graphEdges = append(graphEdges, GraphEdge{ID: edge.ID, From: edge.From, To: edge.To, Relation: edge.Relation})
	}
	sort.Slice(nodes, func(i, j int) bool {
		if nodes[i].ID != nodes[j].ID {
			return nodes[i].ID < nodes[j].ID
		}
		return nodes[i].Kind < nodes[j].Kind
	})
	sort.Slice(graphEdges, func(i, j int) bool {
		if graphEdges[i].ID != graphEdges[j].ID {
			return graphEdges[i].ID < graphEdges[j].ID
		}
		if graphEdges[i].From != graphEdges[j].From {
			return graphEdges[i].From < graphEdges[j].From
		}
		return graphEdges[i].To < graphEdges[j].To
	})
	return ProvenanceGraph{Schema: provenanceGraphSchema, SourceDigest: sourceDigest, Nodes: nodes, Edges: graphEdges}
}

func EvaluateInput(spec SourceSpec, input Input, sourceDigest string) (Projection, error) {
	if input.Schema != inputSchema {
		return Projection{}, fmt.Errorf("unexpected input schema %q", input.Schema)
	}
	if sourceDigest == "" {
		digest, err := CanonicalDigest(input)
		if err != nil {
			return Projection{}, err
		}
		sourceDigest = digest
	}
	inputDigest, err := CanonicalDigest(input)
	if err != nil {
		return Projection{}, err
	}
	semanticIR := BuildSemanticIR(spec, sourceDigest)
	semanticIRDigest, err := CanonicalDigest(semanticIR)
	if err != nil {
		return Projection{}, err
	}
	graph := BuildProvenanceGraph(input, sourceDigest)
	graphDigest, err := CanonicalDigest(graph)
	if err != nil {
		return Projection{}, err
	}

	activities, claims, issues := indexInput(input)
	historicalRefutations := historicalRefutationIDs(input, activities)
	trace := traceInput(input, activities)
	for _, issue := range issues {
		trace = append(trace, TraceEvent{Kind: "REFUTATION_OBSERVED", SubjectID: issue.SubjectID, Reason: issue.Reason, BlockingIDs: nonEmpty(issue.BlockingID)})
	}

	decision := DecisionClosed
	var graphUnknown *Unknown
	if len(issues) > 0 {
		decision = DecisionRefuted
	} else {
		graphUnknown = graphUnknownTuple(input)
		if graphUnknown != nil {
			decision = DecisionUnknown
		}
	}

	frontier, blocked := projectFrontier(input, activities, claims, decision)
	for _, item := range frontier {
		unknown := item.Unknown
		trace = append(trace, TraceEvent{Kind: "FRONTIER_SELECTED", SubjectID: item.ActivityID, State: DecisionUnknown, Unknown: &unknown})
	}
	for _, item := range blocked {
		unknown := item.Unknown
		trace = append(trace, TraceEvent{Kind: "DEPENDENCY_BLOCKED", SubjectID: item.ActivityID, State: DecisionUnknown, BlockingIDs: item.BlockingIDs, Unknown: &unknown})
	}
	if graphUnknown != nil {
		trace = append(trace, TraceEvent{Kind: "GRAPH_EVIDENCE_UNKNOWN", SubjectID: "graph-evidence", State: DecisionUnknown, Unknown: graphUnknown, Reason: graphUnknown.Reason})
	}
	trace = append(trace, TraceEvent{Kind: "PROJECTION_DECISION", SubjectID: "frontier-projection", Decision: decision})
	for index := range trace {
		trace[index].Sequence = index + 1
	}

	subject := Subject{
		InputDigest: inputDigest,
		SourceKind: input.Source.Kind,
		SourceRelease: input.Source.Release,
		SourceImmutable: input.Source.Immutable,
		ExternalUtilityState: input.Source.ExternalUtilityState,
		AcceptanceRequiredGate: input.Source.AcceptanceRequiredGate,
		Toolchain: Toolchain,
		Runner: Runner,
	}
	canonical := CanonicalFrontier{
		Schema: frontierSchema,
		Subject: subject,
		Decision: decision,
		Precedence: append([]string(nil), DecisionPrecedence...),
		Antichain: decision != DecisionRefuted,
		Frontier: frontier,
		HistoricalRefutationsExcluded: historicalRefutations,
	}
	blockedDocument := BlockedFrontier{
		Schema: blockedSchema,
		Subject: subject,
		Decision: decision,
		Precedence: append([]string(nil), DecisionPrecedence...),
		Blocked: blocked,
		DirectCount: len(frontier),
		BlockedCount: len(blocked),
	}
	receipt := Receipt{
		Schema: receiptSchema,
		Subject: subject,
		Decision: decision,
		FrontierCount: len(frontier),
		BlockedCount: len(blocked),
		HistoricalRefutationCount: len(historicalRefutations),
		UnknownFields: append([]string(nil), RequiredUnknownFields...),
		Improvement: ImprovementPolicy{Mode: "PER_INDICATOR_ONLY", IdentityFields: []string{"scenario", "fixture", "contract", "toolchain", "runner"}, Aggregate: "NOT_COMBINED", MissingMetric: "null+UNKNOWN"},
		Authority: AuthorityReport{
			RuntimeAuthority: RuntimeAuthority{},
			OperatorBoundary: OperatorBoundary{SeparateFromRuntime: true},
		},
		LocalExecutionCount: LocalExecution{},
		Artifacts: []string{"canonical-frontier.json", "blocked-frontier.json", "causal-trace.ndjson", "human-report.md", "semantic-ir.json", "provenance-graph.json", "receipt.json"},
		ReplayExact: true,
		SemanticIRDigest: semanticIRDigest,
		GraphDigest: graphDigest,
	}
	report := BuildHumanReport(canonical, blockedDocument, receipt, trace)
	return Projection{Canonical: canonical, Blocked: blockedDocument, Trace: trace, Receipt: receipt, Report: report, SemanticIR: semanticIR, Graph: graph}, nil
}

func indexInput(input Input) (map[string]Activity, map[string]Claim, []graphIssue) {
	activities := make(map[string]Activity, len(input.Activities))
	claims := make(map[string]Claim, len(input.Claims))
	issues := make([]graphIssue, 0)
	for _, activity := range input.Activities {
		if activity.ID == "" {
			issues = append(issues, graphIssue{Reason: "MISSING_ACTIVITY_STABLE_ID", SubjectID: "activity"})
			continue
		}
		if _, exists := activities[activity.ID]; exists {
			issues = append(issues, graphIssue{Reason: "DUPLICATE_ACTIVITY_STABLE_ID", SubjectID: activity.ID})
			continue
		}
		activities[activity.ID] = activity
	}
	for _, claim := range input.Claims {
		if claim.ID == "" {
			issues = append(issues, graphIssue{Reason: "MISSING_CLAIM_STABLE_ID", SubjectID: "claim"})
			continue
		}
		if _, exists := claims[claim.ID]; exists {
			issues = append(issues, graphIssue{Reason: "DUPLICATE_CLAIM_STABLE_ID", SubjectID: claim.ID})
			continue
		}
		claims[claim.ID] = claim
	}

	for _, activity := range input.Activities {
		if !validDecision(activity.State) {
			issues = append(issues, graphIssue{Reason: "UNKNOWN_ACTIVITY_DECISION", SubjectID: activity.ID})
		}
		if activity.State == DecisionUnknown && !activity.Unknown.Valid() {
			issues = append(issues, graphIssue{Reason: "MALFORMED_UNKNOWN_TUPLE", SubjectID: activity.ID})
		}
		if activity.State == DecisionClosed && (activity.Evidence == nil || !activity.Evidence.Complete) {
			issues = append(issues, graphIssue{Reason: "FALSE_CLOSURE", SubjectID: activity.ID})
		}
		if activity.State == DecisionRefuted && !activity.Historical {
			issues = append(issues, graphIssue{Reason: "LIVE_REFUTATION_NOT_ACTIONABLE", SubjectID: activity.ID})
		}
		claim, exists := claims[activity.ClaimID]
		if !exists {
			issues = append(issues, graphIssue{Reason: "CLAIM_BINDING_MISSING", SubjectID: activity.ID})
		} else if claim.ActivityID != activity.ID || claim.State != activity.State || claim.Historical != activity.Historical {
			issues = append(issues, graphIssue{Reason: "CLAIM_ACTIVITY_STATE_CONTRADICTION", SubjectID: activity.ID, BlockingID: claim.ID})
		}
	}
	for _, claim := range input.Claims {
		if !validDecision(claim.State) {
			issues = append(issues, graphIssue{Reason: "UNKNOWN_CLAIM_DECISION", SubjectID: claim.ID})
		}
		if _, exists := activities[claim.ActivityID]; !exists {
			issues = append(issues, graphIssue{Reason: "CLAIM_ACTIVITY_TARGET_MISSING", SubjectID: claim.ID, BlockingID: claim.ActivityID})
		}
	}

	if input.Source.AcceptanceRequiredGate != 0 || input.Source.ExternalUtilityState != "UNKNOWN" || (input.Source.Kind == "shared-ledger" && !input.Source.Immutable) {
		issues = append(issues, graphIssue{Reason: "SOURCE_BOUNDARY_CONTRADICTION", SubjectID: "source"})
	}
	if input.GraphEvidence != nil && (input.GraphEvidence.State != DecisionUnknown || !input.GraphEvidence.Unknown.Valid()) {
		issues = append(issues, graphIssue{Reason: "MALFORMED_GRAPH_UNKNOWN", SubjectID: "graph-evidence"})
	}
	if input.ImmutableHistory {
		for _, entry := range input.History {
			if entry.ID == "" || !entry.Immutable || entry.Mutation != "APPEND" || !entry.Historical {
				issues = append(issues, graphIssue{Reason: "UNAUTHORIZED_MUTATION", SubjectID: entry.ID})
			}
		}
	}

	edgesByID := make(map[string]Edge, len(input.Edges))
	for _, edge := range input.Edges {
		if edge.ID == "" || edge.From == "" || edge.To == "" || (edge.Relation != "PRECEDES" && edge.Relation != "BLOCKED_BY") {
			issues = append(issues, graphIssue{Reason: "MALFORMED_GRAPH_EDGE", SubjectID: edge.ID})
			continue
		}
		if prior, exists := edgesByID[edge.ID]; exists && prior != edge {
			issues = append(issues, graphIssue{Reason: "CONTRADICTORY_EDGE", SubjectID: edge.ID})
			continue
		}
		edgesByID[edge.ID] = edge
		if _, exists := activities[edge.From]; !exists {
			if input.GraphBounded {
				issues = append(issues, graphIssue{Reason: "EDGE_ENDPOINT_MISSING", SubjectID: edge.ID, BlockingID: edge.From})
			}
		}
		if _, exists := activities[edge.To]; !exists {
			if input.GraphBounded {
				issues = append(issues, graphIssue{Reason: "EDGE_ENDPOINT_MISSING", SubjectID: edge.ID, BlockingID: edge.To})
			}
		}
	}

	if input.GraphBounded && input.GraphEvidence == nil {
		for _, activity := range input.Activities {
			if len(activity.BlockedBy) == 0 {
				continue
			}
			for _, blockerID := range activity.BlockedBy {
				blocker, exists := activities[blockerID]
				if !exists {
					issues = append(issues, graphIssue{Reason: "BLOCKER_ACTIVITY_MISSING", SubjectID: activity.ID, BlockingID: blockerID})
					continue
				}
				if !hasEdge(input.Edges, activity.ID, blockerID, "BLOCKED_BY") && !blocker.Historical {
					issues = append(issues, graphIssue{Reason: "BLOCKED_BY_EDGE_MISSING", SubjectID: activity.ID, BlockingID: blockerID})
				}
			}
		}
	}
	if cycle := precedenceCycle(input, activities); cycle != "" {
		issues = append(issues, graphIssue{Reason: "PRECEDENCE_CYCLE", SubjectID: cycle})
	}
	return activities, claims, issues
}

func hasEdge(edges []Edge, from, to, relation string) bool {
	for _, edge := range edges {
		if !edge.Historical && edge.From == from && edge.To == to && edge.Relation == relation {
			return true
		}
	}
	return false
}

func precedenceCycle(input Input, activities map[string]Activity) string {
	adjacency := make(map[string][]string, len(activities))
	for _, edge := range input.Edges {
		if edge.Historical || edge.Relation != "PRECEDES" {
			continue
		}
		if _, ok := activities[edge.From]; !ok {
			continue
		}
		if _, ok := activities[edge.To]; !ok {
			continue
		}
		adjacency[edge.From] = append(adjacency[edge.From], edge.To)
	}
	for key := range adjacency {
		adjacency[key] = sortedStrings(adjacency[key])
	}
	color := make(map[string]int, len(activities))
	var visit func(string) string
	visit = func(id string) string {
		switch color[id] {
		case 1:
			return id
		case 2:
			return ""
		}
		color[id] = 1
		for _, next := range adjacency[id] {
			if cycle := visit(next); cycle != "" {
				return cycle
			}
		}
		color[id] = 2
		return ""
	}
	ids := make([]string, 0, len(activities))
	for id := range activities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if cycle := visit(id); cycle != "" {
			return cycle
		}
	}
	return ""
}

func graphUnknownTuple(input Input) *Unknown {
	if input.GraphEvidence == nil {
		if input.GraphBounded {
			return nil
		}
		return &Unknown{Stage: "GRAPH_INPUT", Step: "BOUND_GRAPH_EVIDENCE", Reason: "GRAPH_BOUNDARY_NOT_FINITE", UnknownClass: "UNBOUNDED_GRAPH_EVIDENCE", NextOperation: "PIN_FINITE_GRAPH_BOUNDARY", BlockedBy: []string{"graph-boundary"}}
	}
	if input.GraphEvidence.State != DecisionUnknown || !input.GraphEvidence.Unknown.Valid() {
		return &Unknown{Stage: "GRAPH_INPUT", Step: "VALIDATE_GRAPH_EVIDENCE", Reason: "MALFORMED_GRAPH_UNKNOWN", UnknownClass: "CONTRADICTION", NextOperation: "RESTORE_GRAPH_EVIDENCE", BlockedBy: []string{"graph-evidence"}}
	}
	unknown := copyUnknown(input.GraphEvidence.Unknown)
	unknown.BlockedBy = sortedStrings(unknown.BlockedBy)
	return &unknown
}

func projectFrontier(input Input, activities map[string]Activity, claims map[string]Claim, decision string) ([]FrontierItem, []BlockedItem) {
	if decision == DecisionRefuted {
		return []FrontierItem{}, []BlockedItem{}
	}
	predecessors := make(map[string][]string, len(activities))
	for _, edge := range input.Edges {
		if edge.Historical {
			continue
		}
		switch edge.Relation {
		case "PRECEDES":
			predecessors[edge.To] = append(predecessors[edge.To], edge.From)
		case "BLOCKED_BY":
			predecessors[edge.From] = append(predecessors[edge.From], edge.To)
		}
	}
	for _, activity := range input.Activities {
		if activity.Historical {
			continue
		}
		predecessors[activity.ID] = append(predecessors[activity.ID], activity.BlockedBy...)
	}
	for id := range predecessors {
		predecessors[id] = uniqueSorted(predecessors[id])
	}

	ids := make([]string, 0, len(activities))
	for id := range activities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	frontier := make([]FrontierItem, 0)
	blocked := make([]BlockedItem, 0)
	for _, id := range ids {
		activity := activities[id]
		if activity.Historical || !activity.Actionable || activity.State != DecisionUnknown || !activity.Unknown.Valid() {
			continue
		}
		unresolved := make([]string, 0)
		states := make([]string, 0)
		for _, blockerID := range predecessors[id] {
			blocker, exists := activities[blockerID]
			if !exists {
				unresolved = append(unresolved, blockerID)
				states = append(states, DecisionUnknown)
				continue
			}
			if blocker.Historical {
				continue
			}
			if blocker.State == DecisionClosed && blocker.Evidence != nil && blocker.Evidence.Complete {
				continue
			}
			unresolved = append(unresolved, blockerID)
			states = append(states, blocker.State)
		}
		if len(unresolved) > 0 {
			blocked = append(blocked, BlockedItem{ActivityID: id, ClaimID: activity.ClaimID, Unknown: normalizedUnknown(activity.Unknown), BlockingIDs: uniqueSorted(unresolved), BlockingStates: uniqueSorted(states)})
			continue
		}
		claim := claims[activity.ClaimID]
		frontier = append(frontier, FrontierItem{ActivityID: id, ClaimID: claim.ID, Next: activity.Unknown.NextOperation, Unknown: normalizedUnknown(activity.Unknown)})
	}

	frontierIDs := make(map[string]bool, len(frontier))
	for _, item := range frontier {
		frontierIDs[item.ActivityID] = true
	}
	for _, item := range append([]FrontierItem(nil), frontier...) {
		for _, predecessor := range allReachablePredecessors(item.ActivityID, predecessors) {
			if frontierIDs[predecessor] {
				frontier = removeFrontier(frontier, item.ActivityID)
				frontierIDs[item.ActivityID] = false
				blocked = append(blocked, BlockedItem{ActivityID: item.ActivityID, ClaimID: item.ClaimID, Unknown: item.Unknown, BlockingIDs: []string{predecessor}, BlockingStates: []string{DecisionUnknown}})
				break
			}
		}
	}
	sort.Slice(frontier, func(i, j int) bool { return frontier[i].ActivityID < frontier[j].ActivityID })
	sort.Slice(blocked, func(i, j int) bool { return blocked[i].ActivityID < blocked[j].ActivityID })
	return frontier, blocked
}

func allReachablePredecessors(id string, predecessors map[string][]string) []string {
	seen := make(map[string]bool)
	var visit func(string)
	visit = func(current string) {
		for _, predecessor := range predecessors[current] {
			if seen[predecessor] {
				continue
			}
			seen[predecessor] = true
			visit(predecessor)
		}
	}
	visit(id)
	result := make([]string, 0, len(seen))
	for predecessor := range seen {
		result = append(result, predecessor)
	}
	sort.Strings(result)
	return result
}

func removeFrontier(items []FrontierItem, id string) []FrontierItem {
	result := make([]FrontierItem, 0, len(items))
	for _, item := range items {
		if item.ActivityID != id {
			result = append(result, item)
		}
	}
	return result
}

func historicalRefutationIDs(input Input, activities map[string]Activity) []string {
	ids := make([]string, 0)
	for _, activity := range activities {
		if activity.Historical && activity.State == DecisionRefuted {
			ids = append(ids, activity.ID)
		}
	}
	for _, entry := range input.History {
		if entry.Historical && entry.State == DecisionRefuted {
			ids = append(ids, entry.ActivityID)
		}
	}
	return uniqueSorted(ids)
}

func traceInput(input Input, activities map[string]Activity) []TraceEvent {
	trace := []TraceEvent{{Kind: "INPUT_ACCEPTED", SubjectID: "claim-activity-graph", State: input.Schema}}
	ids := make([]string, 0, len(activities))
	for id := range activities {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		activity := activities[id]
		var unknown *Unknown
		if activity.Unknown != nil {
			copy := normalizedUnknown(activity.Unknown)
			unknown = &copy
		}
		trace = append(trace, TraceEvent{Kind: "ACTIVITY_OBSERVED", SubjectID: id, State: activity.State, Historical: activity.Historical, Unknown: unknown})
	}
	edges := append([]Edge(nil), input.Edges...)
	sort.Slice(edges, func(i, j int) bool {
		if edges[i].ID != edges[j].ID {
			return edges[i].ID < edges[j].ID
		}
		if edges[i].From != edges[j].From {
			return edges[i].From < edges[j].From
		}
		return edges[i].To < edges[j].To
	})
	for _, edge := range edges {
		trace = append(trace, TraceEvent{Kind: "EDGE_OBSERVED", SubjectID: edge.ID, Historical: edge.Historical, Relation: edge.Relation, BlockingIDs: []string{edge.From, edge.To}})
	}
	for _, id := range historicalRefutationIDs(input, activities) {
		trace = append(trace, TraceEvent{Kind: "HISTORICAL_REFUTATION_EXCLUDED", SubjectID: id, State: DecisionRefuted, Historical: true, Reason: "HISTORICAL_NOT_AUTO_ACTION"})
	}
	return trace
}

func normalizedUnknown(unknown *Unknown) Unknown {
	if unknown == nil {
		return Unknown{}
	}
	copy := *unknown
	copy.BlockedBy = sortedStrings(unknown.BlockedBy)
	return copy
}

func copyUnknown(unknown *Unknown) *Unknown {
	if unknown == nil {
		return nil
	}
	copy := normalizedUnknown(unknown)
	return &copy
}

func uniqueSorted(values []string) []string {
	seen := make(map[string]bool, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" || seen[value] {
			continue
		}
		seen[value] = true
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func nonEmpty(value string) []string {
	if value == "" {
		return nil
	}
	return []string{value}
}
