//go:build deprecated
// +build deprecated

// DEPRECATED: Batch handling has been integrated into protocol.go
// This file is kept for reference but should not be used.

package mcp

import (
	"context"
)

// BatchRequest represents a JSON-RPC 2.0 batch request
type BatchRequest []Request

// BatchResponse represents a JSON-RPC 2.0 batch response
type BatchResponse []Response

// HandleBatchMessage is deprecated - use HandleMessage in protocol.go
// which now handles both single and batch requests
func (h *ProtocolHandler) HandleBatchMessage(ctx context.Context, message []byte) ([]byte, error) {
	// Delegate to the main handler which now supports batch requests
	return h.HandleMessage(ctx, message)
}