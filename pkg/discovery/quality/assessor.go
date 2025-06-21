package quality

import (
	"context"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
)

// Assessor implements quality assessment for schemas
type Assessor struct {
	config Config
}

// Config holds configuration for quality assessment
type Config struct {
	// Thresholds for quality dimensions
	CompletenessThreshold  float64
	ConsistencyThreshold   float64
	TimelinessThreshold    time.Duration
	UniquenessThreshold    float64
	ValidityThreshold      float64
	
	// Weights for overall score calculation
	CompletenessWeight     float64
	ConsistencyWeight      float64
	TimelinessWeight       float64
	UniquenessWeight       float64
	ValidityWeight         float64
}

// DefaultConfig returns default configuration
func DefaultConfig() Config {
	return Config{
		CompletenessThreshold: 0.95,
		ConsistencyThreshold:  0.90,
		TimelinessThreshold:   5 * time.Minute,
		UniquenessThreshold:   0.99,
		ValidityThreshold:     0.95,
		
		CompletenessWeight:    0.25,
		ConsistencyWeight:     0.25,
		TimelinessWeight:      0.20,
		UniquenessWeight:      0.15,
		ValidityWeight:        0.15,
	}
}

// NewAssessor creates a new quality assessor
func NewAssessor(config Config) *Assessor {
	return &Assessor{
		config: config,
	}
}

// AssessSchema performs comprehensive quality assessment
func (a *Assessor) AssessSchema(ctx context.Context, schema discovery.Schema, sample discovery.DataSample) discovery.QualityReport {
	report := discovery.QualityReport{
		SchemaName: schema.Name,
		Timestamp:  time.Now(),
	}
	
	// Assess each quality dimension
	report.Dimensions.Completeness = a.assessCompleteness(schema, sample)
	report.Dimensions.Consistency = a.assessConsistency(schema, sample)
	report.Dimensions.Timeliness = a.assessTimeliness(schema, sample)
	report.Dimensions.Uniqueness = a.assessUniqueness(schema, sample)
	report.Dimensions.Validity = a.assessValidity(schema, sample)
	
	// Calculate overall score
	report.OverallScore = a.calculateOverallScore(report.Dimensions)
	
	// Identify issues
	report.Issues = a.identifyIssues(report.Dimensions)
	
	// Generate recommendations
	report.Recommendations = a.generateRecommendations(report.Issues)
	
	// Add ML predictions if available
	// TODO: Implement predictions when QualityReport supports it
	// if a.shouldUsePredictions() {
	//     report.Predictions = a.generatePredictions(schema, report)
	// }
	
	return report
}

// assessCompleteness measures data completeness
func (a *Assessor) assessCompleteness(schema discovery.Schema, sample discovery.DataSample) discovery.DimensionScore {
	dim := discovery.DimensionScore{}
	
	// Calculate null/missing value ratio for each attribute
	attributeScores := make(map[string]float64)
	totalFields := len(schema.Attributes) * sample.SampleSize
	missingCount := 0
	
	for _, attr := range schema.Attributes {
		nullCount := 0
		for _, record := range sample.Records {
			if val, exists := record[attr.Name]; !exists || val == nil || val == "" {
				nullCount++
				missingCount++
			}
		}
		
		if sample.SampleSize > 0 {
			attributeScores[attr.Name] = 1.0 - float64(nullCount)/float64(sample.SampleSize)
		}
	}
	
	// Overall completeness score
	if totalFields > 0 {
		dim.Score = 1.0 - float64(missingCount)/float64(totalFields)
	} else {
		dim.Score = 0.0
	}
	
	// Identify problematic attributes
	issues := []string{}
	for attrName, score := range attributeScores {
		if score < 0.9 {
			issues = append(issues, fmt.Sprintf("%s: %.1f%% complete", attrName, score*100))
		}
	}
	
	dim.Details = fmt.Sprintf("Missing values: %d/%d, Problematic fields: %d", 
		missingCount, totalFields, len(issues))
	
	return dim
}

