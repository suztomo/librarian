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

package protoparser

import (
	"testing"
)

func TestParsePluginArguments(t *testing.T) {
	param := "grpc-service-config=/tmp/logging_grpc_service_config.json,gapic-config=/tmp/logging_gapic.yaml,api-service-config=/tmp/logging.yaml,transport=grpc+rest,repo=googleapis/google-cloud-java,artifact=google-cloud-logging,metadata,rest-numeric-enums,generate-version-java"
	args := ParsePluginArgumentsString(param)

	if args.GrpcServiceConfigPath != "/tmp/logging_grpc_service_config.json" {
		t.Errorf("unexpected GrpcServiceConfigPath: %s", args.GrpcServiceConfigPath)
	}
	if args.GapicYamlConfigPath != "/tmp/logging_gapic.yaml" {
		t.Errorf("unexpected GapicYamlConfigPath: %s", args.GapicYamlConfigPath)
	}
	if args.ServiceYamlConfigPath != "/tmp/logging.yaml" {
		t.Errorf("unexpected ServiceYamlConfigPath: %s", args.ServiceYamlConfigPath)
	}
	if args.Transport != "grpc+rest" {
		t.Errorf("unexpected Transport: %s", args.Transport)
	}
	if args.Repo != "googleapis/google-cloud-java" {
		t.Errorf("unexpected Repo: %s", args.Repo)
	}
	if args.Artifact != "google-cloud-logging" {
		t.Errorf("unexpected Artifact: %s", args.Artifact)
	}
	if !args.HasMetadata {
		t.Errorf("expected HasMetadata to be true")
	}
	if !args.HasNumericEnum {
		t.Errorf("expected HasNumericEnum to be true")
	}
	if !args.HasGenerateVersionJava {
		t.Errorf("expected HasGenerateVersionJava to be true")
	}
}

func TestParsePluginArgumentsFileEndings(t *testing.T) {
	param := "/tmp/foo_grpc_service_config.json,/tmp/foo_gapic.yaml"
	args := ParsePluginArgumentsString(param)

	if args.GrpcServiceConfigPath != "/tmp/foo_grpc_service_config.json" {
		t.Errorf("unexpected GrpcServiceConfigPath: %s", args.GrpcServiceConfigPath)
	}
	if args.GapicYamlConfigPath != "/tmp/foo_gapic.yaml" {
		t.Errorf("unexpected GapicYamlConfigPath: %s", args.GapicYamlConfigPath)
	}
}
