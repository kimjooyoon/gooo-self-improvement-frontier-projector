package frontier

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type traceRecord struct {
	Schema string      `json:"schema"`
	Event  TraceEvent  `json:"event"`
}

func WriteProjection(output string, projection Projection) error {
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("create output directory: %w", err)
	}
	if err := writeJSON(filepath.Join(output, "canonical-frontier.json"), projection.Canonical); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(output, "blocked-frontier.json"), projection.Blocked); err != nil {
		return err
	}
	if err := writeTrace(filepath.Join(output, "causal-trace.ndjson"), projection.Trace); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(output, "human-report.md"), []byte(projection.Report), 0o644); err != nil {
		return fmt.Errorf("write human report: %w", err)
	}
	if err := writeJSON(filepath.Join(output, "semantic-ir.json"), projection.SemanticIR); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(output, "provenance-graph.json"), projection.Graph); err != nil {
		return err
	}
	if err := writeJSON(filepath.Join(output, "receipt.json"), projection.Receipt); err != nil {
		return err
	}
	return nil
}

func writeJSON(path string, value any) error {
	raw, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal %s: %w", path, err)
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func writeTrace(path string, events []TraceEvent) error {
	var builder strings.Builder
	for _, event := range events {
		raw, err := json.Marshal(traceRecord{Schema: traceSchema, Event: event})
		if err != nil {
			return fmt.Errorf("marshal causal trace: %w", err)
		}
		builder.Write(raw)
		builder.WriteByte('\n')
	}
	if err := os.WriteFile(path, []byte(builder.String()), 0o644); err != nil {
		return fmt.Errorf("write %s: %w", path, err)
	}
	return nil
}

func BuildHumanReport(frontier CanonicalFrontier, blocked BlockedFrontier, receipt Receipt, trace []TraceEvent) string {
	var builder strings.Builder
	builder.WriteString("# Self-improvement frontier projection\n\n")
	fmt.Fprintf(&builder, "Decision: `%s`\n\n", frontier.Decision)
	fmt.Fprintf(&builder, "Input digest: `%s`\n\n", frontier.Subject.InputDigest)
	builder.WriteString("The decision precedence is `REFUTED > UNKNOWN > CLOSED`. The frontier is a proposal only; product commit, merge, and release authority are zero. No score or percentage is emitted.\n\n")
	builder.WriteString("## Canonical frontier\n\n")
	fmt.Fprintf(&builder, "Antichain: `%t`; items: `%d`.\n\n", frontier.Antichain, len(frontier.Frontier))
	builder.WriteString("| activity_id | claim_id | next_operation | unknown_class | blocked_by |\n|---|---|---|---|---|\n")
	for _, item := range frontier.Frontier {
		fmt.Fprintf(&builder, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n", item.ActivityID, item.ClaimID, item.Next, item.Unknown.UnknownClass, strings.Join(item.Unknown.BlockedBy, ","))
	}
	if len(frontier.Frontier) == 0 {
		builder.WriteString("| _none_ |  |  |  |  |\n")
	}
	builder.WriteString("\n## Blocked frontier\n\n")
	fmt.Fprintf(&builder, "Direct frontier items: `%d`; dependency-blocked items: `%d`.\n\n", blocked.DirectCount, blocked.BlockedCount)
	builder.WriteString("| activity_id | claim_id | blocking_ids | blocking_states | unknown_class |\n|---|---|---|---|---|\n")
	for _, item := range blocked.Blocked {
		fmt.Fprintf(&builder, "| `%s` | `%s` | `%s` | `%s` | `%s` |\n", item.ActivityID, item.ClaimID, strings.Join(item.BlockingIDs, ","), strings.Join(item.BlockingStates, ","), item.Unknown.UnknownClass)
	}
	if len(blocked.Blocked) == 0 {
		builder.WriteString("| _none_ |  |  |  |  |\n")
	}
	builder.WriteString("\n## Historical refutation handling\n\n")
	if len(frontier.HistoricalRefutationsExcluded) == 0 {
		builder.WriteString("Excluded historical refutations: none.\n\n")
	} else {
		fmt.Fprintf(&builder, "Excluded historical refutations: `%s`. These IDs are preserved in the causal trace and are not automatic actions.\n\n", strings.Join(frontier.HistoricalRefutationsExcluded, "`, `"))
	}
	builder.WriteString("## Unknown tuple fields\n\n")
	builder.WriteString("Every emitted unknown tuple has the six fields `stage`, `step`, `reason`, `unknown_class`, `next_operation`, and `blocked_by`.\n\n")
	builder.WriteString("## Runtime boundary\n\n")
	fmt.Fprintf(&builder, "repository_writes=`%d`; source_mutations=`%d`; commit=`%d`; merge=`%d`; release=`%d`; local_test_executions=`%d`; cross_project_required_gates=`%d`; acceptance_required_gate=`%d`; external_utility=`%s`.\n\n", receipt.Authority.RuntimeAuthority.RepositoryWrites, receipt.Authority.RuntimeAuthority.SourceMutations, receipt.Authority.RuntimeAuthority.Commit, receipt.Authority.RuntimeAuthority.Merge, receipt.Authority.RuntimeAuthority.Release, receipt.Authority.RuntimeAuthority.LocalTestExecutions, receipt.Authority.RuntimeAuthority.CrossProjectRequiredGates, receipt.Authority.RuntimeAuthority.AcceptanceRequiredGate, frontier.Subject.ExternalUtilityState)
	fmt.Fprintf(&builder, "Improvement policy: mode=`%s`; identity=`%s`; aggregate=`%s`; missing_metric=`%s`.\n\n", receipt.Improvement.Mode, strings.Join(receipt.Improvement.IdentityFields, ","), receipt.Improvement.Aggregate, receipt.Improvement.MissingMetric)
	fmt.Fprintf(&builder, "Operator authoring boundary is separate from runtime: `%t`. Causal trace events: `%d`. Replay exact: `%t`.\n", receipt.Authority.OperatorBoundary.SeparateFromRuntime, len(trace), receipt.ReplayExact)
	return builder.String()
}
