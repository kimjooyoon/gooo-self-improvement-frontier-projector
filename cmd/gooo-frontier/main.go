package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kimjooyoon/gooo-self-improvement-frontier-projector/internal/frontier"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) == 0 {
		return usageError{}
	}
	switch args[0] {
	case "project":
		return runProject(args[1:])
	case "conformance":
		return runConformance(args[1:])
	case "inventory":
		return runInventory(args[1:])
	default:
		return fmt.Errorf("unknown command %q", args[0])
	}
}

type usageError struct{}

func (usageError) Error() string {
	return "usage: gooo-frontier {project|conformance|inventory} [flags]"
}

func runProject(args []string) error {
	flags := flag.NewFlagSet("project", flag.ContinueOnError)
	sourcePath := flags.String("source", ".gooo", "Gooo semantic source")
	contractPath := flags.String("contract", "contracts/frontier-denominator-v1.json", "fixed denominator contract")
	inputPath := flags.String("input", "", "claim/activity graph input")
	outputPath := flags.String("output", "", "caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *inputPath == "" || *outputPath == "" {
		return fmt.Errorf("project requires --input and --output")
	}
	spec, _, sourceDigest, err := loadSemantics(*sourcePath, *contractPath)
	if err != nil {
		return err
	}
	var input frontier.Input
	if _, err := frontier.LoadJSON(*inputPath, &input); err != nil {
		return err
	}
	first, err := frontier.EvaluateInput(spec, input, sourceDigest)
	if err != nil {
		return err
	}
	second, err := frontier.EvaluateInput(spec, input, sourceDigest)
	if err != nil {
		return fmt.Errorf("replay evaluation: %w", err)
	}
	first.Receipt.ReplayExact = frontier.ProjectionBytesEqual(first, second)
	if err := frontier.WriteProjection(filepath.Clean(*outputPath), first); err != nil {
		return err
	}
	if !first.Receipt.ReplayExact {
		return fmt.Errorf("replay exactness failed")
	}
	return nil
}

func runConformance(args []string) error {
	flags := flag.NewFlagSet("conformance", flag.ContinueOnError)
	sourcePath := flags.String("source", ".gooo", "Gooo semantic source")
	contractPath := flags.String("contract", "contracts/frontier-denominator-v1.json", "fixed denominator contract")
	fixturesPath := flags.String("fixtures", "fixtures/cases", "case fixture directory")
	outputPath := flags.String("output", "", "caller-owned output directory")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outputPath == "" {
		return fmt.Errorf("conformance requires --output")
	}
	spec, contract, sourceDigest, err := loadSemantics(*sourcePath, *contractPath)
	if err != nil {
		return err
	}
	projections, report, err := frontier.LoadCaseFixtures(spec, sourceDigest, contract, *fixturesPath)
	if err != nil {
		return err
	}
	if err := frontier.WriteConformance(filepath.Clean(*outputPath), projections, report); err != nil {
		return err
	}
	if !report.AllExpectedMatch || !report.ReplayExact || report.ActualCounts != report.ExpectedCounts {
		return fmt.Errorf("conformance failed: expected matches=%t replay_exact=%t actual_counts=%+v expected_counts=%+v", report.AllExpectedMatch, report.ReplayExact, report.ActualCounts, report.ExpectedCounts)
	}
	return nil
}

func runInventory(args []string) error {
	flags := flag.NewFlagSet("inventory", flag.ContinueOnError)
	rootPath := flags.String("root", ".", "repository root")
	outputPath := flags.String("output", "", "caller-owned output file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if *outputPath == "" {
		return fmt.Errorf("inventory requires --output")
	}
	inventory, err := frontier.CollectInventory(filepath.Clean(*rootPath), filepath.Clean(filepath.Dir(*outputPath)))
	if err != nil {
		return err
	}
	return frontier.WriteInventory(*outputPath, inventory)
}

func loadSemantics(sourcePath, contractPath string) (frontier.SourceSpec, frontier.Contract, string, error) {
	sourceRaw, err := os.ReadFile(sourcePath)
	if err != nil {
		return frontier.SourceSpec{}, frontier.Contract{}, "", err
	}
	spec, err := frontier.ParseSource(sourceRaw)
	if err != nil {
		return frontier.SourceSpec{}, frontier.Contract{}, "", err
	}
	var contract frontier.Contract
	if _, err := frontier.LoadJSON(contractPath, &contract); err != nil {
		return frontier.SourceSpec{}, frontier.Contract{}, "", err
	}
	if err := frontier.ValidateContract(contract); err != nil {
		return frontier.SourceSpec{}, frontier.Contract{}, "", err
	}
	return spec, contract, frontier.DigestBytes(sourceRaw), nil
}
