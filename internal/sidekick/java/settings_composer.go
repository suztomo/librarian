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
	"fmt"

	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
)

// ComposeSettingsClass generates the <Service>Settings.java AST.
func ComposeSettingsClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	settingsType := ast.ObjectType(svc.SettingsName, ann.PackageName)
	stubSettingsType := ast.ObjectType(svc.StubSettingsName, ann.StubPackageName)

	doc := fmt.Sprintf("Settings class to configure an instance of {@link %s}.", svc.ClientName)

	classDef := &ast.ClassDefinition{
		Package: ann.PackageName,
		Name:    svc.SettingsName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Extends: ast.GenericType(TypeClientSettings, settingsType),
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"java.io.IOException",
			"javax.annotation.Generated",
			"com.google.api.gax.rpc.ClientSettings",
			"com.google.api.gax.rpc.ClientContext",
		},
	}

	// Method settings accessors
	for _, m := range svc.Methods {
		if m.IsPaged {
			pagedSettingsType := ast.GenericType(TypePagedCallSettings, m.RequestType, m.ResponseType, ast.ObjectType(pagedResponseClass(m.Name), ""))
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.MethodName + "Settings",
				ReturnType: pagedSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings()).%sSettings()", svc.StubSettingsName, m.MethodName),
					},
				},
			})
		} else if m.IsLRO {
			opSettingsType := ast.GenericType(TypeOperationCallSettings, m.RequestType, m.LroResponseType, m.LroMetadataType)
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.MethodName + "OperationSettings",
				ReturnType: opSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings()).%sOperationSettings()", svc.StubSettingsName, m.MethodName),
					},
				},
			})
		} else if m.IsServerStreaming {
			serverStreamingSettingsType := ast.GenericType(TypeServerStreamingCallSettings, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.MethodName + "Settings",
				ReturnType: serverStreamingSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings()).%sSettings()", svc.StubSettingsName, m.MethodName),
					},
				},
			})
		} else if m.IsBidiStreaming || m.IsClientStreaming {
			streamingSettingsType := ast.GenericType(TypeStreamingCallSettings, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.MethodName + "Settings",
				ReturnType: streamingSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings()).%sSettings()", svc.StubSettingsName, m.MethodName),
					},
				},
			})
		} else {
			unarySettingsType := ast.GenericType(TypeUnaryCallSettings, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.MethodName + "Settings",
				ReturnType: unarySettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings()).%sSettings()", svc.StubSettingsName, m.MethodName),
					},
				},
			})
		}
	}

	// Factory methods: createDefault, create(stubSettings), newBuilder, toBuilder
	builderType := ast.ObjectType("Builder", "")
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "createDefault",
			ReturnType: settingsType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Throws:     []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("newBuilder().build()"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: settingsType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "stubSettings", Type: stubSettingsType, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s.Builder(stubSettings.toBuilder()).build()", svc.SettingsName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "newBuilder",
			ReturnType: builderType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("Builder.createDefault()"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "newBuilder",
			ReturnType: builderType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "clientContext", Type: TypeClientContext, Final: true},
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("new Builder(clientContext)"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "toBuilder",
			ReturnType: builderType,
			Scope:      ast.Public,
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("new Builder(this)"),
				},
			},
		},
	)

	// Protected Constructor
	classDef.Constructors = []*ast.ConstructorDefinition{
		{
			Scope: ast.Protected,
			Parameters: []*ast.ParameterDefinition{
				{Name: "builder", Type: builderType, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				ast.StatementFrom("super(builder);"),
			},
		},
	}

	// Nested Builder class
	builderClass := composeSettingsBuilder(svc, ann)
	classDef.InnerClasses = append(classDef.InnerClasses, builderClass)

	return classDef
}

func composeSettingsBuilder(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	settingsType := ast.ObjectType(svc.SettingsName, ann.PackageName)
	builderType := ast.ObjectType("Builder", "")
	stubSettingsBuilderType := ast.ObjectType(svc.StubSettingsName+".Builder", ann.StubPackageName)

	builderClass := &ast.ClassDefinition{
		Name:      "Builder",
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Static},
		Kind:      ast.ClassKindClass,
		JavaDoc:   fmt.Sprintf("Builder for {@link %s}.", svc.SettingsName),
		Extends: ast.GenericType(
			ast.ObjectType("ClientSettings.Builder", PkgGaxRpc),
			settingsType,
			builderType,
		),
		Constructors: []*ast.ConstructorDefinition{
			{
				Scope:  ast.Protected,
				Throws: []*ast.TypeNode{ast.TypeIOException},
				Body: []ast.Statement{
					ast.StatementFromF("this(((%s) null));", svc.StubSettingsName+".Builder"),
				},
			},
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{Name: "clientContext", Type: TypeClientContext, Final: true},
				},
				Body: []ast.Statement{
					ast.StatementFromF("super(%s.newBuilder(clientContext));", svc.StubSettingsName),
				},
			},
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{Name: "settings", Type: settingsType, Final: true},
				},
				Body: []ast.Statement{
					ast.StatementFromF("super(settings.getStubSettings().toBuilder());"),
				},
			},
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{Name: "stubSettings", Type: stubSettingsBuilderType, Final: true},
				},
				Body: []ast.Statement{
					ast.StatementFrom("super(stubSettings);"),
				},
			},
		},
		Methods: []*ast.MethodDefinition{
			{
				Name:       "createDefault",
				ReturnType: builderType,
				Scope:      ast.Private,
				Modifiers:  []ast.Modifier{ast.Static},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("new Builder(%s.newBuilder())", svc.StubSettingsName),
					},
				},
			},
			{
				Name:       "getStubSettingsBuilder",
				ReturnType: stubSettingsBuilderType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("((%s) getStubSettings())", svc.StubSettingsName+".Builder"),
					},
				},
			},
			{
				Name:       "build",
				ReturnType: settingsType,
				Scope:      ast.Public,
				Throws:     []*ast.TypeNode{ast.TypeIOException},
				Annotations: []*ast.AnnotationNode{
					ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
				},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("new %s(this)", svc.SettingsName),
					},
				},
			},
		},
	}

	return builderClass
}
