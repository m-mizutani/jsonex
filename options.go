package jsonex

// options holds internal configuration options (unexported)
type options struct {
	maxDepth   int // maximum nesting depth (default: 1000)
	bufferSize int // read buffer size (default: 4096)
}

// defaultOptions returns the default configuration
func defaultOptions() options {
	return options{
		maxDepth:   1000,
		bufferSize: 4096,
	}
}

// Option is a function that modifies options
type Option func(*options)

// WithMaxDepth sets the maximum nesting depth
// This helps prevent stack overflow attacks with deeply nested JSON
func WithMaxDepth(depth int) Option {
	return func(o *options) {
		if depth > 0 {
			o.maxDepth = depth
		}
	}
}

// WithBufferSize sets the read buffer size for performance tuning
// Larger buffers may improve performance for large JSON files
func WithBufferSize(size int) Option {
	return func(o *options) {
		if size > 0 {
			o.bufferSize = size
		}
	}
}

// applyOptions applies the given options to the default configuration
func applyOptions(opts ...Option) options {
	o := defaultOptions()
	for _, opt := range opts {
		opt(&o)
	}
	return o
}
