package main

import (
	"testing"
)

func TestUnifiation(t *testing.T) {
	/*
			var buf strings.Builder
			c := newStore()
			q := newQuery(c)
			A := q.newVar("A")
			B := q.newVar("B")
			if !unify(A, c.newStruct("f", c.atom("false"), B)) {
				t.Fatal("First test")
			}
			if !unify(B, c.atom("true")) {
				t.Fatal("Second test")
			}
			D := q.newVar("D")
			if !unify(A, c.newStruct("f", D, c.atom("true"))) {
				t.Fatal("Third test")
			}
			E := q.newVar("E")
			F := q.newVar("F")
			G := q.newVar("G")
			H := q.newVar("H")
			if !unify(c.newStruct("f", E, F), c.newStruct("f", G, H)) {
				t.Fatal("Fourth test")
			}
			if !unify(c.newStruct("f", c.atom("hi"), c.atom("ho")), E) {
				t.Fatal("Fifth test")
			}
			q.toString(&buf)
			expect :=
				`A = f(false,true)
		B = true
		D = false
		E = f(hi,ho)
		F = F
		G = f(hi,ho)
		H = F
		`
			if buf.String() != expect {
				t.Fatal("Bad result")
			}
	*/
}
