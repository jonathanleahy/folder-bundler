package collect

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"folder-bundler/internal/config"
	"folder-bundler/internal/fileutils"
)

type FileCollator struct {
	currentSize  int64
	currentPart  int
	currentFile  *os.File
	baseFileName string
	params       *config.Parameters
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
		if err != nil {
			return err
		}

		if relPath == "." {
			return nil
		}

		return collator.processPath(relPath, info)
	})
}

func (fc *FileCollator) processPath(relPath string, info os.FileInfo) error {
	if info.IsDir() {
		return fc.writeContent(fmt.Sprintf("## Directory: %s\n\n", relPath))
	}

	if info.Size() > fc.params.MaxOutputSize {
		return fc.writeContent(fmt.Sprintf("## File: %s (Skipped - Size exceeds 2MB)\n\n", relPath))
	}

	metadata := fmt.Sprintf("## File: %s\n\nSize: %d bytes\n\nLast Modified: %s\n\n",
		relPath, info.Size(), info.ModTime().Format(time.RFC3339))

	content, err := ioutil.ReadFile(filepath.Join(fc.params.RootDir, relPath))
	if err != nil {
		return err
	}

	if fileutils.IsTextFile(content) {
		language := fileutils.GetLanguage(filepath.Ext(relPath))
		return fc.writeContent(fmt.Sprintf("%s```%s\n%s\n```\n\n", metadata, language, string(content)))
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
	}
}
