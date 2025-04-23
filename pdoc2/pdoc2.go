// Hi there

// Can we use the Go parser framework for process-doc?
//
// Yes, but there needs to be no space between the comment and the item it comments on.
// Preamble/postamble can be handled using synthetic type definitions eg `type _preamble int`.
// Triple slashes would work but go fmt will whack them so should not use them.
package main

import (
	"go/ast"
	"go/parser"
	"go/token"
	"fmt"
	"log"
	"os"
)

// Comment on struct type zappa, should be a doc comment
type zappa struct {
	// comment before field x, should be a doc comment
	x int
	y int // comment after field y, should be a line comment
}

// Comment on function main, should be a doc comment
func main() {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(
		fset,
		"<stdin>",
		os.Stdin,
		parser.ParseComments|parser.SkipObjectResolution,
	)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(f)
	pc(f.Doc)

	for _, d := range f.Decls {
		if fd, ok := d.(*ast.FuncDecl); ok {
			pc(fd.Doc)
			continue
		}
		if gd, ok := d.(*ast.GenDecl); ok && gd.Tok == token.TYPE {
			pc(gd.Doc)
			for _, s := range gd.Specs {
				td := s.(*ast.TypeSpec)
				if st, ok := td.Type.(*ast.StructType); ok {
					fs := st.Fields
					for _, f := range fs.List {
						fmt.Println(f.Names[0].Name)
						pc(f.Doc)
					}
				}
			}
		}
	}
}

// Not a comment on function

func pc(d *ast.CommentGroup) {
	if d != nil {
		//fmt.Println(len(d.List))
		for _, c := range d.List {
			fmt.Println(c)
		}
	}
}
