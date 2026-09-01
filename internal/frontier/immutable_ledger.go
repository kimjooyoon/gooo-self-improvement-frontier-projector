package frontier

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"
)

const (
	immutableLedgerInputSchema = "gooo/self-improvement-frontier/immutable-ledger-input/v1"
	ledgerReportSchema         = "gooo/self-improvement-portfolio/report/v1"
	adapterActivityID          = "immutable-ledger-adapter"
)

type adapterValidation struct {
	State   string
	Reason  string
	Unknown Unknown
}

func LoadProjectInput(path string) (Input, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return Input{}, err
	}
	var header struct {
		Schema string `json:"schema"`
	}
	if err := json.Unmarshal(raw, &header); err != nil {
		return Input{}, fmt.Errorf("decode %s: %w", path, err)
	}
	if header.Schema == immutableLedgerInputSchema {
		var envelope ImmutableLedgerInput
		if err := decodeStrict(raw, &envelope); err != nil {
			return Input{}, fmt.Errorf("decode immutable ledger input %s: %w", path, err)
		}
		return AdaptImmutableLedger(envelope, DigestBytes(raw)), nil
	}
	var input Input
	if err := decodeStrict(raw, &input); err != nil {
		return Input{}, fmt.Errorf("decode %s: %w", path, err)
	}
	return input, nil
}

func decodeStrict(raw []byte, target any) error {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.DisallowUnknownFields()
	return decoder.Decode(target)
}

func AdaptImmutableLedger(envelope ImmutableLedgerInput, envelopeDigest string) Input {
	validation := validateImmutableLedger(envelope)
	metadata := ImmutableLedgerMetadata{
		EnvelopeSchema: immutableLedgerInputSchema,
		EnvelopeDigest: envelopeDigest,
		LedgerVersion:  envelope.LedgerVersion,
		Profile:        envelope.Profile,
		Release:        envelope.Release,
		Tag:            envelope.Tag,
		ReleasedAsset:  envelope.ReleasedAsset,
	}
	input := Input{
		Schema: inputSchema,
		Source: SourceBinding{
			Kind:                   "immutable-ledger",
			Release:                envelope.Release.Tag,
			Immutable:              envelope.Release.Immutable,
			AcceptanceRequiredGate: 0,
			ExternalUtilityState:   "UNKNOWN",
		},
		ImmutableHistory: true,
		GraphBounded:    true,
		ImmutableLedger: &metadata,
		OperationalEvents: append([]OperationalEvent(nil), envelope.OperationalEvents...),
	}
	if validation.State != "" {
		input.ImmutableLedger.InputStatus = validation.State
		failureKind := validation.State
		if envelope.Failure != nil && envelope.Failure.Kind != "" {
			failureKind = strings.ToUpper(envelope.Failure.Kind)
		}
		input.ImmutableLedger.Failure = &AdapterFailure{
			Kind:          failureKind,
			Stage:         validation.Unknown.Stage,
			Step:          validation.Unknown.Step,
			Reason:        validation.Unknown.Reason,
			UnknownClass:  validation.Unknown.UnknownClass,
			NextOperation: validation.Unknown.NextOperation,
			BlockedBy:     append([]string(nil), validation.Unknown.BlockedBy...),
		}
		return adapterFailureInput(input, validation)
	}

	cellIDs := make(map[string]bool, len(envelope.Cells))
	for _, cell := range envelope.Cells {
		cellIDs[cell.ID] = true
	}
	for _, cell := range sortedLedgerCells(envelope.Cells) {
		activity := activityFromCell(cell, cellIDs)
		input.Claims = append(input.Claims, Claim{
			ID:         "claim:" + cell.ID,
			ActivityID: activity.ID,
			State:      activity.State,
			Historical: activity.Historical,
			Immutable:  true,
		})
		input.Activities = append(input.Activities, activity)
		if cell.State == DecisionRefuted {
			input.History = append(input.History, HistoryEntry{
				ID:         "history:cell:" + cell.ID,
				ActivityID: cell.ID,
				State:      DecisionRefuted,
				Historical: true,
				Immutable:  true,
				Mutation:   "APPEND",
			})
		}
	}
	for _, event := range sortedOperationalEvents(envelope.OperationalEvents) {
		switch event.State {
		case "OPERATIONAL_REFUTED":
			input.History = append(input.History, HistoryEntry{
				ID:         "history:operational:" + event.ID,
				ActivityID: event.ID,
				State:      DecisionRefuted,
				Historical: event.Historical,
				Immutable:  true,
				Mutation:   "APPEND",
			})
		case DecisionUnknown, DecisionClosed:
			activity := activityFromOperationalEvent(event, cellIDs)
			input.Claims = append(input.Claims, Claim{
				ID:         "claim:" + event.ID,
				ActivityID: activity.ID,
				State:      activity.State,
				Historical: activity.Historical,
				Immutable:  true,
			})
			input.Activities = append(input.Activities, activity)
		}
	}
	input.Edges = edgesFromActivities(input.Activities)
	metadata.InputStatus = inputStatusFromActivities(input.Activities)
	metadata.CellCount = len(envelope.Cells)
	metadata.UnknownCellIDs, metadata.RefutedCellIDs = cellStateIDs(envelope.Cells)
	metadata.OperationalRefutedCount = countOperationalRefuted(envelope.OperationalEvents)
	return input
}

