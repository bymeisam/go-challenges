package main

import (
	"crypto/md5"
	"errors"
	"fmt"
	"sync"
	"time"
)

// ========== Feature Flag Models ==========

type FeatureFlag struct {
	ID              string            `json:"id"`
	Name            string            `json:"name"`
	Description     string            `json:"description"`
	Enabled         bool              `json:"enabled"`
	RolloutPercent  int               `json:"rollout_percent"`
	TargetedUsers   map[string]bool   `json:"targeted_users"`
	Segments        []string          `json:"segments"`
	CreatedAt       time.Time         `json:"created_at"`
	UpdatedAt       time.Time         `json:"updated_at"`
	CreatedBy       string            `json:"created_by"`
}

type FlagEvaluation struct {
	FlagID        string            `json:"flag_id"`
	UserID        string            `json:"user_id"`
	Enabled       bool              `json:"enabled"`
	Reason        string            `json:"reason"`
	EvaluatedAt   time.Time         `json:"evaluated_at"`
	VariantID     string            `json:"variant_id,omitempty"`
	RuleMatched   string            `json:"rule_matched,omitempty"`
}

type UserSegment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Rules           []SegmentRule          `json:"rules"`
	UserCount       int                    `json:"user_count"`
	CreatedAt       time.Time              `json:"created_at"`
}

type SegmentRule struct {
	Attribute string      `json:"attribute"`
	Operator  string      `json:"operator"` // equals, contains, greaterThan, etc
	Value     interface{} `json:"value"`
}

// ========== A/B Testing Models ==========

type Experiment struct {
	ID              string                 `json:"id"`
	Name            string                 `json:"name"`
	Description     string                 `json:"description"`
	Status          string                 `json:"status"` // draft, running, paused, completed
	StartTime       time.Time              `json:"start_time"`
	EndTime         time.Time              `json:"end_time"`
	FlagID          string                 `json:"flag_id"`
	Variants        map[string]*Variant    `json:"variants"`
	TargetSegments  []string               `json:"target_segments"`
	TrafficPercent  int                    `json:"traffic_percent"`
	PrimaryMetric   string                 `json:"primary_metric"`
	CreatedAt       time.Time              `json:"created_at"`
}

type Variant struct {
	ID              string      `json:"id"`
	Name            string      `json:"name"`
	TrafficPercent  int         `json:"traffic_percent"`
	Config          interface{} `json:"config"`
}

