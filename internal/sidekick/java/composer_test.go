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
	"strings"
	"testing"

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/writer"
)

func TestComposeAll(t *testing.T) {
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

	codec := NewCodec(map[string]string{
		"transport":             "grpc+rest",
		"metadata":              "true",
		"generate-version-java": "true",
	})

	ann, err := AnnotateModel(model, codec)
	if err != nil {
		t.Fatalf("AnnotateModel failed: %v", err)
	}

	artifacts, err := ComposeAll(ann)
	if err != nil {
		t.Fatalf("ComposeAll failed: %v", err)
	}

	if len(artifacts.Classes) == 0 {
		t.Fatalf("Expected classes to be composed, got 0")
	}

	// Verify each class formats properly through the AST writer
	for _, cls := range artifacts.Classes {
		src, err := writer.WriteClass(cls)
		if err != nil {
			t.Errorf("WriteClass failed for %s: %v", cls.Name, err)
		}
		if len(src) == 0 {
			t.Errorf("Generated empty source code for %s", cls.Name)
		}
	}

	// Check package infos
	if len(artifacts.PackageInfos) != 2 {
		t.Errorf("Expected 2 package infos, got %d", len(artifacts.PackageInfos))
	}
	for _, pkgInfo := range artifacts.PackageInfos {
		src := WritePackageInfo(pkgInfo)
		if len(src) == 0 {
			t.Errorf("WritePackageInfo produced empty source")
		}
	}

	// Check metadata
	if artifacts.GapicMetadata == nil {
		t.Errorf("Expected GapicMetadata to be non-nil")
	} else {
		metaBytes, err := WriteGapicMetadata(artifacts.GapicMetadata)
		if err != nil || len(metaBytes) == 0 {
			t.Errorf("WriteGapicMetadata failed: %v", err)
		}
	}

	// Check reflect config
	if len(artifacts.ReflectConfigs) == 0 {
		t.Errorf("Expected ReflectConfigs to be populated")
	} else {
		reflectBytes, err := WriteReflectConfig(artifacts.ReflectConfigs)
		if err != nil || len(reflectBytes) == 0 {
			t.Errorf("WriteReflectConfig failed: %v", err)
		}
	}
}

func TestComposeMultiPatternResourceName(t *testing.T) {
	res := &ResourceAnnotation{
		Type:        "secretmanager.googleapis.com/Secret",
		ClassName:   "SecretName",
		PackageName: "com.google.cloud.secretmanager.v1",
		Patterns: []string{
			"projects/{project}/secrets/{secret}",
			"projects/{project}/locations/{location}/secrets/{secret}",
		},
	}

	cls := ComposeResourceNameClass(res)
	if cls == nil {
		t.Fatalf("ComposeResourceNameClass returned nil")
	}

	src, err := writer.WriteClass(cls)
	if err != nil {
		t.Fatalf("WriteClass failed: %v", err)
	}

	if !strings.Contains(src, "PROJECT_SECRET_PATH_TEMPLATE") {
		t.Errorf("Missing PROJECT_SECRET_PATH_TEMPLATE in:\n%s", src)
	}
	if !strings.Contains(src, "PROJECT_LOCATION_SECRET_PATH_TEMPLATE") {
		t.Errorf("Missing PROJECT_LOCATION_SECRET_PATH_TEMPLATE in:\n%s", src)
	}
	if !strings.Contains(src, "ofProjectSecretName") {
		t.Errorf("Missing ofProjectSecretName in:\n%s", src)
	}
	if !strings.Contains(src, "ofProjectLocationSecretName") {
		t.Errorf("Missing ofProjectLocationSecretName in:\n%s", src)
	}
	if !strings.Contains(src, "PROJECT_SECRET_PATH_TEMPLATE.matches(formattedString) || PROJECT_LOCATION_SECRET_PATH_TEMPLATE.matches(formattedString)") {
		t.Errorf("Missing isParsableFrom multi-pattern check in:\n%s", src)
	}
}
