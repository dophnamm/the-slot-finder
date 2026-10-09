package constants

// Pagination constants
const (
	// NoLimit indicates no pagination should be applied
	NoLimit = 0
	// DefaultPage is the default page number if not specified
	DefaultPage = 1
	// DefaultPageSize is the default page size if not specified
	DefaultPageSize = 10
	// MaxPageSize caps page_size so a client cannot request the whole table
	MaxPageSize = 100
)
