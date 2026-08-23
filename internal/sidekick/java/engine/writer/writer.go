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

// Package writer formats Java AST into idiomatic Java source files with automated import management.
package writer

import (
	"fmt"
	"slices"
	"strings"

	"github.com/googleapis/librarian/internal/license"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

// WriteClass formats an ast.ClassDefinition into a Java source code string.
func WriteClass(classDef *ast.ClassDefinition) (string, error) {
	var sb strings.Builder

	// 1. License header
	for _, line := range license.HeaderBulk() {
		sb.WriteString(line)
		sb.WriteString("\n")
	}
	sb.WriteString("\n")

	// 2. Package declaration
	if classDef.Package != "" {
		fmt.Fprintf(&sb, "package %s;\n\n", classDef.Package)
	}

	// 3. Import resolution
	imports, staticImports := resolveImports(classDef)
	if len(staticImports) > 0 {
		for _, imp := range staticImports {
			fmt.Fprintf(&sb, "import static %s;\n", imp)
		}
		sb.WriteString("\n")
	}
	if len(imports) > 0 {
		for _, imp := range imports {
			fmt.Fprintf(&sb, "import %s;\n", imp)
		}
		sb.WriteString("\n")
	}

	// 4. Class body
	writeClassDefinition(&sb, classDef, "")

	return sb.String(), nil
}

func writeClassDefinition(sb *strings.Builder, classDef *ast.ClassDefinition, indent string) {
	// Javadoc
	if classDef.JavaDoc != "" {
		writeJavaDoc(sb, classDef.JavaDoc, indent)
	}

	// Annotations
	for _, ann := range classDef.Annotations {
		writeAnnotation(sb, ann, indent)
	}

	// Class declaration line
	sb.WriteString(indent)
	if classDef.Scope != ast.PackagePrivate {
		sb.WriteString(string(classDef.Scope) + " ")
	}
	for _, m := range classDef.Modifiers {
		sb.WriteString(string(m) + " ")
	}
	kind := classDef.Kind
	if kind == "" {
		kind = ast.ClassKindClass
	}
	sb.WriteString(string(kind) + " " + classDef.Name)

	if classDef.Extends != nil {
		sb.WriteString(" extends " + classDef.Extends.TypeString())
	}
	if len(classDef.Implements) > 0 {
		var implStrs []string
		for _, imp := range classDef.Implements {
			implStrs = append(implStrs, imp.TypeString())
		}
		sb.WriteString(" implements " + strings.Join(implStrs, ", "))
	}
	sb.WriteString(" {\n")

	innerIndent := indent + "  "

	// Fields
	for _, f := range classDef.Fields {
		sb.WriteString("\n")
		if f.JavaDoc != "" {
			writeJavaDoc(sb, f.JavaDoc, innerIndent)
		}
		for _, ann := range f.Annotations {
			writeAnnotation(sb, ann, innerIndent)
		}
		sb.WriteString(innerIndent)
		if f.Scope != ast.PackagePrivate {
			sb.WriteString(string(f.Scope) + " ")
		}
		for _, m := range f.Modifiers {
			sb.WriteString(string(m) + " ")
		}
		sb.WriteString(f.Type.TypeString() + " " + f.Name)
		if f.Initializer != nil {
			sb.WriteString(" = " + f.Initializer.FormatExpression())
		}
		sb.WriteString(";\n")
	}

	// Constructors
	for _, c := range classDef.Constructors {
		sb.WriteString("\n")
		if c.JavaDoc != "" {
			writeJavaDoc(sb, c.JavaDoc, innerIndent)
		}
		for _, ann := range c.Annotations {
			writeAnnotation(sb, ann, innerIndent)
		}
		sb.WriteString(innerIndent)
		if c.Scope != ast.PackagePrivate {
			sb.WriteString(string(c.Scope) + " ")
		}
		sb.WriteString(classDef.Name + "(")
		var paramStrs []string
		for _, p := range c.Parameters {
			paramStr := ""
			if p.Final {
				paramStr += "final "
			}
			paramStr += p.Type.TypeString() + " " + p.Name
			paramStrs = append(paramStrs, paramStr)
		}
		sb.WriteString(strings.Join(paramStrs, ", ") + ")")
		if len(c.Throws) > 0 {
			var throwStrs []string
			for _, th := range c.Throws {
				throwStrs = append(throwStrs, th.TypeString())
			}
			sb.WriteString(" throws " + strings.Join(throwStrs, ", "))
		}
		sb.WriteString(" {\n")
		for _, stmt := range c.Body {
			formatted := stmt.FormatStatement(innerIndent + "  ")
			if formatted != "" {
				sb.WriteString(formatted + "\n")
			}
		}
		sb.WriteString(innerIndent + "}\n")
	}

	// Methods
	for _, m := range classDef.Methods {
		sb.WriteString("\n")
		if m.JavaDoc != "" {
			writeJavaDoc(sb, m.JavaDoc, innerIndent)
		}
		for _, ann := range m.Annotations {
			writeAnnotation(sb, ann, innerIndent)
		}
		sb.WriteString(innerIndent)
		if m.Scope != ast.PackagePrivate {
			sb.WriteString(string(m.Scope) + " ")
		}
		for _, mod := range m.Modifiers {
			sb.WriteString(string(mod) + " ")
		}
		returnType := "void"
		if m.ReturnType != nil {
			returnType = m.ReturnType.TypeString()
		}
		sb.WriteString(returnType + " " + m.Name + "(")
		var paramStrs []string
		for _, p := range m.Parameters {
			paramStr := ""
			if p.Final {
				paramStr += "final "
			}
			paramStr += p.Type.TypeString() + " " + p.Name
			paramStrs = append(paramStrs, paramStr)
		}
		sb.WriteString(strings.Join(paramStrs, ", ") + ")")
		if len(m.Throws) > 0 {
			var throwStrs []string
			for _, th := range m.Throws {
				throwStrs = append(throwStrs, th.TypeString())
			}
			sb.WriteString(" throws " + strings.Join(throwStrs, ", "))
		}

		if slices.Contains(m.Modifiers, ast.Abstract) || classDef.Kind == ast.ClassKindInterface {
			sb.WriteString(";\n")
		} else {
			sb.WriteString(" {\n")
			for _, stmt := range m.Body {
				formatted := stmt.FormatStatement(innerIndent + "  ")
				if formatted != "" {
					sb.WriteString(formatted + "\n")
				}
			}
			sb.WriteString(innerIndent + "}\n")
		}
	}

	// Inner classes
	for _, inner := range classDef.InnerClasses {
		sb.WriteString("\n")
		writeClassDefinition(sb, inner, innerIndent)
	}

	sb.WriteString(indent + "}\n")
}

func writeJavaDoc(sb *strings.Builder, doc string, indent string) {
	sanitized := lexicon.SanitizeComment(doc)
	lines := strings.Split(sanitized, "\n")
	sb.WriteString(indent + "/**\n")
	for _, l := range lines {
		trimmed := strings.TrimRight(l, " \t\r")
		if trimmed == "" {
			sb.WriteString(indent + " *\n")
		} else {
			sb.WriteString(indent + " * " + trimmed + "\n")
		}
	}
	sb.WriteString(indent + " */\n")
}

func writeAnnotation(sb *strings.Builder, ann *ast.AnnotationNode, indent string) {
	sb.WriteString(indent + "@" + ann.Type.Name)
	if ann.SingleValue != nil {
		sb.WriteString("(" + ann.SingleValue.FormatExpression() + ")")
	} else if len(ann.Properties) > 0 {
		var propStrs []string
		var keys []string
		for k := range ann.Properties {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		for _, k := range keys {
			propStrs = append(propStrs, fmt.Sprintf("%s = %s", k, ann.Properties[k].FormatExpression()))
		}
		sb.WriteString("(" + strings.Join(propStrs, ", ") + ")")
	}
	sb.WriteString("\n")
}

func resolveImports(classDef *ast.ClassDefinition) ([]string, []string) {
	typeMap := make(map[string]bool)
	collectTypesFromClass(classDef, typeMap)

	var imports []string
	seen := make(map[string]bool)

	for fullType := range typeMap {
		if fullType == "" {
			continue
		}
		idx := strings.LastIndex(fullType, ".")
		if idx == -1 {
			continue
		}
		pkg := fullType[:idx]
		if pkg == "java.lang" || pkg == classDef.Package {
			continue
		}
		if !seen[fullType] {
			seen[fullType] = true
			imports = append(imports, fullType)
		}
	}

	for _, extra := range classDef.ExtraImports {
		if extra != "" && !seen[extra] {
			seen[extra] = true
			imports = append(imports, extra)
		}
	}

	slices.Sort(imports)
	staticImports := slices.Clone(classDef.StaticImports)
	slices.Sort(staticImports)

	return imports, staticImports
}

func collectTypesFromClass(classDef *ast.ClassDefinition, typeMap map[string]bool) {
	if classDef.Extends != nil {
		collectType(classDef.Extends, typeMap)
	}
	for _, imp := range classDef.Implements {
		collectType(imp, typeMap)
	}
	for _, ann := range classDef.Annotations {
		collectType(ann.Type, typeMap)
	}
	for _, f := range classDef.Fields {
		collectType(f.Type, typeMap)
		for _, ann := range f.Annotations {
			collectType(ann.Type, typeMap)
		}
	}
	for _, c := range classDef.Constructors {
		for _, p := range c.Parameters {
			collectType(p.Type, typeMap)
		}
		for _, th := range c.Throws {
			collectType(th, typeMap)
		}
		for _, ann := range c.Annotations {
			collectType(ann.Type, typeMap)
		}
	}
	for _, m := range classDef.Methods {
		if m.ReturnType != nil {
			collectType(m.ReturnType, typeMap)
		}
		for _, p := range m.Parameters {
			collectType(p.Type, typeMap)
		}
		for _, th := range m.Throws {
			collectType(th, typeMap)
		}
		for _, ann := range m.Annotations {
			collectType(ann.Type, typeMap)
		}
	}
	for _, inner := range classDef.InnerClasses {
		collectTypesFromClass(inner, typeMap)
	}
}

func collectType(t *ast.TypeNode, typeMap map[string]bool) {
	if t == nil {
		return
	}
	if t.Package != "" && t.Kind == ast.KindObject {
		typeMap[t.FullName()] = true
	}
	for _, g := range t.Generics {
		collectType(g, typeMap)
	}
	if t.ExtendsBound != nil {
		collectType(t.ExtendsBound, typeMap)
	}
	if t.SuperBound != nil {
		collectType(t.SuperBound, typeMap)
	}
}
