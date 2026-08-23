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
	"encoding/json"
)

// GapicMetadata represents the gapic_metadata.json structure.
type GapicMetadata struct {
	Schema         string                      `json:"schema"`
	Comment        string                      `json:"comment"`
	Language       string                      `json:"language"`
	ProtoPackage   string                      `json:"proto_package"`
	LibraryPackage string                      `json:"library_package"`
	Services       map[string]*MetadataService `json:"services"`
}

// MetadataService represents service entries in gapic_metadata.json.
type MetadataService struct {
	LibraryClient string                  `json:"library_client"`
	Rpcs          map[string]*MetadataRPC `json:"rpcs"`
}

// MetadataRPC represents RPC entries in gapic_metadata.json.
type MetadataRPC struct {
	Methods []string `json:"methods"`
}

// ComposeGapicMetadata creates the GapicMetadata data structure.
func ComposeGapicMetadata(ann *ModelAnnotations) *GapicMetadata {
	meta := &GapicMetadata{
		Schema:         "1.0",
		Comment:        "This file maps proto services/RPCs to the corresponding library clients/methods.",
		Language:       "java",
		ProtoPackage:   ann.PackageName,
		LibraryPackage: ann.PackageName,
		Services:       make(map[string]*MetadataService),
	}

	for _, svc := range ann.Services {
		svcMeta := &MetadataService{
			LibraryClient: svc.ClientName,
			Rpcs:          make(map[string]*MetadataRPC),
		}
		for _, m := range svc.Methods {
			var methodNames []string
			methodNames = append(methodNames, m.MethodName, m.CallableName)
			if m.IsPaged {
				methodNames = append(methodNames, m.PagedCallableName)
			}
			if m.IsLRO {
				methodNames = append(methodNames, m.MethodName+"Async", m.OperationCallableName)
			}
			svcMeta.Rpcs[m.Name] = &MetadataRPC{Methods: methodNames}
		}
		meta.Services[svc.Name] = svcMeta
	}

	return meta
}

// WriteGapicMetadata serializes GapicMetadata to formatted JSON.
func WriteGapicMetadata(meta *GapicMetadata) ([]byte, error) {
	return json.MarshalIndent(meta, "", "  ")
}
