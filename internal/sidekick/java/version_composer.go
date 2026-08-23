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

package java

import (
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
)

// ComposeVersionClass generates Version.java AST.
func ComposeVersionClass(pkgName string) *ast.ClassDefinition {
	return &ast.ClassDefinition{
		Package:   pkgName,
		Name:      "Version",
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Final},
		Kind:      ast.ClassKindClass,
		JavaDoc:   "Version class for library metadata.",
		Fields: []*ast.FieldDefinition{
			{
				Name:        "DEFAULT_VERSION",
				Type:        ast.TypeString,
				Scope:       ast.Public,
				Modifiers:   []ast.Modifier{ast.Static, ast.Final},
				Initializer: ast.StringLiteralExpr("0.0.0"),
			},
		},
		Constructors: []*ast.ConstructorDefinition{
			{
				Scope: ast.Private,
				Body:  nil,
			},
		},
		Methods: []*ast.MethodDefinition{
			{
				Name:       "getVersion",
				ReturnType: ast.TypeString,
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Static},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.Expr("DEFAULT_VERSION"),
					},
				},
			},
		},
	}
}
