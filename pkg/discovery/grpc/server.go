package grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery"
	"github.com/deepaucksharma/mcp-server-newrelic/pkg/discovery/telemetry"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
)

// DiscoveryServer implements a gRPC server for the discovery engine
type DiscoveryServer struct {
	engine discovery.DiscoveryEngine
	tracer *telemetry.Tracer
	server *grpc.Server
}

// Config holds gRPC server configuration
type Config struct {
	Port              int
	MaxMessageSize    int
	ConnectionTimeout time.Duration
	EnableReflection  bool
	EnableHealth      bool
}

// DefaultConfig returns default gRPC configuration
func DefaultConfig() Config {
	return Config{
		Port:              8081,
		MaxMessageSize:    10 * 1024 * 1024, // 10MB
		ConnectionTimeout: 30 * time.Second,
		EnableReflection:  true,
		EnableHealth:      true,
	}
}

// NewDiscoveryServer creates a new gRPC server for discovery
func NewDiscoveryServer(engine discovery.DiscoveryEngine, config Config) (*DiscoveryServer, error) {
	// Create tracer
	tracerConfig := telemetry.DefaultConfig()
	tracer, err := telemetry.NewTracer(tracerConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create tracer: %w", err)
	}

	// Create gRPC server with options
	opts := []grpc.ServerOption{
		grpc.MaxRecvMsgSize(config.MaxMessageSize),
		grpc.MaxSendMsgSize(config.MaxMessageSize),
		grpc.ConnectionTimeout(config.ConnectionTimeout),
		grpc.UnaryInterceptor(unaryInterceptor(tracer)),
		grpc.StreamInterceptor(streamInterceptor(tracer)),
	}

	server := grpc.NewServer(opts...)

	// Create discovery server
	ds := &DiscoveryServer{
		engine: engine,
		tracer: tracer,
		server: server,
	}

	// Register services
	RegisterDiscoveryServiceServer(server, ds)

	// Enable reflection for debugging
	if config.EnableReflection {
		reflection.Register(server)
	}

	// Enable health checks
	if config.EnableHealth {
		healthServer := health.NewServer()
		grpc_health_v1.RegisterHealthServer(server, healthServer)
		healthServer.SetServingStatus("discovery.DiscoveryService", grpc_health_v1.HealthCheckResponse_SERVING)
	}

	return ds, nil
}

// Start starts the gRPC server
func (s *DiscoveryServer) Start(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	go func() {
		if err := s.server.Serve(lis); err != nil {
			fmt.Printf("gRPC server error: %v\n", err)
		}
	}()

	return nil
}

// Stop gracefully stops the server
func (s *DiscoveryServer) Stop(ctx context.Context) error {
	s.server.GracefulStop()
	return s.tracer.Shutdown(ctx)
}

