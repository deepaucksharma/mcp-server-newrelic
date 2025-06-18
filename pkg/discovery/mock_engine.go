package discovery

import (
	"context"
	"time"
)

// MockEngine is a mock implementation of DiscoveryEngine for testing
type MockEngine struct{}

// NewMockEngine creates a new mock discovery engine
func NewMockEngine() *MockEngine {
	return &MockEngine{}
}

// DiscoverSchemas returns mock schemas
func (m *MockEngine) DiscoverSchemas(ctx context.Context, filter DiscoveryFilter) ([]Schema, error) {
	return []Schema{
		{
			ID:          "mock-transaction",
			Name:        "Transaction",
			EventType:   "Transaction",
			SampleCount: 1000,
			Attributes: []Attribute{
				{
					Name:     "duration",
					DataType: DataTypeNumeric,
				},
				{
					Name:     "name",
					DataType: DataTypeString,
				},
			},
			DiscoveredAt: time.Now(),
		},
	}, nil
}

// DiscoverWithIntelligence returns mock discovery result
func (m *MockEngine) DiscoverWithIntelligence(ctx context.Context, hints DiscoveryHints) (*DiscoveryResult, error) {
	schemas, _ := m.DiscoverSchemas(ctx, DiscoveryFilter{})
	return &DiscoveryResult{
		Schemas:         schemas,
		Patterns:        []CrossSchemaPattern{},
		Insights:        []Insight{},
		Recommendations: []string{"Mock recommendation"},
	}, nil
}

// ProfileSchema returns a mock schema profile
func (m *MockEngine) ProfileSchema(ctx context.Context, eventType string, depth ProfileDepth) (*Schema, error) {
	return &Schema{
		ID:          "mock-" + eventType,
		Name:        eventType,
		EventType:   eventType,
		SampleCount: 1000,
		DiscoveredAt: time.Now(),
	}, nil
}

// GetSamplingStrategy returns a mock sampling strategy
func (m *MockEngine) GetSamplingStrategy(ctx context.Context, eventType string) (SamplingStrategy, error) {
	return &mockSamplingStrategy{}, nil
}

// SampleData returns mock data samples
func (m *MockEngine) SampleData(ctx context.Context, params SamplingParams) (*DataSample, error) {
	return &DataSample{
		EventType:    params.EventType,
		Records:      []map[string]interface{}{},
		SampleSize:   100,
		TotalSize:    1000,
		SamplingRate: 0.1,
		Strategy:     "mock",
	}, nil
}

// AssessQuality returns a mock quality report
func (m *MockEngine) AssessQuality(ctx context.Context, schema string) (*QualityReport, error) {
	return &QualityReport{
		SchemaName:   schema,
		Timestamp:    time.Now(),
		OverallScore: 0.85,
		Dimensions: QualityDimensions{
			Completeness: DimensionScore{Score: 0.9},
			Consistency:  DimensionScore{Score: 0.8},
		},
	}, nil
}

// FindRelationships returns mock relationships
func (m *MockEngine) FindRelationships(ctx context.Context, schemas []Schema) ([]Relationship, error) {
	return []Relationship{}, nil
}

// Start does nothing for mock
func (m *MockEngine) Start(ctx context.Context) error {
	return nil
}

// Stop does nothing for mock
func (m *MockEngine) Stop(ctx context.Context) error {
	return nil
}

// Health returns healthy status
func (m *MockEngine) Health() HealthStatus {
	return HealthStatus{
		Status:     "healthy",
		Version:    "1.0.0",
		Uptime:     time.Hour,
		Components: map[string]ComponentHealth{},
		Metrics:    map[string]interface{}{},
	}
}

// mockSamplingStrategy implements SamplingStrategy for mock
type mockSamplingStrategy struct{}

func (s *mockSamplingStrategy) Sample(ctx context.Context, params SamplingParams) (*DataSample, error) {
	return &DataSample{
		EventType:    params.EventType,
		Records:      []map[string]interface{}{},
		SampleSize:   100,
		TotalSize:    1000,
		SamplingRate: 0.1,
		Strategy:     "mock",
	}, nil
}

func (s *mockSamplingStrategy) EstimateSampleSize(totalRecords int64) int64 {
	if totalRecords < 1000 {
		return totalRecords
	}
	return 1000
}

func (s *mockSamplingStrategy) GetStrategyName() string {
	return "mock"
}