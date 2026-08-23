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

// ComposeStubClass generates the abstract stub/<Service>Stub.java AST.
func ComposeStubClass(svc *ServiceAnnotation, ann *ModelAnnotations) *ast.ClassDefinition {
	doc := fmt.Sprintf("Base stub class for the %s API service.\n\nThis class is for advanced usage and reflects the underlying API directly.", svc.Name)

	classDef := &ast.ClassDefinition{
		Package:   ann.StubPackageName,
		Name:      svc.StubName,
		Scope:     ast.Public,
		Modifiers: []ast.Modifier{ast.Abstract},
		Kind:      ast.ClassKindClass,
		JavaDoc:   doc,
		Implements: []*ast.TypeNode{
			TypeBackgroundResource,
		},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"javax.annotation.Generated",
			"com.google.api.gax.core.BackgroundResource",
		},
	}

	// OperationsStub getter if service has LROs
	if svc.HasLRO {
		operationsStubType := TypeOperationsStub
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       "getOperationsStub",
			ReturnType: operationsStubType,
			Scope:      ast.Public,
			JavaDoc:    "Returns the OperationsStub to poll long-running operations.",
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.Expr("null"),
				},
			},
		})
	}

	// Callable abstract methods
	for _, m := range svc.Methods {
		if m.IsPaged {
			pagedResponseType := ast.ObjectType(fmt.Sprintf("%s.%s", svc.ClientName, pagedResponseClass(m.Name)), ann.PackageName)
			pagedCallableType := ast.GenericType(TypePagedCallable, m.RequestType, pagedResponseType)
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)

			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.PagedCallableName,
					ReturnType: pagedCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.PagedCallableName + "()\");"),
					},
				},
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: unaryCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.CallableName + "()\");"),
					},
				},
			)
		} else if m.IsLRO {
			opCallableType := ast.GenericType(TypeOperationCallable, m.RequestType, m.LroResponseType, m.LroMetadataType)
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, TypeOperation)

			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.OperationCallableName,
					ReturnType: opCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.OperationCallableName + "()\");"),
					},
				},
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: unaryCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.CallableName + "()\");"),
					},
				},
			)
		} else if m.IsServerStreaming {
			serverStreamingCallableType := ast.GenericType(TypeServerStreamingCallable, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: serverStreamingCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.CallableName + "()\");"),
					},
				},
			)
		} else if m.IsBidiStreaming {
			bidiStreamingCallableType := ast.GenericType(TypeBidiStreamingCallable, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: bidiStreamingCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.CallableName + "()\");"),
					},
				},
			)
		} else {
			unaryCallableType := ast.GenericType(TypeUnaryCallable, m.RequestType, m.ResponseType)
			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       m.CallableName,
					ReturnType: unaryCallableType,
					Scope:      ast.Public,
					Body: []ast.Statement{
						ast.StatementFrom("throw new UnsupportedOperationException(\"Not implemented: " + m.CallableName + "()\");"),
					},
				},
			)
		}
	}

	// Abstract close method
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "close",
		ReturnType: ast.TypeVoid,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Abstract},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
		},
	})

	return classDef
}