func validateImmutableLedger(envelope ImmutableLedgerInput) adapterValidation {
	if envelope.Schema != immutableLedgerInputSchema {
		return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VALIDATE_INPUT_SCHEMA", "INPUT_SCHEMA_DOES_NOT_MATCH_ADAPTER", []string{"schema"})
	}
	if envelope.Failure != nil {
		return validationFromFailure(*envelope.Failure)
	}
	if envelope.LedgerVersion == "" {
		return adapterUnknown("IMMUTABLE_LEDGER_ADAPTER", "READ_EXPLICIT_LEDGER_VERSION", "LEDGER_VERSION_NOT_PROVIDED", "DIRECT_MISSING", "PROVIDE_EXPLICIT_LEDGER_VERSION", []string{"ledger_version"})
	}
	if envelope.LedgerVersion != ledgerReportSchema {
		return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VALIDATE_EXPLICIT_LEDGER_VERSION", "LEDGER_VERSION_DOES_NOT_MATCH_PROFILE_SCHEMA", []string{"ledger_version", "profile.schema"})
	}
	if missingReleaseMetadata(envelope.Release, envelope.Tag, envelope.ReleasedAsset) {
		return adapterUnknown("IMMUTABLE_LEDGER_ADAPTER", "READ_RELEASE_IDENTITY", "RELEASE_IDENTITY_NOT_COMPLETE", "DIRECT_MISSING", "PROVIDE_RELEASE_TAG_TARGET_ASSET_DIGEST", []string{"release", "tag", "released_asset"})
	}
	if !envelope.Release.Immutable {
		return adapterRefuted("RELEASE_IDENTITY_CONTRADICTION", "VERIFY_RELEASE_IMMUTABILITY", "RELEASE_IS_NOT_IMMUTABLE", []string{"release.immutable"})
	}
	if envelope.Release.Tag != envelope.Tag.Name || envelope.Release.TagObjectSHA != envelope.Tag.ObjectSHA || envelope.Release.TargetCommitSHA != envelope.Tag.TargetCommitSHA {
		return adapterRefuted("RELEASE_IDENTITY_CONTRADICTION", "VERIFY_RELEASE_TAG_BINDING", "RELEASE_TAG_OBJECT_TARGET_CONTRADICTED", []string{"release", "tag"})
	}
	if !matchingReleaseAsset(envelope.Release.Assets, envelope.ReleasedAsset) {
		return adapterRefuted("RELEASE_IDENTITY_CONTRADICTION", "VERIFY_RELEASE_ASSET_BINDING", "RELEASED_ASSET_NOT_BOUND_TO_RELEASE_DIGEST", []string{"release.assets", "released_asset"})
	}
	if envelope.Profile.Schema == "" || envelope.Profile.ProfileID == "" || envelope.Profile.AssessmentID == "" {
		return adapterUnknown("IMMUTABLE_LEDGER_ADAPTER", "READ_PROFILE_IDENTITY", "PROFILE_IDENTITY_NOT_COMPLETE", "DIRECT_MISSING", "PROVIDE_PROFILE_SCHEMA_AND_STABLE_IDS", []string{"profile"})
	}
	if envelope.Profile.Schema != ledgerReportSchema {
		return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_PROFILE_SCHEMA", "PROFILE_SCHEMA_DOES_NOT_MATCH_EXPLICIT_LEDGER_VERSION", []string{"ledger_version", "profile.schema"})
	}
	if envelope.Profile.SubjectSHA != "" && envelope.Profile.SubjectSHA != envelope.Release.TargetCommitSHA {
		return adapterRefuted("RELEASE_IDENTITY_CONTRADICTION", "VERIFY_PROFILE_SUBJECT", "PROFILE_SUBJECT_DOES_NOT_MATCH_RELEASE_TARGET", []string{"profile.subject_sha", "tag.target_commit_sha"})
	}
	if !equalStrings(envelope.Profile.Precedence, DecisionPrecedence) {
		return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_PROFILE_PRECEDENCE", "PROFILE_PRECEDENCE_DOES_NOT_MATCH_CANONICAL_PRECEDENCE", []string{"profile.precedence"})
	}
	if len(envelope.Cells) == 0 {
		return adapterUnknown("IMMUTABLE_LEDGER_ADAPTER", "READ_PROFILE_CELLS", "PROFILE_CELLS_NOT_OBSERVED", "DIRECT_MISSING", "PROVIDE_PROFILE_CELLS", []string{"cells"})
	}
	seenIDs := make(map[string]bool, len(envelope.Cells))
	seenOrdinals := make(map[int]bool, len(envelope.Cells))
	counts := LedgerSummary{}
	for _, cell := range envelope.Cells {
		if cell.ID == "" || cell.Activity == "" || cell.Ordinal < 1 || seenIDs[cell.ID] || seenOrdinals[cell.Ordinal] || !validDecision(cell.State) || cell.Denominator != 1 || (cell.Numerator != 0 && cell.Numerator != 1) {
			return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_PROFILE_CELL_IDENTITY", "PROFILE_CELL_SCHEMA_OR_IDENTITY_CONTRADICTED", []string{cell.ID})
		}
		if cell.State == DecisionUnknown && !cell.Unknown.Valid() {
			return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_PROFILE_UNKNOWN_TUPLE", "PROFILE_UNKNOWN_TUPLE_MALFORMED", []string{cell.ID})
		}
		seenIDs[cell.ID] = true
		seenOrdinals[cell.Ordinal] = true
		incrementLedgerSummary(&counts, cell.State)
	}
	if counts != envelope.Profile.Summary {
		return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_PROFILE_SUMMARY", "PROFILE_SUMMARY_DOES_NOT_MATCH_CELLS", []string{"profile.summary", "cells"})
	}
	seenEvents := make(map[string]bool, len(envelope.OperationalEvents))
	for _, event := range envelope.OperationalEvents {
		if event.ID == "" || seenEvents[event.ID] {
			return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_OPERATIONAL_EVENT_IDENTITY", "OPERATIONAL_EVENT_IDENTITY_CONTRADICTED", []string{event.ID})
		}
		seenEvents[event.ID] = true
		switch event.State {
		case "OPERATIONAL_REFUTED":
			if !event.Historical || event.Reason == "" || event.Stage == "" || event.Step == "" || event.NextOperation == "" || event.BlockedBy == nil {
				return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_OPERATIONAL_HISTORY", "OPERATIONAL_REFUTATION_HISTORY_MALFORMED", []string{event.ID})
			}
		case DecisionUnknown:
			if !operationalUnknown(event).Valid() {
				return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_OPERATIONAL_UNKNOWN", "OPERATIONAL_UNKNOWN_TUPLE_MALFORMED", []string{event.ID})
			}
		case DecisionClosed:
		default:
			return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_OPERATIONAL_EVENT_STATE", "OPERATIONAL_EVENT_STATE_UNSUPPORTED", []string{event.ID})
		}
	}
	return adapterValidation{}
}