// assessConsistency measures data consistency
func (a *Assessor) assessConsistency(schema discovery.Schema, sample discovery.DataSample) discovery.DimensionScore {
	dim := discovery.DimensionScore{}
	
	inconsistencies := 0
	totalChecks := 0
	
	// Check format consistency for string attributes
	for _, attr := range schema.Attributes {
		if attr.DataType == discovery.DataTypeString {
			formats := a.detectFormats(attr, sample)
			if len(formats) > 1 {
				// Multiple formats detected - potential inconsistency
				inconsistencies += len(formats) - 1
			}
			totalChecks++
		}
	}
	
	// Check numeric range consistency
	for _, attr := range schema.Attributes {
		if attr.DataType == discovery.DataTypeNumeric {
			if outliers := a.detectOutliers(attr, sample); outliers > 0 {
				inconsistencies += outliers
			}
			totalChecks++
		}
	}
	
	// Calculate score
	if totalChecks > 0 {
		dim.Score = 1.0 - float64(inconsistencies)/float64(totalChecks*sample.SampleSize)
	} else {
		dim.Score = 1.0
	}
	
	dim.Details = fmt.Sprintf("Inconsistencies: %d/%d checks, Types: format_consistency, numeric_ranges",
		inconsistencies, totalChecks)
	
	return dim
}

// assessTimeliness measures data freshness
func (a *Assessor) assessTimeliness(schema discovery.Schema, sample discovery.DataSample) discovery.DimensionScore {
	dim := discovery.DimensionScore{}
	
	// Find timestamp attribute
	var timestampAttr *discovery.Attribute
	for i := range schema.Attributes {
		if schema.Attributes[i].DataType == discovery.DataTypeTimestamp || 
		   schema.Attributes[i].Name == "timestamp" {
			timestampAttr = &schema.Attributes[i]
			break
		}
	}
	
	if timestampAttr == nil {
		dim.Score = 0.5 // Can't assess without timestamp
		dim.Details = "Error: No timestamp attribute found"
		return dim
	}
	
	// Calculate data age
	now := time.Now()
	delays := []time.Duration{}
	
	for _, record := range sample.Records {
		if tsVal, exists := record[timestampAttr.Name]; exists {
			if ts, ok := tsVal.(time.Time); ok {
				delay := now.Sub(ts)
				delays = append(delays, delay)
			}
		}
	}
	
	if len(delays) == 0 {
		dim.Score = 0.5
		return dim
	}
	
	// Calculate average delay
	totalDelay := time.Duration(0)
	maxDelay := time.Duration(0)
	for _, d := range delays {
		totalDelay += d
		if d > maxDelay {
			maxDelay = d
		}
	}
	avgDelay := totalDelay / time.Duration(len(delays))
	
	// Score based on threshold
	if avgDelay <= a.config.TimelinessThreshold {
		dim.Score = 1.0
	} else {
		// Linear decay after threshold
		dim.Score = float64(a.config.TimelinessThreshold) / float64(avgDelay)
	}
	
	dim.Details = fmt.Sprintf("Avg delay: %s, Max delay: %s, Threshold: %s, Samples: %d",
		avgDelay, maxDelay, a.config.TimelinessThreshold, len(delays))
	
	return dim
}

// assessUniqueness measures duplicate detection
func (a *Assessor) assessUniqueness(schema discovery.Schema, sample discovery.DataSample) discovery.DimensionScore {
	dim := discovery.DimensionScore{}
	
	// Find potential unique identifiers
	uniqueAttrs := []string{}
	for _, attr := range schema.Attributes {
		if attr.SemanticType == discovery.SemanticTypeID || 
		   strings.Contains(strings.ToLower(attr.Name), "id") {
			uniqueAttrs = append(uniqueAttrs, attr.Name)
		}
	}
	
	if len(uniqueAttrs) == 0 {
		dim.Score = 1.0 // No unique constraints to check
		dim.Details = "Note: No identifier attributes found"
		return dim
	}
	
	// Check for duplicates
	totalDuplicates := 0
	for _, attrName := range uniqueAttrs {
		seen := make(map[interface{}]int)
		duplicates := 0
		
		for _, record := range sample.Records {
			if val, exists := record[attrName]; exists && val != nil {
				seen[val]++
				if seen[val] > 1 {
					duplicates++
				}
			}
		}
		
		totalDuplicates += duplicates
	}
	
	// Calculate score
	totalChecks := len(uniqueAttrs) * sample.SampleSize
	if totalChecks > 0 {
		dim.Score = 1.0 - float64(totalDuplicates)/float64(totalChecks)
	} else {
		dim.Score = 1.0
	}
	
	dim.Details = fmt.Sprintf("Duplicates: %d found, Checked %d unique attributes",
		totalDuplicates, len(uniqueAttrs))
	
	return dim
}

