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
		filepath.Join(tmpDir, "subdir", "file3.go"),
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

	// Test findGoFiles function
	foundFiles, err := findGoFiles(tmpDir)
	if err != nil {
		t.Fatalf("findGoFiles failed: %v", err)
	}

	// Check if the correct number of Go files were found
	expectedGoFiles := 3 // We created 3 .go files
	if len(foundFiles) != expectedGoFiles {
		t.Errorf("Expected to find %d Go files, but found %d", expectedGoFiles, len(foundFiles))
	}

	// Check if all found files have .go extension
	for _, file := range foundFiles {
		if !strings.HasSuffix(file, ".go") {
			t.Errorf("Found non-Go file: %s", file)
		}
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
		"ProcessData",
		"DoSomething",
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