func validationFromFailure(failure AdapterFailure) adapterValidation {
	kind := strings.ToUpper(failure.Kind)
	if kind == "MISSING" || kind == "STALE" || kind == "AMBIGUOUS" {
		stage := failure.Stage
		if stage == "" {
			stage = "IMMUTABLE_LEDGER_ADAPTER"
		}
		step := failure.Step
		if step == "" {
			step = "READ_RELEASED_ASSET"
		}
		reason := failure.Reason
		if reason == "" {
			reason = "ADAPTER_INPUT_" + kind
		}
		unknownClass := failure.UnknownClass
		if unknownClass == "" {
			unknownClass = kind
		}
		next := failure.NextOperation
		if next == "" {
			next = "RESTORE_IMMUTABLE_LEDGER_INPUT"
		}
		blockedBy := failure.BlockedBy
		if blockedBy == nil {
			blockedBy = []string{}
		}
		return adapterUnknown(stage, step, reason, unknownClass, next, blockedBy)
	}
	return adapterRefuted("LEDGER_SCHEMA_CONTRADICTION", "VERIFY_ADAPTER_FAILURE_KIND", "ADAPTER_FAILURE_KIND_UNSUPPORTED", []string{"failure.kind"})
}

func missingReleaseMetadata(release ReleaseObservation, tag TagObservation, asset ReleaseAsset) bool {
	return release.Repository == "" || release.ReleaseID == 0 || release.Tag == "" || release.TagObjectSHA == "" || release.TargetCommitSHA == "" || tag.Name == "" || tag.ObjectSHA == "" || tag.TargetCommitSHA == "" || asset.ID == 0 || asset.Name == "" || asset.SizeBytes == 0 || asset.Digest == "" || len(release.Assets) == 0
}

