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

// ComposeHttpJsonStubClass generates the stub/HttpJson<Service>Stub.java AST.
func ComposeHttpJsonStubClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	stubType := ast.ObjectType(svc.StubName, ann.StubPackageName)
	httpJsonStubType := ast.ObjectType(svc.HttpJsonStubName, ann.StubPackageName)
	stubSettingsType := ast.ObjectType(svc.StubSettingsName, ann.StubPackageName)

	doc := fmt.Sprintf("REST (HTTP/JSON) stub transport implementation for the %s API.", svc.Name)

	classDef := &ast.ClassDefinition{
		Package: ann.StubPackageName,
		Name:    svc.HttpJsonStubName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Extends: stubType,
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"java.io.IOException",
			"java.util.concurrent.TimeUnit",
			"javax.annotation.Generated",
			"com.google.api.gax.core.BackgroundResource",
			"com.google.api.gax.core.BackgroundResourceAggregation",
			"com.google.api.gax.httpjson.ApiMethodDescriptor",
			"com.google.api.gax.httpjson.HttpJsonCallSettings",
			"com.google.api.gax.httpjson.HttpJsonStubCallableFactory",
			"com.google.api.gax.httpjson.HttpJsonTransportChannel",
			"com.google.api.gax.httpjson.ProtoMessageRequestFormatter",
			"com.google.api.gax.httpjson.ProtoMessageResponseParser",
			"com.google.api.gax.httpjson.ProtoRestSerializer",
			"com.google.api.gax.rpc.ClientContext",
			"com.google.api.gax.rpc.UnaryCallable",
		},
	}

	// Static method descriptors for HTTP/JSON RPCs
	for _, m := range svc.Methods {
		descType := ast.GenericType(TypeApiMethodDescriptor, m.RequestType, m.ResponseType)
		descName := m.MethodName + "MethodDescriptor"
		classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
			Name:      descName,
			Type:      descType,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Static, ast.Final},
			Initializer: ast.ExprF(
				"ApiMethodDescriptor.<%s, %s>newBuilder()\n"+
					"  .setFullMethodName(\"%s/%s\")\n"+
					"  .setHttpMethod(\"POST\")\n"+
					"  .setType(ApiMethodDescriptor.MethodType.UNARY)\n"+
					"  .setRequestFormatter(ProtoMessageRequestFormatter.<%s>newBuilder()\n"+
					"    .setPath(\"/v1/\" + \"%s\", request -> java.util.Collections.emptyMap())\n"+
					"    .setQueryParamsExtractor(request -> java.util.Collections.emptyMap())\n"+
					"    .setRequestBodyExtractor(request -> ProtoRestSerializer.create().toBody(\"*\", request, false))\n"+
					"    .build())\n"+
					"  .setResponseParser(ProtoMessageResponseParser.<%s>newBuilder()\n"+
					"    .setDefaultInstance(%s.getDefaultInstance())\n"+
					"    .build())\n"+
					"  .build()",
				m.RequestType.TypeString(), m.ResponseType.TypeString(),
				svc.Service.ID, m.Name,
				m.RequestType.TypeString(),
				m.MethodName,
				m.ResponseType.TypeString(),
				m.ResponseType.TypeString(),
			),
		})
	}

	// Callable fields
	classDef.Fields = append(classDef.Fields,
		&ast.FieldDefinition{
			Name:      "backgroundResources",
			Type:      TypeBackgroundResource,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		},
		&ast.FieldDefinition{
			Name:      "callableFactory",
			Type:      TypeHttpJsonStubCallableFactory,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		},
	)

	for _, m := range svc.Methods {
		if m.IsPaged {
			pagedResponseType := ast.ObjectType(fmt.Sprintf("%s.%s", svc.ClientName, pagedResponseClass(m.Name)), ann.PackageName)
			pagedCallableType := ast.GenericType(TypePagedCallable, m.RequestType, pagedResponseType)
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)

			classDef.Fields = append(classDef.Fields,
				&ast.FieldDefinition{
					Name:      m.CallableName,
					Type:      unaryCallableType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
				&ast.FieldDefinition{
					Name:      m.PagedCallableName,
					Type:      pagedCallableType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
			)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.PagedCallableName,
					ReturnType: pagedCallableType,
					Scope:      ast.Public,
					Annotations: []*ast.AnnotationNode{
						ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
					},
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(m.PagedCallableName)},
					},
				},
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: unaryCallableType,
					Scope:      ast.Public,
					Annotations: []*ast.AnnotationNode{
						ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
					},
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(m.CallableName)},
					},
				},
			)
		} else if m.IsLRO {
			opCallableType := ast.GenericType(TypeOperationCallable, m.RequestType, m.LroResponseType, m.LroMetadataType)
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, TypeOperation)

			classDef.Fields = append(classDef.Fields,
				&ast.FieldDefinition{
					Name:      m.CallableName,
					Type:      unaryCallableType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
				&ast.FieldDefinition{
					Name:      m.OperationCallableName,
					Type:      opCallableType,
					Scope:     ast.Private,
					Modifiers: []ast.Modifier{ast.Final},
				},
			)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: unaryCallableType,
					Scope:      ast.Public,
					Annotations: []*ast.AnnotationNode{
						ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
					},
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(m.CallableName)},
					},
				},
				&ast.MethodDefinition{
					Name:       m.OperationCallableName,
					ReturnType: opCallableType,
					Scope:      ast.Public,
					Annotations: []*ast.AnnotationNode{
						ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
					},
					Body: []ast.Statement{
						&ast.ReturnStatement{Expr: ast.Expr(m.OperationCallableName)},
					},
				},
			)
		} else {
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      m.CallableName,
				Type:      unaryCallableType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.CallableName,
				ReturnType: unaryCallableType,
				Scope:      ast.Public,
				Annotations: []*ast.AnnotationNode{
					ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
				},
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(m.CallableName)},
				},
			})
		}
	}

	// Factory create methods
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: httpJsonStubType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "settings", Type: stubSettingsType, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("create(settings.toBuilder().build().createDefaultStubCallableFactory(), ClientContext.create(settings))"),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: httpJsonStubType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "callableFactory", Type: TypeHttpJsonStubCallableFactory, Final: true},
				{Name: "clientContext", Type: TypeClientContext, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s(callableFactory, clientContext)", svc.HttpJsonStubName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "getHttpJsonTransportName",
			ReturnType: ast.TypeString,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.StringLiteralExpr("httpjson")},
			},
		},
	)

	// Protected Constructors
	classDef.Constructors = []*ast.ConstructorDefinition{
		{
			Scope: ast.Protected,
			Parameters: []*ast.ParameterDefinition{
				{Name: "callableFactory", Type: TypeHttpJsonStubCallableFactory, Final: true},
				{Name: "clientContext", Type: TypeClientContext, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body:   composeHttpJsonStubConstructorBody(svc),
		},
	}

	// BackgroundResource lifecycle methods
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "close",
			ReturnType: ast.TypeVoid,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				ast.StatementFrom("backgroundResources.close();"),
			},
		},
		&ast.MethodDefinition{
			Name:       "shutdown",
			ReturnType: ast.TypeVoid,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				ast.StatementFrom("backgroundResources.shutdown();"),
			},
		},
		&ast.MethodDefinition{
			Name:       "isShutdown",
			ReturnType: ast.TypeBoolean,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("backgroundResources.isShutdown()")},
			},
		},
		&ast.MethodDefinition{
			Name:       "isTerminated",
			ReturnType: ast.TypeBoolean,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("backgroundResources.isTerminated()")},
			},
		},
		&ast.MethodDefinition{
			Name:       "shutdownNow",
			ReturnType: ast.TypeVoid,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				ast.StatementFrom("backgroundResources.shutdownNow();"),
			},
		},
		&ast.MethodDefinition{
			Name:       "awaitTermination",
			ReturnType: ast.TypeBoolean,
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "duration", Type: ast.TypeLong},
				{Name: "unit", Type: ast.TypeTimeUnit},
			},
			Throws: []*ast.TypeNode{ast.ObjectType("InterruptedException", "java.lang")},
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("backgroundResources.awaitTermination(duration, unit)")},
			},
		},
	)

	return classDef
}

