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
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
)

func TestFieldTypeToJavaType(t *testing.T) {
	strField := &api.Field{
		Name:  "name",
		Typez: api.TypezString,
	}
	wrapper := FieldTypeToJavaType(strField)
	if wrapper.Type.Name != "String" {
		t.Errorf("got %q, want String", wrapper.Type.Name)
	}

	repField := &api.Field{
		Name:     "labels",
		Typez:    api.TypezString,
		Repeated: true,
	}
	repWrapper := FieldTypeToJavaType(repField)
	if !repWrapper.IsRepeated {
		t.Errorf("expected IsRepeated true")
	}
	if got := repWrapper.Type.TypeString(); got != "List<String>" {
		t.Errorf("got %q, want List<String>", got)
	}

	intField := &api.Field{
		Name:  "count",
		Typez: api.TypezInt32,
	}
	intWrapper := FieldTypeToJavaType(intField)
	if intWrapper.Type.Kind != ast.KindPrimitive || intWrapper.Type.Name != "int" {
		t.Errorf("got %v, want primitive int", intWrapper.Type)
	}
}

func TestWellKnownTypes(t *testing.T) {
	emptyMsg := &api.Message{
		ID:   ".google.protobuf.Empty",
		Name: "Empty",
	}
	emptyType := MessageToJavaType(emptyMsg)
	if got, want := emptyType.FullName(), "com.google.protobuf.Empty"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}

	customMsg := &api.Message{
		ID:      ".google.cloud.secretmanager.v1.Secret",
		Name:    "Secret",
		Package: "google.cloud.secretmanager.v1",
	}
	customType := MessageToJavaType(customMsg)
	if got, want := customType.FullName(), "com.google.cloud.secretmanager.v1.Secret"; got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}
