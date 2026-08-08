package platform

import "errors"

// ErrorCategory is a bounded, log-safe runtime failure classification.
type ErrorCategory string

const (
	// CategoryConfiguration identifies configuration validation failures.
	CategoryConfiguration ErrorCategory = "configuration"
	// CategoryDependencies identifies dependency initialization failures.
	CategoryDependencies ErrorCategory = "dependencies"
	// CategoryStartup identifies process startup events.
	CategoryStartup ErrorCategory = "startup"
	// CategoryHTTPInternal identifies internal HTTP server failures.
	CategoryHTTPInternal ErrorCategory = "http_internal"
	// CategoryHTTPListenOrServe identifies HTTP listener or serving failures.
	CategoryHTTPListenOrServe ErrorCategory = "http_listen_or_serve"
	// CategoryHTTPShutdown identifies graceful HTTP shutdown failures.
	CategoryHTTPShutdown ErrorCategory = "http_shutdown"
	// CategoryHTTPForcedClose identifies failures while forcing an HTTP server closed.
	CategoryHTTPForcedClose ErrorCategory = "http_forced_close"
	// CategoryInternal identifies otherwise uncategorized internal failures.
	CategoryInternal ErrorCategory = "internal"
)

// CategorizedError exposes a safe category while retaining its private cause for inspection.
type CategorizedError struct {
	category ErrorCategory
	cause    error
}

// NewCategorizedError wraps cause with a safe external category.
func NewCategorizedError(category ErrorCategory, cause error) *CategorizedError {
	return &CategorizedError{category: category, cause: cause}
}

func (e *CategorizedError) Error() string {
	return string(e.category)
}

func (e *CategorizedError) Unwrap() error {
	return e.cause
}

// ErrorCategoryOf extracts a categorized error or returns fallback.
func ErrorCategoryOf(err error, fallback ErrorCategory) ErrorCategory {
	var categorized *CategorizedError
	if errors.As(err, &categorized) {
		return categorized.category
	}
	return fallback
}
