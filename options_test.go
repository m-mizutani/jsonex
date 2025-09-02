package jsonex

import (
	"testing"
)

func TestDefaultOptions(t *testing.T) {
	opts := defaultOptions()

	if opts.maxDepth != 1000 {
		t.Errorf("defaultOptions().maxDepth = %d, expected 1000", opts.maxDepth)
	}
	if opts.bufferSize != 4096 {
		t.Errorf("defaultOptions().bufferSize = %d, expected 4096", opts.bufferSize)
	}
}

func TestWithMaxDepth(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{500, 500},
		{1, 1},
		{0, 1000},  // Invalid value, should keep default
		{-1, 1000}, // Invalid value, should keep default
	}

	for _, test := range tests {
		opts := defaultOptions()
		WithMaxDepth(test.input)(&opts)

		if opts.maxDepth != test.expected {
			t.Errorf("WithMaxDepth(%d) resulted in maxDepth = %d, expected %d",
				test.input, opts.maxDepth, test.expected)
		}
	}
}

func TestWithBufferSize(t *testing.T) {
	tests := []struct {
		input    int
		expected int
	}{
		{8192, 8192},
		{1024, 1024},
		{0, 4096},  // Invalid value, should keep default
		{-1, 4096}, // Invalid value, should keep default
	}

	for _, test := range tests {
		opts := defaultOptions()
		WithBufferSize(test.input)(&opts)

		if opts.bufferSize != test.expected {
			t.Errorf("WithBufferSize(%d) resulted in bufferSize = %d, expected %d",
				test.input, opts.bufferSize, test.expected)
		}
	}
}

func TestApplyOptions(t *testing.T) {
	opts := applyOptions(
		WithMaxDepth(500),
		WithBufferSize(8192),
	)

	if opts.maxDepth != 500 {
		t.Errorf("applyOptions maxDepth = %d, expected 500", opts.maxDepth)
	}
	if opts.bufferSize != 8192 {
		t.Errorf("applyOptions bufferSize = %d, expected 8192", opts.bufferSize)
	}
}

func TestApplyOptionsEmpty(t *testing.T) {
	opts := applyOptions()

	// Should get default values
	if opts.maxDepth != 1000 {
		t.Errorf("applyOptions() maxDepth = %d, expected 1000", opts.maxDepth)
	}
	if opts.bufferSize != 4096 {
		t.Errorf("applyOptions() bufferSize = %d, expected 4096", opts.bufferSize)
	}
}
