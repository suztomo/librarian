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

// ComposeGrpcStubClass generates the stub/Grpc<Service>Stub.java AST.
func ComposeGrpcStubClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	stubType := ast.ObjectType(svc.StubName, ann.StubPackageName)
	grpcStubType := ast.ObjectType(svc.GrpcStubName, ann.StubPackageName)
	stubSettingsType := ast.ObjectType(svc.StubSettingsName, ann.StubPackageName)

	doc := fmt.Sprintf("gRPC stub transport implementation for the %s API.", svc.Name)

	classDef := &ast.ClassDefinition{
		Package: ann.StubPackageName,
		Name:    svc.GrpcStubName,
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
			"com.google.api.gax.grpc.GrpcCallSettings",
			"com.google.api.gax.grpc.GrpcStubCallableFactory",
			"com.google.api.gax.grpc.GrpcTransportChannel",
			"com.google.api.gax.rpc.ClientContext",
			"com.google.api.gax.rpc.RequestParamsBuilder",
			"io.grpc.MethodDescriptor",
			"io.grpc.protobuf.ProtoUtils",
		},
	}

	// Static method descriptors for gRPC RPCs
	for _, m := range svc.Methods {
		descType := ast.GenericType(TypeMethodDescriptor, m.RequestType, m.ResponseType)
		descName := m.MethodName + "MethodDescriptor"
		classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
			Name:      descName,
			Type:      descType,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Static, ast.Final},
			Initializer: ast.ExprF(
				"MethodDescriptor.<%s, %s>newBuilder()\n"+
					"  .setType(MethodDescriptor.MethodType.UNARY)\n"+
					"  .setFullMethodName(\"%s/%s\")\n"+
					"  .setSampledToLocalTracing(true)\n"+
					"  .setRequestMarshaller(ProtoUtils.marshaller(%s.getDefaultInstance()))\n"+
					"  .setResponseMarshaller(ProtoUtils.marshaller(%s.getDefaultInstance()))\n"+
					"  .build()",
				m.RequestType.TypeString(), m.ResponseType.TypeString(),
				svc.Service.ID, m.Name,
				m.RequestType.TypeString(), m.ResponseType.TypeString(),
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
			Type:      TypeGrpcStubCallableFactory,
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
		} else if m.IsServerStreaming {
			serverStreamingCallableType := ast.GenericType(TypeServerStreamingCallable, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      m.CallableName,
				Type:      serverStreamingCallableType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.CallableName,
				ReturnType: serverStreamingCallableType,
				Scope:      ast.Public,
				Annotations: []*ast.AnnotationNode{
					ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
				},
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(m.CallableName)},
				},
			})
		} else if m.IsBidiStreaming {
			bidiStreamingCallableType := ast.GenericType(TypeBidiStreamingCallable, m.RequestType, m.ResponseType)
			classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
				Name:      m.CallableName,
				Type:      bidiStreamingCallableType,
				Scope:     ast.Private,
				Modifiers: []ast.Modifier{ast.Final},
			})
			classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
				Name:       m.CallableName,
				ReturnType: bidiStreamingCallableType,
				Scope:      ast.Public,
				Annotations: []*ast.AnnotationNode{
					ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
				},
				Body: []ast.Statement{
					&ast.ReturnStatement{Expr: ast.Expr(m.CallableName)},
				},
			})
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
			ReturnType: grpcStubType,
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
			ReturnType: grpcStubType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "callableFactory", Type: TypeGrpcStubCallableFactory, Final: true},
				{Name: "clientContext", Type: TypeClientContext, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s(callableFactory, clientContext)", svc.GrpcStubName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "getGrpcTransportName",
			ReturnType: ast.TypeString,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.StringLiteralExpr("grpc")},
			},
		},
	)

	// Protected Constructors
	classDef.Constructors = []*ast.ConstructorDefinition{
		{
			Scope: ast.Protected,
			Parameters: []*ast.ParameterDefinition{
				{Name: "callableFactory", Type: TypeGrpcStubCallableFactory, Final: true},
				{Name: "clientContext", Type: TypeClientContext, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body:   composeGrpcStubConstructorBody(svc),
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

func composeGrpcStubConstructorBody(svc *ServiceAnnotation) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("this.callableFactory = callableFactory;"),
		ast.StatementFrom("GrpcTransportChannel channel = (GrpcTransportChannel) clientContext.getTransportChannel();"),
	)

	for _, m := range svc.Methods {
		callSettingsName := m.MethodName + "GrpcCallSettings"
		descName := m.MethodName + "MethodDescriptor"
		stmts = append(stmts,
			ast.StatementFromF("GrpcCallSettings<%s, %s> %s = GrpcCallSettings.<%s, %s>newBuilder()\n"+
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
		} else if m.IsServerStreaming {
			stmts = append(stmts,
				ast.StatementFromF("this.%s = callableFactory.createServerStreamingCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.CallableName, callSettingsName, svc.StubSettingsName, fieldName),
			)
		} else if m.IsBidiStreaming {
			stmts = append(stmts,
				ast.StatementFromF("this.%s = callableFactory.createBidiStreamingCallable(%s, ((%s) clientContext.getStubSettings()).%s(), clientContext);",
					m.CallableName, callSettingsName, svc.StubSettingsName, fieldName),
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

// ComposeGrpcCallableFactoryClass generates the stub/Grpc<Service>CallableFactory.java AST.
func ComposeGrpcCallableFactoryClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	doc := fmt.Sprintf("gRPC callable factory for the %s service API.", svc.Name)

	classDef := &ast.ClassDefinition{
		Package: ann.StubPackageName,
		Name:    svc.GrpcFactoryName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Implements: []*ast.TypeNode{
			TypeGrpcStubCallableFactory,
		},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"javax.annotation.Generated",
			"com.google.api.gax.grpc.GrpcStubCallableFactory",
			"com.google.api.gax.grpc.GrpcCallableFactory",
			"com.google.api.gax.grpc.GrpcCallSettings",
			"com.google.api.gax.rpc.UnaryCallSettings",
			"com.google.api.gax.rpc.PagedCallSettings",
			"com.google.api.gax.rpc.ServerStreamingCallSettings",
			"com.google.api.gax.rpc.StreamingCallSettings",
			"com.google.api.gax.rpc.OperationCallSettings",
			"com.google.api.gax.rpc.ClientContext",
			"com.google.api.gax.rpc.UnaryCallable",
			"com.google.api.gax.rpc.ServerStreamingCallable",
			"com.google.api.gax.rpc.BidiStreamingCallable",
			"com.google.api.gax.rpc.OperationCallable",
			"com.google.api.gax.core.BackgroundResource",
		},
	}

	// Implementation of GrpcStubCallableFactory methods
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "createUnaryCallable",
			ReturnType: ast.GenericType(TypeUnaryCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "grpcCallSettings", Type: ast.GenericType(TypeGrpcCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "unaryCallSettings", Type: ast.GenericType(TypeUnaryCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("GrpcCallableFactory.createUnaryCallable(grpcCallSettings, unaryCallSettings, clientContext)")},
			},
		},
		&ast.MethodDefinition{
			Name:       "createPagedCallable",
			ReturnType: ast.GenericType(TypePagedCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("PagedListResponseT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "grpcCallSettings", Type: ast.GenericType(TypeGrpcCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""))},
				{Name: "pagedCallSettings", Type: ast.GenericType(TypePagedCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("PagedListResponseT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("GrpcCallableFactory.createPagedCallable(grpcCallSettings, pagedCallSettings, clientContext)")},
			},
		},
		&ast.MethodDefinition{
			Name:       "createOperationCallable",
			ReturnType: ast.GenericType(TypeOperationCallable, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("MetadataT", "")),
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "grpcCallSettings", Type: ast.GenericType(TypeGrpcCallSettings, ast.ObjectType("RequestT", ""), TypeOperation)},
				{Name: "operationCallSettings", Type: ast.GenericType(TypeOperationCallSettings, ast.ObjectType("RequestT", ""), ast.ObjectType("ResponseT", ""), ast.ObjectType("MetadataT", ""))},
				{Name: "clientContext", Type: TypeClientContext},
				{Name: "operationsStub", Type: TypeOperationsStub},
			},
			Annotations: []*ast.AnnotationNode{ast.NewAnnotation(ast.ObjectType("Override", "java.lang"))},
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("GrpcCallableFactory.createOperationCallable(grpcCallSettings, operationCallSettings, clientContext, operationsStub)")},
			},
		},
	)

	return classDef
}