func matchingReleaseAsset(assets []ReleaseAsset, selected ReleaseAsset) bool {
	for _, asset := range assets {
		if asset.ID == selected.ID && asset.Name == selected.Name && asset.SizeBytes == selected.SizeBytes && asset.Digest == selected.Digest {
			return true
		}
	}
	return false
}

func adapterUnknown(stage, step, reason, class, next string, blockedBy []string) adapterValidation {
	return adapterValidation{State: DecisionUnknown, Reason: reason, Unknown: Unknown{Stage: stage, Step: step, Reason: reason, UnknownClass: class, NextOperation: next, BlockedBy: append([]string{}, blockedBy...)}}
}

func adapterRefuted(stage, step, reason string, blockedBy []string) adapterValidation {
	return adapterValidation{State: DecisionRefuted, Reason: reason, Unknown: Unknown{Stage: stage, Step: step, Reason: reason, UnknownClass: "CONTRADICTION", NextOperation: "REJECT_CONTRADICTED_IMMUTABLE_INPUT", BlockedBy: append([]string{}, blockedBy...)}}
}

func adapterFailureInput(input Input, validation adapterValidation) Input {
	activity := Activity{
		ID:         adapterActivityID,
		ClaimID:    "claim:" + adapterActivityID,
		State:      validation.State,
		Actionable: validation.State == DecisionUnknown,
		Historical: false,
		BlockedBy:  []string{},
	}
	if validation.State == DecisionUnknown {
		activity.Unknown = &validation.Unknown
	} else {
		activity.Actionable = false
	}
	input.Claims = []Claim{{ID: activity.ClaimID, ActivityID: activity.ID, State: activity.State, Historical: false, Immutable: true}}
	input.Activities = []Activity{activity}
	input.Edges = nil
	input.ImmutableLedger.CellCount = 0
	return input
}

func sortedLedgerCells(cells []LedgerCell) []LedgerCell {
	result := append([]LedgerCell(nil), cells...)
	sort.Slice(result, func(i, j int) bool {
		if result[i].Ordinal != result[j].Ordinal {
			return result[i].Ordinal < result[j].Ordinal
		}
		return result[i].ID < result[j].ID
	})
	return result
}

func sortedOperationalEvents(events []OperationalEvent) []OperationalEvent {
	result := append([]OperationalEvent(nil), events...)
	sort.Slice(result, func(i, j int) bool { return result[i].ID < result[j].ID })
	return result
}

func activityFromCell(cell LedgerCell, known map[string]bool) Activity {
	activity := Activity{ID: cell.ID, ClaimID: "claim:" + cell.ID, State: cell.State, Actionable: cell.State == DecisionUnknown, Historical: cell.State == DecisionRefuted, BlockedBy: []string{}}
	if cell.State == DecisionClosed {
		activity.Evidence = &Evidence{Complete: true}
	}
	if cell.Unknown != nil {
		unknown := normalizedUnknown(cell.Unknown)
		activity.Unknown = &unknown
		for _, blocker := range unknown.BlockedBy {
			if known[blocker] {
				activity.BlockedBy = append(activity.BlockedBy, blocker)
			}
		}
	}
	return activity
}

func activityFromOperationalEvent(event OperationalEvent, known map[string]bool) Activity {
	activity := Activity{ID: event.ID, ClaimID: "claim:" + event.ID, State: event.State, Actionable: event.State == DecisionUnknown, Historical: event.Historical, BlockedBy: []string{}}
	if event.State == DecisionClosed {
		activity.Evidence = &Evidence{Complete: true}
	}
	if event.State == DecisionUnknown {
		unknown := operationalUnknown(event)
		activity.Unknown = &unknown
		for _, blocker := range event.BlockedBy {
			if known[blocker] {
				activity.BlockedBy = append(activity.BlockedBy, blocker)
			}
		}
	}
	return activity
}

func operationalUnknown(event OperationalEvent) Unknown {
	return Unknown{Stage: event.Stage, Step: event.Step, Reason: event.Reason, UnknownClass: event.UnknownClass, NextOperation: event.NextOperation, BlockedBy: append([]string{}, event.BlockedBy...)}
}

