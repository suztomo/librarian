// Copyright 2026 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package ruby

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUpdateRepoMetadata(t *testing.T) {
	for _, test := range []struct {
		name        string
		gemName     string
		stagingJSON string
		outputJSON  string
		wantJSON    string
	}{
		{
			name:        "preserves release_level, library_type, and custom fields from existing metadata",
			gemName:     "google-cloud-asset-v1",
			stagingJSON: `{"api_id":"cloudasset.googleapis.com","release_level":"unreleased","ruby-cloud-description":"new description","library_type":"GAPIC_AUTO"}`,
			outputJSON:  `{"api_id":"cloudasset.googleapis.com","release_level":"stable","product_documentation":"https://cloud.google.com/asset-inventory/","ruby-cloud-product-url":"https://cloud.google.com/asset-inventory/","library_type":"GAPIC_COMBO"}`,
			wantJSON: `{
    "api_id": "cloudasset.googleapis.com",
    "library_type": "GAPIC_COMBO",
    "product_documentation": "https://cloud.google.com/asset-inventory/",
    "release_level": "stable",
    "ruby-cloud-description": "new description",
    "ruby-cloud-product-url": "https://cloud.google.com/asset-inventory/"
}`,
		},
		{
			name:        "new library without existing metadata uses staging values",
			gemName:     "google-cloud-vision-v1",
			stagingJSON: `{"api_id":"vision.googleapis.com","release_level":"unreleased","library_type":"GAPIC_AUTO"}`,
			outputJSON:  "",
			wantJSON: `{
    "api_id": "vision.googleapis.com",
    "library_type": "GAPIC_AUTO",
    "release_level": "unreleased"
}`,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			outputDir := t.TempDir()
			stagingDir := t.TempDir()

			stagingFile := filepath.Join(stagingDir, ".repo-metadata.json")
			if err := os.WriteFile(stagingFile, []byte(test.stagingJSON), 0o644); err != nil {
				t.Fatal(err)
			}

			if test.outputJSON != "" {
				outputFile := filepath.Join(outputDir, ".repo-metadata.json")
				if err := os.WriteFile(outputFile, []byte(test.outputJSON), 0o644); err != nil {
					t.Fatal(err)
				}
			}

			if err := updateRepoMetadata(outputDir, stagingDir, test.gemName); err != nil {
				t.Fatal(err)
			}

			gotData, err := os.ReadFile(stagingFile)
			if err != nil {
				t.Fatal(err)
			}

			var gotMap, wantMap map[string]any
			if err := json.Unmarshal(gotData, &gotMap); err != nil {
				t.Fatal(err)
			}
			if err := json.Unmarshal([]byte(test.wantJSON), &wantMap); err != nil {
				t.Fatal(err)
			}

			if diff := cmp.Diff(wantMap, gotMap); diff != "" {
				t.Errorf("mismatch (-want +got):\n%s", diff)
			}
		})
	}
}

func TestUpdateRepoMetadata_Subdirectory(t *testing.T) {
	outputDir := t.TempDir()
	stagingDir := t.TempDir()
	gemName := "google-cloud-asset-v1"

	if err := os.MkdirAll(filepath.Join(stagingDir, gemName), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(outputDir, gemName), 0o755); err != nil {
		t.Fatal(err)
	}

	stagingFile := filepath.Join(stagingDir, gemName, ".repo-metadata.json")
	stagingJSON := `{"api_id":"cloudasset.googleapis.com","release_level":"unreleased","library_type":"GAPIC_AUTO"}`
	if err := os.WriteFile(stagingFile, []byte(stagingJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	outputFile := filepath.Join(outputDir, gemName, ".repo-metadata.json")
	outputJSON := `{"api_id":"cloudasset.googleapis.com","release_level":"stable","product_documentation":"https://cloud.google.com/asset-inventory/","library_type":"GAPIC_COMBO"}`
	if err := os.WriteFile(outputFile, []byte(outputJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := updateRepoMetadata(outputDir, stagingDir, gemName); err != nil {
		t.Fatal(err)
	}

	gotData, err := os.ReadFile(stagingFile)
	if err != nil {
		t.Fatal(err)
	}

	var gotMap, wantMap map[string]any
	if err := json.Unmarshal(gotData, &gotMap); err != nil {
		t.Fatal(err)
	}
	wantJSON := `{
    "api_id": "cloudasset.googleapis.com",
    "library_type": "GAPIC_COMBO",
    "product_documentation": "https://cloud.google.com/asset-inventory/",
    "release_level": "stable"
}`
	if err := json.Unmarshal([]byte(wantJSON), &wantMap); err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(wantMap, gotMap); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
}

func TestUpdateRepoMetadata_NoStagingFile(t *testing.T) {
	outputDir := t.TempDir()
	stagingDir := t.TempDir()

	if err := updateRepoMetadata(outputDir, stagingDir, "google-cloud-asset-v1"); err != nil {
		t.Fatal(err)
	}
}
