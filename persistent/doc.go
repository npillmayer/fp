/*
Immutable persistent data structures are data structures which can be copied and modified
efficiently, leaving the original unchanged. Functional programming languages like Lisp have long
relied on using them.
This package offers a selection of data structures with similar properties.

Immutable data structures in many cases offer benefits over mutable data structures in terms
of concurrent access and functional reasoning.  *Persistent* immutable data-structures offer
structural sharing, which means that if two data structures are mostly copies of each other,
most of the memory they take up will be shared between them. This implies that making copies
of an immutable data structure is relatively cheap in terms of space- and time-complexity.

License

Governed by a 3-Clause BSD license. License file may be found in the root
folder of this module.

Copyright © 2022 Norbert Pillmayer <norbert@pillmayer.com>

*/
package persistent
