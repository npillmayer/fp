/*
Package styledtree is a straightforward default implementation of a styled document tree.

Overview

This is an implementation of style.TreeNode and of cssom.StyledNode.
Using a builder type, cssom.Style() will create a styled tree from an
HTML parse tree and a CSSOM. The resulting styled tree exposes interface
style.TreeNode for every node and may be manipulated via an API.

This is the default implementation used by the engine. However, for
interactive use it may be appropriate to create a styled tree derived
from another type of styled node. The engine's design should fully
support this kind of switch.

___________________________________________________________________________

License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2017–2022 Norbert Pillmayer <norbert@pillmayer.com>

*/
package styledtree

import (
	"github.com/npillmayer/schuko/tracing"
)

// tracer traces with key 'tyse.dom'.
func tracer() tracing.Trace {
	return tracing.Select("tyse.dom")
}
