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

package writer

import (
	"strings"
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
)

func TestWriteClass(t *testing.T) {
	classDef := &ast.ClassDefinition{
		Package:   "com.google.cloud.secretmanager.v1",
		Name:      "SecretManagerServiceClient",
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Final},
		Implements: []*ast.TypeNode{
			ast.TypeAutoCloseable,
		},
		JavaDoc: "Service description client.",
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by sidekick")),
		},
		Fields: []*ast.FieldDefinition{
			{
				Name:        "settings",
				Type:        ast.ObjectType("SecretManagerServiceSettings", "com.google.cloud.secretmanager.v1"),
				Scope:       ast.Private,
				Modifiers:   []ast.Modifier{ast.Final},
				Initializer: nil,
			},
		},
		Constructors: []*ast.ConstructorDefinition{
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{
						Name:  "settings",
						Type:  ast.ObjectType("SecretManagerServiceSettings", "com.google.cloud.secretmanager.v1"),
						Final: true,
					},
				},
				Body: []ast.Statement{
					ast.StatementFrom("this.settings = settings;"),
				},
			},
		},
		Methods: []*ast.MethodDefinition{
			{
				Name:       "getSettings",
				ReturnType: ast.ObjectType("SecretManagerServiceSettings", "com.google.cloud.secretmanager.v1"),
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr("settings")},
				},
			},
		},
	}

	src, err := WriteClass(classDef)
	if err != nil {
		t.Fatalf("WriteClass failed: %v", err)
	}

	if !strings.Contains(src, "package com.google.cloud.secretmanager.v1;") {
		t.Errorf("Missing package line in:\n%s", src)
	}
	if !strings.Contains(src, "import javax.annotation.Generated;") {
		t.Errorf("Missing Generated import in:\n%s", src)
	}
	if !strings.Contains(src, "public final class SecretManagerServiceClient implements AutoCloseable {") {
		t.Errorf("Missing class header in:\n%s", src)
	}
	if !strings.Contains(src, "private final SecretManagerServiceSettings settings;") {
		t.Errorf("Missing field in:\n%s", src)
	}
	if !strings.Contains(src, "protected SecretManagerServiceClient(final SecretManagerServiceSettings settings) {") {
		t.Errorf("Missing constructor in:\n%s", src)
	}
	if !strings.Contains(src, "public SecretManagerServiceSettings getSettings() {") {
		t.Errorf("Missing getSettings method in:\n%s", src)
	}
}