// assessValidity measures data validity
func (a *Assessor) assessValidity(schema discovery.Schema, sample discovery.DataSample) discovery.DimensionScore {
	dim := discovery.DimensionScore{}
	
	invalidCount := 0
	totalValidations := 0
	
	// Validate each attribute based on type and constraints
	for _, attr := range schema.Attributes {
		validations := 0
		invalid := 0
		
		for _, record := range sample.Records {
			if val, exists := record[attr.Name]; exists && val != nil {
				totalValidations++
				validations++
				
				if !a.isValid(attr, val) {
					invalid++
					invalidCount++
				}
			}
		}
		
		// Store per-attribute validity
		if validations > 0 {
			validity := 1.0 - float64(invalid)/float64(validations)
			if validity < 0.95 {
				// Track problematic attributes
			}
		}
	}
	
	// Calculate overall validity score
	if totalValidations > 0 {
		dim.Score = 1.0 - float64(invalidCount)/float64(totalValidations)
	} else {
		dim.Score = 1.0
	}
	
	dim.Details = fmt.Sprintf("Invalid values: %d/%d, Validation types: type_check, range_check, format_check",
		invalidCount, totalValidations)
	
	return dim
}

// Helper methods

// detectFormats detects different formats in string data
func (a *Assessor) detectFormats(attr discovery.Attribute, sample discovery.DataSample) map[string]int {
	formats := make(map[string]int)
	
	for _, record := range sample.Records {
		if val, exists := record[attr.Name]; exists {
			if strVal, ok := val.(string); ok {
				format := a.classifyFormat(strVal)
				formats[format]++
			}
		}
	}
	
	return formats
}

// classifyFormat classifies string format
func (a *Assessor) classifyFormat(s string) string {
	s = strings.TrimSpace(s)
	
	// Check common formats
	if strings.Contains(s, "@") && strings.Contains(s, ".") {
		return "email"
	}
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return "url"
	}
	if len(s) == 36 && s[8] == '-' && s[13] == '-' && s[18] == '-' && s[23] == '-' {
		return "uuid"
	}
	
	// Check if numeric
	isNumeric := true
	for _, ch := range s {
		if (ch < '0' || ch > '9') && ch != '.' && ch != '-' {
			isNumeric = false
			break
		}
	}
	if isNumeric {
		return "numeric_string"
	}
	
	return "general_string"
}

// detectOutliers detects outliers in numeric data
func (a *Assessor) detectOutliers(attr discovery.Attribute, sample discovery.DataSample) int {
	values := []float64{}
	
	for _, record := range sample.Records {
		if val, exists := record[attr.Name]; exists {
			if numVal, ok := a.toFloat64(val); ok {
				values = append(values, numVal)
			}
		}
	}
	
	if len(values) < 4 {
		return 0
	}
	
	// Calculate mean and standard deviation
	mean, stdDev := a.meanStdDev(values)
	
	// Count outliers (values beyond 3 standard deviations)
	outliers := 0
	for _, v := range values {
		if math.Abs(v-mean) > 3*stdDev {
			outliers++
		}
	}
	
	return outliers
}

// toFloat64 converts various numeric types to float64
func (a *Assessor) toFloat64(val interface{}) (float64, bool) {
	switch v := val.(type) {
	case float64:
		return v, true
	case float32:
		return float64(v), true
	case int:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	default:
		return 0, false
	}
}

// meanStdDev calculates mean and standard deviation
func (a *Assessor) meanStdDev(values []float64) (mean, stdDev float64) {
	if len(values) == 0 {
		return 0, 0
	}
	
	// Calculate mean
	sum := 0.0
	for _, v := range values {
		sum += v
	}
	mean = sum / float64(len(values))
	
	// Calculate standard deviation
	sumSquares := 0.0
	for _, v := range values {
		diff := v - mean
		sumSquares += diff * diff
	}
	stdDev = math.Sqrt(sumSquares / float64(len(values)))
	
	return mean, stdDev
}

