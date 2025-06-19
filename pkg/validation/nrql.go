package validation

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// NRQLValidator provides comprehensive NRQL query validation and sanitization
type NRQLValidator struct {
	maxQueryLength int
	allowedPattern *regexp.Regexp
	dangerousOps   *regexp.Regexp
	sqlInjection   *regexp.Regexp
}

// NewNRQLValidator creates a new NRQL validator with default settings
func NewNRQLValidator() *NRQLValidator {
	return &NRQLValidator{
		maxQueryLength: 10000,
		// Allow alphanumeric, spaces, and common NRQL operators
		allowedPattern: regexp.MustCompile(`^[a-zA-Z0-9\s\-_\.\*\(\),=<>'":/%\[\]{}]+$`),
		// Dangerous operations that should never appear in NRQL
		dangerousOps: regexp.MustCompile(`(?i)\b(DROP|DELETE|UPDATE|INSERT|CREATE|ALTER|EXEC|EXECUTE|GRANT|REVOKE|TRUNCATE)\b`),
		// Common SQL injection patterns
		sqlInjection: regexp.MustCompile(`(?i)(--|/\*|\*/|;|\bunion\b|\bor\b\s+\d+=\d+|\band\b\s+\d+=\d+|'='|"="|'or'|"or")`),
	}
}

// Sanitize cleans and validates an NRQL query
func (v *NRQLValidator) Sanitize(query string) (string, error) {
	// Trim whitespace
	query = strings.TrimSpace(query)
	
	// Check if empty
	if query == "" {
		return "", fmt.Errorf("query cannot be empty")
	}
	
	// Check length
	if len(query) > v.maxQueryLength {
		return "", fmt.Errorf("query exceeds maximum length of %d characters", v.maxQueryLength)
	}
	
	// Remove any null bytes
	query = strings.ReplaceAll(query, "\x00", "")
	
	// Normalize whitespace
	query = normalizeWhitespace(query)
	
	// Check for dangerous operations
	if v.dangerousOps.MatchString(query) {
		return "", fmt.Errorf("query contains potentially dangerous operations")
	}
	
	// Check for SQL injection patterns
	if v.sqlInjection.MatchString(query) {
		return "", fmt.Errorf("query contains potential SQL injection patterns")
	}
	
	// Validate NRQL syntax
	if err := v.validateNRQLSyntax(query); err != nil {
		return "", err
	}
	
	return query, nil
}

// SanitizeIdentifier cleans an identifier (event type, attribute name, etc.)
func (v *NRQLValidator) SanitizeIdentifier(identifier string) (string, error) {
	// Trim whitespace
	identifier = strings.TrimSpace(identifier)
	
	// Check if empty
	if identifier == "" {
		return "", fmt.Errorf("identifier cannot be empty")
	}
	
	// Check length
	if len(identifier) > 255 {
		return "", fmt.Errorf("identifier exceeds maximum length of 255 characters")
	}
	
	// Only allow alphanumeric, underscore, dot, and hyphen
	if !regexp.MustCompile(`^[a-zA-Z0-9_\.\-]+$`).MatchString(identifier) {
		return "", fmt.Errorf("identifier contains invalid characters")
	}
	
	return identifier, nil
}

// SanitizeStringValue cleans a string value for use in NRQL
func (v *NRQLValidator) SanitizeStringValue(value string) string {
	// Escape single quotes by doubling them (NRQL standard)
	value = strings.ReplaceAll(value, "'", "''")
	
	// Remove any control characters
	value = removeControlCharacters(value)
	
	// Limit length
	if len(value) > 4000 {
		value = value[:4000]
	}
	
	return value
}

