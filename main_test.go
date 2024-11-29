package main

import (
	"bytes"
	"fmt"
	"io"
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

	// Test with explicit directory
	testExplicitDirectory(t, tempDir)

	// Test with current directory
	testCurrentDirectory(t)

	// Test backup functionality
	testBackupFunctionality(t, tempDir)
}

func testExplicitDirectory(t *testing.T, tempDir string) {
	os.Args = []string{"cmd", tempDir}
	output := captureOutput(main)
	t.Logf("Debug output:\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	verifyOutput(t, outputPath, tempDir, false)
}

func testCurrentDirectory(t *testing.T) {
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}
	defer os.Chdir(currentDir)

	tempDir, err := ioutil.TempDir("", "test_current_dir")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)

	os.Chdir(tempDir)
	createTestFiles(t, tempDir)

	os.Args = []string{"cmd"}
	output := captureOutput(main)
	t.Logf("Debug output:\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	verifyOutput(t, outputPath, tempDir, false)
}

func testBackupFunctionality(t *testing.T, tempDir string) {
	os.Args = []string{"cmd", tempDir}

	// Run once to create the initial file
	output := captureOutput(main)
	t.Logf("Debug output (first run):\n%s", output)

	// Run again to test backup functionality
	output = captureOutput(main)
	t.Logf("Debug output (second run):\n%s", output)

	outputFileName := fmt.Sprintf("%s_collated.md", filepath.Base(tempDir))
	outputPath := filepath.Join(".", outputFileName)
	backupPath := outputPath + ".bak"

	// Verify both the new file and the backup
	verifyOutput(t, outputPath, tempDir, false)
	verifyOutput(t, backupPath, tempDir, true)
}

func verifyOutput(t *testing.T, outputPath string, tempDir string, isBackup bool) {
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
		"## File: binary.dat",
		"Unable to parse: File appears to be in a non-text format",
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

	// Check for backup-specific content
	if isBackup {
		expectedBackupMessage := "# Project Files Summary"
		if !strings.Contains(contentStr, expectedBackupMessage) {
			t.Errorf("Backup file does not contain the expected content: %s", expectedBackupMessage)
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

	// Create a binary file
	binaryData := []byte{0x00, 0x01, 0x02, 0x03}
	err = ioutil.WriteFile(filepath.Join(root, "binary.dat"), binaryData, 0644)
	if err != nil {
		t.Fatalf("Failed to create binary.dat: %v", err)
	}

	t.Logf("Test files created in %s", root)
	filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		relPath, _ := filepath.Rel(root, path)
		t.Logf("Created: %s", relPath)
		return nil
	})
}

func captureOutput(f func()) string {
	old := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	f()

	w.Close()
	os.Stdout = old

	var buf bytes.Buffer
	io.Copy(&buf, r)
	return buf.String()
}
