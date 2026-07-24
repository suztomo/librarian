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
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestUpdateSnippetMetadataVersion(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "snippet_metadata_test.json")
	input := `{
  "client_library": {
    "name": "google-cloud-asset-v1",
    "version": "",
    "language": "RUBY"
  }
}`
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := updateSnippetMetadataVersion(path, "1.2.3"); err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	want := `{
  "client_library": {
    "language": "RUBY",
    "name": "google-cloud-asset-v1",
    "version": "1.2.3"
  }
}`
	if diff := cmp.Diff(want, string(got)); diff != "" {
		t.Errorf("mismatch (-want +got):\n%s", diff)
	}
	var metadata SnippetMetadata
	if err := json.Unmarshal(got, &metadata); err != nil {
		t.Fatal(err)
	}
	if metadata.ClientLibrary.Version != "1.2.3" {
		t.Errorf("ClientLibrary.Version = %q, want %q", metadata.ClientLibrary.Version, "1.2.3")
	}
}

func TestUpdateSnippetMetadataVersion_Error(t *testing.T) {
	for _, test := range []struct {
		name    string
		setup   func(t *testing.T, path string)
		wantErr error
	}{
		{
			name: "missing client_library field",
			setup: func(t *testing.T, path string) {
				if err := os.WriteFile(path, []byte("{}"), 0o644); err != nil {
					t.Fatal(err)
				}
			},
			wantErr: errNoClientLibraryField,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "snippet_metadata_test.json")
			test.setup(t, path)
			gotErr := updateSnippetMetadataVersion(path, "1.2.3")
			if !errors.Is(gotErr, test.wantErr) {
				t.Errorf("updateSnippetMetadataVersion error = %v, wantErr %v", gotErr, test.wantErr)
			}
		})
	}
}
