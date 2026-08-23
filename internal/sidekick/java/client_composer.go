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
	"strings"

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

// ComposeClientClass generates the main <Service>Client.java AST.
func ComposeClientClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	settingsType := ast.ObjectType(svc.SettingsName, ann.PackageName)
	stubType := ast.ObjectType(svc.StubName, ann.StubPackageName)
	clientType := ast.ObjectType(svc.ClientName, ann.PackageName)

	doc := fmt.Sprintf("Service Description: %s\n\nThis class provides the client interface to %s.", svc.Name, svc.Name)
	if svc.Service.Documentation != "" {
		doc = FormatJavaDocComment(svc.Service.Documentation)
	}

	classDef := &ast.ClassDefinition{
		Package: ann.PackageName,
		Name:    svc.ClientName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: doc,
		Implements: []*ast.TypeNode{
			TypeBackgroundResource,
		},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"java.io.IOException",
			"java.util.concurrent.TimeUnit",
			"javax.annotation.Generated",
			"com.google.api.gax.core.BackgroundResource",
		},
	}

	// Member fields
	classDef.Fields = []*ast.FieldDefinition{
		{
			Name:      "settings",
			Type:      settingsType,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		},
		{
			Name:      "stub",
			Type:      stubType,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		},
	}

	// Factory methods: create(), create(settings), create(stub)
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: clientType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Throws:     []*ast.TypeNode{ast.TypeIOException},
			JavaDoc:    fmt.Sprintf("Constructs an instance of %s with default settings.", svc.ClientName),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("create(%s.createDefault())", svc.SettingsName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: clientType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "settings", Type: settingsType, Final: true},
			},
			Throws:  []*ast.TypeNode{ast.TypeIOException},
			JavaDoc: fmt.Sprintf("Constructs an instance of %s, using the given settings.", svc.ClientName),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s(settings)", svc.ClientName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       "create",
			ReturnType: clientType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Static},
			Parameters: []*ast.ParameterDefinition{
				{Name: "stub", Type: stubType, Final: true},
			},
			JavaDoc: fmt.Sprintf("Constructs an instance of %s, using the given stub for making calls.", svc.ClientName),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("new %s(stub)", svc.ClientName),
				},
			},
		},
	)

	// Constructors
	classDef.Constructors = []*ast.ConstructorDefinition{
		{
			Scope: ast.Protected,
			Parameters: []*ast.ParameterDefinition{
				{Name: "settings", Type: settingsType, Final: true},
			},
			Throws: []*ast.TypeNode{ast.TypeIOException},
			Body: []ast.Statement{
				ast.StatementFrom("this.settings = settings;"),
				ast.StatementFrom("this.stub = ((" + svc.StubSettingsName + ") settings.getStubSettings()).createStub();"),
			},
		},
		{
			Scope: ast.Protected,
			Parameters: []*ast.ParameterDefinition{
				{Name: "stub", Type: stubType, Final: true},
			},
			Body: []ast.Statement{
				ast.StatementFrom("this.settings = null;"),
				ast.StatementFrom("this.stub = stub;"),
			},
		},
	}

	// Settings & Stub getters
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "getSettings",
			ReturnType: settingsType,
			Scope:      ast.Public,
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("settings")},
			},
		},
		&ast.MethodDefinition{
			Name:       "getStub",
			ReturnType: stubType,
			Scope:      ast.Public,
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr("stub")},
			},
		},
	)

	// OperationsClient getter if service has LROs
	if svc.HasLRO {
		operationsClientType := TypeOperationsClient
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       "getOperationsClient",
			ReturnType: operationsClientType,
			Scope:      ast.Public,
			JavaDoc:    "Returns the OperationsClient that can be used to query the status of a long-running operation returned by another API method.",
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("stub.getOperationsClient()"),
				},
			},
		})
	}

	// Methods for each RPC in the service
	for _, m := range svc.Methods {
		composeClientMethods(classDef, m, svc, ann)
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
				ast.StatementFrom("stub.close();"),
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
				ast.StatementFrom("stub.shutdown();"),
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
				&ast.ReturnStatement{Expr: ast.Expr("stub.isShutdown()")},
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
				&ast.ReturnStatement{Expr: ast.Expr("stub.isTerminated()")},
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
				ast.StatementFrom("stub.shutdownNow();"),
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
				&ast.ReturnStatement{Expr: ast.Expr("stub.awaitTermination(duration, unit)")},
			},
		},
	)

	return classDef
}

