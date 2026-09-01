package frontier

import (
	"os"
	"path/filepath"
	"testing"
)

func TestOutputNamesAndReplayBundle(t *testing.T) {
	root := filepath.Join("..", "..")
	sourceRaw, err := os.ReadFile(filepath.Join(root, ".gooo"))
	if err != nil {
		t.Fatal(err)
	}
	spec, err := ParseSource(sourceRaw)
	if err != nil {
		t.Fatal(err)
	}
	var input Input
	if _, err := LoadJSON(filepath.Join(root, "fixtures", "inputs", "shared-ledger-v0480.json"), &input); err != nil {
		t.Fatal(err)
	}
	first, err := EvaluateInput(spec, input, DigestBytes(sourceRaw))
	if err != nil {
		t.Fatal(err)
	}
	second, err := EvaluateInput(spec, input, DigestBytes(sourceRaw))
	if err != nil {
		t.Fatal(err)
	}
	if !ProjectionBytesEqual(first, second) {
		t.Fatal("projection bundle is not replay exact")
	}
	if err := WriteProjection(t.TempDir(), first); err != nil {
		t.Fatal(err)
	}
}
