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

// ComposeStubSettingsClass generates stub/<Service>StubSettings.java AST.
func ComposeStubSettingsClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	stubSettingsType := ast.ObjectType(svc.StubSettingsName, ann.StubPackageName)
	builderType := ast.ObjectType("Builder", "")

	doc := fmt.Sprintf("Settings class to configure an instance of {@link %s}.", svc.StubName)

	classDef := &ast.ClassDefinition{
		Package: ann.StubPackageName,
		Name:    svc.StubSettingsName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Extends: ast.GenericType(TypeStubSettings, stubSettingsType),
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"java.io.IOException",
			"java.util.List",
			"java.util.Arrays",
			"java.util.Collections",
			"javax.annotation.Generated",
			"com.google.api.gax.rpc.StubSettings",
			"com.google.api.gax.rpc.ClientContext",
			"com.google.api.gax.rpc.TransportChannelProvider",
			"com.google.api.gax.core.GoogleCredentialsProvider",
		},
	}

	// Settings fields and getters
	for _, m := range svc.Methods {
		fieldName := m.MethodName + "Settings"
		if m.IsPaged {
			pagedResponseType := ast.ObjectType(fmt.Sprintf("%s.%s", svc.ClientName, pagedResponseClass(m.Name)), ann.PackageName)
			pagedSettingsType := ast.GenericType(TypePagedCallSettings, m.RequestType, m.ResponseType, pagedResponseType)

			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      pagedSettingsType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: pagedSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			opSettingsType := ast.GenericType(TypeOperationCallSettings, m.RequestType, m.LroResponseType, m.LroMetadataType)
			unarySettingsType := ast.GenericType(TypeUnaryCallSettings, m.RequestType, TypeOperation)

			classDef.Fields = append(classDef.Fields,
				&ast.FieldDefinition{
					Name:      fieldName,
					Type:      unarySettingsType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
				&ast.FieldDefinition{
					Name:      opFieldName,
					Type:      opSettingsType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
			)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       fieldName,
					ReturnType: unarySettingsType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
					},
				},
				&ast.MethodDefinition{
					Name:       opFieldName,
					ReturnType: opSettingsType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(opFieldName)},
					},
				},
			)
		} else if m.IsServerStreaming {
			serverStreamingSettingsType := ast.GenericType(TypeServerStreamingCallSettings, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      serverStreamingSettingsType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: serverStreamingSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else if m.IsBidiStreaming || m.IsClientStreaming {
			streamingSettingsType := ast.GenericType(TypeStreamingCallSettings, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      streamingSettingsType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: streamingSettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else {
			unarySettingsType := ast.GenericType(TypeUnaryCallSettings, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      unarySettingsType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: unarySettingsType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		}
	}

	// createStub method
	stubType := ast.ObjectType(svc.StubName, ann.StubPackageName)
	stubCreationStmt := "throw new UnsupportedOperationException();"
	if svc.HasGrpc && !svc.HasHttpJson {
		stubCreationStmt = fmt.Sprintf("return %s.create(this);", svc.GrpcStubName)
	} else if svc.HasHttpJson && !svc.HasGrpc {
		stubCreationStmt = fmt.Sprintf("return %s.create(this);", svc.HttpJsonStubName)
	} else if svc.HasGrpc && svc.HasHttpJson {
		stubCreationStmt = fmt.Sprintf("if (getTransportChannelProvider().getTransportName().equals(%s.getGrpcTransportName())) {\n"+
			"  return %s.create(this);\n"+
			"}\n"+
			"if (getTransportChannelProvider().getTransportName().equals(%s.getHttpJsonTransportName())) {\n"+
			"  return %s.create(this);\n"+
			"}\n"+
			"throw new UnsupportedOperationException(String.format(\"Transport not supported: %%s\", getTransportChannelProvider().getTransportName()));",
			svc.GrpcStubName, svc.GrpcStubName, svc.HttpJsonStubName, svc.HttpJsonStubName)
	}

	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "createStub",
		ReturnType: stubType,
		Scope:      ast.Public,
		Throws:     []*ast.TypeNode{ast.TypeIOException},
		JavaDoc:    fmt.Sprintf("Creates an instance of {@link %s}.", svc.StubName),
		Body: []ast.Statement{
			ast.StatementFrom(stubCreationStmt),
		},
	})

	// Default endpoint, scopes, and transport providers
	endpoint := svc.DefaultEndpoint
	if endpoint == "" {
		endpoint = "localhost:8080"
	}
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "getDefaultEndpoint",
			ReturnType: ast.TypeString,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.StringLiteralExpr(endpoint),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "getDefaultServiceScopes",
			ReturnType: ast.GenericType(ast.TypeList, ast.TypeString),
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("Collections.unmodifiableList(Arrays.asList(\"https://www.googleapis.com/auth/cloud-platform\"))"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "defaultCredentialsProviderBuilder",
			ReturnType: ast.ObjectType("GoogleCredentialsProvider.Builder", PkgGaxCore),
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("GoogleCredentialsProvider.newBuilder().setScopesToApply(getDefaultServiceScopes())"),
				},
			},
		},
	)

	if svc.HasGrpc {
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       "defaultGrpcTransportProviderBuilder",
			ReturnType: ast.ObjectType("InstantiatingGrpcChannelProvider.Builder", PkgGaxGrpc),
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("InstantiatingGrpcChannelProvider.newBuilder().setEndpoint(getDefaultEndpoint())"),
				},
			},
		})
	}

	if svc.HasHttpJson {
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       "defaultHttpJsonTransportProviderBuilder",
			ReturnType: ast.ObjectType("InstantiatingHttpJsonChannelProvider.Builder", PkgGaxHttpJson),
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("InstantiatingHttpJsonChannelProvider.newBuilder().setEndpoint(getDefaultEndpoint())"),
				},
			},
		})
	}

	defaultTransportStmt := "defaultGrpcTransportProviderBuilder().build()"
	if !svc.HasGrpc && svc.HasHttpJson {
		defaultTransportStmt = "defaultHttpJsonTransportProviderBuilder().build()"
	}
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "defaultTransportChannelProvider",
			ReturnType: TypeTransportChannelProvider,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr(defaultTransportStmt),
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
			Body:   composeStubSettingsConstructorBody(svc),
		},
	}

	// Nested Builder class
	builderClass := composeStubSettingsBuilder(svc, ann)
	classDef.InnerClasses = append(classDef.InnerClasses, builderClass)

	return classDef
}

