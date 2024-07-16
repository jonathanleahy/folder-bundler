package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestDirectoryFileProcessor(t *testing.T) {
	tempDir, err := ioutil.TempDir("", "test_dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	createTestFiles(t, tempDir)

	os.Args = []string{"cmd", tempDir}
	main()

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	if _, err := os.Stat(outputPath); os.IsNotExist(err) {
		t.Fatalf("Output file was not created: %s", err)
	}

	content, err := ioutil.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("Failed to read output file: %v", err)
	}

	contentStr := string(content)

	t.Logf("Output file content:\n%s", contentStr)

	expectedContents := []string{
		"# Project Files Summary",
		"## Directory: subdir",
		"## File: test.go",
		"package main",
		"func main() {}",
		"## File: subdir", // We'll check for "subdir" and "subfile.txt" separately
		"subfile.txt",
		"This is a subfile",
	}

	for _, expected := range expectedContents {
		if !strings.Contains(contentStr, expected) {
			t.Errorf("Output does not contain expected content: %s", expected)
		}
	}

	// Check specifically for the subdirectory file, allowing for different path separators
	if !strings.Contains(contentStr, "## File: subdir/subfile.txt") && !strings.Contains(contentStr, "## File: subdir\\subfile.txt") {
		t.Errorf("Output does not contain the subdirectory file (subdir/subfile.txt or subdir\\subfile.txt)")
	}

	unexpectedContents := []string{
		"excluded.exe",
		".gitignore",
	}

	for _, unexpected := range unexpectedContents {
		if strings.Contains(contentStr, unexpected) {
			t.Errorf("Output contains unexpected content: %s", unexpected)
		}
	}

	// Print the directory structure for debugging
	t.Log("Directory structure:")
	filepath.Walk(tempDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(tempDir, path)
		t.Logf("%s", relPath)
		return nil
	})

	os.Remove(outputPath)
}

func createTestFiles(t *testing.T, root string) {
	err := ioutil.WriteFile(filepath.Join(root, "test.go"), []byte("package main\n\nfunc main() {}\n"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test.go: %v", err)
	}

	subdir := filepath.Join(root, "subdir")
	err = os.Mkdir(subdir, 0755)
	if err != nil {
		t.Fatalf("Failed to create subdir: %v", err)
	}
	err = ioutil.WriteFile(filepath.Join(subdir, "subfile.txt"), []byte("This is a subfile"), 0644)
	if err != nil {
		t.Fatalf("Failed to create subfile.txt: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(root, "excluded.exe"), []byte("This should be excluded"), 0644)
	if err != nil {
		t.Fatalf("Failed to create excluded.exe: %v", err)
	}

	err = ioutil.WriteFile(filepath.Join(root, ".gitignore"), []byte("*.exe"), 0644)
	if err != nil {
		t.Fatalf("Failed to create .gitignore: %v", err)
	}
}