// validateNRQLSyntax performs basic NRQL syntax validation
func (v *NRQLValidator) validateNRQLSyntax(query string) error {
	query = strings.ToUpper(query)
	
	// Must start with SELECT
	if !strings.HasPrefix(query, "SELECT") {
		return fmt.Errorf("NRQL query must start with SELECT")
	}
	
	// Must have FROM clause
	if !strings.Contains(query, " FROM ") {
		return fmt.Errorf("NRQL query must have a FROM clause")
	}
	
	// Check for balanced parentheses
	parenCount := 0
	inString := false
	var stringChar rune
	
	for i, ch := range query {
		// Handle string literals
		if !inString && (ch == '\'' || ch == '"') {
			inString = true
			stringChar = ch
		} else if inString && ch == stringChar {
			// Check if it's escaped
			if i == 0 || query[i-1] != '\\' {
				inString = false
			}
		}
		
		// Count parentheses only outside strings
		if !inString {
			if ch == '(' {
				parenCount++
			} else if ch == ')' {
				parenCount--
			}
			if parenCount < 0 {
				return fmt.Errorf("unbalanced parentheses in query")
			}
		}
	}
	
	if parenCount != 0 {
		return fmt.Errorf("unbalanced parentheses in query")
	}
	
	if inString {
		return fmt.Errorf("unclosed string literal in query")
	}
	
	return nil
}

// normalizeWhitespace replaces multiple spaces with single space
func normalizeWhitespace(s string) string {
	// Replace tabs and newlines with spaces
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	s = re.ReplaceAllString(s, " ")
	
	return strings.TrimSpace(s)
}

// removeControlCharacters removes non-printable characters
func removeControlCharacters(s string) string {
	return strings.Map(func(r rune) rune {
		if unicode.IsControl(r) && r != '\t' && r != '\n' {
			return -1
		}
		return r
	}, s)
}

// ValidateTimeRange validates a time range string
func (v *NRQLValidator) ValidateTimeRange(timeRange string) error {
	timeRange = strings.TrimSpace(strings.ToLower(timeRange))
	
	// Check common time range patterns
	validPatterns := []string{
		`^\d+\s+(second|minute|hour|day|week|month)s?\s+ago$`,
		`^since\s+\d+\s+(second|minute|hour|day|week|month)s?\s+ago$`,
		`^\d{4}-\d{2}-\d{2}(\s+\d{2}:\d{2}:\d{2})?$`,
		`^yesterday$`,
		`^today$`,
		`^this\s+(week|month|quarter|year)$`,
		`^last\s+(week|month|quarter|year)$`,
	}
	
	for _, pattern := range validPatterns {
		if matched, _ := regexp.MatchString(pattern, timeRange); matched {
			return nil
		}
	}
	
	return fmt.Errorf("invalid time range format")
}

// ExtractEventTypes extracts event types from a query safely
func (v *NRQLValidator) ExtractEventTypes(query string) ([]string, error) {
	// First sanitize the query
	query, err := v.Sanitize(query)
	if err != nil {
		return nil, err
	}
	
	// Find FROM clause
	re := regexp.MustCompile(`(?i)\bFROM\s+([a-zA-Z0-9_,\s]+)(?:\s+WHERE|\s+SINCE|\s+UNTIL|\s+FACET|\s+LIMIT|$)`)
	matches := re.FindStringSubmatch(query)
	
	if len(matches) < 2 {
		return nil, fmt.Errorf("could not extract event types from query")
	}
	
	// Split by comma and clean each event type
	eventTypesStr := matches[1]
	eventTypes := strings.Split(eventTypesStr, ",")
	
	cleanedTypes := make([]string, 0, len(eventTypes))
	for _, et := range eventTypes {
		et = strings.TrimSpace(et)
		if et != "" {
			sanitized, err := v.SanitizeIdentifier(et)
			if err != nil {
				return nil, fmt.Errorf("invalid event type '%s': %w", et, err)
			}
			cleanedTypes = append(cleanedTypes, sanitized)
		}
	}
	
	if len(cleanedTypes) == 0 {
		return nil, fmt.Errorf("no valid event types found")
	}
	
	return cleanedTypes, nil
}