package autoresolver

import (
	"context"
	"fmt"
	"log/slog"
	"strings"
)

// Brain defines the interface for resolving semantic intents to Playwright selectors.
type DOMElementResolver interface {
	Query(ctx context.Context, intent string, cat string) (string, error)
}

// Mock is a placeholder implementation of the Brain interface.
// In a real scenario, this would involve SLM/VLM inference.
type Mock struct {
	// For testing purposes, can store predefined mappings
	PredefinedResponses map[string]string
}

// NewMock creates a new Mock instance.
func NewMock(responses map[string]string) *Mock {
	return &Mock{
		PredefinedResponses: responses,
	}
}

// Query attempts to resolve a semantic intent using the CAT.
// For now, it checks predefined responses or returns a dummy selector based on keywords.
func (mb *Mock) Query(ctx context.Context, intent string, cat string) (string, error) {
	slog.Debug("Mock: Querying for intent", "intent", intent, "cat_length", len(cat))

	// 1. Check predefined responses (simulates a very basic "learned" mapping)
	if selector, ok := mb.PredefinedResponses[intent]; ok {
		slog.Debug("Mock: Found predefined response", "intent", intent, "selector", selector)
		return selector, nil
	}

	// 2. Simulate Local SLM inference (simple keyword matching for now)
	// In a real implementation, this would involve a local WASM model.
	lowerIntent := strings.ToLower(intent)
	if strings.Contains(lowerIntent, "sign-up button") || strings.Contains(lowerIntent, "register button") {
		slog.Debug("Mock: SLM-like inference: found 'sign-up button'", "intent", intent)
		return "button[aria-label*='sign-up'], button[id*='register']", nil
	}
	if strings.Contains(lowerIntent, "login button") {
		slog.Debug("Mock: SLM-like inference: found 'login button'", "intent", intent)
		return "button[aria-label*='login'], button[id*='login']", nil
	}
	if strings.Contains(lowerIntent, "search input") {
		slog.Debug("Mock: SLM-like inference: found 'search input'", "intent", intent)
		return "input[type='search'], input[placeholder*='search']", nil
	}
	if strings.Contains(lowerIntent, "submit button") {
		slog.Debug("Mock: SLM-like inference: found 'submit button'", "intent", intent)
		return "button[type='submit'], input[type='submit']", nil
	}
	if strings.Contains(lowerIntent, "text field") || strings.Contains(lowerIntent, "input field") {
		slog.Debug("Mock: SLM-like inference: found 'text field'", "intent", intent)
		return "input[type='text'], textarea", nil
	}
	if strings.Contains(lowerIntent, "image") {
		slog.Debug("Mock: SLM-like inference: found 'image'", "intent", intent)
		return "img", nil
	}

	// 3. Fallback to Cloud VLM (currently just a failure for mock)
	// In a real implementation, this would involve:
	// - Taking a screenshot
	// - Sending intent, CAT, and screenshot to a cloud VLM (e.g., Gemini API)
	// - Parsing the VLM's response for the selector.
	slog.Warn("Mock: No SLM-like inference, falling back to VLM (mock failure)", "intent", intent)
	return "", fmt.Errorf("Mock: could not resolve intent '%s' via SLM or VLM", intent)
}
