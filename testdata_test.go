package sunatlib

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestValidateTestDataXMLs runs the UBLValidator against all XML files in the testdata directory.
// This ensures that our "golden masters" used for Beta testing are also structurally valid
// according to our local validation rules.
func TestValidateTestDataXMLs(t *testing.T) {
	v := NewUBLValidator()
	testDataDir := "testdata"

	files, err := os.ReadDir(testDataDir)
	if err != nil {
		t.Fatalf("failed to read testdata directory: %v", err)
	}

	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".xml") {
			continue
		}

		t.Run(file.Name(), func(t *testing.T) {
			path := filepath.Join(testDataDir, file.Name())
			content, err := os.ReadFile(path)
			if err != nil {
				t.Fatalf("failed to read file %s: %v", path, err)
			}

			// Perform structural validation
			if err := v.Validate(content); err != nil {
				t.Errorf("Validation failed for %s: %v", file.Name(), err)
			}
		})
	}
}
