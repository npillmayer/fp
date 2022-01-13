package btree

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'fp.btree'.
func tracer() tracing.Trace {
	return tracing.Select("fp.btree")
}