// DiscoverSchemas implements the gRPC method
func (s *DiscoveryServer) DiscoverSchemas(ctx context.Context, req *DiscoverSchemasRequest) (*DiscoverSchemasResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.DiscoverSchemas")
	defer span.End()

	// Convert request to discovery filter
	filter := discovery.DiscoveryFilter{
		EventTypes:  req.EventTypes,
		MaxSchemas:  int(req.MaxSchemas),
		Tags:        req.Tags,
	}

	// Call engine
	start := time.Now()
	schemas, err := s.engine.DiscoverSchemas(ctx, filter)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Convert response
	resp := &DiscoverSchemasResponse{
		Schemas:           convertSchemas(schemas),
		TotalCount:        int32(len(schemas)),
		DiscoveryDuration: duration.Milliseconds(),
		Metadata: map[string]string{
			"trace_id": telemetry.ExtractTraceID(ctx),
		},
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// ProfileSchema implements the gRPC method
func (s *DiscoveryServer) ProfileSchema(ctx context.Context, req *ProfileSchemaRequest) (*ProfileSchemaResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.ProfileSchema")
	defer span.End()

	// Convert profile depth
	depth := discovery.ProfileDepth(req.ProfileDepth)
	if depth == "" {
		depth = discovery.ProfileDepthStandard
	}

	// Call engine
	start := time.Now()
	schema, err := s.engine.ProfileSchema(ctx, req.EventType, depth)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Convert response
	resp := &ProfileSchemaResponse{
		Schema:            convertSchema(schema),
		ProfilingDuration: duration.Milliseconds(),
		Metadata: map[string]string{
			"trace_id": telemetry.ExtractTraceID(ctx),
		},
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// IntelligentDiscovery implements the gRPC method
func (s *DiscoveryServer) IntelligentDiscovery(ctx context.Context, req *IntelligentDiscoveryRequest) (*IntelligentDiscoveryResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.IntelligentDiscovery")
	defer span.End()

	// Convert hints
	hints := discovery.DiscoveryHints{
		Keywords:       req.FocusAreas, // Map FocusAreas to Keywords
		Purpose:        "", // Not directly provided in request
		PreferredTypes: req.EventTypes,
		Domain:         "", // Not provided in request
		Examples:       []string{}, // Not provided in request
		Constraints:    make(map[string]interface{}), // Could add confidence threshold here
	}
	
	// Add context to constraints if provided
	if len(req.Context) > 0 {
		hints.Constraints["context"] = req.Context
	}
	
	// Add confidence threshold to constraints if provided
	if req.ConfidenceThreshold > 0 {
		hints.Constraints["confidence_threshold"] = req.ConfidenceThreshold
	}

	// Call engine
	start := time.Now()
	result, err := s.engine.DiscoverWithIntelligence(ctx, hints)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Convert response
	var schemas []discovery.Schema
	if result != nil {
		schemas = result.Schemas
	}
	
	resp := &IntelligentDiscoveryResponse{
		Schemas:           convertSchemas(schemas),
		DiscoveryDuration: duration.Milliseconds(),
		Insights:          generateInsights(schemas),
		Recommendations:   generateRecommendations(schemas),
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// FindRelationships implements the gRPC method
func (s *DiscoveryServer) FindRelationships(ctx context.Context, req *FindRelationshipsRequest) (*FindRelationshipsResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.FindRelationships")
	defer span.End()

	// Get schemas for relationship discovery
	schemas := make([]discovery.Schema, 0, len(req.SchemaNames))
	for _, name := range req.SchemaNames {
		schema, err := s.engine.ProfileSchema(ctx, name, discovery.ProfileDepthBasic)
		if err != nil {
			continue // Skip schemas that can't be profiled
		}
		schemas = append(schemas, *schema)
	}

	// Call engine
	start := time.Now()
	relationships, err := s.engine.FindRelationships(ctx, schemas)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Convert response
	resp := &FindRelationshipsResponse{
		Relationships:    convertRelationships(relationships),
		Graph:            buildRelationshipGraph(schemas, relationships),
		AnalysisDuration: duration.Milliseconds(),
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// AssessQuality implements the gRPC method
func (s *DiscoveryServer) AssessQuality(ctx context.Context, req *AssessQualityRequest) (*AssessQualityResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.AssessQuality")
	defer span.End()

	// Call engine
	start := time.Now()
	report, err := s.engine.AssessQuality(ctx, req.EventType)
	duration := time.Since(start)

	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return nil, err
	}

	// Convert response
	resp := &AssessQualityResponse{
		Report:             convertQualityReport(report),
		AssessmentDuration: duration.Milliseconds(),
	}

	span.SetStatus(codes.Ok, "")
	return resp, nil
}

// GetHealth implements the gRPC method
func (s *DiscoveryServer) GetHealth(ctx context.Context, req *GetHealthRequest) (*GetHealthResponse, error) {
	ctx, span := s.tracer.Start(ctx, "grpc.GetHealth")
	defer span.End()

	health := s.engine.Health()

	resp := &GetHealthResponse{
		IsHealthy: health.Status == "healthy",
		Status:    health.Status,
		Timestamp: time.Now().Unix(),
		Checks:    convertHealthChecks(&health),
	}

	if req.IncludeMetrics {
		// Extract metrics from health.Metrics if available
		metrics := &HealthMetrics{
			Uptime: int64(health.Uptime.Seconds()),
		}
		
		// Extract other metrics from health.Metrics map if they exist
		if qp, ok := health.Metrics["queries_processed"].(int64); ok {
			metrics.QueriesProcessed = qp
		}
		if ec, ok := health.Metrics["errors_count"].(int64); ok {
			metrics.ErrorsCount = ec
		}
		if chr, ok := health.Metrics["cache_hit_rate"].(float64); ok {
			metrics.CacheHitRate = chr
		}
		if aqt, ok := health.Metrics["average_query_time_ms"].(int64); ok {
			metrics.AverageQueryTimeMs = aqt
		}
		
		resp.Metrics = metrics
	}

	if health.Status == "healthy" {
		span.SetStatus(codes.Ok, "Healthy")
	} else {
		span.SetStatus(codes.Error, health.Status)
	}

	return resp, nil
}

// Helper functions for conversions

func convertSchemas(schemas []discovery.Schema) []*Schema {
	result := make([]*Schema, len(schemas))
	for i, s := range schemas {
		result[i] = convertSchema(&s)
	}
	return result
}

func convertSchema(s *discovery.Schema) *Schema {
	if s == nil {
		return nil
	}

	metadata, _ := json.Marshal(s.Metadata)
	
	return &Schema{
		Id:             s.ID,
		Name:           s.Name,
		EventType:      s.EventType,
		Attributes:     convertAttributes(s.Attributes),
		SampleCount:    s.SampleCount,
		DataVolume:     convertDataVolume(s.DataVolume),
		Quality:        convertQualityMetrics(s.Quality),
		Patterns:       convertPatterns(s.Patterns),
		DiscoveredAt:   s.DiscoveredAt.Unix(),
		LastAnalyzedAt: s.LastAnalyzedAt.Unix(),
		Metadata:       string(metadata),
	}
}

func convertAttributes(attrs []discovery.Attribute) []*Attribute {
	result := make([]*Attribute, len(attrs))
	for i, a := range attrs {
		sampleValues := make([]string, 0, len(a.SampleValues))
		for _, v := range a.SampleValues {
			sampleValues = append(sampleValues, fmt.Sprintf("%v", v))
		}
		result[i] = &Attribute{
			Name:         a.Name,
			DataType:     string(a.DataType),
			SemanticType: string(a.SemanticType),
			IsRequired:   a.NullRatio == 0, // Inferred from null ratio
			IsUnique:     a.Cardinality.Ratio > 0.9, // High cardinality suggests uniqueness
			IsIndexed:    false, // Not available in the struct
			Cardinality:  a.Cardinality.Ratio,
			SampleValues: sampleValues,
		}
	}
	return result
}

func convertDataVolume(dv discovery.DataVolumeProfile) *DataVolumeProfile {
	return &DataVolumeProfile{
		TotalEvents:      dv.TotalRecords,
		EventsPerMinute:  float64(dv.RecordsPerHour) / 60.0,
		DataSizeBytes:    int64(dv.EstimatedSizeGB * 1024 * 1024 * 1024),
		FirstSeen:        time.Now().Unix(), // Not available in the struct
		LastSeen:         time.Now().Unix(), // Not available in the struct
	}
}

func convertQualityMetrics(qm discovery.QualityMetrics) *QualityMetrics {
	return &QualityMetrics{
		OverallScore: qm.OverallScore,
		Dimensions:   convertQualityDimensionsFromMetrics(qm),
		Issues:       convertQualityIssues(qm.Issues),
	}
}

func convertQualityDimensionsFromMetrics(qm discovery.QualityMetrics) *QualityDimensions {
	return &QualityDimensions{
		Completeness: &QualityDimension{
			Score:  qm.Completeness,
			Issues: []string{},
		},
		Consistency: &QualityDimension{
			Score:  qm.Consistency,
			Issues: []string{},
		},
		Timeliness: &QualityDimension{
			Score:  qm.Timeliness,
			Issues: []string{},
		},
		Uniqueness: &QualityDimension{
			Score:  qm.Uniqueness,
			Issues: []string{},
		},
		Validity: &QualityDimension{
			Score:  qm.Validity,
			Issues: []string{},
		},
	}
}

func convertQualityDimensions(qd discovery.QualityDimensions) *QualityDimensions {
	return &QualityDimensions{
		Completeness: convertQualityDimension(qd.Completeness),
		Consistency:  convertQualityDimension(qd.Consistency),
		Timeliness:   convertQualityDimension(qd.Timeliness),
		Uniqueness:   convertQualityDimension(qd.Uniqueness),
		Validity:     convertQualityDimension(qd.Validity),
	}
}

func convertQualityDimension(qd discovery.DimensionScore) *QualityDimension {
	return &QualityDimension{
		Score:  qd.Score,
		Issues: qd.Issues,
	}
}

func convertQualityIssues(issues []discovery.QualityIssue) []*QualityIssue {
	result := make([]*QualityIssue, len(issues))
	for i, issue := range issues {
		affectedAttrs := []string{}
		if issue.Attribute != "" {
			affectedAttrs = append(affectedAttrs, issue.Attribute)
		}
		result[i] = &QualityIssue{
			Severity:           issue.Severity,
			Type:               issue.Type,
			Description:        issue.Description,
			AffectedAttributes: affectedAttrs,
			OccurrenceCount:    int64(issue.Impact), // Using impact as occurrence count proxy
		}
	}
	return result
}

func convertPatterns(patterns []discovery.DetectedPattern) []*DetectedPattern {
	result := make([]*DetectedPattern, len(patterns))
	for i, p := range patterns {
		params, _ := json.Marshal(p.Evidence)
		result[i] = &DetectedPattern{
			Type:               p.Type,
			Subtype:            p.Name, // Using Name as subtype
			Confidence:         p.Confidence,
			Description:        p.Description,
			Parameters:         string(params),
			AffectedAttributes: p.Attributes,
		}
	}
	return result
}

func convertRelationships(relationships []discovery.Relationship) []*Relationship {
	result := make([]*Relationship, len(relationships))
	for i, r := range relationships {
		// Create join conditions from source and target attributes
		joinConditions := []*JoinCondition{{
			SourceAttribute: r.SourceAttribute,
			TargetAttribute: r.TargetAttribute,
			Operator:        "=", // Default operator
		}}
		
		result[i] = &Relationship{
			Id:              r.ID,
			Type:            string(r.Type),
			SourceSchema:    r.SourceSchema,
			TargetSchema:    r.TargetSchema,
			JoinConditions:  joinConditions,
			Strength:        r.Confidence, // Using confidence as strength
			Confidence:      r.Confidence,
			SampleMatches:   0, // Not available in the struct
		}
	}
	return result
}


func convertQualityReport(report *discovery.QualityReport) *QualityReport {
	if report == nil {
		return nil
	}
	
	var metadata []byte
	// Create empty metadata if not available
	metadataMap := make(map[string]string)
	metadata, _ = json.Marshal(metadataMap)
	
	return &QualityReport{
		EventType:    report.SchemaName,
		OverallScore: report.OverallScore,
		Dimensions:   convertQualityDimensions(report.Dimensions),
		Issues:       convertQualityIssues(report.Issues),
		AssessedAt:   report.Timestamp.Unix(),
		Metadata:     string(metadata),
	}
}

func convertHealthChecks(health *discovery.HealthStatus) []*HealthCheck {
	checks := make([]*HealthCheck, 0)
	
	// Add engine health check
	checks = append(checks, &HealthCheck{
		Name:      "engine",
		IsHealthy: health.Status == "healthy",
		Message:   health.Status,
	})
	
	// Add component health checks
	for name, comp := range health.Components {
		checks = append(checks, &HealthCheck{
			Name:      name,
			IsHealthy: comp.Status == "healthy",
			Message:   comp.Message,
		})
	}
	
	return checks
}

func generateInsights(schemas []discovery.Schema) []*DiscoveryInsight {
	// This would analyze schemas and generate insights
	// For now, return empty
	return []*DiscoveryInsight{}
}

func generateRecommendations(schemas []discovery.Schema) []string {
	// This would analyze schemas and generate recommendations
	// For now, return empty
	return []string{}
}

func buildRelationshipGraph(schemas []discovery.Schema, relationships []discovery.Relationship) *RelationshipGraph {
	// Build graph representation
	nodes := make([]*GraphNode, len(schemas))
	for i, s := range schemas {
		nodes[i] = &GraphNode{
			Id:         s.ID,
			SchemaName: s.Name,
		}
	}

	edges := make([]*GraphEdge, len(relationships))
	for i, r := range relationships {
		edges[i] = &GraphEdge{
			Source:         r.SourceSchema,
			Target:         r.TargetSchema,
			RelationshipId: r.ID,
			Weight:         r.Confidence, // Using confidence as weight
		}
	}

	return &RelationshipGraph{
		Nodes: nodes,
		Edges: edges,
	}
}

// Interceptors for tracing

func unaryInterceptor(tracer *telemetry.Tracer) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		ctx, span := tracer.Start(ctx, info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.String("rpc.service", "discovery"),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		resp, err := handler(ctx, req)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

func streamInterceptor(tracer *telemetry.Tracer) grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx, span := tracer.Start(ss.Context(), info.FullMethod,
			trace.WithAttributes(
				attribute.String("rpc.system", "grpc"),
				attribute.String("rpc.method", info.FullMethod),
				attribute.String("rpc.service", "discovery"),
				attribute.Bool("rpc.stream", true),
			),
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		err := handler(srv, wrapped)
		if err != nil {
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}