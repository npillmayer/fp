/*
Package vector implements an immutable persistent vector.

An immutable persistent vector has copy-on-write behaviour: Each “modification” of the vector
(insertion, replacement or deletion) creates a copy, leaving the original unmodified.
Under the hood, copy-on-write retains most of the memory held by the original, and creates
a new incarnation of parts of the structure only. Thus, most of the structure/memory
is shared between original and copy, transparently to clients.

Immutable vectors are inherently concurrency-safe.

Status

Awaiting Go 1.18 with generics.

License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2022 Norbert Pillmayer <norbert@pillmayer.com>

*/
package vector

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'persistent.vector'.
func tracer() tracing.Trace {
	return tracing.Select("persistent.vector")
}
