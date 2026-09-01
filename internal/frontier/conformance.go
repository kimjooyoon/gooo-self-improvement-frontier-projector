package frontier

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type CaseOutcome struct {
	Ordinal          int    `json:"ordinal"`
	ID               string `json:"id"`
	Class            string `json:"class"`
	Coverage         string `json:"coverage"`
	ExpectedDecision string `json:"expected_decision"`
	Decision         string `json:"decision"`
	ExpectedMatch    bool   `json:"expected_match"`
	FrontierCount    int    `json:"frontier_count"`
	BlockedCount     int    `json:"blocked_count"`
	ReplayExact      bool   `json:"replay_exact"`
}

type ConformanceReport struct {
	Schema            string          `json:"schema"`
	DenominatorID     string          `json:"denominator_id"`
	Total             int             `json:"total"`
	ExpectedCounts    ClassCounts     `json:"expected_counts"`
	ActualCounts      ClassCounts     `json:"actual_counts"`
	Cases             []CaseOutcome   `json:"cases"`
	ReplayExact       bool            `json:"replay_exact"`
	AllExpectedMatch  bool            `json:"all_expected_match"`
	Authority         AuthorityReport `json:"authority"`
	LocalExecution    LocalExecution  `json:"local_execution"`
	NoScoresOrPercent bool            `json:"no_scores_or_percentages"`
}

type FixtureProjection struct {
	Case       ContractCase
	Fixture    CaseFixture
	Projection Projection
}

func LoadCaseFixtures(spec SourceSpec, sourceDigest string, contract Contract, directory string) ([]FixtureProjection, ConformanceReport, error) {
	projections := make([]FixtureProjection, 0, len(contract.Cases))
	report := ConformanceReport{
		Schema: "gooo/self-improvement-frontier/conformance-report/v1",
		DenominatorID: contract.DenominatorID,
		Total: contract.Total,
		ExpectedCounts: contract.ClassCounts,
		AllExpectedMatch: true,
		ReplayExact: true,
		Authority: AuthorityReport{RuntimeAuthority: RuntimeAuthority{}, OperatorBoundary: OperatorBoundary{SeparateFromRuntime: true}},
		LocalExecution: LocalExecution{},
		NoScoresOrPercent: true,
	}
	for _, contractCase := range contract.Cases {
		path := filepath.Join(directory, contractCase.Fixture)
		var fixture CaseFixture
		_, err := LoadJSON(path, &fixture)
		if err != nil {
			return nil, ConformanceReport{}, err
		}
		if err := ValidateCaseFixture(contractCase, fixture); err != nil {
			return nil, ConformanceReport{}, err
		}
		first, err := EvaluateInput(spec, fixture.Input, sourceDigest)
		if err != nil {
			return nil, ConformanceReport{}, fmt.Errorf("evaluate %s: %w", fixture.CaseID, err)
		}
		second, err := EvaluateInput(spec, fixture.Input, sourceDigest)
		if err != nil {
			return nil, ConformanceReport{}, fmt.Errorf("replay %s: %w", fixture.CaseID, err)
		}
		replayExact := ProjectionBytesEqual(first, second)
		first.Receipt.ReplayExact = replayExact
		outcome := CaseOutcome{
			Ordinal: contractCase.Ordinal,
			ID: contractCase.ID,
			Class: contractCase.Class,
			Coverage: contractCase.Coverage,
			ExpectedDecision: contractCase.ExpectedDecision,
			Decision: first.Canonical.Decision,
			ExpectedMatch: first.Canonical.Decision == contractCase.ExpectedDecision,
			FrontierCount: len(first.Canonical.Frontier),
			BlockedCount: len(first.Blocked.Blocked),
			ReplayExact: replayExact,
		}
		report.Cases = append(report.Cases, outcome)
		report.ReplayExact = report.ReplayExact && replayExact
		report.AllExpectedMatch = report.AllExpectedMatch && outcome.ExpectedMatch
		incrementClass(&report.ActualCounts, first.Canonical.Decision)
		projections = append(projections, FixtureProjection{Case: contractCase, Fixture: fixture, Projection: first})
	}
	sort.Slice(report.Cases, func(i, j int) bool { return report.Cases[i].Ordinal < report.Cases[j].Ordinal })
	return projections, report, nil
}

