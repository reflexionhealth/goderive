package main

import (
	"bytes"
	"go/ast"
	"go/printer"
	"io"

	"github.com/reflexionhealth/goderive/derive"
)

type Data struct{ Type, Subtype string }

const Template = `
func (dups {{.Type}}) Unique() {{.Type}} {
	// Create a map of all unique elements.
	encountered := make(map[{{.Subtype}}]bool)
	for v := range dups {
		encountered[dups[v]] = true
	}

	// Place all keys from the map into a slice.
	result := make({{.Type}}, 0, len(encountered))
	for key, _ := range encountered {
		result = append(result, key)
	}

	return result
}`

func main() {
	targets := derive.Load()
	targets.WriteEach("unique_gen.go", func(out io.Writer, node ast.Node) {
		// get the array type as a string
		typeSpec, ok := node.(*ast.TypeSpec)
		derive.Assert(ok, `Cannot derive "Unique" for non-type declarations`)
		typ := typeSpec.Name.Name

		// get the element type as a string
		var buf bytes.Buffer
		arrayType, ok := typeSpec.Type.(*ast.ArrayType)
		derive.Assert(ok, `Cannot derive "Unique" for non-array/non-slice types`)
		printer.Fprint(&buf, targets.FileSet, arrayType.Elt)
		subtype := buf.String()

		// output the template
		derive.Template(out, Data{typ, subtype}, Template)
	})
}