func edgesFromActivities(activities []Activity) []Edge {
	known := make(map[string]bool, len(activities))
	for _, activity := range activities {
		known[activity.ID] = true
	}
	edges := make([]Edge, 0)
	for _, activity := range activities {
		for _, blocker := range activity.BlockedBy {
			if known[blocker] {
				edges = append(edges, Edge{ID: "edge:" + activity.ID + ":blocked-by:" + blocker, From: activity.ID, To: blocker, Relation: "BLOCKED_BY", Historical: false, Immutable: true})
			}
		}
	}
	sort.Slice(edges, func(i, j int) bool { return edges[i].ID < edges[j].ID })
	return edges
}

func incrementLedgerSummary(summary *LedgerSummary, state string) {
	summary.Total++
	switch state {
	case DecisionClosed:
		summary.Closed++
	case DecisionUnknown:
		summary.Unknown++
	case DecisionRefuted:
		summary.Refuted++
	}
}

func cellStateIDs(cells []LedgerCell) ([]string, []string) {
	unknown := make([]string, 0)
	refuted := make([]string, 0)
	for _, cell := range cells {
		switch cell.State {
		case DecisionUnknown:
			unknown = append(unknown, cell.ID)
		case DecisionRefuted:
			refuted = append(refuted, cell.ID)
		}
	}
	return sortedStrings(unknown), sortedStrings(refuted)
}

func countOperationalRefuted(events []OperationalEvent) int {
	count := 0
	for _, event := range events {
		if event.State == "OPERATIONAL_REFUTED" {
			count++
		}
	}
	return count
}

func inputStatusFromActivities(activities []Activity) string {
	for _, activity := range activities {
		if activity.State == DecisionRefuted && !activity.Historical {
			return DecisionRefuted
		}
	}
	for _, activity := range activities {
		if activity.State == DecisionUnknown && !activity.Historical {
			return DecisionUnknown
		}
	}
	return DecisionClosed
}

func inputStatusFor(input Input, activities map[string]Activity, issues []graphIssue) string {
	if len(issues) > 0 {
		return DecisionRefuted
	}
	if input.ImmutableLedger != nil && input.ImmutableLedger.InputStatus != "" {
		return input.ImmutableLedger.InputStatus
	}
	values := make([]Activity, 0, len(activities))
	for _, activity := range activities {
		values = append(values, activity)
	}
	return inputStatusFromActivities(values)
}

func ledgerOrigin(input Input, subjectID string) string {
	if input.ImmutableLedger == nil {
		return "claim-activity-graph"
	}
	for _, event := range input.OperationalEvents {
		if event.ID == subjectID {
			return "operational_event"
		}
	}
	return "ledger_cell"
}

func ledgerMetadataTrace(input Input, inputStatus string) []TraceEvent {
	if input.ImmutableLedger == nil {
		return nil
	}
	metadata := input.ImmutableLedger
	trace := []TraceEvent{
		{Kind: "RELEASE_IDENTITY_OBSERVED", SubjectID: fmt.Sprintf("release:%s#%d", metadata.Release.Repository, metadata.Release.ReleaseID), State: DecisionClosed, Class: "IMMUTABLE_RELEASE_IDENTITY", Origin: "immutable-ledger-adapter", BlockingIDs: []string{metadata.Tag.ObjectSHA, metadata.Tag.TargetCommitSHA, metadata.ReleasedAsset.Digest}},
		{Kind: "PROFILE_PARSED", SubjectID: metadata.Profile.ProfileID, State: inputStatus, Class: "INPUT_PROFILE", Origin: "immutable-ledger-adapter", Reason: "PROFILE_AND_CELLS_PARSED_FROM_RELEASED_ASSET"},
	}
	for _, event := range sortedOperationalEvents(input.OperationalEvents) {
		if event.State != "OPERATIONAL_REFUTED" {
			continue
		}
		trace = append(trace, TraceEvent{Kind: "OPERATIONAL_REFUTATION_OBSERVED", SubjectID: event.ID, State: event.State, Historical: event.Historical, BlockingIDs: append([]string{}, event.BlockedBy...), Reason: event.Reason, Class: "HISTORICAL_REFUTATION", Origin: "operational_event"})
	}
	return trace
}

func operationalRefutedCount(input Input) int {
	if input.ImmutableLedger != nil && input.ImmutableLedger.OperationalRefutedCount != 0 {
		return input.ImmutableLedger.OperationalRefutedCount
	}
	return countOperationalRefuted(input.OperationalEvents)
}
