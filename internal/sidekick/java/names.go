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

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

// JavaPackage returns the Java package name for an API.
func JavaPackage(model *api.API, pkgOverride string) string {
	if pkgOverride != "" {
		return pkgOverride
	}
	return lexicon.JavaPackageFromProto(model.PackageName)
}

// StubPackage returns the stub subpackage for a Java package.
func StubPackage(javaPkg string) string {
	return javaPkg + ".stub"
}

// ClientClassName returns the main client class name for a service.
func ClientClassName(serviceName string) string {
	return lexicon.ToUpperCamel(serviceName) + "Client"
}

// SettingsClassName returns the settings class name for a service.
func SettingsClassName(serviceName string) string {
	return lexicon.ToUpperCamel(serviceName) + "Settings"
}

// StubClassName returns the abstract stub class name for a service.
func StubClassName(serviceName string) string {
	return lexicon.ToUpperCamel(serviceName) + "Stub"
}

// StubSettingsClassName returns the stub settings class name for a service.
func StubSettingsClassName(serviceName string) string {
	return lexicon.ToUpperCamel(serviceName) + "StubSettings"
}

// GrpcStubClassName returns the gRPC stub class name for a service.
func GrpcStubClassName(serviceName string) string {
	return "Grpc" + lexicon.ToUpperCamel(serviceName) + "Stub"
}

// GrpcCallableFactoryClassName returns the gRPC callable factory class name.
func GrpcCallableFactoryClassName(serviceName string) string {
	return "Grpc" + lexicon.ToUpperCamel(serviceName) + "CallableFactory"
}

// HttpJsonStubClassName returns the HTTP/JSON stub class name for a service.
func HttpJsonStubClassName(serviceName string) string {
	return "HttpJson" + lexicon.ToUpperCamel(serviceName) + "Stub"
}

// HttpJsonCallableFactoryClassName returns the HTTP/JSON callable factory class name.
func HttpJsonCallableFactoryClassName(serviceName string) string {
	return "HttpJson" + lexicon.ToUpperCamel(serviceName) + "CallableFactory"
}

// ResourceClassName returns the helper class name for a resource definition.
func ResourceClassName(resType string) string {
	// e.g. "secretmanager.googleapis.com/Secret" -> "SecretName"
	// or "Topic" -> "TopicName"
	parts := strings.Split(resType, "/")
	rawName := parts[len(parts)-1]
	camel := lexicon.ToUpperCamel(rawName)
	if strings.HasSuffix(camel, "Name") {
		return camel
	}
	return camel + "Name"
}

// MethodName returns the Java method name for an RPC.
func MethodName(rpcName string) string {
	return lexicon.ToLowerCamel(rpcName)
}

// CallableMethodName returns the callable accessor method name for an RPC.
func CallableMethodName(rpcName string) string {
	return lexicon.ToLowerCamel(rpcName) + "Callable"
}

// PagedCallableMethodName returns the paged callable accessor method name for an RPC.
func PagedCallableMethodName(rpcName string) string {
	return lexicon.ToLowerCamel(rpcName) + "PagedCallable"
}

// OperationCallableMethodName returns the operation callable accessor method name for an RPC.
func OperationCallableMethodName(rpcName string) string {
	return lexicon.ToLowerCamel(rpcName) + "OperationCallable"
}

// FieldName returns the Java field name for a proto field.
func FieldName(protoFieldName string) string {
	return lexicon.ToLowerCamel(protoFieldName)
}

// GetterName returns the Java getter method name for a field.
func GetterName(protoFieldName string) string {
	return "get" + lexicon.ToUpperCamel(protoFieldName)
}

// SetterName returns the Java setter method name for a field.
func SetterName(protoFieldName string) string {
	return "set" + lexicon.ToUpperCamel(protoFieldName)
}
