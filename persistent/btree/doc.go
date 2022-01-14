/*
Package btree implements a persistent (immutable) in-memory version of B-trees.

A good introduction to B-trees and their algorithms may be found at
https://algorithmtutor.com/Data-Structures/Tree/B-Trees/.
*/
package btree

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'fp.btree'.
func tracer() tracing.Trace {
	return tracing.Select("fp.btree")
}
