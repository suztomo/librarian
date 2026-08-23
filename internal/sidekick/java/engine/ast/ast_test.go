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

package ast

import (
	"testing"
)

func TestTypeNode_TypeString(t *testing.T) {
	tests := []struct {
		name string
		typ  *TypeNode
		want string
	}{
		{
			name: "primitive int",
			typ:  TypeInt,
			want: "int",
		},
		{
			name: "object String",
			typ:  TypeString,
			want: "String",
		},
		{
			name: "generic List<String>",
			typ:  GenericType(TypeList, TypeString),
			want: "List<String>",
		},
		{
			name: "generic Map<String, Integer>",
			typ:  GenericType(TypeMap, TypeString, TypeBoxedInteger),
			want: "Map<String, Integer>",
		},
		{
			name: "array String[]",
			typ:  ArrayType(TypeString),
			want: "String[]",
		},
		{
			name: "wildcard ? extends Bound",
			typ:  WildcardType(TypeString),
			want: "? extends String",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typ.TypeString()
			if got != tt.want {
				t.Errorf("TypeString() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestTypeNode_FullName(t *testing.T) {
	tests := []struct {
		name string
		typ  *TypeNode
		want string
	}{
		{
			name: "primitive",
			typ:  TypeInt,
			want: "int",
		},
		{
			name: "object",
			typ:  TypeString,
			want: "java.lang.String",
		},
		{
			name: "protobuf ByteString",
			typ:  TypeByteString,
			want: "com.google.protobuf.ByteString",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.typ.FullName()
			if got != tt.want {
				t.Errorf("FullName() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestStatementsAndExpressions(t *testing.T) {
	decl := &VariableDeclarationStatement{
		Type:        TypeString,
		Name:        "name",
		Initializer: StringLiteralExpr("hello"),
		Final:       true,
	}
	wantDecl := "  final String name = \"hello\";"
	if got := decl.FormatStatement("  "); got != wantDecl {
		t.Errorf("VariableDeclarationStatement = %q, want %q", got, wantDecl)
	}

	ret := &ReturnStatement{Expr: Expr("name")}
	wantRet := "  return name;"
	if got := ret.FormatStatement("  "); got != wantRet {
		t.Errorf("ReturnStatement = %q, want %q", got, wantRet)
	}

	call := &MethodInvocationExpr{
		Target: Expr("client"),
		Method: "close",
		Args:   nil,
	}
	wantCall := "client.close()"
	if got := call.FormatExpression(); got != wantCall {
		t.Errorf("MethodInvocationExpr = %q, want %q", got, wantCall)
	}
}
