package collect

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/jonathanleahy/folder-bundler/internal/config"
	"github.com/jonathanleahy/folder-bundler/internal/fileutils"
)

type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	params       *config.Parameters
}

func hasHiddenComponent(path string) bool {
	parts := strings.Split(path, string(os.PathSeparator))
	for _, part := range parts {
		if strings.HasPrefix(part, ".") {
			return true
		}
	}
	return false
}

func ProcessDirectory(params *config.Parameters) error {
	collator := &FileCollator{
		currentPart:  1,
		baseFileName: fmt.Sprintf("%s_collated", filepath.Base(params.RootDir)),
		params:       params,
	}
	defer collator.closeCurrentFile()

	if err := collator.createNewFile(); err != nil {
		return err
	}

	return filepath.Walk(params.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(params.RootDir, path)
		if err != nil || relPath == "." {
			return nil
		}

		// Skip entire hidden paths unless explicitly included
		if !params.IncludeHidden && hasHiddenComponent(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		// Skip excluded directories
		if info.IsDir() && params.ExcludedDirs[info.Name()] {
			return filepath.SkipDir
		}

		return collator.processPath(relPath, info)
	})
}

func (fc *FileCollator) processPath(relPath string, info os.FileInfo) error {
	// Normalize path to use forward slashes for cross-platform compatibility
	normalizedPath := filepath.ToSlash(relPath)
	
	fullPath := filepath.Join(fc.params.RootDir, relPath)
	
	// Check if it's a symlink
	if info.Mode()&os.ModeSymlink != 0 {
		target, err := os.Readlink(fullPath)
		if err != nil {
			return fc.writeContent(fmt.Sprintf("## Symlink: %s (Error reading target: %v)\n\n", normalizedPath, err))
		}
		return fc.writeContent(fmt.Sprintf("## Symlink: %s\n\nTarget: %s\n\n", normalizedPath, target))
	}
	
	if info.IsDir() {
		return fc.writeContent(fmt.Sprintf("## Directory: %s\n\n", normalizedPath))
	}

	if info.Size() > fc.params.MaxOutputSize {
		return fc.writeContent(fmt.Sprintf("## File: %s (Skipped - Size exceeds 2MB)\n\n", normalizedPath))
	}

	metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
		normalizedPath, info.Size(), info.ModTime().Format(time.RFC3339))

	content, err := ioutil.ReadFile(fullPath)
	if err != nil {
		return err
	}

	if fileutils.IsTextFile(content) {
		language := fileutils.GetLanguage(filepath.Ext(relPath))
		contentStr := string(content)
		// Ensure content ends with exactly one newline before the closing backticks
		if !strings.HasSuffix(contentStr, "\n") {
			contentStr += "\n"
		}
		return fc.writeContent(fmt.Sprintf("%s```%s\n%s```\n\n", metadata, language, contentStr))
	}

	return fc.writeContent(fmt.Sprintf("%sBinary file - content not shown\n\n", metadata))
}

func (fc *FileCollator) writeContent(content string) error {
	contentSize := int64(len(content))
	if fc.currentSize+contentSize > fc.params.MaxOutputSize {
		fc.currentPart++
		if err := fc.createNewFile(); err != nil {
			return err
		}
	}

	_, err := fc.currentFile.WriteString(content)
	if err != nil {
		return err
	}

	fc.currentSize += contentSize
	return nil
}

func (fc *FileCollator) createNewFile() error {
	fc.closeCurrentFile()

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	file, err := os.Create(fileName)
	if err != nil {
		return err
	}

	fc.currentFile = file
	fc.currentSize = 0

	header := fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
		fc.currentPart, time.Now().Format(time.RFC3339), fc.params.RootDir)

	_, err = fc.currentFile.WriteString(header)
	return err
}

func (fc *FileCollator) closeCurrentFile() {
	if fc.currentFile != nil {
		fc.currentFile.Close()
		fc.currentFile = nil
	}
}
