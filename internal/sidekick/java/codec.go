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

	"github.com/googleapis/librarian/internal/sidekick/parser"
)

// Codec holds Java generator configuration.
type Codec struct {
	PackageOverride     string
	Artifact            string
	Repo                string
	Transport           string
	HasMetadata         bool
	GenerateVersionJava bool
	RestNumericEnums    bool
	SkipFormat          bool
}

// NewCodec creates a Codec from a key-value map.
func NewCodec(codecMap map[string]string) *Codec {
	c := &Codec{
		Transport: "grpc",
	}
	if codecMap == nil {
		return c
	}
	if v, ok := codecMap["package"]; ok {
		c.PackageOverride = v
	}
	if v, ok := codecMap["java-package"]; ok {
		c.PackageOverride = v
	}
	if v, ok := codecMap["artifact"]; ok {
		c.Artifact = v
	}
	if v, ok := codecMap["repo"]; ok {
		c.Repo = v
	}
	if v, ok := codecMap["transport"]; ok {
		c.Transport = strings.ToLower(v)
	}
	if v, ok := codecMap["metadata"]; ok && (v == "true" || v == "1") {
		c.HasMetadata = true
	}
	if v, ok := codecMap["generate-version-java"]; ok && (v == "true" || v == "1") {
		c.GenerateVersionJava = true
	}
	if v, ok := codecMap["rest-numeric-enums"]; ok && (v == "true" || v == "1") {
		c.RestNumericEnums = true
	}
	if v, ok := codecMap["skip-format"]; ok && (v == "true" || v == "1") {
		c.SkipFormat = true
	}
	return c
}

// NewCodecFromModelConfig creates a Codec from a parser.ModelConfig.
func NewCodecFromModelConfig(cfg *parser.ModelConfig) *Codec {
	if cfg == nil {
		return NewCodec(nil)
	}
	return NewCodec(cfg.Codec)
}