func ProjectionBytesEqual(left, right Projection) bool {
	leftRaw := projectionBytes(left)
	rightRaw := projectionBytes(right)
	return string(leftRaw) == string(rightRaw)
}

func projectionBytes(projection Projection) []byte {
	semantic, _ := CanonicalJSON(projection.SemanticIR)
	graph, _ := CanonicalJSON(projection.Graph)
	canonical, _ := CanonicalJSON(projection.Canonical)
	blocked, _ := CanonicalJSON(projection.Blocked)
	receipt, _ := CanonicalJSON(projection.Receipt)
	return []byte(strings.Join([]string{string(semantic), string(graph), string(canonical), string(blocked), string(receipt), projection.Report, traceJSON(projection.Trace)}, "\n"))
}

func traceJSON(events []TraceEvent) string {
	var builder strings.Builder
	for _, event := range events {
		raw, _ := CanonicalJSON(traceRecord{Schema: traceSchema, Event: event})
		builder.Write(raw)
		builder.WriteByte('\n')
	}
	return builder.String()
}

func incrementClass(counts *ClassCounts, decision string) {
	switch decision {
	case DecisionClosed:
		counts.Closed++
	case DecisionUnknown:
		counts.Unknown++
	case DecisionRefuted:
		counts.Refuted++
	}
}

func WriteConformance(output string, projections []FixtureProjection, report ConformanceReport) error {
	if err := os.MkdirAll(output, 0o755); err != nil {
		return fmt.Errorf("create conformance output: %w", err)
	}
	for _, fixtureProjection := range projections {
		caseOutput := filepath.Join(output, "cases", fmt.Sprintf("%02d-%s", fixtureProjection.Case.Ordinal, fixtureProjection.Case.ID))
		if err := WriteProjection(caseOutput, fixtureProjection.Projection); err != nil {
			return err
		}
	}
	if err := writeJSON(filepath.Join(output, "conformance-report.json"), report); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(output, "conformance-report.md"), []byte(BuildConformanceReport(report)), 0o644); err != nil {
		return fmt.Errorf("write conformance report: %w", err)
	}
	return nil
}

func BuildConformanceReport(report ConformanceReport) string {
	var builder strings.Builder
	builder.WriteString("# Frontier projector conformance\n\n")
	fmt.Fprintf(&builder, "Denominator: `%d`; CLOSED=`%d`; UNKNOWN=`%d`; REFUTED=`%d`.\n\n", report.Total, report.ExpectedCounts.Closed, report.ExpectedCounts.Unknown, report.ExpectedCounts.Refuted)
	fmt.Fprintf(&builder, "Actual decisions: CLOSED=`%d`; UNKNOWN=`%d`; REFUTED=`%d`.\n\n", report.ActualCounts.Closed, report.ActualCounts.Unknown, report.ActualCounts.Refuted)
	fmt.Fprintf(&builder, "All expected matches: `%t`; replay exact: `%t`; no scores or percentages: `%t`.\n\n", report.AllExpectedMatch, report.ReplayExact, report.NoScoresOrPercent)
	builder.WriteString("| ordinal | case | coverage | expected | actual | match | replay | frontier | blocked |\n|---:|---|---|---|---|---|---|---:|---:|\n")
	for _, item := range report.Cases {
		fmt.Fprintf(&builder, "| %d | `%s` | `%s` | `%s` | `%s` | `%t` | `%t` | %d | %d |\n", item.Ordinal, item.ID, item.Coverage, item.ExpectedDecision, item.Decision, item.ExpectedMatch, item.ReplayExact, item.FrontierCount, item.BlockedCount)
	}
	builder.WriteString("\nRuntime authority: repository_writes=`0`; local_test_executions=`0`; cross_project_required_gates=`0`; acceptance_required_gate=`0`; product_commit_merge_release=`0`.\n")
	return builder.String()
}
