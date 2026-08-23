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

func TestNames(t *testing.T) {
	model := &api.API{
		PackageName: "google.cloud.secretmanager.v1",
	}

	if got, want := JavaPackage(model, ""), "com.google.cloud.secretmanager.v1"; got != want {
		t.Errorf("JavaPackage() = %q, want %q", got, want)
	}

	if got, want := JavaPackage(model, "custom.pkg"), "custom.pkg"; got != want {
		t.Errorf("JavaPackage(override) = %q, want %q", got, want)
	}

	if got, want := StubPackage("com.google.cloud.secretmanager.v1"), "com.google.cloud.secretmanager.v1.stub"; got != want {
		t.Errorf("StubPackage() = %q, want %q", got, want)
	}

	if got, want := ClientClassName("SecretManagerService"), "SecretManagerServiceClient"; got != want {
		t.Errorf("ClientClassName() = %q, want %q", got, want)
	}

	if got, want := SettingsClassName("SecretManagerService"), "SecretManagerServiceSettings"; got != want {
		t.Errorf("SettingsClassName() = %q, want %q", got, want)
	}

	if got, want := StubClassName("SecretManagerService"), "SecretManagerServiceStub"; got != want {
		t.Errorf("StubClassName() = %q, want %q", got, want)
	}

	if got, want := StubSettingsClassName("SecretManagerService"), "SecretManagerServiceStubSettings"; got != want {
		t.Errorf("StubSettingsClassName() = %q, want %q", got, want)
	}

	if got, want := GrpcStubClassName("SecretManagerService"), "GrpcSecretManagerServiceStub"; got != want {
		t.Errorf("GrpcStubClassName() = %q, want %q", got, want)
	}

	if got, want := HttpJsonStubClassName("SecretManagerService"), "HttpJsonSecretManagerServiceStub"; got != want {
		t.Errorf("HttpJsonStubClassName() = %q, want %q", got, want)
	}

	if got, want := ResourceClassName("secretmanager.googleapis.com/Secret"), "SecretName"; got != want {
		t.Errorf("ResourceClassName() = %q, want %q", got, want)
	}

	if got, want := ResourceClassName("TopicName"), "TopicName"; got != want {
		t.Errorf("ResourceClassName() = %q, want %q", got, want)
	}

	if got, want := MethodName("GetSecret"), "getSecret"; got != want {
		t.Errorf("MethodName() = %q, want %q", got, want)
	}

	if got, want := CallableMethodName("GetSecret"), "getSecretCallable"; got != want {
		t.Errorf("CallableMethodName() = %q, want %q", got, want)
	}

	if got, want := PagedCallableMethodName("ListSecrets"), "listSecretsPagedCallable"; got != want {
		t.Errorf("PagedCallableMethodName() = %q, want %q", got, want)
	}

	if got, want := OperationCallableMethodName("CreateSecret"), "createSecretOperationCallable"; got != want {
		t.Errorf("OperationCallableMethodName() = %q, want %q", got, want)
	}

	if got, want := GetterName("secret_payload"), "getSecretPayload"; got != want {
		t.Errorf("GetterName() = %q, want %q", got, want)
	}

	if got, want := SetterName("secret_payload"), "setSecretPayload"; got != want {
		t.Errorf("SetterName() = %q, want %q", got, want)
	}
}
