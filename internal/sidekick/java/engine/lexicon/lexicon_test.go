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

package lexicon

import "testing"

func TestIsReservedKeyword(t *testing.T) {
	tests := []struct {
		input string
		want  bool
	}{
		{"class", true},
		{"package", true},
		{"int", true},
		{"public", true},
		{"return", true},
		{"myVariable", false},
		{"Secret", false},
		{"", false},
	}

	for _, tt := range tests {
		got := IsReservedKeyword(tt.input)
		if got != tt.want {
			t.Errorf("IsReservedKeyword(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}

func TestEscapeIdentifier(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"package", "package_"},
		{"class", "class_"},
		{"int", "int_"},
		{"name", "name"},
		{"parent", "parent"},
	}

	for _, tt := range tests {
		got := EscapeIdentifier(tt.input)
		if got != tt.want {
			t.Errorf("EscapeIdentifier(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToLowerCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"project_id", "projectId"},
		{"SecretManagerService", "secretManagerService"},
		{"package", "package_"},
		{"", ""},
	}

	for _, tt := range tests {
		got := ToLowerCamel(tt.input)
		if got != tt.want {
			t.Errorf("ToLowerCamel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestToUpperCamel(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"project_id", "ProjectId"},
		{"secret_manager_service", "SecretManagerService"},
		{"secret", "Secret"},
		{"", ""},
	}

	for _, tt := range tests {
		got := ToUpperCamel(tt.input)
		if got != tt.want {
			t.Errorf("ToUpperCamel(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestJavaPackageFromProto(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"google.cloud.secretmanager.v1", "com.google.cloud.secretmanager.v1"},
		{".google.cloud.secretmanager.v1", "com.google.cloud.secretmanager.v1"},
		{"com.google.cloud.secretmanager.v1", "com.google.cloud.secretmanager.v1"},
		{"custom.company.v1", "com.custom.company.v1"},
	}

	for _, tt := range tests {
		got := JavaPackageFromProto(tt.input)
		if got != tt.want {
			t.Errorf("JavaPackageFromProto(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSanitizeComment(t *testing.T) {
	input := "This is a comment with /* and */ nested markers."
	want := "This is a comment with / * and * / nested markers."
	got := SanitizeComment(input)
	if got != want {
		t.Errorf("SanitizeComment(%q) = %q, want %q", input, got, want)
	}
}