func composeHttpJsonStubConstructorBody(svc *ServiceAnnotation) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("this.callableFactory = callableFactory;"),
	)

	for _, m := range svc.Methods {
		callSettingsName := m.MethodName + "HttpJsonCallSettings"
		descName := m.MethodName + "MethodDescriptor"
		stmts = append(stmts,
			ast.StatementFromF("HttpJsonCallSettings<%s, %s> %s = HttpJsonCallSettings.<%s, %s>newBuilder()\n"+
				"  .setMethodDescriptor(%s)\n"+
				"  .build();",
				m.RequestType.TypeString(), m.ResponseType.TypeString(),
				callSettingsName,
				m.RequestType.TypeString(), m.ResponseType.TypeString(),
				descName,
			),
		)

		fieldName := m.MethodName + "Settings"
		if m.IsPaged {
			stmts = append(stmts,
				ast.StatementFromF("this.%s = callableFactory.createUnaryCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.CallableName, callSettingsName, svc.StubSettingsName, fieldName),
				ast.StatementFromF("this.%s = callableFactory.createPagedCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.PagedCallableName, callSettingsName, svc.StubSettingsName, fieldName),
			)
		} else if m.IsLRO {
			opFieldName := m.MethodName + "OperationSettings"
			stmts = append(stmts,
				ast.StatementFromF("this.%s = callableFactory.createUnaryCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.CallableName, callSettingsName, svc.StubSettingsName, fieldName),
				ast.StatementFromF("this.%s = callableFactory.createOperationCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext, null);",
					m.OperationCallableName, callSettingsName, svc.StubSettingsName, opFieldName),
			)
		} else {
			stmts = append(stmts,
				ast.StatementFromF("this.%s = callableFactory.createUnaryCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.CallableName, callSettingsName, svc.StubSettingsName, fieldName),
			)
		}
	}

	stmts = append(stmts, ast.StatementFrom("this.backgroundResources = new BackgroundResourceAggregation(clientContext.getBackgroundResources());"))
	return stmts
}

// ComposeHttpJsonCallableFactoryClass generates the stub/HttpJson<Service>CallableFactory.java AST.
func ComposeHttpJsonCallableFactoryClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	doc := fmt.Sprintf("REST (HTTP/JSON) callable factory for the %s service API.", svc.Name)

	classDef := &ast.ClassDefinition{
		Package: ann.StubPackageName,
		Name:    svc.HttpJsonFactoryName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Implements: []*ast.TypeNode{
			TypeHttpJsonStubCallableFactory,
		},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"javax.annotation.Generated",
			"com.google.api.gax.httpjson.HttpJsonStubCallableFactory",
			"com.google.api.gax.httpjson.HttpJsonCallableFactory",
			"com.google.api.gax.httpjson.HttpJsonCallSettings",
			"com.google.api.gax.rpc.UnaryCallSettings",
			"com.google.api.gax.rpc.PagedCallSettings",
			"com.google.api.gax.rpc.OperationCallSettings",
			"com.google.api.gax.rpc.ClientContext",
			"com.google.api.gax.rpc.UnaryCallable",
			"com.google.api.gax.rpc.OperationCallable",
			"com.google.api.gax.core.BackgroundResource",
		},
	}

	// Implementation of HttpJsonStubCallableFactory methods
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "createUnaryCallable",
			ReturnType: ast.GenericType(TypeUnaryCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "httpJsonCallSettings", Type: ast.GenericType(TypeHttpJsonCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "unaryCallSettings", Type: ast.GenericType(TypeUnaryCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("HttpJsonCallableFactory.createUnaryCallable(httpJsonCallSettings, unaryCallSettings, clientContext)")},
			},
		},
		&ast.MethodDefinition{
			Name:       "createPagedCallable",
			ReturnType: ast.GenericType(TypePagedCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("PagedListResponseT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "httpJsonCallSettings", Type: ast.GenericType(TypeHttpJsonCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "pagedCallSettings", Type: ast.GenericType(TypePagedCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("PagedListResponseT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("HttpJsonCallableFactory.createPagedCallable(httpJsonCallSettings, pagedCallSettings, clientContext)")},
			},
		},
		&ast.MethodDefinition{
			Name:       "createOperationCallable",
			ReturnType: ast.GenericType(TypeOperationCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("MetadataT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "httpJsonCallSettings", Type: ast.GenericType(TypeHttpJsonCallSettings, ast.ObjectType("RequestT", ""), TypeOperation)},
				{Name: "operationCallSettings", Type: ast.GenericType(TypeOperationCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("MetadataT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
				{Name: "operationsStub", Type: TypeOperationsStub},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("HttpJsonCallableFactory.createOperationCallable(httpJsonCallSettings, operationCallSettings, clientContext, operationsStub)")},
			},
		},
	)

	return classDef
}
