package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFindGoFiles tests the findGoFiles function
func TestFindGoFiles(t *testing.T) {
	// Create a temporary directory for test files
	tmpDir, err := os.MkdirTemp("", "gosummarize-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create some test files
	files := []string{
		filepath.Join(tmpDir, "file1.go"),
		filepath.Join(tmpDir, "file2.go"),
		filepath.Join(tmpDir, "file3_test.go"),
		filepath.Join(tmpDir, "subdir", "file4.go"),
		filepath.Join(tmpDir, "subdir", "file5_test.go"),
		filepath.Join(tmpDir, "notgo.txt"),
	}

	// Create subdirectory
	err = os.Mkdir(filepath.Join(tmpDir, "subdir"), 0755)
	if err != nil {
		t.Fatalf("Failed to create subdirectory: %v", err)
	}

	// Create the files
	for _, file := range files {
		// Skip the .txt file when creating files
		if !strings.HasSuffix(file, ".go") {
			continue
		}
		dir := filepath.Dir(file)
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatalf("Failed to create directory %s: %v", dir, err)
		}
		if err := os.WriteFile(file, []byte("package test"), 0644); err != nil {
			t.Fatalf("Failed to create file %s: %v", file, err)
		}
	}

	// Test findGoFiles function without ignoring test files
	foundFiles, err := findGoFiles(tmpDir, false)
	if err != nil {
		t.Fatalf("findGoFiles failed: %v", err)
	}

	// Check if the correct number of Go files were found
	expectedGoFiles := 5 // We created 5 .go files
	if len(foundFiles) != expectedGoFiles {
		t.Errorf("Without -t flag: Expected to find %d Go files, but found %d", expectedGoFiles, len(foundFiles))
	}

	// Check if all found files have .go extension
	for _, file := range foundFiles {
		if !strings.HasSuffix(file, ".go") {
			t.Errorf("Found non-Go file: %s", file)
		}
	}

	// Test findGoFiles function while ignoring test files
	foundFiles, err = findGoFiles(tmpDir, true)
	if err != nil {
		t.Fatalf("findGoFiles with ignoreTests=true failed: %v", err)
	}

	// Check if the correct number of Go files were found
	expectedNonTestGoFiles := 3 // We created 3 non-test .go files
	if len(foundFiles) != expectedNonTestGoFiles {
		t.Errorf("With -t flag: Expected to find %d non-test Go files, but found %d", expectedNonTestGoFiles, len(foundFiles))
	}

	// Check if all found files have .go extension and none are test files
	for _, file := range foundFiles {
		if !strings.HasSuffix(file, ".go") {
			t.Errorf("Found non-Go file: %s", file)
		}
		if strings.HasSuffix(file, "_test.go") {
			t.Errorf("Found test file despite ignoreTests=true: %s", file)
		}
	}
}

// TestEmptyDirectoryWithSubdirs tests that findGoFiles works with empty top-level directories
func TestEmptyDirectoryWithSubdirs(t *testing.T) {
	// Create a temporary directory structure
	tmpDir, err := os.MkdirTemp("", "gosummarize-empty-test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a nested subdirectory structure with Go files only in subdirectories
	nestedDir := filepath.Join(tmpDir, "empty", "nested")
	if err := os.MkdirAll(nestedDir, 0755); err != nil {
		t.Fatalf("Failed to create nested directories: %v", err)
	}

	// Create a Go file in the nested directory only
	nestedGoFile := filepath.Join(nestedDir, "nested.go")
	if err := os.WriteFile(nestedGoFile, []byte("package nested"), 0644); err != nil {
		t.Fatalf("Failed to create Go file in nested directory: %v", err)
	}

	// Test finding files from the empty top directory
	emptyDir := filepath.Join(tmpDir, "empty")
	foundFiles, err := findGoFiles(emptyDir, false)
	if err != nil {
		t.Fatalf("findGoFiles failed on empty directory: %v", err)
	}

	// Should find 1 Go file in the nested subdirectory
	if len(foundFiles) != 1 {
		t.Errorf("Expected to find 1 Go file in subdirectories, but found %d", len(foundFiles))
	}

	// Verify the file found is actually our nested Go file
	if len(foundFiles) > 0 && foundFiles[0] != nestedGoFile {
		t.Errorf("Expected to find %s, but found %s", nestedGoFile, foundFiles[0])
	}
}

// TestSummarizeFile tests the summarizeFile function
func TestSummarizeFile(t *testing.T) {
	// Capture stdout to check output
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	// Call summarizeFile on our sample file
	testFile := "./testdata/sample.go"
	err := summarizeFile(testFile)
	if err != nil {
		t.Fatalf("summarizeFile failed: %v", err)
	}

	// Restore stdout and get the output
	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, err = buf.ReadFrom(r)
	if err != nil {
		t.Fatalf("Failed to read output: %v", err)
	}
	output := buf.String()

	// Verify the output contains expected elements
	expectedElements := []string{
		"<<<FILE_START>>>",
		"<<<FILE_END>>>",
		"PublicStruct",
		"PublicInterface",
		"PublicFunction",
		"func (s *StructWithInterface) ProcessData",
		"func (s *StructWithInterface) DoSomething",
		"PublicConst",
		"FirstConst",
		"SecondConst",
	}

	for _, expected := range expectedElements {
		if !strings.Contains(output, expected) {
			t.Errorf("Expected output to contain %q, but it didn't", expected)
		}
	}

	// Verify that private elements are not included
	unexpectedElements := []string{
		"privateFunction",
		"privateField",
		"privateConst",
	}

	for _, unexpected := range unexpectedElements {
		if strings.Contains(output, unexpected) {
			t.Errorf("Output should not contain %q, but it did", unexpected)
		}
	}
}

// TestMain runs the tests
func TestMain(m *testing.M) {
	// Setup code if needed
	code := m.Run()
	// Cleanup code if needed
	os.Exit(code)
}
