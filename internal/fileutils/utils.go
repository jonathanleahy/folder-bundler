package fileutils

import (
	"fmt"
	"os"
	"strings"
	"unicode/utf8"
)

func IsTextFile(content []byte) bool {
	if len(content) == 0 {
		return true
	}
	return utf8.Valid(content) && !containsNullByte(content)
}

func containsNullByte(data []byte) bool {
	for _, b := range data {
		if b == 0 {
			return true
		}
	}
	return false
}

func GetLanguage(extension string) string {
	languageMap := map[string]string{
		".go":         "go",
		".js":         "javascript",
		".jsx":        "javascript",
		".ts":         "typescript",
		".tsx":        "typescript",
		".py":         "python",
		".java":       "java",
		".cpp":        "cpp",
		".hpp":        "cpp",
		".h":          "cpp",
		".cc":         "cpp",
		".c":          "c",
		".cs":         "csharp",
		".html":       "html",
		".css":        "css",
		".scss":       "scss",
		".sass":       "scss",
		".md":         "markdown",
		".sh":         "bash",
		".bash":       "bash",
		".json":       "json",
		".xml":        "xml",
		".yaml":       "yaml",
		".yml":        "yaml",
		".sql":        "sql",
		".rs":         "rust",
		".rb":         "ruby",
		".php":        "php",
		".swift":      "swift",
		".kt":         "kotlin",
		".scala":      "scala",
		".r":          "r",
		".dart":       "dart",
		".lua":        "lua",
		".pl":         "perl",
		".gradle":     "gradle",
		".dockerfile": "dockerfile",
		".tf":         "terraform",
		".vue":        "vue",
		".ini":        "ini",
		".conf":       "ini",
		".toml":       "toml",
		".proto":      "protobuf",
	}

	return languageMap[strings.ToLower(extension)]
}

func BackupExistingFile(filename string) error {
	if _, err := os.Stat(filename); err == nil {
		backupName := filename + ".bak"
		if err := os.Rename(filename, backupName); err != nil {
			return fmt.Errorf("failed to create backup: %v", err)
		}
		fmt.Printf("Existing file backed up to %s\n", backupName)
	}
	return nil
}