func composeClientMethods(classDef *ast.ClassDefinition, m *MethodAnnotation, svc *ServiceAnnotation, ann *ModelAnnotations) {
	if m.IsPaged {
		pagedResponseTypeName := pagedResponseClass(m.Name)
		pagedResponseType := ast.ObjectType(fmt.Sprintf("%s.%s", svc.ClientName, pagedResponseTypeName), ann.PackageName)
		callableType := ast.GenericType(TypePagedCallable, m.RequestType, pagedResponseType)

		// 1. Paged Callable getter
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       m.PagedCallableName,
			ReturnType: callableType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			JavaDoc:    fmt.Sprintf("Returns the object with the settings used for key %s calls.", m.Name),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("stub.%s()", m.PagedCallableName),
				},
			},
		})

		// 2. Unary Callable getter
		unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       m.CallableName,
			ReturnType: unaryCallableType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			JavaDoc:    fmt.Sprintf("Returns the object with the settings used for %s calls.", m.Name),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("stub.%s()", m.CallableName),
				},
			},
		})

		// 3. Direct paged RPC method
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       m.MethodName,
			ReturnType: pagedResponseType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			Parameters: []*ast.ParameterDefinition{
				{Name: "request", Type: m.RequestType, Final: true},
			},
			JavaDoc: m.Description,
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("%s().call(request)", m.PagedCallableName),
				},
			},
		})

		// 4. Paged Signature Overload methods
		for _, sig := range m.Signatures {
			composeSignatureMethod(classDef, m, sig, pagedResponseType)
		}

		// 5. Compose nested PagedResponse View classes
		composePagedClasses(classDef, m)
		return
	}

	if m.IsLRO {
		operationFutureType := ast.GenericType(TypeOperationFuture, m.LroResponseType, m.LroMetadataType)
		operationCallableType := ast.GenericType(TypeOperationCallable, m.RequestType, m.LroResponseType, m.LroMetadataType)
		unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, TypeOperation)

		// 1. Operation Callable getter
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       m.OperationCallableName,
			ReturnType: operationCallableType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			JavaDoc:    fmt.Sprintf("Returns the operation callable for %s calls.", m.Name),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("stub.%s()", m.OperationCallableName),
				},
			},
		})

		// 2. Unary Callable getter
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       m.CallableName,
			ReturnType: unaryCallableType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			JavaDoc:    fmt.Sprintf("Returns the callable for %s calls.", m.Name),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("stub.%s()", m.CallableName),
				},
			},
		})

		// 3. Direct Async LRO method
		asyncMethodName := m.MethodName + "Async"
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       asyncMethodName,
			ReturnType: operationFutureType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			Parameters: []*ast.ParameterDefinition{
				{Name: "request", Type: m.RequestType, Final: true},
			},
			JavaDoc: m.Description,
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("%s().futureCall(request)", m.OperationCallableName),
				},
			},
		})

		// 4. LRO Signature Overload methods
		for _, sig := range m.Signatures {
			composeSignatureMethod(classDef, m, sig, operationFutureType)
		}
		return
	}

	if m.IsServerStreaming {
		callableType := ast.GenericType(TypeServerStreamingCallable, m.RequestType, m.ResponseType)
		classDef.Methods = append(classDef.Methods,
			&ast.MethodDefinition{
				Name:       m.CallableName,
				ReturnType: callableType,
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Final},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("stub.%s()", m.CallableName),
					},
				},
			},
			&ast.MethodDefinition{
				Name:       m.MethodName,
				ReturnType: ast.TypeVoid,
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Final},
				Parameters: []*ast.ParameterDefinition{
					{Name: "request", Type: m.RequestType, Final: true},
					{Name: "responseObserver", Type: ast.GenericType(TypeResponseObserver, m.ResponseType), Final: true},
				},
				JavaDoc: m.Description,
				Body: []ast.Statement{
					ast.StatementFromF("%s().call(request, responseObserver);", m.CallableName),
				},
			},
		)
		return
	}

	if m.IsBidiStreaming {
		callableType := ast.GenericType(TypeBidiStreamingCallable, m.RequestType, m.ResponseType)
		classDef.Methods = append(classDef.Methods,
			&ast.MethodDefinition{
				Name:       m.CallableName,
				ReturnType: callableType,
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Final},
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("stub.%s()", m.CallableName),
					},
				},
			},
			&ast.MethodDefinition{
				Name:       m.MethodName,
				ReturnType: ast.GenericType(TypeClientStream, m.RequestType),
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Final},
				Parameters: []*ast.ParameterDefinition{
					{Name: "responseObserver", Type: ast.GenericType(TypeResponseObserver, m.ResponseType), Final: true},
				},
				JavaDoc: m.Description,
				Body: []ast.Statement{
					&ast.ReturnStatement{
						Expr: ast.ExprF("%s().splitCall(responseObserver)", m.CallableName),
					},
				},
			},
		)
		return
	}

	// Default: Unary RPC
	unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       m.CallableName,
			ReturnType: unaryCallableType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			JavaDoc:    fmt.Sprintf("Returns the callable for %s calls.", m.Name),
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("stub.%s()", m.CallableName),
				},
			},
		},
		&ast.MethodDefinition{
			Name:       m.MethodName,
			ReturnType: m.ResponseType,
			Scope:      ast.Public,
			Modifiers:  []ast.Modifier{ast.Final},
			Parameters: []*ast.ParameterDefinition{
				{Name: "request", Type: m.RequestType, Final: true},
			},
			JavaDoc: m.Description,
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("%s().call(request)", m.CallableName),
				},
			},
		},
	)

	// Signature Overloads for Unary RPC
	for _, sig := range m.Signatures {
		composeSignatureMethod(classDef, m, sig, m.ResponseType)
	}
}

