package frontier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestTwelveCaseCorpusAndExactCounts(t *testing.T) {
	root := filepath.Join("..", "..")
	sourceRaw, err := os.ReadFile(filepath.Join(root, ".gooo"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseSource(sourceRaw)
	if err != nil {
		t.Fatal(err)
	}
	var contract Contract
	if _, err := LoadJSON(filepath.Join(root, "contracts", "frontier-denominator-v1.json"), &contract); err != nil {
		t.Fatal(err)
	}
	if err := ValidateContract(contract); err != nil {
		t.Fatal(err)
	}
	projections, report, err := LoadCaseFixtures(spec, DigestBytes(sourceRaw), contract, filepath.Join(root, "fixtures", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	if len(projections) != 12 || report.Total != 12 {
		t.Fatalf("case count = %d/%d, want 12/12", len(projections), report.Total)
	}
	if report.ActualCounts != (ClassCounts{Closed: 4, Unknown: 4, Refuted: 4}) {
		t.Fatalf("actual counts = %+v", report.ActualCounts)
	}
	if !report.AllExpectedMatch || !report.ReplayExact {
		t.Fatalf("conformance flags = match:%t replay:%t", report.AllExpectedMatch, report.ReplayExact)
	}
}

func TestNormalFrontierShapes(t *testing.T) {
	root := filepath.Join("..", "..")
	sourceRaw, err := os.ReadFile(filepath.Join(root, ".gooo"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseSource(sourceRaw)
	if err != nil {
		t.Fatal(err)
	}
	var contract Contract
	if _, err := LoadJSON(filepath.Join(root, "contracts", "frontier-denominator-v1.json"), &contract); err != nil {
		t.Fatal(err)
	}
	projections, _, err := LoadCaseFixtures(spec, DigestBytes(sourceRaw), contract, filepath.Join(root, "fixtures", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	if len(projections[1].Projection.Canonical.Frontier) != 2 {
		t.Fatalf("parallel frontier size = %d, want 2", len(projections[1].Projection.Canonical.Frontier))
	}
	if projections[1].Projection.Canonical.Frontier[0].ActivityID != "activity-parallel-a" || projections[1].Projection.Canonical.Frontier[1].ActivityID != "activity-parallel-b" {
		t.Fatalf("parallel frontier is not stable-ID sorted: %+v", projections[1].Projection.Canonical.Frontier)
	}
	if len(projections[2].Projection.Canonical.Frontier) != 1 || projections[2].Projection.Canonical.Frontier[0].ActivityID != "activity-proof" {
		t.Fatalf("dependency frontier = %+v", projections[2].Projection.Canonical.Frontier)
	}
	if len(projections[2].Projection.Blocked.Blocked) != 1 || projections[2].Projection.Blocked.Blocked[0].ActivityID != "activity-publish" {
		t.Fatalf("dependency blocked frontier = %+v", projections[2].Projection.Blocked)
	}
	if len(projections[3].Projection.Canonical.HistoricalRefutationsExcluded) != 1 || projections[3].Projection.Canonical.HistoricalRefutationsExcluded[0] != "activity-old-refuted" {
		t.Fatalf("historical refutation exclusion = %+v", projections[3].Projection.Canonical.HistoricalRefutationsExcluded)
	}
}

func TestUnknownTupleAndRefutationPrecedence(t *testing.T) {
	root := filepath.Join("..", "..")
	sourceRaw, err := os.ReadFile(filepath.Join(root, ".gooo"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseSource(sourceRaw)
	if err != nil {
		t.Fatal(err)
	}
	var contract Contract
	if _, err := LoadJSON(filepath.Join(root, "contracts", "frontier-denominator-v1.json"), &contract); err != nil {
		t.Fatal(err)
	}
	projections, _, err := LoadCaseFixtures(spec, DigestBytes(sourceRaw), contract, filepath.Join(root, "fixtures", "cases"))
	if err != nil {
		t.Fatal(err)
	}
	unknown := projections[5].Projection.Trace[1].Unknown
	if unknown == nil || !unknown.Valid() || unknown.UnknownClass != "STALE" {
		t.Fatalf("unknown tuple not preserved: %+v", unknown)
	}
	if projections[8].Projection.Canonical.Decision != DecisionRefuted || projections[9].Projection.Canonical.Decision != DecisionRefuted || projections[10].Projection.Canonical.Decision != DecisionRefuted || projections[11].Projection.Canonical.Decision != DecisionRefuted {
		t.Fatalf("refutation precedence was not preserved")
	}
}
