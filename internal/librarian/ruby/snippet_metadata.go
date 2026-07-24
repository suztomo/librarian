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
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

var (
	errNoClientLibraryField     = errors.New("no client_library field at the top level")
	errSnippetMetadataDirectory = errors.New("expected file; was a directory")
	errSnippetMetadataLink      = errors.New("expected regular file; was a link")
)

// SnippetMetadata represents the top-level structure of a Ruby snippet metadata file.
type SnippetMetadata struct {
	ClientLibrary ClientLibrary `json:"client_library"`
}

// ClientLibrary represents the client_library object inside Ruby snippet metadata.
type ClientLibrary struct {
	Name     string `json:"name,omitempty"`
	Version  string `json:"version,omitempty"`
	Language string `json:"language,omitempty"`
}

func updateAllSnippetMetadataVersions(dir, version string) error {
	files, err := findAllSnippetMetadataFiles(dir)
	if err != nil {
		return err
	}
	for _, file := range files {
		if err := updateSnippetMetadataVersion(file, version); err != nil {
			return err
		}
	}
	return nil
}

func updateSnippetMetadataVersion(path, version string) error {
	content, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("error reading snippet metadata file %s: %w", path, err)
	}
	var metadata map[string]any
	if err := json.Unmarshal(content, &metadata); err != nil {
		return fmt.Errorf("error parsing snippet metadata file %s: %w", path, err)
	}
	clientLib, ok := metadata["client_library"].(map[string]any)
	if !ok {
		return fmt.Errorf("error updating snippet metadata file %s: %w", path, errNoClientLibraryField)
	}
	clientLib["version"] = version

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(metadata); err != nil {
		return fmt.Errorf("error encoding snippet metadata file %s: %w", path, err)
	}
	out := buf.Bytes()
	if len(out) > 0 && out[len(out)-1] == '\n' {
		out = out[:len(out)-1]
	}
	if err := os.WriteFile(path, out, 0o644); err != nil {
		return fmt.Errorf("error writing snippet metadata file %s: %w", path, err)
	}
	return nil
}

func findAllSnippetMetadataFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		name := entry.Name()
		if !strings.HasPrefix(name, "snippet_metadata") || !strings.HasSuffix(name, ".json") {
			return nil
		}
		if entry.IsDir() {
			return fmt.Errorf("error for possible snippet metadata file %s: %w", path, errSnippetMetadataDirectory)
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("error for possible snippet metadata file %s: %w", path, errSnippetMetadataLink)
		}
		files = append(files, path)
		return nil
	})
	if err != nil {
		return nil, err
	}
	return files, nil
}
