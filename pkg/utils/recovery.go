package utils

import (
	"fmt"
	"runtime/debug"
	"sync/atomic"

	"github.com/deepaucksharma/mcp-server-newrelic/pkg/logger"
)

// PanicHandler is a function that handles panics
type PanicHandler func(recovered interface{}, stack []byte)

// DefaultPanicHandler logs the panic and stack trace
var DefaultPanicHandler PanicHandler = func(recovered interface{}, stack []byte) {
	logger.Error("Panic recovered: %v\nStack trace:\n%s", recovered, stack)
}

// PanicStats tracks panic statistics
type PanicStats struct {
	Total   int64
	handler PanicHandler
}

// GlobalPanicStats tracks global panic statistics
var GlobalPanicStats = &PanicStats{
	handler: DefaultPanicHandler,
}

// SetPanicHandler sets a custom panic handler
func (ps *PanicStats) SetPanicHandler(handler PanicHandler) {
	ps.handler = handler
}

// GetTotal returns the total number of panics recovered
func (ps *PanicStats) GetTotal() int64 {
	return atomic.LoadInt64(&ps.Total)
}

// SafeGo runs a function in a goroutine with panic recovery
func SafeGo(fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt64(&GlobalPanicStats.Total, 1)
				stack := debug.Stack()
				if GlobalPanicStats.handler != nil {
					GlobalPanicStats.handler(r, stack)
				}
			}
		}()
		fn()
	}()
}

// SafeGoWithContext runs a function with a context-specific panic handler
func SafeGoWithContext(name string, fn func()) {
	go func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt64(&GlobalPanicStats.Total, 1)
				stack := debug.Stack()
				logger.Error("Panic in %s: %v\nStack trace:\n%s", name, r, stack)
				if GlobalPanicStats.handler != nil {
					GlobalPanicStats.handler(r, stack)
				}
			}
		}()
		fn()
	}()
}

// SafeGoWithRestart runs a function and restarts it if it panics
func SafeGoWithRestart(name string, fn func(), maxRestarts int) {
	restarts := 0
	
	var runWithRecovery func()
	runWithRecovery = func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt64(&GlobalPanicStats.Total, 1)
				stack := debug.Stack()
				logger.Error("Panic in %s (restart %d/%d): %v\nStack trace:\n%s", 
					name, restarts, maxRestarts, r, stack)
				
				if restarts < maxRestarts {
					restarts++
					logger.Info("Restarting %s after panic (attempt %d/%d)", name, restarts, maxRestarts)
					go runWithRecovery()
				} else {
					logger.Error("Max restarts reached for %s, giving up", name)
				}
			}
		}()
		fn()
	}
	
	go runWithRecovery()
}

// SafeFunc wraps a function to recover from panics and return an error
func SafeFunc(fn func() error) error {
	var err error
	func() {
		defer func() {
			if r := recover(); r != nil {
				atomic.AddInt64(&GlobalPanicStats.Total, 1)
				stack := debug.Stack()
				err = fmt.Errorf("panic recovered: %v\nStack trace:\n%s", r, stack)
				if GlobalPanicStats.handler != nil {
					GlobalPanicStats.handler(r, stack)
				}
			}
		}()
		err = fn()
	}()
	return err
}

// SafeCall wraps a function call to recover from panics
func SafeCall(fn func()) (err error) {
	defer func() {
		if r := recover(); r != nil {
			atomic.AddInt64(&GlobalPanicStats.Total, 1)
			stack := debug.Stack()
			err = fmt.Errorf("panic recovered: %v", r)
			logger.Error("Panic recovered: %v\nStack trace:\n%s", r, stack)
			if GlobalPanicStats.handler != nil {
				GlobalPanicStats.handler(r, stack)
			}
		}
	}()
	fn()
	return nil
}