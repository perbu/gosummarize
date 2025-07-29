package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/printer"
	"go/token"
	"os"
	"path/filepath"
	"strings"
)

// findGoFiles returns a slice of all Go files in the given directory and its subdirectories
// If ignoreTests is true, files ending with _test.go will be ignored
func findGoFiles(root string, ignoreTests bool) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
			// Skip test files if ignoreTests flag is set
			if ignoreTests && strings.HasSuffix(path, "_test.go") {
				return nil
			}
			files = append(files, path)
		}
		return nil
	})
	return files, err
}

// summarizeFile parses a Go file and extracts information about exported declarations
func summarizeFile(filePath string) error {
	fset := token.NewFileSet()

	// Parse the file
	file, err := parser.ParseFile(fset, filePath, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("error parsing %s: %v", filePath, err)
	}

	// Print clear file start marker with filename
	fmt.Printf("<<<FILE_START>>> %s\n\n", filePath)

	// Extract exported declarations
	for _, decl := range file.Decls {
		processDeclaration(decl, file, fset)
	}

	// Print clear file end marker
	fmt.Printf("<<<FILE_END>>> %s\n\n", filePath)
	return nil
}

// processDeclaration extracts and prints information about exported declarations
func processDeclaration(decl ast.Decl, file *ast.File, fset *token.FileSet) {
	// Handle functions and methods
	if fn, ok := decl.(*ast.FuncDecl); ok {
		if fn.Name.IsExported() {
			doc := ""
			if fn.Doc != nil {
				doc = fn.Doc.Text()
			}

			// We no longer need to distinguish between method/func in the output
			// as we're using the standard Go syntax format for both

			// Get the full function signature
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, fn.Type)
			signature := buf.String()

			// Remove "func" from the signature as we'll add it manually
			signature = strings.TrimPrefix(signature, "func")

			// For methods, add the receiver and display in a format similar to source code
			if fn.Recv != nil {
				// Let's manually extract the receiver details
				recvList := fn.Recv.List[0] // Get the first (and only) receiver parameter

				// Get the type of the receiver
				var typeNameBuf bytes.Buffer
				printer.Fprint(&typeNameBuf, fset, recvList.Type)
				recvType := typeNameBuf.String()

				// Get the variable name of the receiver (if any)
				recvVarName := ""
				if len(recvList.Names) > 0 {
					recvVarName = recvList.Names[0].Name
				}

				// Format like Go source: func (s *Type) Method()
				if recvVarName != "" {
					fmt.Printf("func (%s %s) %s%s\n", recvVarName, recvType, fn.Name.Name, signature)
				} else {
					fmt.Printf("func (%s) %s%s\n", recvType, fn.Name.Name, signature)
				}
			} else {
				fmt.Printf("func %s%s\n", fn.Name.Name, signature)
			}

			if doc != "" {
				fmt.Printf("    %s\n", strings.TrimSpace(doc))
			}
			fmt.Println()
		}
		return
	}

	// Handle types, vars, and consts
	if gen, ok := decl.(*ast.GenDecl); ok {
		for _, spec := range gen.Specs {
			switch s := spec.(type) {
			case *ast.TypeSpec:
				if s.Name.IsExported() {
					doc := ""
					if gen.Doc != nil {
						doc = gen.Doc.Text()
					}

					// For structs, we need to filter out unexported fields
					if structType, ok := s.Type.(*ast.StructType); ok && structType.Fields != nil {
						fmt.Printf("type %s struct {\n", s.Name.Name)

						// Print only exported fields
						for _, field := range structType.Fields.List {
							if len(field.Names) > 0 && field.Names[0].IsExported() {
								var fieldBuf bytes.Buffer
								printer.Fprint(&fieldBuf, fset, field.Type)
								fieldType := fieldBuf.String()

								// Print field doc if exists
								if field.Doc != nil && field.Doc.Text() != "" {
									fieldDoc := strings.TrimSpace(field.Doc.Text())
									fmt.Printf("\t// %s\n", fieldDoc)
								}

								fmt.Printf("\t%s\t%s\n", field.Names[0].Name, fieldType)
							}
						}
						fmt.Printf("}\n")
					} else {
						// For non-struct types, print the full definition
						var buf bytes.Buffer
						printer.Fprint(&buf, fset, s.Type)
						typeDefinition := buf.String()

						fmt.Printf("type %s %s\n", s.Name.Name, typeDefinition)
					}

					if doc != "" {
						fmt.Printf("    %s\n", strings.TrimSpace(doc))
					}
					fmt.Println()
				}

			case *ast.ValueSpec:
				for i, name := range s.Names {
					if name.IsExported() {
						doc := ""
						if gen.Doc != nil {
							doc = gen.Doc.Text()
						}

						var declType string
						if gen.Tok == token.CONST {
							declType = "const"
						} else {
							declType = "var"
						}

						// Get the type if present
						typeStr := ""
						if s.Type != nil {
							var buf bytes.Buffer
							printer.Fprint(&buf, fset, s.Type)
							typeStr = " " + buf.String()
						}

						// Get the value if present
						valueStr := ""
						if i < len(s.Values) && s.Values[i] != nil {
							var buf bytes.Buffer
							printer.Fprint(&buf, fset, s.Values[i])
							valueStr = " = " + buf.String()
						}

						fmt.Printf("%s %s%s%s\n", declType, name.Name, typeStr, valueStr)
						if doc != "" {
							fmt.Printf("    %s\n", strings.TrimSpace(doc))
						}
						fmt.Println()
					}
				}
			}
		}
	}
}

func main() {
	// Define command-line flags
	ignoreTests := flag.Bool("t", false, "Ignore test files (files ending with _test.go)")
	flag.Parse()

	// Get the directory path from command line arguments
	args := flag.Args()
	if len(args) < 1 {
		fmt.Println("Usage: gosummarize [-t] <directory>")
		flag.PrintDefaults()
		os.Exit(1)
	}

	dirPath := args[0]

	// Find all Go files in the directory and subdirectories
	files, err := findGoFiles(dirPath, *ignoreTests)
	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No Go files found in the specified directory or its subdirectories")
		os.Exit(0)
	}

	// Process each file
	for _, file := range files {
		err := summarizeFile(file)
		if err != nil {
			fmt.Println(err)
		}
	}
}