type ExperimentEvent struct {
	EventID         string                 `json:"event_id"`
	ExperimentID    string                 `json:"experiment_id"`
	UserID          string                 `json:"user_id"`
	VariantID       string                 `json:"variant_id"`
	EventType       string                 `json:"event_type"` // exposure, conversion, etc
	Timestamp       time.Time              `json:"timestamp"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// ========== Analytics Models ==========

type AnalyticsEvent struct {
	EventID         string                 `json:"event_id"`
	UserID          string                 `json:"user_id"`
	EventType       string                 `json:"event_type"`
	FeatureFlagID   string                 `json:"feature_flag_id,omitempty"`
	ExperimentID    string                 `json:"experiment_id,omitempty"`
	VariantID       string                 `json:"variant_id,omitempty"`
	Timestamp       time.Time              `json:"timestamp"`
	Properties      map[string]interface{} `json:"properties"`
}

type ExperimentMetrics struct {
	ExperimentID    string                 `json:"experiment_id"`
	VariantID       string                 `json:"variant_id"`
	Exposures       int64                  `json:"exposures"`
	Conversions     int64                  `json:"conversions"`
	ConversionRate  float64                `json:"conversion_rate"`
	ConfidenceScore float64                `json:"confidence_score"`
}

// ========== Feature Flag Manager ==========

type FeatureFlagManager struct {
	flags           map[string]*FeatureFlag
	flagsMu         sync.RWMutex
	segments        map[string]*UserSegment
	segmentsMu      sync.RWMutex
	evaluationCache map[string]*FlagEvaluation
	cacheMu         sync.RWMutex
	cacheTTL        time.Duration
}

// NewFeatureFlagManager creates a new feature flag manager
func NewFeatureFlagManager(cacheTTL time.Duration) *FeatureFlagManager {
	return &FeatureFlagManager{
		flags:           make(map[string]*FeatureFlag),
		segments:        make(map[string]*UserSegment),
		evaluationCache: make(map[string]*FlagEvaluation),
		cacheTTL:        cacheTTL,
	}
}

// CreateFlag creates a new feature flag
func (fm *FeatureFlagManager) CreateFlag(name, description string, enabled bool, createdBy string) (*FeatureFlag, error) {
	fm.flagsMu.Lock()
	defer fm.flagsMu.Unlock()

	// Check for duplicates
	for _, flag := range fm.flags {
		if flag.Name == name {
			return nil, errors.New("flag name already exists")
		}
	}

	flag := &FeatureFlag{
		ID:             generateFlagID(),
		Name:           name,
		Description:    description,
		Enabled:        enabled,
		RolloutPercent: 0,
		TargetedUsers:  make(map[string]bool),
		Segments:       []string{},
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		CreatedBy:      createdBy,
	}

	fm.flags[flag.ID] = flag
	return flag, nil
}

// GetFlag retrieves a flag by ID
func (fm *FeatureFlagManager) GetFlag(flagID string) (*FeatureFlag, error) {
	fm.flagsMu.RLock()
	defer fm.flagsMu.RUnlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return nil, errors.New("flag not found")
	}

	return flag, nil
}

// SetRolloutPercent sets the rollout percentage for a flag
func (fm *FeatureFlagManager) SetRolloutPercent(flagID string, percent int) error {
	if percent < 0 || percent > 100 {
		return errors.New("rollout percent must be between 0 and 100")
	}

	fm.flagsMu.Lock()
	defer fm.flagsMu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return errors.New("flag not found")
	}

	flag.RolloutPercent = percent
	flag.UpdatedAt = time.Now()

	fm.invalidateCache(flagID)
	return nil
}

// TargetUser adds a user to targeted users
func (fm *FeatureFlagManager) TargetUser(flagID, userID string) error {
	fm.flagsMu.Lock()
	defer fm.flagsMu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return errors.New("flag not found")
	}

	flag.TargetedUsers[userID] = true
	flag.UpdatedAt = time.Now()

	fm.invalidateCache(flagID)
	return nil
}

// RemoveTargetedUser removes a user from targeted users
func (fm *FeatureFlagManager) RemoveTargetedUser(flagID, userID string) error {
	fm.flagsMu.Lock()
	defer fm.flagsMu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return errors.New("flag not found")
	}

	delete(flag.TargetedUsers, userID)
	flag.UpdatedAt = time.Now()

	fm.invalidateCache(flagID)
	return nil
}

// EvaluateFlag evaluates a flag for a user
func (fm *FeatureFlagManager) EvaluateFlag(flagID, userID string) (*FlagEvaluation, error) {
	// Check cache first
	cacheKey := fmt.Sprintf("%s:%s", flagID, userID)
	fm.cacheMu.RLock()
	if cached, exists := fm.evaluationCache[cacheKey]; exists {
		fm.cacheMu.RUnlock()
		return cached, nil
	}
	fm.cacheMu.RUnlock()

	fm.flagsMu.RLock()
	flag, exists := fm.flags[flagID]
	fm.flagsMu.RUnlock()

	if !exists {
		return nil, errors.New("flag not found")
	}

	evaluation := &FlagEvaluation{
		FlagID:      flagID,
		UserID:      userID,
		EvaluatedAt: time.Now(),
	}

	// Check if flag is disabled
	if !flag.Enabled {
		evaluation.Enabled = false
		evaluation.Reason = "flag_disabled"
		return evaluation, nil
	}

	// Check targeted users
	if flag.TargetedUsers[userID] {
		evaluation.Enabled = true
		evaluation.Reason = "user_targeted"
		fm.cacheEvaluation(cacheKey, evaluation)
		return evaluation, nil
	}

	// Check rollout percentage
	if hashUserID(userID, flagID)%100 < uint32(flag.RolloutPercent) {
		evaluation.Enabled = true
		evaluation.Reason = "rollout_percentage"
	} else {
		evaluation.Enabled = false
		evaluation.Reason = "failed_rollout"
	}

	fm.cacheEvaluation(cacheKey, evaluation)
	return evaluation, nil
}

// ========== Segment Management ==========

// CreateSegment creates a user segment
func (fm *FeatureFlagManager) CreateSegment(name, description string, rules []SegmentRule) (*UserSegment, error) {
	fm.segmentsMu.Lock()
	defer fm.segmentsMu.Unlock()

	segment := &UserSegment{
		ID:          generateSegmentID(),
		Name:        name,
		Description: description,
		Rules:       rules,
		CreatedAt:   time.Now(),
	}

	fm.segments[segment.ID] = segment
	return segment, nil
}

// AddSegmentToFlag adds a segment to a flag
func (fm *FeatureFlagManager) AddSegmentToFlag(flagID, segmentID string) error {
	fm.flagsMu.Lock()
	defer fm.flagsMu.Unlock()

	flag, exists := fm.flags[flagID]
	if !exists {
		return errors.New("flag not found")
	}

	// Check if segment exists
	fm.segmentsMu.RLock()
	_, segExists := fm.segments[segmentID]
	fm.segmentsMu.RUnlock()

	if !segExists {
		return errors.New("segment not found")
	}

	for _, seg := range flag.Segments {
		if seg == segmentID {
			return errors.New("segment already added to flag")
		}
	}

	flag.Segments = append(flag.Segments, segmentID)
	flag.UpdatedAt = time.Now()

	fm.invalidateCache(flagID)
	return nil
}

// ========== A/B Testing Manager ==========

type ABTestManager struct {
	experiments      map[string]*Experiment
	experimentsMu    sync.RWMutex
	analyticsEvents  []*AnalyticsEvent
	eventsMu         sync.RWMutex
	variantAssign    map[string]map[string]string // experiment -> user -> variant
	variantMu        sync.RWMutex
	metrics          map[string]*ExperimentMetrics
	metricsMu        sync.RWMutex
}

// NewABTestManager creates a new A/B test manager
func NewABTestManager() *ABTestManager {
	return &ABTestManager{
		experiments:   make(map[string]*Experiment),
		analyticsEvents: []*AnalyticsEvent{},
		variantAssign: make(map[string]map[string]string),
		metrics:       make(map[string]*ExperimentMetrics),
	}
}

// CreateExperiment creates a new experiment
func (am *ABTestManager) CreateExperiment(name, description, flagID string, variants map[string]*Variant) (*Experiment, error) {
	am.experimentsMu.Lock()
	defer am.experimentsMu.Unlock()

	experiment := &Experiment{
		ID:             generateExperimentID(),
		Name:           name,
		Description:    description,
		Status:         "draft",
		FlagID:         flagID,
		Variants:       variants,
		TargetSegments: []string{},
		TrafficPercent: 100,
		CreatedAt:      time.Now(),
	}

	am.experiments[experiment.ID] = experiment
	am.variantAssign[experiment.ID] = make(map[string]string)

	return experiment, nil
}

// StartExperiment starts an experiment
func (am *ABTestManager) StartExperiment(experimentID string) error {
	am.experimentsMu.Lock()
	defer am.experimentsMu.Unlock()

	exp, exists := am.experiments[experimentID]
	if !exists {
		return errors.New("experiment not found")
	}

	if exp.Status != "draft" {
		return errors.New("experiment must be in draft status")
	}

	exp.Status = "running"
	exp.StartTime = time.Now()

	return nil
}

// AssignVariant assigns a variant to a user for an experiment
func (am *ABTestManager) AssignVariant(experimentID, userID string) (string, error) {
	am.variantMu.Lock()
	defer am.variantMu.Unlock()

	// Check if already assigned
	if variantID, exists := am.variantAssign[experimentID][userID]; exists {
		return variantID, nil
	}

	am.experimentsMu.RLock()
	exp, exists := am.experiments[experimentID]
	am.experimentsMu.RUnlock()

	if !exists {
		return "", errors.New("experiment not found")
	}

	// Assign variant based on hash
	variantID := selectVariant(userID, experimentID, exp.Variants)

	am.variantAssign[experimentID][userID] = variantID

	// Record event
	am.recordExperimentEvent(experimentID, userID, variantID, "exposure")

	return variantID, nil
}

func selectVariant(userID, experimentID string, variants map[string]*Variant) string {
	hash := hashUserID(userID, experimentID) % 100
	cumulativePercent := 0

	for _, variant := range variants {
		cumulativePercent += variant.TrafficPercent
		if int(hash) < cumulativePercent {
			return variant.ID
		}
	}

	// Fallback to first variant
	for _, variant := range variants {
		return variant.ID
	}
	return ""
}

// RecordConversion records a conversion for a user
func (am *ABTestManager) RecordConversion(experimentID, userID string, metadata map[string]interface{}) error {
	am.variantMu.RLock()
	variantID, exists := am.variantAssign[experimentID][userID]
	am.variantMu.RUnlock()

	if !exists {
		return errors.New("user not assigned to experiment")
	}

	am.recordExperimentEvent(experimentID, userID, variantID, "conversion")

	am.metricsMu.Lock()
	key := fmt.Sprintf("%s:%s", experimentID, variantID)
	if metric, exists := am.metrics[key]; exists {
		metric.Conversions++
	}
	am.metricsMu.Unlock()

	return nil
}

// GetMetrics gets metrics for an experiment variant
func (am *ABTestManager) GetMetrics(experimentID, variantID string) (*ExperimentMetrics, error) {
	am.metricsMu.RLock()
	defer am.metricsMu.RUnlock()

	key := fmt.Sprintf("%s:%s", experimentID, variantID)
	metrics, exists := am.metrics[key]
	if !exists {
		return nil, errors.New("metrics not found")
	}

	return metrics, nil
}

func (am *ABTestManager) recordExperimentEvent(experimentID, userID, variantID, eventType string) {
	event := &AnalyticsEvent{
		EventID:      generateEventID(),
		UserID:       userID,
		EventType:    eventType,
		ExperimentID: experimentID,
		VariantID:    variantID,
		Timestamp:    time.Now(),
	}

	am.eventsMu.Lock()
	am.analyticsEvents = append(am.analyticsEvents, event)
	am.eventsMu.Unlock()

	// Update metrics
	am.metricsMu.Lock()
	key := fmt.Sprintf("%s:%s", experimentID, variantID)
	if _, exists := am.metrics[key]; !exists {
		am.metrics[key] = &ExperimentMetrics{
			ExperimentID: experimentID,
			VariantID:    variantID,
		}
	}
	if eventType == "exposure" {
		am.metrics[key].Exposures++
	}
	am.metricsMu.Unlock()
}

// ========== Helper Functions ==========

func (fm *FeatureFlagManager) cacheEvaluation(key string, evaluation *FlagEvaluation) {
	fm.cacheMu.Lock()
	defer fm.cacheMu.Unlock()
	fm.evaluationCache[key] = evaluation
}

func (fm *FeatureFlagManager) invalidateCache(flagID string) {
	fm.cacheMu.Lock()
	defer fm.cacheMu.Unlock()

	// Simple invalidation: remove entries for this flag
	for key := range fm.evaluationCache {
		if len(key) > 0 && key[:len(flagID)] == flagID {
			delete(fm.evaluationCache, key)
		}
	}
}

func hashUserID(userID, flagID string) uint32 {
	hash := md5.Sum([]byte(userID + flagID))
	result := uint32(hash[0])<<24 | uint32(hash[1])<<16 | uint32(hash[2])<<8 | uint32(hash[3])
	return result
}

func generateFlagID() string {
	return fmt.Sprintf("flag_%d", time.Now().UnixNano())
}

func generateSegmentID() string {
	return fmt.Sprintf("segment_%d", time.Now().UnixNano())
}

func generateExperimentID() string {
	return fmt.Sprintf("exp_%d", time.Now().UnixNano())
}

func generateEventID() string {
	return fmt.Sprintf("event_%d", time.Now().UnixNano())
}

// GetFlagEvaluationLog returns recent evaluations
func (fm *FeatureFlagManager) GetFlagEvaluationLog(flagID string, limit int) []*FlagEvaluation {
	fm.cacheMu.RLock()
	defer fm.cacheMu.RUnlock()

	var evals []*FlagEvaluation
	count := 0
	for _, eval := range fm.evaluationCache {
		if eval.FlagID == flagID {
			evals = append(evals, eval)
			count++
			if count >= limit {
				break
			}
		}
	}

	return evals
}

func main() {
	// Example feature flag
	fm := NewFeatureFlagManager(1 * time.Hour)
	fm.CreateFlag("new-feature", "Test feature", true, "admin")
}
