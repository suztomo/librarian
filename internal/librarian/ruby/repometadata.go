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
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/googleapis/librarian/internal/repometadata"
)

func updateRepoMetadata(outputDir, stagingDir, gemName string) error {
	stagingPath := filepath.Join(stagingDir, gemName, ".repo-metadata.json")
	outputPath := filepath.Join(outputDir, gemName, ".repo-metadata.json")
	writeDir := filepath.Join(stagingDir, gemName)

	stagingData, err := os.ReadFile(stagingPath)
	if err != nil {
		if !errors.Is(err, fs.ErrNotExist) {
			return fmt.Errorf("failed to read staging repo metadata: %w", err)
		}
		stagingPath = filepath.Join(stagingDir, ".repo-metadata.json")
		outputPath = filepath.Join(outputDir, ".repo-metadata.json")
		writeDir = stagingDir
		stagingData, err = os.ReadFile(stagingPath)
		if err != nil {
			if errors.Is(err, fs.ErrNotExist) {
				return nil
			}
			return fmt.Errorf("failed to read staging repo metadata: %w", err)
		}
	}

	var newMap map[string]any
	if err := json.Unmarshal(stagingData, &newMap); err != nil {
		return fmt.Errorf("failed to unmarshal staging repo metadata: %w", err)
	}
	if newMap == nil {
		newMap = make(map[string]any)
	}

	if outputData, err := os.ReadFile(outputPath); err == nil {
		var existingMap map[string]any
		if err := json.Unmarshal(outputData, &existingMap); err == nil && existingMap != nil {
			for key, val := range existingMap {
				if _, ok := newMap[key]; !ok {
					newMap[key] = val
				}
			}
			if rl, ok := existingMap["release_level"].(string); ok && rl != "" {
				newMap["release_level"] = rl
			}
			if lt, ok := existingMap["library_type"].(string); ok && lt != "" {
				newMap["library_type"] = lt
			}
		}
	}

	return repometadata.WriteJSON(newMap, "    ", writeDir, ".repo-metadata.json")
}
