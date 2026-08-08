package platform

import "errors"

type ErrorCategory string

const (
	CategoryConfiguration     ErrorCategory = "configuration"
	CategoryDependencies      ErrorCategory = "dependencies"
	CategoryStartup           ErrorCategory = "startup"
	CategoryHTTPInternal      ErrorCategory = "http_internal"
	CategoryHTTPListenOrServe ErrorCategory = "http_listen_or_serve"
	CategoryHTTPShutdown      ErrorCategory = "http_shutdown"
	CategoryHTTPForcedClose   ErrorCategory = "http_forced_close"
	CategoryInternal          ErrorCategory = "internal"
)

type CategorizedError struct {
	category ErrorCategory
	cause    error
}

func NewCategorizedError(category ErrorCategory, cause error) *CategorizedError {
	return &CategorizedError{category: category, cause: cause}
}

func (e *CategorizedError) Error() string {
	return string(e.category)
}

func (e *CategorizedError) Unwrap() error {
	return e.cause
}

func ErrorCategoryOf(err error, fallback ErrorCategory) ErrorCategory {
	var categorized *CategorizedError
	if errors.As(err, &categorized) {
		return categorized.category
	}
	return fallback
}
