package frontier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestImmutableLedgerAdapterPreservesLiveAndHistoricalClasses(t *testing.T) {
	root := filepath.Join("..", "..")
	input, err := LoadProjectInput(filepath.Join(root, "fixtures", "inputs", "immutable-ledger-v0490.json"))
	if err != nil {
		t.Fatal(err)
	}
	if input.ImmutableLedger == nil {
		t.Fatal("immutable ledger metadata was not preserved")
	}
	if input.ImmutableLedger.Release.ReleaseID != 380810861 || input.ImmutableLedger.ReleasedAsset.ID != 540115901 {
		t.Fatalf("release identity was not preserved: %+v", input.ImmutableLedger)
	}
	if input.ImmutableLedger.OperationalRefutedCount != 5 {
		t.Fatalf("operational refutation count = %d, want 5", input.ImmutableLedger.OperationalRefutedCount)
	}
	if input.ImmutableLedger.InputStatus != DecisionUnknown {
		t.Fatalf("input status = %q, want UNKNOWN", input.ImmutableLedger.InputStatus)
	}

	sourceRaw, err := os.ReadFile(filepath.Join(root, ".gooo"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseSource(sourceRaw)
	if err != nil {
		t.Fatal(err)
	}
	projection, err := EvaluateInput(spec, input, DigestBytes(sourceRaw))
	if err != nil {
		t.Fatal(err)
	}
	if projection.Canonical.Decision != DecisionClosed || projection.Canonical.Subject.InputStatus != DecisionUnknown {
		t.Fatalf("projector/input statuses = %q/%q", projection.Canonical.Decision, projection.Canonical.Subject.InputStatus)
	}
	if len(projection.Canonical.Frontier) != 1 || projection.Canonical.Frontier[0].ActivityID != "EXTERNAL_UTILITY_EVIDENCE" {
		t.Fatalf("actionable frontier = %+v", projection.Canonical.Frontier)
	}
	if len(projection.Blocked.Blocked) != 1 || projection.Blocked.Blocked[0].ActivityID != "operational:v049-parent-cache" {
		t.Fatalf("blocked frontier = %+v", projection.Blocked.Blocked)
	}
	if projection.Receipt.OperationalRefutedCount != 5 {
		t.Fatalf("receipt operational refutations = %d, want 5", projection.Receipt.OperationalRefutedCount)
	}
	if len(projection.Canonical.HistoricalRefutationsExcluded) != 7 {
		t.Fatalf("historical exclusions = %d, want 7", len(projection.Canonical.HistoricalRefutationsExcluded))
	}
}

func TestImmutableLedgerAdapterFailureClasses(t *testing.T) {
	missing := ImmutableLedgerInput{Schema: immutableLedgerInputSchema, Failure: &AdapterFailure{Kind: "MISSING"}}
	missingInput := AdaptImmutableLedger(missing, "sha256:missing")
	if missingInput.ImmutableLedger.InputStatus != DecisionUnknown || missingInput.Activities[0].Unknown == nil || !missingInput.Activities[0].Unknown.Valid() {
		t.Fatalf("missing failure was not six-field UNKNOWN: %+v", missingInput)
	}

	contradiction := ImmutableLedgerInput{Schema: immutableLedgerInputSchema, LedgerVersion: "wrong-ledger-schema"}
	refutedInput := AdaptImmutableLedger(contradiction, "sha256:contradiction")
	if refutedInput.ImmutableLedger.InputStatus != DecisionRefuted || refutedInput.Activities[0].State != DecisionRefuted {
		t.Fatalf("schema contradiction was not REFUTED: %+v", refutedInput)
	}
}
