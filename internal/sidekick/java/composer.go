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
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
)

// ComposedArtifacts contains all AST classes, package infos, and metadata files generated for an API.
type ComposedArtifacts struct {
	Classes        []*ast.ClassDefinition
	PackageInfos   []*PackageInfo
	GapicMetadata  *GapicMetadata
	ReflectConfigs []*ReflectConfig
}

// PackageInfo represents package-info.java metadata.
type PackageInfo struct {
	PackageName string
	Description string
	IsStub      bool
}

// ComposeAll runs all class and metadata composers for the given model annotations.
func ComposeAll(ann *ModelAnnotations) (*ComposedArtifacts, error) {
	artifacts := &ComposedArtifacts{}

	// 1. Compose Package-Info
	artifacts.PackageInfos = append(artifacts.PackageInfos,
		ComposePackageInfo(ann.PackageName, ann.Title, false),
		ComposePackageInfo(ann.StubPackageName, ann.Title, true),
	)

	// 2. Compose per-service classes
	for _, svc := range ann.Services {
		// Main Client class
		clientClass := ComposeClientClass(svc, ann)
		artifacts.Classes = append(artifacts.Classes, clientClass)

		// Main Settings class
		settingsClass := ComposeSettingsClass(svc, ann)
		artifacts.Classes = append(artifacts.Classes, settingsClass)

		// Stub abstract base class
		stubClass := ComposeStubClass(svc, ann)
		artifacts.Classes = append(artifacts.Classes, stubClass)

		// StubSettings class
		stubSettingsClass := ComposeStubSettingsClass(svc, ann)
		artifacts.Classes = append(artifacts.Classes, stubSettingsClass)

		// gRPC transport stub and factory
		if svc.HasGrpc {
			grpcStub := ComposeGrpcStubClass(svc, ann)
			grpcFactory := ComposeGrpcCallableFactoryClass(svc, ann)
			artifacts.Classes = append(artifacts.Classes, grpcStub, grpcFactory)
		}

		// HTTP/JSON transport stub and factory
		if svc.HasHttpJson {
			httpJsonStub := ComposeHttpJsonStubClass(svc, ann)
			httpJsonFactory := ComposeHttpJsonCallableFactoryClass(svc, ann)
			artifacts.Classes = append(artifacts.Classes, httpJsonStub, httpJsonFactory)
		}
	}

	// 3. Compose Resource Name helper classes
	for _, res := range ann.Resources {
		resClass := ComposeResourceNameClass(res)
		if resClass != nil {
			artifacts.Classes = append(artifacts.Classes, resClass)
		}
	}

	// 4. Compose Version.java if requested
	if ann.HasVersionJava && len(ann.Services) > 0 {
		versionClass := ComposeVersionClass(ann.PackageName)
		artifacts.Classes = append(artifacts.Classes, versionClass)
	}

	// 5. Compose GAPIC Metadata if requested
	if ann.HasMetadata {
		artifacts.GapicMetadata = ComposeGapicMetadata(ann)
		artifacts.ReflectConfigs = ComposeReflectConfig(ann)
	}

	return artifacts, nil
}
