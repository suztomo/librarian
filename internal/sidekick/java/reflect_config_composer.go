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

// ReflectConfig represents a GraalVM native image reflection configuration entry.
type ReflectConfig struct {
	Name                    string `json:"name"`
	AllDeclaredConstructors bool   `json:"allDeclaredConstructors,omitempty"`
	AllPublicConstructors   bool   `json:"allPublicConstructors,omitempty"`
	AllDeclaredMethods      bool   `json:"allDeclaredMethods,omitempty"`
	AllPublicMethods        bool   `json:"allPublicMethods,omitempty"`
}

// ComposeReflectConfig generates the reflect-config.json entries for all generated classes.
func ComposeReflectConfig(ann *ModelAnnotations) []*ReflectConfig {
	var configs []*ReflectConfig

	for _, svc := range ann.Services {
		clientFQCN := ann.PackageName + "." + svc.ClientName
		settingsFQCN := ann.PackageName + "." + svc.SettingsName
		stubFQCN := ann.StubPackageName + "." + svc.StubName
		stubSettingsFQCN := ann.StubPackageName + "." + svc.StubSettingsName

		configs = append(configs,
			&ReflectConfig{
				Name:                  clientFQCN,
				AllPublicConstructors: true,
				AllPublicMethods:      true,
			},
			&ReflectConfig{
				Name:                  settingsFQCN,
				AllPublicConstructors: true,
				AllPublicMethods:      true,
			},
			&ReflectConfig{
				Name:                  stubFQCN,
				AllPublicConstructors: true,
				AllPublicMethods:      true,
			},
			&ReflectConfig{
				Name:                  stubSettingsFQCN,
				AllPublicConstructors: true,
				AllPublicMethods:      true,
			},
		)

		if svc.HasGrpc {
			configs = append(configs,
				&ReflectConfig{
					Name:                  ann.StubPackageName + "." + svc.GrpcStubName,
					AllPublicConstructors: true,
					AllPublicMethods:      true,
				},
				&ReflectConfig{
					Name:                  ann.StubPackageName + "." + svc.GrpcFactoryName,
					AllPublicConstructors: true,
					AllPublicMethods:      true,
				},
			)
		}

		if svc.HasHttpJson {
			configs = append(configs,
				&ReflectConfig{
					Name:                  ann.StubPackageName + "." + svc.HttpJsonStubName,
					AllPublicConstructors: true,
					AllPublicMethods:      true,
				},
				&ReflectConfig{
					Name:                  ann.StubPackageName + "." + svc.HttpJsonFactoryName,
					AllPublicConstructors: true,
					AllPublicMethods:      true,
				},
			)
		}
	}

	return configs
}

// WriteReflectConfig serializes []*ReflectConfig to formatted JSON bytes.
func WriteReflectConfig(configs []*ReflectConfig) ([]byte, error) {
	return json.MarshalIndent(configs, "", "  ")
}
