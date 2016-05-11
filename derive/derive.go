package derive

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"golang.org/x/tools/imports"
)

type Targets struct {
	Pkg     string
	FileSet *token.FileSet
	Files   []string
	Names   []string
	Nodes   []ast.Node
}

func Load() *Targets {
	log.SetFlags(0)
	var pkgFlag = flag.String("pkg", "", "the package to generate code for")
	var filesFlag = flag.String("files", "", "a comma-separated list of filenames")
	var namesFlag = flag.String("names", "", "a comma-separated list of identifiers")
	flag.Parse()

	targets := &Targets{Pkg: *pkgFlag}
	targets.Names = strings.Split(*namesFlag, ",")
	targets.Files = strings.Split(*filesFlag, ",")
	targets.FileSet = token.NewFileSet()

	files := []*ast.File{}
	for _, filename := range targets.Files {
		parsed, err := parser.ParseFile(targets.FileSet, filename, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("can't parse file: %s: %s", filename, err)
		}
		files = append(files, parsed)
	}

	for _, file := range files {
		ast.Walk(targets, file)
	}

	return targets
}

func (t *Targets) Include(name string) bool {
	for _, n := range t.Names {
		if n == name {
			return true
		}
	}
	return false
}

// Implements ast.Visit interface
func (t *Targets) Visit(node ast.Node) ast.Visitor {
	switch n := node.(type) {
	case *ast.TypeSpec:
		if t.Include(n.Name.Name) {
			t.Nodes = append(t.Nodes, node)
			return nil
		}
	case *ast.FuncDecl:
		if t.Include(n.Name.Name) {
			t.Nodes = append(t.Nodes, node)
			return nil
		}
	}
	return t
}

func (t *Targets) WriteEach(file string, transform func(io.Writer, ast.Node)) {
	// for node := range derive.Nodes() {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "package %s\n\n", t.Pkg)
	for i, node := range t.Nodes {
		if i > 0 {
			buf.WriteRune('\n')
		}

		transform(&buf, node)
		buf.WriteRune('\n')
	}

	os.Remove(file)
	out, err := imports.Process(file, buf.Bytes(), nil)
	if err != nil {
		log.Fatalf("can't format generated code: %s\n", err)
	}
	ioutil.WriteFile(file, out, 0600)
}

func Assert(ok bool, format string, args ...interface{}) {
	if !ok {
		log.Fatalf(format, args...)
	}
}

func Template(out io.Writer, data interface{}, format string) {
	t := template.New("scanMethod")
	if _, err := t.Parse(strings.TrimSpace(format)); err != nil {
		log.Fatalf("can't parse code template: %s", err)
	}
	t.Execute(out, data)
}
