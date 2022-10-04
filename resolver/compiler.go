package main

import "fmt"

type astNode interface {
	fmt.Stringer
}

type astVar struct {
	name string
}

func (n *astVar) String() string {
	return n.name
}

type astAtom struct {
	name string
}

func (n *astAtom) String() string {
	return n.name
}

type astNumber struct {
	value int64
}

func (n *astNumber) String() string {
	return fmt.Sprint(n.value)
}

type astStruct struct {
	name       string
	components []astNode
}

func (n *astStruct) String() string {
	args := ""
	for _, c := range n.components {
		if args != "" {
			args = args + ","
		}
		args = args + c.String()
	}
	return n.name + "(" + args + ")"
}

// f(G) :- h(G,A,"x"), j(A)
// is compiled into
//
//	rule{locals: 2,
//	     head: &structure{"f",[]term{&local{0}},
//	     body: []*structure{&structure{"h", []term{&local{0},&local{1},intern("x")}},
//	                        &structure{"j", []term{&local{1}}}}
//
// The input to the compiler is a head structure and a slice of subterms, also structures.
// These structures are fully ground - this is source code!  There is a
/*
func compileRule(head *astStruct, subterms []*astNode) *rule {
	//rib := make(map[string]int)

	panic("NYI")
}
*/