func composeStubSettingsConstructorBody(svc *ServiceAnnotation) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts, ast.StatementFrom("super(builder);"))
	for _, m := range svc.Methods {
		fieldName := m.MethodName + "Settings"
		stmts = append(stmts, ast.StatementFromF("%s = builder.%s().build();", fieldName, fieldName))
		if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			stmts = append(stmts, ast.StatementFromF("%s = builder.%s().build();", opFieldName, opFieldName))
		}
	}
	return stmts
}

func composeStubSettingsBuilder(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	stubSettingsType := ast.ObjectType(svc.StubSettingsName, ann.StubPackageName)
	builderType := ast.ObjectType("Builder", "")

	builderClass := &ast.ClassDefinition{
		Name:      "Builder",
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Static},
		Kind:      ast.ClassKindClass,
		JavaDoc:   fmt.Sprintf("Builder for {@link %s}.", svc.StubSettingsName),
		Extends: ast.GenericType(
			ast.ObjectType("StubSettings.Builder", PkgGaxRpc),
			stubSettingsType,
			builderType,
		),
		Constructors: []*ast.ConstructorDefinition{
			{
				Scope: ast.Protected,
				Body: []ast.Statement{
					ast.StatementFrom("this(((ClientContext) null));"),
				},
			},
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{Name: "clientContext", Type: TypeClientContext, Final: true},
				},
				Body: composeStubSettingsBuilderInitBody(svc),
			},
			{
				Scope: ast.Protected,
				Parameters: []*ast.ParameterDefinition{
					{Name: "settings", Type: stubSettingsType, Final: true},
				},
				Body: composeStubSettingsBuilderFromSettingsBody(svc),
			},
		},
	}

	// Builder fields and getters
	for _, m := range svc.Methods {
		fieldName := m.MethodName + "Settings"
		if m.IsPaged {
			pagedResponseType := ast.ObjectType(fmt.Sprintf("%s.%s", svc.ClientName, pagedResponseClass(m.Name)), ann.PackageName)
			pagedSettingsBuilderType := ast.GenericType(ast.ObjectType("PagedCallSettings.Builder", PkgGaxRpc), m.RequestType, m.ResponseType, pagedResponseType)

			builderClass.Fields = append(builderClass.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      pagedSettingsBuilderType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			builderClass.Methods = append(builderClass.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: pagedSettingsBuilderType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			opSettingsBuilderType := ast.GenericType(ast.ObjectType("OperationCallSettings.Builder", PkgGaxRpc), m.RequestType, m.LroResponseType, m.LroMetadataType)
			unarySettingsBuilderType := ast.GenericType(ast.ObjectType("UnaryCallSettings.Builder", PkgGaxRpc), m.RequestType, TypeOperation)

			builderClass.Fields = append(builderClass.Fields,
				&ast.FieldDefinition{
					Name:      fieldName,
					Type:      unarySettingsBuilderType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
				&ast.FieldDefinition{
					Name:      opFieldName,
					Type:      opSettingsBuilderType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
			)
			builderClass.Methods = append(builderClass.Methods,
				&ast.MethodDefinition{
					Name:       fieldName,
					ReturnType: unarySettingsBuilderType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
					},
				},
				&ast.MethodDefinition{
					Name:       opFieldName,
					ReturnType: opSettingsBuilderType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(opFieldName)},
					},
				},
			)
		} else if m.IsServerStreaming {
			serverStreamingSettingsBuilderType := ast.GenericType(ast.ObjectType("ServerStreamingCallSettings.Builder", PkgGaxRpc), m.RequestType, m.ResponseType)
			builderClass.Fields = append(builderClass.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      serverStreamingSettingsBuilderType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			builderClass.Methods = append(builderClass.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: serverStreamingSettingsBuilderType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else if m.IsBidiStreaming || m.IsClientStreaming {
			streamingSettingsBuilderType := ast.GenericType(ast.ObjectType("StreamingCallSettings.Builder", PkgGaxRpc), m.RequestType, m.ResponseType)
			builderClass.Fields = append(builderClass.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      streamingSettingsBuilderType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			builderClass.Methods = append(builderClass.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: streamingSettingsBuilderType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		} else {
			unarySettingsBuilderType := ast.GenericType(ast.ObjectType("UnaryCallSettings.Builder", PkgGaxRpc), m.RequestType, m.ResponseType)
			builderClass.Fields = append(builderClass.Fields, &ast.FieldDefinition{
				Name:      fieldName,
				Type:      unarySettingsBuilderType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			builderClass.Methods = append(builderClass.Methods, &ast.MethodDefinition{
				Name:       fieldName,
				ReturnType: unarySettingsBuilderType,
				Scope:      ast.Public,
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
				},
			})
		}
	}

	// build() and createDefault()
	builderClass.Methods = append(builderClass.Methods,
		&ast.MethodDefinition{
			Name:       "createDefault",
			ReturnType: builderType,
			Scope:      ast.Private,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("new Builder(((ClientContext) null))"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "build",
			ReturnType: stubSettingsType,
			Scope:      ast.Public,
			Throws:     []*ast.TypeNode{ast.TypeIOException},
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s(this)", svc.StubSettingsName),
				},
			},
		},
	)

	return builderClass
}

func composeStubSettingsBuilderInitBody(svc *ServiceAnnotation) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("super(clientContext);"),
		ast.StatementFrom("setEndpoint(getDefaultEndpoint());"),
		ast.StatementFrom("setCredentialsProvider(defaultCredentialsProviderBuilder().build());"),
		ast.StatementFrom("setTransportChannelProvider(defaultTransportChannelProvider());"),
	)
	for _, m := range svc.Methods {
		fieldName := m.MethodName + "Settings"
		if m.IsPaged {
			stmts = append(stmts, ast.StatementFromF("%s = PagedCallSettings.newBuilder(%s.createAsync);", fieldName, pagedResponseClass(m.Name)))
		} else if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			stmts = append(stmts,
				ast.StatementFromF("%s = UnaryCallSettings.newUnaryCallSettingsBuilder();", fieldName),
				ast.StatementFromF("%s = OperationCallSettings.newBuilder();", opFieldName),
			)
		} else if m.IsServerStreaming {
			stmts = append(stmts, ast.StatementFromF("%s = ServerStreamingCallSettings.newBuilder();", fieldName))
		} else if m.IsBidiStreaming || m.IsClientStreaming {
			stmts = append(stmts, ast.StatementFromF("%s = StreamingCallSettings.newBuilder();", fieldName))
		} else {
			stmts = append(stmts, ast.StatementFromF("%s = UnaryCallSettings.newUnaryCallSettingsBuilder();", fieldName))
		}
	}
	return stmts
}

func composeStubSettingsBuilderFromSettingsBody(svc *ServiceAnnotation) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts, ast.StatementFrom("super(settings);"))
	for _, m := range svc.Methods {
		fieldName := m.MethodName + "Settings"
		stmts = append(stmts, ast.StatementFromF("%s = settings.%s.toBuilder();", fieldName, fieldName))
		if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			stmts = append(stmts, ast.StatementFromF("%s = settings.%s.toBuilder();", opFieldName, opFieldName))
		}
	}
	return stmts
}
