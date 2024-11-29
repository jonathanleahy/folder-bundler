package collect

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"time"

	"folder-bundler/internal/config"
	"folder-bundler/internal/fileutils"

	ignore "github.com/sabhiram/go-gitignore"
)

type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	rootDir      string
	params       *config.Parameters
}

func NewFileCollator(rootDir string, params *config.Parameters) *FileCollator {
	return &FileCollator{
		currentPart:  1,
		baseFileName: fmt.Sprintf("%s_collated", filepath.Base(rootDir)),
		rootDir:      rootDir,
		params:       params,
	}
}

func ProcessDirectory(params *config.Parameters) error {
	collator := NewFileCollator(params.RootDir, params)
	defer collator.Close()

	if err := collator.createNewFile(); err != nil {
		return fmt.Errorf("error creating output file: %v", err)
	}

	var ignorer *ignore.GitIgnore
	if !params.SkipGitignore {
		gitignoreFile := filepath.Join(params.RootDir, ".gitignore")
		if _, err := os.Stat(gitignoreFile); err == nil {
			ignorer, _ = ignore.CompileIgnoreFile(gitignoreFile)
		}
	}

	return filepath.Walk(params.RootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(params.RootDir, path)
		if err != nil {
			return err
		}

		if !params.IncludeHidden && strings.HasPrefix(filepath.Base(path), ".") {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if ignorer != nil && ignorer.MatchesPath(relPath) {
			if info.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		return collator.processPath(relPath, info)
	})
}

func (fc *FileCollator) processPath(relPath string, info os.FileInfo) error {
	if info.IsDir() {
		if fc.params.ExcludedDirs[info.Name()] {
			return filepath.SkipDir
		}
		return fc.writeContent(fmt.Sprintf("## Directory: %s\n\n", relPath))
	}

	if fc.params.ExcludedFiles[info.Name()] {
		return nil
	}

	ext := strings.ToLower(filepath.Ext(relPath))
	if fc.params.ExcludedExts[ext] {
		return nil
	}

	return fc.processFile(relPath, info, ext)
}

func (fc *FileCollator) processFile(relPath string, info os.FileInfo, ext string) error {
	if info.Size() > fc.params.MaxFileSize {
		return fc.writeContent(fmt.Sprintf("## File: %s (Skipped - Size: %d bytes)\n\n", relPath, info.Size()))
	}

	metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
		relPath, info.Size(), info.ModTime().Format(time.RFC3339))

	if err := fc.writeContent(metadata); err != nil {
		return err
	}

	content, err := ioutil.ReadFile(filepath.Join(fc.rootDir, relPath))
	if err != nil {
		return err
	}

	if fileutils.IsTextFile(content) {
		language := fileutils.GetLanguage(ext)
		return fc.writeContent(fmt.Sprintf("```%s\n%s\n```\n\n", language, string(content)))
	}

	return fc.writeContent("Unable to parse: File appears to be in a non-text format.\n\n")
}

func (fc *FileCollator) createNewFile() error {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}

	fileName := fmt.Sprintf("%s_part%d.md", fc.baseFileName, fc.currentPart)
	if err := fileutils.BackupExistingFile(fileName); err != nil {
		return err
	}

	file, err := os.Create(fileName)
	if err != nil {
		return fmt.Errorf("failed to create new file part: %v", err)
	}

	fc.currentFile = file
	fc.currentSize = 0

	header := fmt.Sprintf("# Project Files Summary - Part %d\n\nGenerated on: %s\n\nRoot Directory: %s\n\n---\n\n",
		fc.currentPart, time.Now().Format(time.RFC3339), fc.rootDir)

	_, err = fc.currentFile.WriteString(header)
	return err
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

func (fc *FileCollator) Close() {
	if fc.currentFile != nil {
		fc.currentFile.Close()
	}
}
