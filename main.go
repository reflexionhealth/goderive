package main

import (
	"fmt"
	"go/ast"
	"go/build"
	"go/parser"
	"go/token"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
)

type FileWithName struct {
	Name string
	File *ast.File
}

func main() {
	log.SetFlags(0)

	// Parse arguments of the form "{Name}={path/to/cmd}".
	commands := make(map[string]string)
	for _, arg := range os.Args[1:] {
		parts := strings.Split(arg, "=")
		if len(parts) != 2 {
			log.Fatalf(`error: arguments to derive must use the format "{Name}={path/to/cmd}" (unexpected "%s")\n`, arg)
		} else if len(parts[0]) == 0 || len(parts[1]) == 0 {
			log.Fatalf(`error: arguments to derive must use the format "{Name}={path/to/cmd}" (for "%s")\n`, arg)
		} else {
			commands[parts[0]] = parts[1]
		}
	}

	// Lookup the package being processed.
	abspath, _ := filepath.Abs(".")
	pkg, err := build.Default.ImportDir(abspath, 0) // TODO: accept optionally as argument
	if err != nil {
		log.Fatalf("error: can't import directory %s: %s\n", abspath, err)
	}

	// Parse the files in the package (with comments).
	allFiles := []FileWithName{}
	fileset := token.NewFileSet()
	for _, filename := range pkg.GoFiles {
		parsed, err := parser.ParseFile(fileset, filename, nil, parser.ParseComments)
		if err != nil {
			log.Fatalf("error: can't parse package: %s: %s\n", pkg.Name, err)
		} else {
			allFiles = append(allFiles, FileWithName{filename, parsed})
		}
	}

	// Sort declarations by derived trait.
	files := make(map[string][]string)
	identifiers := make(map[string][]string)
	for _, file := range allFiles {
		for _, decl := range file.File.Decls {
			ident, traits := parseDerive(decl)
			if traits != nil {
				for _, trait := range traits {
					files[trait] = append(files[trait], file.Name)
					identifiers[trait] = append(identifiers[trait], ident)
				}
			}
		}
	}

	// Run code generators.
	for trait, cmd := range commands {
		if len(files[trait]) > 0 {
			traitPkg, err := build.Import(cmd, pkg.Dir, 0)
			if err != nil {
				log.Fatalf("%s: %s\n", trait, err)
			}

			args := []string{"run"}
			for _, goFile := range traitPkg.GoFiles {
				args = append(args, filepath.Join(traitPkg.Dir, goFile))
			}
			args = append(args, "-pkg", pkg.Name)
			args = append(args, "-files", strings.Join(uniqueStrings(files[trait]), ","))
			args = append(args, "-names", strings.Join(uniqueStrings(identifiers[trait]), ","))

			fmt.Println("go", strings.Join(args, " "))
			cmd := exec.Command("go", args...)
			output, err := cmd.CombinedOutput()
			if err != nil {
				fmt.Print(string(output))
			}
		}
	}
}

var matchDerive = regexp.MustCompile(`\[Derive\(([^)]+)\)\]`)

func parseDerive(decl ast.Decl) (string, []string) {
	var name string
	var comments *ast.CommentGroup
	switch d := decl.(type) {
	case *ast.FuncDecl:
		name = d.Name.Name
		comments = d.Doc
	case *ast.GenDecl:
		// FIXME: Remove this hack (make it more generic!)
		if d.Tok == token.TYPE {
			name = d.Specs[0].(*ast.TypeSpec).Name.Name
			comments = d.Doc
		}
	}

	if comments != nil {
		for _, comment := range comments.List {
			matches := matchDerive.FindStringSubmatch(comment.Text)
			if matches != nil {
				return name, parseTraits(matches[1])
			}
		}
	}

	return "", nil
}

func parseTraits(deriveArgs string) []string {
	untrimmedTraits := strings.Split(deriveArgs, ",")
	traits := make([]string, len(untrimmedTraits))
	for _, untrimmed := range untrimmedTraits {
		traits = append(traits, strings.TrimSpace(untrimmed))
	}
	return traits
}

func uniqueStrings(elements []string) []string {
	// Create a map of all unique elements.
	encountered := make(map[string]bool)
	for v := range elements {
		encountered[elements[v]] = true
	}

	// Place all keys from the map into a slice.
	result := make([]string, 0, len(encountered))
	for key, _ := range encountered {
		result = append(result, key)
	}
	return result
}
