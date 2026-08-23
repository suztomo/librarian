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
	"os"
	"path/filepath"
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/parser"
)

func TestGenerateEndToEnd(t *testing.T) {
	reqMsg := api.NewTestMessage("GetSecretRequest").
		WithPackage("google.cloud.secretmanager.v1").
		WithFields(api.NewTestField("name").WithType(api.TypezString))

	respMsg := api.NewTestMessage("Secret").
		WithPackage("google.cloud.secretmanager.v1").
		WithFields(api.NewTestField("name").WithType(api.TypezString))

	res := api.NewTestResource("secretmanager.googleapis.com/Secret").
		WithPatterns(api.ResourcePattern{
			{Literal: "projects"},
			{Variable: &api.PathVariable{FieldPath: []string{"project"}}},
			{Literal: "secrets"},
			{Variable: &api.PathVariable{FieldPath: []string{"secret"}}},
		})

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
	model.ResourceDefinitions = []*api.Resource{res}

	outdir := t.TempDir()
	codecMap := map[string]string{
		"transport":             "grpc+rest",
		"metadata":              "true",
		"generate-version-java": "true",
	}

	if err := Generate(t.Context(), model, outdir, codecMap); err != nil {
		t.Fatalf("Generate failed: %v", err)
	}

	expectedFiles := []string{
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "SecretManagerServiceClient.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "SecretManagerServiceSettings.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "SecretName.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "Version.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "package-info.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "SecretManagerServiceStub.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "SecretManagerServiceStubSettings.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "GrpcSecretManagerServiceStub.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "GrpcSecretManagerServiceCallableFactory.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "HttpJsonSecretManagerServiceStub.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "HttpJsonSecretManagerServiceCallableFactory.java"),
		filepath.Join(outdir, "com", "google", "cloud", "secretmanager", "v1", "stub", "package-info.java"),
		filepath.Join(outdir, "gapic_metadata.json"),
		filepath.Join(outdir, "resources", "META-INF", "native-image", "com", "google", "cloud", "secretmanager", "v1", "reflect-config.json"),
	}

	for _, f := range expectedFiles {
		if _, err := os.Stat(f); os.IsNotExist(err) {
			t.Errorf("Expected generated file %s does not exist", f)
		}
	}
}

func TestGenerateWithConfig(t *testing.T) {
	reqMsg := api.NewTestMessage("GetEchoRequest").
		WithPackage("google.showcase.v1beta1").
		WithFields(api.NewTestField("content").WithType(api.TypezString))

	respMsg := api.NewTestMessage("EchoResponse").
		WithPackage("google.showcase.v1beta1").
		WithFields(api.NewTestField("content").WithType(api.TypezString))

	method := api.NewTestMethod("Echo").
		WithInput(reqMsg).
		WithOutput(respMsg)

	svc := api.NewTestService("Echo").
		WithPackage("google.showcase.v1beta1").
		WithMethods(method)
	svc.DefaultHost = "localhost:7469"

	model := api.NewTestAPI([]*api.Message{reqMsg, respMsg}, nil, []*api.Service{svc}).
		WithPackageName("google.showcase.v1beta1")

	cfg := &parser.ModelConfig{
		Codec: map[string]string{
			"transport": "grpc",
		},
	}

	outdir := t.TempDir()
	if err := GenerateWithConfig(t.Context(), model, outdir, cfg); err != nil {
		t.Fatalf("GenerateWithConfig failed: %v", err)
	}

	clientPath := filepath.Join(outdir, "com", "google", "showcase", "v1beta1", "EchoClient.java")
	if _, err := os.Stat(clientPath); os.IsNotExist(err) {
		t.Errorf("EchoClient.java was not generated at %s", clientPath)
	}
}
