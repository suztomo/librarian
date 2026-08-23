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
)

func TestAnnotateModel(t *testing.T) {
	reqMsg := api.NewTestMessage("GetSecretRequest").
		WithPackage("google.cloud.secretmanager.v1").
		WithFields(api.NewTestField("name").WithType(api.TypezString))

	respMsg := api.NewTestMessage("Secret").
		WithPackage("google.cloud.secretmanager.v1").
		WithFields(api.NewTestField("name").WithType(api.TypezString))

	method := api.NewTestMethod("GetSecret").
		WithInput(reqMsg).
		WithOutput(respMsg).
		WithSignatures(&api.MethodSignature{
			Names: []string{"name"},
		})

	svc := api.NewTestService("SecretManagerService").
		WithPackage("google.cloud.secretmanager.v1").
		WithMethods(method)
	svc.DefaultHost = "secretmanager.googleapis.com"

	model := api.NewTestAPI([]*api.Message{reqMsg, respMsg}, nil, []*api.Service{svc}).
		WithPackageName("google.cloud.secretmanager.v1")

	codec := NewCodec(map[string]string{
		"transport": "grpc+rest",
		"metadata":  "true",
	})

	ann, err := AnnotateModel(model, codec)
	if err != nil {
		t.Fatalf("AnnotateModel failed: %v", err)
	}

	if got, want := ann.PackageName, "com.google.cloud.secretmanager.v1"; got != want {
		t.Errorf("PackageName = %q, want %q", got, want)
	}
	if got, want := ann.StubPackageName, "com.google.cloud.secretmanager.v1.stub"; got != want {
		t.Errorf("StubPackageName = %q, want %q", got, want)
	}
	if len(ann.Services) != 1 {
		t.Fatalf("Services count = %d, want 1", len(ann.Services))
	}

	svcAnn := ann.Services[0]
	if got, want := svcAnn.ClientName, "SecretManagerServiceClient"; got != want {
		t.Errorf("ClientName = %q, want %q", got, want)
	}
	if got, want := svcAnn.SettingsName, "SecretManagerServiceSettings"; got != want {
		t.Errorf("SettingsName = %q, want %q", got, want)
	}
	if got, want := svcAnn.GrpcStubName, "GrpcSecretManagerServiceStub"; got != want {
		t.Errorf("GrpcStubName = %q, want %q", got, want)
	}
	if got, want := svcAnn.HttpJsonStubName, "HttpJsonSecretManagerServiceStub"; got != want {
		t.Errorf("HttpJsonStubName = %q, want %q", got, want)
	}
	if !svcAnn.HasGrpc || !svcAnn.HasHttpJson {
		t.Errorf("Expected HasGrpc=true, HasHttpJson=true, got grpc=%v, rest=%v", svcAnn.HasGrpc, svcAnn.HasHttpJson)
	}

	if len(svcAnn.Methods) != 1 {
		t.Fatalf("Methods count = %d, want 1", len(svcAnn.Methods))
	}
	mAnn := svcAnn.Methods[0]
	if got, want := mAnn.MethodName, "getSecret"; got != want {
		t.Errorf("MethodName = %q, want %q", got, want)
	}
	if got, want := mAnn.CallableName, "getSecretCallable"; got != want {
		t.Errorf("CallableName = %q, want %q", got, want)
	}
	if !mAnn.IsUnary {
		t.Errorf("Expected IsUnary=true")
	}
	if len(mAnn.Signatures) != 1 {
		t.Errorf("Signatures count = %d, want 1", len(mAnn.Signatures))
	}
}