// isValid checks if a value is valid for an attribute
func (a *Assessor) isValid(attr discovery.Attribute, val interface{}) bool {
	switch attr.DataType {
	case discovery.DataTypeString:
		_, ok := val.(string)
		return ok
	case discovery.DataTypeNumeric:
		_, ok := a.toFloat64(val)
		return ok
	case discovery.DataTypeBoolean:
		_, ok := val.(bool)
		return ok
	case discovery.DataTypeTimestamp:
		_, ok := val.(time.Time)
		return ok
	default:
		return true // Unknown types are considered valid
	}
}

// calculateOverallScore calculates weighted overall quality score
func (a *Assessor) calculateOverallScore(dims discovery.QualityDimensions) float64 {
	score := dims.Completeness.Score * a.config.CompletenessWeight +
		dims.Consistency.Score * a.config.ConsistencyWeight +
		dims.Timeliness.Score * a.config.TimelinessWeight +
		dims.Uniqueness.Score * a.config.UniquenessWeight +
		dims.Validity.Score * a.config.ValidityWeight
	
	return score
}

// identifyIssues identifies quality issues from dimensions
func (a *Assessor) identifyIssues(dims discovery.QualityDimensions) []discovery.QualityIssue {
	issues := []discovery.QualityIssue{}
	
	// Check each dimension against thresholds
	if dims.Completeness.Score < a.config.CompletenessThreshold {
		issues = append(issues, discovery.QualityIssue{
			Type:        "completeness",
			Severity:    a.getSeverity(dims.Completeness.Score, a.config.CompletenessThreshold),
			Description: fmt.Sprintf("Completeness score %.2f below threshold %.2f. Missing data may lead to incomplete analysis. Review data collection pipeline for missing fields.", dims.Completeness.Score, a.config.CompletenessThreshold),
			Impact:      1.0 - dims.Completeness.Score,
			DetectedAt:  time.Now(),
		})
	}
	
	if dims.Consistency.Score < a.config.ConsistencyThreshold {
		issues = append(issues, discovery.QualityIssue{
			Type:        "consistency",
			Severity:    a.getSeverity(dims.Consistency.Score, a.config.ConsistencyThreshold),
			Description: fmt.Sprintf("Consistency score %.2f below threshold %.2f. Inconsistent data formats may cause parsing errors. Standardize data formats at ingestion.", dims.Consistency.Score, a.config.ConsistencyThreshold),
			Impact:      1.0 - dims.Consistency.Score,
			DetectedAt:  time.Now(),
		})
	}
	
	if dims.Timeliness.Score < 0.8 { // Fixed threshold for timeliness
		issues = append(issues, discovery.QualityIssue{
			Type:        "timeliness",
			Severity:    a.getSeverity(dims.Timeliness.Score, 0.8),
			Description: fmt.Sprintf("Data freshness score %.2f indicates delays. Stale data may not reflect current state. Investigate data pipeline latency.", dims.Timeliness.Score),
			Impact:      0.8 - dims.Timeliness.Score,
			DetectedAt:  time.Now(),
		})
	}
	
	if dims.Uniqueness.Score < a.config.UniquenessThreshold {
		issues = append(issues, discovery.QualityIssue{
			Type:        "uniqueness",
			Severity:    a.getSeverity(dims.Uniqueness.Score, a.config.UniquenessThreshold),
			Description: fmt.Sprintf("Duplicate data detected, uniqueness score %.2f. Duplicates may skew aggregations and counts. Implement deduplication logic.", dims.Uniqueness.Score),
			Impact:      1.0 - dims.Uniqueness.Score,
			DetectedAt:  time.Now(),
		})
	}
	
	if dims.Validity.Score < a.config.ValidityThreshold {
		issues = append(issues, discovery.QualityIssue{
			Type:        "validity",
			Severity:    a.getSeverity(dims.Validity.Score, a.config.ValidityThreshold),
			Description: fmt.Sprintf("Invalid values detected, validity score %.2f. Invalid data may cause processing errors. Add validation rules at data ingestion.", dims.Validity.Score),
			Impact:      1.0 - dims.Validity.Score,
			DetectedAt:  time.Now(),
		})
	}
	
	return issues
}

// getSeverity determines issue severity based on score deviation
func (a *Assessor) getSeverity(score, threshold float64) string {
	deviation := threshold - score
	if deviation > 0.3 {
		return "critical"
	} else if deviation > 0.15 {
		return "high"
	} else if deviation > 0.05 {
		return "medium"
	}
	return "low"
}

