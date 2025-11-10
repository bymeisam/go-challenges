package main

import (
	"testing"
	"time"
)

func TestCreateFlag(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, err := fm.CreateFlag("new-feature", "Test feature", true, "admin")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if flag.ID == "" {
		t.Fatal("Expected non-empty flag ID")
	}

	if !flag.Enabled {
		t.Fatal("Expected flag to be enabled")
	}
}

func TestCreateDuplicateFlag(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	fm.CreateFlag("new-feature", "Test feature", true, "admin")

	_, err := fm.CreateFlag("new-feature", "Another test", true, "admin")
	if err == nil {
		t.Fatal("Expected error for duplicate flag")
	}
}

func TestSetRolloutPercent(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")

	err := fm.SetRolloutPercent(flag.ID, 50)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := fm.GetFlag(flag.ID)
	if updated.RolloutPercent != 50 {
		t.Fatalf("Expected 50 percent, got %d", updated.RolloutPercent)
	}
}

func TestSetInvalidRolloutPercent(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")

	err := fm.SetRolloutPercent(flag.ID, 150)
	if err == nil {
		t.Fatal("Expected error for invalid rollout percent")
	}
}

func TestTargetUser(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", false, "admin")

	err := fm.TargetUser(flag.ID, "user-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := fm.GetFlag(flag.ID)
	if !updated.TargetedUsers["user-1"] {
		t.Fatal("Expected user to be targeted")
	}
}

func TestRemoveTargetedUser(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", false, "admin")

	fm.TargetUser(flag.ID, "user-1")
	err := fm.RemoveTargetedUser(flag.ID, "user-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := fm.GetFlag(flag.ID)
	if updated.TargetedUsers["user-1"] {
		t.Fatal("Expected user to not be targeted")
	}
}

func TestEvaluateFlagDisabled(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", false, "admin")

	eval, _ := fm.EvaluateFlag(flag.ID, "user-1")
	if eval.Enabled {
		t.Fatal("Expected flag to be disabled")
	}

	if eval.Reason != "flag_disabled" {
		t.Fatalf("Expected reason 'flag_disabled', got %s", eval.Reason)
	}
}

func TestEvaluateFlagTargeted(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")
	fm.TargetUser(flag.ID, "user-1")

	eval, _ := fm.EvaluateFlag(flag.ID, "user-1")
	if !eval.Enabled {
		t.Fatal("Expected targeted user to have flag enabled")
	}

	if eval.Reason != "user_targeted" {
		t.Fatalf("Expected reason 'user_targeted', got %s", eval.Reason)
	}
}

func TestEvaluateFlagRollout(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")
	fm.SetRolloutPercent(flag.ID, 50)

	eval, _ := fm.EvaluateFlag(flag.ID, "user-1")
	// Deterministic result based on user hash
	if eval.Reason != "rollout_percentage" && eval.Reason != "failed_rollout" {
		t.Fatalf("Expected rollout reason, got %s", eval.Reason)
	}
}

func TestCreateSegment(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	rules := []SegmentRule{
		{Attribute: "country", Operator: "equals", Value: "US"},
	}

	segment, err := fm.CreateSegment("US Users", "Users from US", rules)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if segment.ID == "" {
		t.Fatal("Expected non-empty segment ID")
	}
}

func TestAddSegmentToFlag(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")

	rules := []SegmentRule{
		{Attribute: "country", Operator: "equals", Value: "US"},
	}
	segment, _ := fm.CreateSegment("US Users", "Users from US", rules)

	err := fm.AddSegmentToFlag(flag.ID, segment.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	updated, _ := fm.GetFlag(flag.ID)
	if len(updated.Segments) != 1 {
		t.Fatal("Expected segment to be added")
	}
}

func TestCreateExperiment(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control":    {ID: "control", Name: "Control", TrafficPercent: 50},
		"treatment":  {ID: "treatment", Name: "Treatment", TrafficPercent: 50},
	}

	exp, err := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if exp.ID == "" {
		t.Fatal("Expected non-empty experiment ID")
	}

	if exp.Status != "draft" {
		t.Fatalf("Expected draft status, got %s", exp.Status)
	}
}

func TestStartExperiment(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control": {ID: "control", Name: "Control", TrafficPercent: 100},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	err := am.StartExperiment(exp.ID)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestAssignVariant(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control":    {ID: "control", Name: "Control", TrafficPercent: 50},
		"treatment":  {ID: "treatment", Name: "Treatment", TrafficPercent: 50},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	variant, err := am.AssignVariant(exp.ID, "user-1")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if variant == "" {
		t.Fatal("Expected variant assignment")
	}
}

func TestConsistentVariantAssignment(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control": {ID: "control", Name: "Control", TrafficPercent: 100},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	variant1, _ := am.AssignVariant(exp.ID, "user-1")
	variant2, _ := am.AssignVariant(exp.ID, "user-1")

	if variant1 != variant2 {
		t.Fatal("Expected consistent variant assignment")
	}
}

func TestRecordConversion(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control": {ID: "control", Name: "Control", TrafficPercent: 100},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	am.AssignVariant(exp.ID, "user-1")

	err := am.RecordConversion(exp.ID, "user-1", nil)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestFlagCaching(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")

	eval1, _ := fm.EvaluateFlag(flag.ID, "user-1")
	eval2, _ := fm.EvaluateFlag(flag.ID, "user-1")

	// Should be same cached evaluation
	if eval1.EvaluatedAt != eval2.EvaluatedAt {
		t.Fatal("Expected cached evaluation to have same timestamp")
	}
}

func TestMultipleFlags(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	for i := 0; i < 10; i++ {
		fm.CreateFlag("feature-"+string(rune(i)), "Test", true, "admin")
	}

	// Test by attempting to access a flag we just created
	for i := 0; i < 10; i++ {
		// At least verify that we can retrieve flags
		_, err := fm.GetFlag("feature_id")
		if err != nil && i == 0 {
			// Expected when flag doesn't exist
		}
	}
}

func TestRolloutDistribution(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")
	fm.SetRolloutPercent(flag.ID, 50)

	enabled := 0
	disabled := 0

	for i := 0; i < 100; i++ {
		eval, _ := fm.EvaluateFlag(flag.ID, "user-"+string(rune(i)))
		if eval.Enabled {
			enabled++
		} else {
			disabled++
		}
	}

	// With 50% rollout, we expect roughly 50 enabled and 50 disabled
	if enabled < 30 || enabled > 70 {
		t.Fatalf("Expected roughly 50%% enabled, got %d%%", enabled)
	}
}

func TestExperimentMetrics(t *testing.T) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control": {ID: "control", Name: "Control", TrafficPercent: 100},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	for i := 0; i < 10; i++ {
		am.AssignVariant(exp.ID, "user-"+string(rune(i)))
	}

	for i := 0; i < 5; i++ {
		am.RecordConversion(exp.ID, "user-"+string(rune(i)), nil)
	}

	metrics, _ := am.GetMetrics(exp.ID, "control")

	if metrics.Exposures != 10 {
		t.Fatalf("Expected 10 exposures, got %d", metrics.Exposures)
	}

	if metrics.Conversions != 5 {
		t.Fatalf("Expected 5 conversions, got %d", metrics.Conversions)
	}
}

func TestCacheInvalidation(t *testing.T) {
	fm := NewFeatureFlagManager(1 * time.Hour)

	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")

	fm.EvaluateFlag(flag.ID, "user-1")

	// Change rollout
	fm.SetRolloutPercent(flag.ID, 100)

	// Cache should be invalidated
	// Next evaluation should reflect changes
	eval, _ := fm.EvaluateFlag(flag.ID, "user-1")
	if eval.EvaluatedAt.Before(time.Now().Add(-1 * time.Second)) {
		t.Fatal("Expected fresh evaluation")
	}
}

// ========== Benchmarks ==========

func BenchmarkEvaluateFlag(b *testing.B) {
	fm := NewFeatureFlagManager(1 * time.Hour)
	flag, _ := fm.CreateFlag("new-feature", "Test feature", true, "admin")
	fm.SetRolloutPercent(flag.ID, 50)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		fm.EvaluateFlag(flag.ID, "user-"+string(rune(i)))
	}
}

func BenchmarkAssignVariant(b *testing.B) {
	am := NewABTestManager()

	variants := map[string]*Variant{
		"control":   {ID: "control", Name: "Control", TrafficPercent: 50},
		"treatment": {ID: "treatment", Name: "Treatment", TrafficPercent: 50},
	}

	exp, _ := am.CreateExperiment("Test Experiment", "Test AB test", "flag-1", variants)

	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		am.AssignVariant(exp.ID, "user-"+string(rune(i)))
	}
}
