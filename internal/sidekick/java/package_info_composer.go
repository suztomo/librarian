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
	"fmt"
	"strings"

	"github.com/googleapis/librarian/internal/license"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

// ComposePackageInfo creates package-info metadata for main and stub packages.
func ComposePackageInfo(pkgName, title string, isStub bool) *PackageInfo {
	desc := fmt.Sprintf("A client to %s", title)
	if isStub {
		desc = fmt.Sprintf("Provides stub implementations for %s", title)
	}
	return &PackageInfo{
		PackageName: pkgName,
		Description: desc,
		IsStub:      isStub,
	}
}

// WritePackageInfo generates the contents of a package-info.java file.
func WritePackageInfo(info *PackageInfo) string {
	var sb strings.Builder

	// License
	for _, line := range license.HeaderBulk() {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// Javadoc
	sb.WriteString("/**\n")
	fmt.Fprintf(&sb, " * %s\n", lexicon.SanitizeComment(info.Description))
	sb.WriteString(" */\n")

	// Annotations
	sb.WriteString("@Generated(\"by gapic-generator-java\")\n")
	fmt.Fprintf(&sb, "package %s;\n\n", info.PackageName)
	sb.WriteString("import javax.annotation.Generated;\n")

	return sb.String()
}
