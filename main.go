package main

import (
	"bytes"
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
func findGoFiles(root string) ([]string, error) {
	var files []string
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && strings.HasSuffix(path, ".go") {
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

			// Determine if it's a method or function
			var funcType string
			if fn.Recv == nil {
				funcType = "func"
			} else {
				funcType = "method"
			}

			// Get the full function signature
			var buf bytes.Buffer
			printer.Fprint(&buf, fset, fn.Type)
			signature := buf.String()

			// For methods, add the receiver
			if fn.Recv != nil {
				var recvBuf bytes.Buffer
				printer.Fprint(&recvBuf, fset, fn.Recv)
				recv := recvBuf.String()
				fmt.Printf("%s %s%s %s\n", funcType, recv, fn.Name.Name, signature)
			} else {
				fmt.Printf("%s %s%s\n", funcType, fn.Name.Name, signature)
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
	// Get the directory path from command line arguments
	if len(os.Args) < 2 {
		fmt.Println("Usage: gosummarize <directory>")
		os.Exit(1)
	}

	dirPath := os.Args[1]

	// Find all Go files in the directory
	files, err := findGoFiles(dirPath)
	if err != nil {
		fmt.Printf("Error finding Go files: %v\n", err)
		os.Exit(1)
	}

	if len(files) == 0 {
		fmt.Println("No Go files found in the specified directory")
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
