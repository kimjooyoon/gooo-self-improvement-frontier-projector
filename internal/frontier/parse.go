package frontier

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

func ParseSource(raw []byte) (SourceSpec, error) {
	spec := SourceSpec{Language: "gooo"}
	seenActivities := make(map[string]bool)
	seenRules := make(map[string]bool)
	for _, rawLine := range strings.Split(string(raw), "\n") {
		line := strings.TrimSpace(rawLine)
		if strings.HasPrefix(line, "namespace ") {
			spec.Namespace = strings.TrimSpace(strings.TrimPrefix(line, "namespace "))
		}
		if strings.HasPrefix(line, "activity ") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "activity "))
			nameEnd := strings.IndexByte(rest, '(')
			if nameEnd <= 0 {
				return SourceSpec{}, fmt.Errorf("activity declaration has no stable ID: %q", line)
			}
			name := strings.TrimSpace(rest[:nameEnd])
			if seenActivities[name] {
				return SourceSpec{}, fmt.Errorf("duplicate activity %q", name)
			}
			seenActivities[name] = true
			spec.Activities = append(spec.Activities, name)
		}
		if strings.HasPrefix(line, "rule ") {
			rest := strings.TrimSpace(strings.TrimPrefix(line, "rule "))
			parts := strings.SplitN(rest, " ", 2)
			if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" || strings.TrimSpace(parts[1]) == "" {
				return SourceSpec{}, fmt.Errorf("rule declaration is incomplete: %q", line)
			}
			id := strings.TrimSpace(parts[0])
			if seenRules[id] {
				return SourceSpec{}, fmt.Errorf("duplicate rule %q", id)
			}
			seenRules[id] = true
			spec.Rules = append(spec.Rules, SemanticRule{ID: id, Value: strings.TrimSpace(parts[1])})
		}
	}
	if spec.Namespace == "" {
		return SourceSpec{}, fmt.Errorf("source namespace is required")
	}
	if len(spec.Activities) == 0 {
		return SourceSpec{}, fmt.Errorf("source must declare activities")
	}
	if !equalStrings(spec.Activities, RequiredActivityIDs) {
		return SourceSpec{}, fmt.Errorf(".gooo activities do not match the canonical frontier projector activities")
	}
	if err := validateSourceRules(spec.Rules); err != nil {
		return SourceSpec{}, err
	}
	return spec, nil
}

func validateSourceRules(rules []SemanticRule) error {
	values := make(map[string]string, len(rules))
	for _, rule := range rules {
		values[rule.ID] = rule.Value
	}
	for _, id := range RequiredSourceRules {
		if values[id] == "" {
			return fmt.Errorf(".gooo rule %q is required", id)
		}
	}
	if values["precedence"] != "REFUTED > UNKNOWN > CLOSED" {
		return fmt.Errorf(".gooo precedence rule is not canonical")
	}
	if values["partial_order"] != "PRECEDES:from->to" {
		return fmt.Errorf(".gooo partial order rule is not canonical")
	}
	if values["blocked_by"] != "BLOCKED_BY:blocked->blocker" {
		return fmt.Errorf(".gooo blocked_by rule is not canonical")
	}
	if values["frontier"] != "MINIMAL_ANTICHAIN:live+actionable+nonhistorical" {
		return fmt.Errorf(".gooo frontier rule is not canonical")
	}
	if values["historical_refutation"] != "AUTO_ACTION=false" {
		return fmt.Errorf(".gooo historical refutation rule is not canonical")
	}
	if values["unknown_fields"] != strings.Join(RequiredUnknownFields, ",") {
		return fmt.Errorf(".gooo unknown field rule is not canonical")
	}
	if values["acceptance_required_gate"] != "0" {
		return fmt.Errorf(".gooo acceptance gate must be zero")
	}
	if values["runtime_authority"] != "repository_writes=0;local_test_executions=0;cross_project_required_gates=0" {
		return fmt.Errorf(".gooo runtime authority is not zeroed")
	}
	if values["product_authority"] != "commit=0;merge=0;release=0" {
		return fmt.Errorf(".gooo product authority is not zeroed")
	}
	if values["improvement"] != "exact-pair:scenario+fixture+contract+toolchain+runner;per-indicator-only;aggregate=NOT_COMBINED" {
		return fmt.Errorf(".gooo improvement policy is not canonical")
	}
	if values["missing_metric"] != "null+UNKNOWN" {
		return fmt.Errorf(".gooo missing metric policy is not canonical")
	}
	if values["external_utility"] != "NO_INTERNAL_CLOSE" {
		return fmt.Errorf(".gooo external utility policy is not canonical")
	}
	return nil
}

