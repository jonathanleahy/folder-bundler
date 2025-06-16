package adapters

import (
	"bytes"
	"strings"
	"testing"
)

func TestTemplateCompression_InvalidUTF8(t *testing.T) {
	tc := NewTemplateCompression()
	
	// Test with invalid UTF-8 bytes
	invalidUTF8 := []byte{0xFF, 0xFE, 0xFD}
	content := append([]byte("valid line\n"), invalidUTF8...)
	content = append(content, []byte("\nanother valid line")...)
	
	// This should not panic
	compressed, metadata, err := tc.Compress(content)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// The compression should handle invalid UTF-8 gracefully
	if compressed == nil {
		t.Error("Expected compressed content, got nil")
	}
	
	if !strings.HasPrefix(metadata, "template:") {
		t.Errorf("Expected template metadata, got: %s", metadata)
	}
}

func TestTemplateCompression_ValidUTF8(t *testing.T) {
	tc := NewTemplateCompression()
	
	// Test with valid repeated patterns - need longer lines for template detection
	content := []byte(`function getDataFromServer() { return fetch('/api/data'); }
function getUserFromServer() { return fetch('/api/user'); }
function getConfigFromServer() { return fetch('/api/config'); }
function getSettingsFromServer() { return fetch('/api/settings'); }
function getProfileFromServer() { return fetch('/api/profile'); }`)
	
	compressed, metadata, err := tc.Compress(content)
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	// Check if compression found templates
	if strings.HasPrefix(metadata, "template:0") {
		t.Log("No templates found (content too short or not enough similarity)")
	} else {
		t.Logf("Found templates: %s", metadata)
		
		// Test decompression
		decompressed, err := tc.Decompress(compressed, metadata)
		if err != nil {
			t.Errorf("Decompression failed: %v", err)
		}
		
		if !bytes.Equal(content, decompressed) {
			t.Errorf("Decompressed content doesn't match original.\nOriginal:\n%s\nDecompressed:\n%s", 
				string(content), string(decompressed))
		}
	}
}

func TestTemplateCompression_MixedUTF8Content(t *testing.T) {
	tc := NewTemplateCompression()
	
	// Create content with some invalid UTF-8 in the middle
	validPart1 := "Line 1: valid content here\n"
	validPart2 := "Line 2: more valid content\n"
	validPart3 := "Line 3: and even more content\n"
	
	// Build content with invalid UTF-8 sequence
	var content bytes.Buffer
	content.WriteString(validPart1)
	content.Write([]byte{0xFF, 0xFE}) // Invalid UTF-8
	content.WriteString("\n")
	content.WriteString(validPart2)
	content.WriteString(validPart3)
	
	// This should not panic
	compressed, _, err := tc.Compress(content.Bytes())
	if err != nil {
		t.Errorf("Expected no error, got: %v", err)
	}
	
	if compressed == nil {
		t.Error("Expected compressed content, got nil")
	}
}