func composeSignatureMethod(classDef *ast.ClassDefinition, m *MethodAnnotation, sig []*api.Field, returnType *ast.TypeNode) {
	methodName := m.MethodName
	if m.IsLRO {
		methodName = m.MethodName + "Async"
	}

	var params []*ast.ParameterDefinition
	var builderCalls []string

	for _, f := range sig {
		paramType := FieldTypeToJavaType(f).Type
		paramName := lexicon.ToLowerCamel(f.Name)
		params = append(params, &ast.ParameterDefinition{
			Name:  paramName,
			Type:  paramType,
			Final: true,
		})
		builderCalls = append(builderCalls, fmt.Sprintf(".%s(%s)", SetterName(f.Name), paramName))
	}

	reqConstruction := fmt.Sprintf("%s request = %s.newBuilder()%s.build();",
		m.RequestType.TypeString(),
		m.RequestType.TypeString(),
		strings.Join(builderCalls, ""),
	)

	invocation := fmt.Sprintf("%s(request)", methodName)

	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       methodName,
		ReturnType: returnType,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Final},
		Parameters: params,
		JavaDoc:    m.Description,
		Body: []ast.Statement{
			ast.StatementFrom(reqConstruction),
			&ast.ReturnStatement{
				Expr: ast.Expr(invocation),
			},
		},
	})
}

func pagedResponseClass(methodName string) string {
	return lexicon.ToUpperCamel(methodName) + "PagedResponse"
}

func composePagedClasses(classDef *ast.ClassDefinition, m *MethodAnnotation) {
	pagedRespName := pagedResponseClass(m.Name)
	pageName := lexicon.ToUpperCamel(m.Name) + "Page"
	fixedCollectionName := lexicon.ToUpperCamel(m.Name) + "FixedSizeCollection"

	pagedClass := &ast.ClassDefinition{
		Name:      pagedRespName,
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Static},
		Kind:      ast.ClassKindClass,
		Extends: ast.GenericType(
			ast.ObjectType("AbstractPagedListResponse", PkgGaxRpc),
			m.RequestType,
			m.ResponseType,
			m.PageItemType,
			ast.ObjectType(pageName, ""),
			ast.ObjectType(fixedCollectionName, ""),
		),
		Methods: []*ast.MethodDefinition{
			{
				Name:       "createAsync",
				ReturnType: ast.GenericType(TypeApiFuture, ast.ObjectType(pagedRespName, "")),
				Scope:      ast.Public,
				Modifiers:  []ast.Modifier{ast.Static},
				Parameters: []*ast.ParameterDefinition{
					{Name: "context", Type: ast.GenericType(ast.ObjectType("PageContext", PkgGaxRpc), m.RequestType, m.ResponseType, m.PageItemType), Final: true},
					{Name: "futureResponse", Type: ast.GenericType(TypeApiFuture, m.ResponseType), Final: true},
				},
				Body: []ast.Statement{
					ast.StatementFromF("return ApiFutures.transform(futureResponse, input -> new %s(context, input), MoreExecutors.directExecutor());", pagedRespName),
				},
			},
		},
		Constructors: []*ast.ConstructorDefinition{
			{
				Scope: ast.Private,
				Parameters: []*ast.ParameterDefinition{
					{Name: "context", Type: ast.GenericType(ast.ObjectType("PageContext", PkgGaxRpc), m.RequestType, m.ResponseType, m.PageItemType)},
					{Name: "response", Type: m.ResponseType},
				},
				Body: []ast.Statement{
					ast.StatementFrom("super(context, response);"),
				},
			},
		},
	}

	classDef.InnerClasses = append(classDef.InnerClasses, pagedClass)
}