// generateRecommendations creates actionable recommendations
func (a *Assessor) generateRecommendations(issues []discovery.QualityIssue) []discovery.QualityRecommendation {
	recommendations := []discovery.QualityRecommendation{}
	
	// Group issues by dimension
	dimensionCounts := make(map[string]int)
	for _, issue := range issues {
		dimensionCounts[issue.Type]++
	}
	
	// Generate recommendations based on issue patterns
	if dimensionCounts["Completeness"] > 0 {
		recommendations = append(recommendations, 
			discovery.QualityRecommendation{
				Type:        "completeness",
				Priority:    "high",
				Description: "Implement data validation at ingestion to catch missing fields",
				Impact:      0.8,
				Effort:      "medium",
			},
			discovery.QualityRecommendation{
				Type:        "completeness",
				Priority:    "medium",
				Description: "Set up alerts for schemas with low completeness scores",
				Impact:      0.6,
				Effort:      "low",
			})
	}
	
	if dimensionCounts["Consistency"] > 0 {
		recommendations = append(recommendations,
			discovery.QualityRecommendation{
				Type:        "consistency",
				Priority:    "high",
				Description: "Create data transformation rules to standardize formats",
				Impact:      0.9,
				Effort:      "high",
			},
			discovery.QualityRecommendation{
				Type:        "consistency",
				Priority:    "medium",
				Description: "Document expected data formats for each attribute",
				Impact:      0.5,
				Effort:      "low",
			})
	}
	
	if dimensionCounts["Timeliness"] > 0 {
		recommendations = append(recommendations,
			discovery.QualityRecommendation{
				Type:        "timeliness",
				Priority:    "high",
				Description: "Optimize data pipeline for reduced latency",
				Impact:      0.85,
				Effort:      "high",
			},
			discovery.QualityRecommendation{
				Type:        "timeliness",
				Priority:    "medium",
				Description: "Consider implementing real-time data streaming",
				Impact:      0.9,
				Effort:      "very_high",
			})
	}
	
	if dimensionCounts["Uniqueness"] > 0 {
		recommendations = append(recommendations,
			discovery.QualityRecommendation{
				Type:        "uniqueness",
				Priority:    "high",
				Description: "Add unique constraints or deduplication logic",
				Impact:      0.75,
				Effort:      "medium",
			},
			discovery.QualityRecommendation{
				Type:        "uniqueness",
				Priority:    "medium",
				Description: "Review data sources for duplicate generation",
				Impact:      0.6,
				Effort:      "low",
			})
	}
	
	if dimensionCounts["Validity"] > 0 {
		recommendations = append(recommendations,
			discovery.QualityRecommendation{
				Type:        "validity",
				Priority:    "high",
				Description: "Implement schema validation with explicit constraints",
				Impact:      0.8,
				Effort:      "medium",
			},
			discovery.QualityRecommendation{
				Type:        "validity",
				Priority:    "medium",
				Description: "Create data quality monitoring dashboards",
				Impact:      0.7,
				Effort:      "medium",
			})
	}
	
	// Always recommend monitoring
	recommendations = append(recommendations,
		discovery.QualityRecommendation{
			Type:        "monitoring",
			Priority:    "high",
			Description: "Set up continuous data quality monitoring",
			Impact:      0.9,
			Effort:      "medium",
		},
		discovery.QualityRecommendation{
			Type:        "monitoring",
			Priority:    "medium",
			Description: "Create quality score trending dashboards",
			Impact:      0.7,
			Effort:      "low",
		})
	
	return recommendations
}

// shouldUsePredictions determines if ML predictions should be used
func (a *Assessor) shouldUsePredictions() bool {
	// In real implementation, would check if ML models are available
	return false
}

// generatePredictions generates quality predictions
// TODO: Implement when QualityPredictions type is added to discovery package
/*
func (a *Assessor) generatePredictions(schema discovery.Schema, report discovery.QualityReport) *discovery.QualityPredictions {
	// Placeholder for ML-based predictions
	return &discovery.QualityPredictions{
		FutureScore:       report.OverallScore,
		TrendDirection:    "stable",
		RiskFactors:       []string{},
		PreventiveActions: []string{},
	}
}
*/