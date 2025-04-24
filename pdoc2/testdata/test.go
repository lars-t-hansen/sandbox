// This is just some junk

package something

// This is the preamble
// And another preamble line
type _preamble int

// This is a doc comment above a parenthesized const, but it should be ignored,
// the individual consts are commented.
const (
	// This is an int constant in a group
	C1 = 10
	// This is a string constant in a group
	C2 = "hello"
)

// This is another string constant
const C3 = "hoho"

// Functions and their docs are ignored
func f() {
}

// Vars and their docs are ignored
var v = 10

// This is a not a doc comment

// This is also not a doc comment, and the next (undocumented) structure has no members and should
// probably have no output?

type X struct { }

// This is a struct definition
type S struct {
	// This is an int field
	F1 int
	// This is a string field
	F2 string
}

// This is a type alias, also documented
type W int

// This is the postamble
type _postamble int
