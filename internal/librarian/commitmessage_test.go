// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	https://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package librarian

import (
	"testing"
)

func TestIsReleaseWorthy(t *testing.T) {
	tests := []struct {
		name      string
		message   *CommitMessage
		libraryId string
		expected  bool
	}{

		{
			name:      "No message",
			message:   &CommitMessage{},
			libraryId: "lib-B",
			expected:  false,
		},
		// --- NoTriggerLibraries specific cases ---
		{
			name:      "Explicitly no-trigger: library ID in NoTriggerLibraries",
			message:   &CommitMessage{NoTriggerLibraries: []string{"lib-A"}},
			libraryId: "lib-A",
			expected:  false,
		},
		{
			name:      "Explicitly no-trigger (with feature): library ID in NoTriggerLibraries overrides features",
			message:   &CommitMessage{Features: []string{"new feature"}, NoTriggerLibraries: []string{"lib-A"}},
			libraryId: "lib-A",
			expected:  false,
		},
		{
			name:      "Explicitly no-trigger (with fix): library ID in NoTriggerLibraries overrides fixes",
			message:   &CommitMessage{Fixes: []string{"bug fix"}, NoTriggerLibraries: []string{"lib-A"}},
			libraryId: "lib-A",
			expected:  false,
		},
		{
			name:      "Explicitly no-trigger (with breaking): library ID in NoTriggerLibraries overrides breaking",
			message:   &CommitMessage{Breaking: true, NoTriggerLibraries: []string{"lib-A"}},
			libraryId: "lib-A",
			expected:  false,
		},

		// --- TriggerLibraries specific cases ---
		{
			name:      "Explicitly trigger: library ID in TriggerLibraries",
			message:   &CommitMessage{TriggerLibraries: []string{"lib-B"}},
			libraryId: "lib-B",
			expected:  true,
		},
		{
			name:      "Explicitly trigger (with no-trigger): TriggerLibNoTriggerLibrariesraries takes precedence",
			message:   &CommitMessage{TriggerLibraries: []string{"lib-B"}, NoTriggerLibraries: []string{"lib-B"}},
			libraryId: "lib-B",
			expected:  false,
		},

		// --- Default logic (Features || Fixes || Breaking) ---
		{
			name:      "Default: Feature present, no explicit trigger/no-trigger",
			message:   &CommitMessage{Features: []string{"new feature"}},
			libraryId: "lib-C",
			expected:  true,
		},
		{
			name:      "Default: Fix present, no explicit trigger/no-trigger",
			message:   &CommitMessage{Fixes: []string{"bug fix"}},
			libraryId: "lib-D",
			expected:  true,
		},
		{
			name:      "Default: Breaking change, no explicit trigger/no-trigger",
			message:   &CommitMessage{Breaking: true},
			libraryId: "lib-E",
			expected:  true,
		},
		{
			name:      "Default: Features and Fixes present",
			message:   &CommitMessage{Features: []string{"f1"}, Fixes: []string{"fx1"}},
			libraryId: "lib-F",
			expected:  true,
		},
		{
			name:      "Default: Features, Fixes, and Breaking change present",
			message:   &CommitMessage{Features: []string{"f1"}, Fixes: []string{"fx1"}, Breaking: true},
			libraryId: "lib-G",
			expected:  true,
		},

		// --- Default: Not release worthy ---
		{
			name:      "Default: Docs and other fields populated, not release worthy",
			message:   &CommitMessage{Docs: []string{"update docs"}, PiperOrigins: []string{"p1"}, SourceLinks: []string{"s1"}, CommitHash: "abc"},
			libraryId: "lib-H",
			expected:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := IsReleaseWorthy(tt.message, tt.libraryId)

			if actual != tt.expected {
				t.Errorf("Failed test name: \"%s\". IsReleaseWorthy(message: %+v, libraryId: %q) = %t, want %t",
					tt.name, tt.message, tt.libraryId, actual, tt.expected)
			}
		})
	}
}
