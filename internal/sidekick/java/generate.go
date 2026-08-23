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

// Package java implements the Java code generator for the Sidekick model.
package java

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/writer"
	"github.com/googleapis/librarian/internal/sidekick/parser"
)

// Generate generates Java client library files from an api.API model.
func Generate(ctx context.Context, model *api.API, outdir string, codecMap map[string]string) error {
	codec := NewCodec(codecMap)
	return generateWithCodec(ctx, model, outdir, codec)
}

// GenerateWithConfig generates Java client library files using a ModelConfig.
func GenerateWithConfig(ctx context.Context, model *api.API, outdir string, cfg *parser.ModelConfig) error {
	codec := NewCodecFromModelConfig(cfg)
	return generateWithCodec(ctx, model, outdir, codec)
}

func generateWithCodec(ctx context.Context, model *api.API, outdir string, codec *Codec) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if model == nil {
		return fmt.Errorf("cannot generate Java code: model is nil")
	}

	ann, err := AnnotateModel(model, codec)
	if err != nil {
		return fmt.Errorf("failed to annotate model for Java: %w", err)
	}

	artifacts, err := ComposeAll(ann)
	if err != nil {
		return fmt.Errorf("failed to compose Java classes: %w", err)
	}

	// 1. Write Java class files
	for _, cls := range artifacts.Classes {
		src, err := writer.WriteClass(cls)
		if err != nil {
			return fmt.Errorf("failed to write Java class %s: %w", cls.Name, err)
		}
		targetPath := filepath.Join(outdir, packageToPath(cls.Package), cls.Name+".java")
		if err := writeFile(targetPath, []byte(src)); err != nil {
			return err
		}
	}

	// 2. Write package-info.java files
	for _, pkgInfo := range artifacts.PackageInfos {
		src := WritePackageInfo(pkgInfo)
		targetPath := filepath.Join(outdir, packageToPath(pkgInfo.PackageName), "package-info.java")
		if err := writeFile(targetPath, []byte(src)); err != nil {
			return err
		}
	}

	// 3. Write gapic_metadata.json
	if artifacts.GapicMetadata != nil {
		metaBytes, err := WriteGapicMetadata(artifacts.GapicMetadata)
		if err != nil {
			return fmt.Errorf("failed to write gapic_metadata.json: %w", err)
		}
		targetPath := filepath.Join(outdir, "gapic_metadata.json")
		if err := writeFile(targetPath, metaBytes); err != nil {
			return err
		}
	}

	// 4. Write reflect-config.json
	if len(artifacts.ReflectConfigs) > 0 {
		reflectBytes, err := WriteReflectConfig(artifacts.ReflectConfigs)
		if err != nil {
			return fmt.Errorf("failed to write reflect-config.json: %w", err)
		}
		targetPath := filepath.Join(outdir, "resources", "META-INF", "native-image", packageToPath(ann.PackageName), "reflect-config.json")
		if err := writeFile(targetPath, reflectBytes); err != nil {
			return err
		}
	}

	return nil
}

func packageToPath(pkgName string) string {
	return strings.ReplaceAll(pkgName, ".", string(filepath.Separator))
}

func writeFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dir, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		return fmt.Errorf("failed to write file %q: %w", path, err)
	}
	return nil
}
