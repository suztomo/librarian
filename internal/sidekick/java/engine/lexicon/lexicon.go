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

// Package lexicon provides Java lexicon utilities, naming conversions, and keyword escaping.
package lexicon

import (
	"strings"
	"unicode"

	"github.com/iancoleman/strcase"
)

var javaKeywords = map[string]bool{
	"abstract":     true,
	"assert":       true,
	"boolean":      true,
	"break":        true,
	"byte":         true,
	"case":         true,
	"catch":        true,
	"char":         true,
	"class":        true,
	"const":        true,
	"continue":     true,
	"default":      true,
	"do":           true,
	"double":       true,
	"else":         true,
	"enum":         true,
	"extends":      true,
	"final":        true,
	"finally":      true,
	"float":        true,
	"for":          true,
	"goto":         true,
	"if":           true,
	"implements":   true,
	"import":       true,
	"instanceof":   true,
	"int":          true,
	"interface":    true,
	"long":         true,
	"native":       true,
	"new":          true,
	"package":      true,
	"private":      true,
	"protected":    true,
	"public":       true,
	"return":       true,
	"short":        true,
	"static":       true,
	"strictfp":     true,
	"super":        true,
	"switch":       true,
	"synchronized": true,
	"this":         true,
	"throw":        true,
	"throws":       true,
	"transient":    true,
	"try":          true,
	"void":         true,
	"volatile":     true,
	"while":        true,
	"true":         true,
	"false":        true,
	"null":         true,
	"record":       true,
	"var":          true,
	"yield":        true,
	"sealed":       true,
	"permits":      true,
	"non-sealed":   true,
}

// IsReservedKeyword returns true if the string is a Java keyword or literal.
func IsReservedKeyword(s string) bool {
	return javaKeywords[s]
}

// EscapeIdentifier escapes Java reserved keywords by appending an underscore.
func EscapeIdentifier(name string) string {
	if IsReservedKeyword(name) {
		return name + "_"
	}
	return name
}

// ToLowerCamel converts a string to lowerCamelCase and escapes Java keywords.
func ToLowerCamel(s string) string {
	if s == "" {
		return ""
	}
	camel := strcase.ToLowerCamel(s)
	return EscapeIdentifier(camel)
}

// ToUpperCamel converts a string to UpperCamelCase (PascalCase).
func ToUpperCamel(s string) string {
	if s == "" {
		return ""
	}
	return strcase.ToCamel(s)
}

// ToUpperSnake converts a string to UPPER_SNAKE_CASE.
func ToUpperSnake(s string) string {
	if s == "" {
		return ""
	}
	return strcase.ToScreamingSnake(s)
}

// SanitizeComment cleans up proto/model comments for Javadoc formatting.
func SanitizeComment(s string) string {
	s = strings.TrimSpace(s)
	s = strings.ReplaceAll(s, "/*", "/ *")
	s = strings.ReplaceAll(s, "*/", "* /")
	return s
}

// JavaPackageFromProto maps a proto package name (e.g. google.cloud.secretmanager.v1)
// to a Java package (com.google.cloud.secretmanager.v1).
func JavaPackageFromProto(protoPkg string) string {
	protoPkg = strings.TrimPrefix(protoPkg, ".")
	if strings.HasPrefix(protoPkg, "google.") {
		return "com." + protoPkg
	}
	if strings.HasPrefix(protoPkg, "com.") || strings.HasPrefix(protoPkg, "org.") || strings.HasPrefix(protoPkg, "net.") {
		return protoPkg
	}
	return "com." + protoPkg
}

// IsValidJavaIdentifier returns true if s is a valid Java identifier.
func IsValidJavaIdentifier(s string) bool {
	if s == "" || IsReservedKeyword(s) {
		return false
	}
	runes := []rune(s)
	if !unicode.IsLetter(runes[0]) && runes[0] != '_' && runes[0] != '$' {
		return false
	}
	for _, r := range runes[1:] {
		if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' && r != '$' {
			return false
		}
	}
	return true
}