func LoadJSON(path string, target any) ([]byte, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(target); err != nil {
		return nil, fmt.Errorf("decode %s: %w", path, err)
	}
	return raw, nil
}

func ValidateContract(contract Contract) error {
	if contract.Schema != "gooo/self-improvement-frontier/denominator/v1" {
		return fmt.Errorf("unexpected contract schema %q", contract.Schema)
	}
	if contract.DenominatorID == "" || contract.Total != 12 || len(contract.Cases) != 12 {
		return fmt.Errorf("contract must contain exactly 12 cases")
	}
	if !equalStrings(contract.DecisionPrecedence, DecisionPrecedence) {
		return fmt.Errorf("contract precedence must be REFUTED > UNKNOWN > CLOSED")
	}
	if contract.ClassCounts != (ClassCounts{Closed: 4, Unknown: 4, Refuted: 4}) {
		return fmt.Errorf("contract class counts must be CLOSED=4 UNKNOWN=4 REFUTED=4")
	}
	if !equalStrings(contract.UnknownFields, RequiredUnknownFields) {
		return fmt.Errorf("contract unknown fields are not the six canonical fields")
	}
	if contract.Authority != (Authority{RepositoryWrites: 0, LocalTestExecutions: 0, CrossProjectRequiredGates: 0, AcceptanceRequiredGate: 0, ProductCommitMergeRelease: 0}) {
		return fmt.Errorf("contract authority boundary is not zeroed")
	}
	if !validImprovementPolicy(contract.Improvement) {
		return fmt.Errorf("contract improvement policy is not exact-pair/per-indicator-only")
	}
	seenIDs := make(map[string]bool, len(contract.Cases))
	seenOrdinals := make(map[int]bool, len(contract.Cases))
	for _, item := range contract.Cases {
		if item.Ordinal < 1 || item.Ordinal > 12 || seenOrdinals[item.Ordinal] || item.ID == "" || seenIDs[item.ID] {
			return fmt.Errorf("contract case identity is not unique")
		}
		if item.Class != item.ExpectedDecision || !validDecision(item.ExpectedDecision) {
			return fmt.Errorf("contract case %q has an invalid class/decision", item.ID)
		}
		if item.Fixture == "" || item.Coverage == "" {
			return fmt.Errorf("contract case %q is missing fixture metadata", item.ID)
		}
		seenIDs[item.ID] = true
		seenOrdinals[item.Ordinal] = true
	}
	return nil
}

func validImprovementPolicy(policy ImprovementPolicy) bool {
	return policy.Mode == "PER_INDICATOR_ONLY" &&
		equalStrings(policy.IdentityFields, []string{"scenario", "fixture", "contract", "toolchain", "runner"}) &&
		policy.Aggregate == "NOT_COMBINED" && policy.MissingMetric == "null+UNKNOWN"
}

func ValidateCaseFixture(contractCase ContractCase, fixture CaseFixture) error {
	if fixture.CaseID != contractCase.ID || fixture.Class != contractCase.Class || fixture.ExpectedDecision != contractCase.ExpectedDecision {
		return fmt.Errorf("fixture %q does not match contract case %q", fixture.CaseID, contractCase.ID)
	}
	if fixture.Input.Schema != "gooo/self-improvement-frontier/input/v1" {
		return fmt.Errorf("fixture %q has unexpected input schema", fixture.CaseID)
	}
	return nil
}

func sortedStrings(values []string) []string {
	result := append([]string(nil), values...)
	sort.Strings(result)
	return result
}

func equalStrings(left, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
