package frontier

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func CollectInventory(root, output string) (Inventory, error) {
	root = filepath.Clean(root)
	output = filepath.Clean(output)
	result := Inventory{RootREADMEExcluded: true}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if path != root && (path == output || pathWithin(path, output)) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path != root && excludedDirectory(entry.Name()) {
			if entry.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}
		if path == filepath.Join(root, "README.md") {
			return nil
		}
		if path != root && entry.IsDir() {
			result.DescendantDirectories++
			return nil
		}
		if !entry.Type().IsRegular() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		result.RegularFiles++
		switch filepath.Ext(path) {
		case ".go":
			result.GoFiles++
			result.GoPhysicalLines += physicalLines(data)
		case ".gooo":
			result.GoooFiles++
			result.GoooPhysicalLines += physicalLines(data)
		}
		return nil
	})
	return result, err
}

func WriteInventory(path string, inventory Inventory) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create inventory output: %w", err)
	}
	raw, err := json.MarshalIndent(inventory, "", "  ")
	if err != nil {
		return err
	}
	raw = append(raw, '\n')
	if err := os.WriteFile(path, raw, 0o644); err != nil {
		return err
	}
	return nil
}

func pathWithin(path, parent string) bool {
	relative, err := filepath.Rel(parent, path)
	if err != nil || relative == "." {
		return relative == "."
	}
	return relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
}

func excludedDirectory(name string) bool {
	switch name {
	case ".git", ".cache", "cache", "tmp", "temp", "output", "out", "vendor", "toolchain", "toolchains", ".toolchain":
		return true
	default:
		return false
	}
}

func physicalLines(data []byte) int {
	if len(data) == 0 {
		return 0
	}
	lines := bytes.Count(data, []byte{'\n'})
	if data[len(data)-1] != '\n' {
		lines++
	}
	return lines
}
