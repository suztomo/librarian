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
	"regexp"
	"strings"

	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

var variableRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?:=[^}]+)?\}`)

// ComposeResourceNameClass generates <Resource>Name.java AST.
func ComposeResourceNameClass(res *ResourceAnnotation) *ast.ClassDefinition {
	if res == nil || res.ClassName == "" {
		return nil
	}

	pattern := "projects/{project}/" + strings.ToLower(strings.TrimSuffix(res.ClassName, "Name")) + "s/{" + strings.ToLower(strings.TrimSuffix(res.ClassName, "Name")) + "}"
	if len(res.Patterns) > 0 {
		pattern = res.Patterns[0]
	}

	// Extract variables from pattern
	matches := variableRegex.FindAllStringSubmatch(pattern, -1)
	var varNames []string
	for _, m := range matches {
		if len(m) > 1 {
			varNames = append(varNames, m[1])
		}
	}
	if len(varNames) == 0 {
		varNames = []string{"project", strings.ToLower(strings.TrimSuffix(res.ClassName, "Name"))}
	}

	classType := ast.ObjectType(res.ClassName, res.PackageName)

	classDef := &ast.ClassDefinition{
		Package: res.PackageName,
		Name:    res.ClassName,
		Scope:   ast.Public,
		Kind:    ast.ClassKindClass,
		JavaDoc: fmt.Sprintf("Resource name class for {@link %s}.", res.Type),
		Implements: []*ast.TypeNode{
			TypeResourceName,
		},
		Annotations: []*ast.AnnotationNode{
			ast.NewAnnotation(ast.TypeGenerated).WithValue(ast.StringLiteralExpr("by gapic-generator-java")),
		},
		ExtraImports: []string{
			"java.util.Map",
			"java.util.HashMap",
			"java.util.Objects",
			"javax.annotation.Generated",
			"com.google.api.pathtemplate.PathTemplate",
			"com.google.api.resourcenames.ResourceName",
		},
	}

	// Static PATH_TEMPLATE field
	classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
		Name:        "PATH_TEMPLATE",
		Type:        ast.ObjectType("PathTemplate", "com.google.api.pathtemplate"),
		Scope:       ast.Private,
		Modifiers:   []ast.Modifier{ast.Static, ast.Final},
		Initializer: ast.ExprF("PathTemplate.createWithoutUrlEncoding(\"%s\")", pattern),
	})

	// Member fields
	for _, v := range varNames {
		classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
			Name:      lexicon.ToLowerCamel(v),
			Type:      ast.TypeString,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		})
	}

	// Private Constructor
	var ctorParams []*ast.ParameterDefinition
	var ctorStmts []ast.Statement
	for _, v := range varNames {
		paramName := lexicon.ToLowerCamel(v)
		ctorParams = append(ctorParams, &ast.ParameterDefinition{
			Name: paramName,
			Type: ast.TypeString,
		})
		ctorStmts = append(ctorStmts, ast.StatementFromF("this.%s = Objects.requireNonNull(%s);", paramName, paramName))
	}
	classDef.Constructors = append(classDef.Constructors, &ast.ConstructorDefinition{
		Scope:      ast.Private,
		Parameters: ctorParams,
		Body:       ctorStmts,
	})

	// of(...) static factory method
	var ofParams []*ast.ParameterDefinition
	var ofArgs []string
	for _, v := range varNames {
		paramName := lexicon.ToLowerCamel(v)
		ofParams = append(ofParams, &ast.ParameterDefinition{
			Name: paramName,
			Type: ast.TypeString,
		})
		ofArgs = append(ofArgs, paramName)
	}
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "of",
		ReturnType: classType,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: ofParams,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Expr: ast.ExprF("new %s(%s)", res.ClassName, strings.Join(ofArgs, ", ")),
			},
		},
	})

	// format(...) static factory method
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "format",
		ReturnType: ast.TypeString,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: ofParams,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Expr: ast.ExprF("of(%s).toString()", strings.Join(ofArgs, ", ")),
			},
		},
	})

	// parse(...) static factory method
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "parse",
		ReturnType: classType,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: []*ast.ParameterDefinition{
			{Name: "formattedString", Type: ast.TypeString},
		},
		Body: composeParseBody(varNames),
	})

	// isParsableFrom(...) static helper
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "isParsableFrom",
		ReturnType: ast.TypeBoolean,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: []*ast.ParameterDefinition{
			{Name: "formattedString", Type: ast.TypeString},
		},
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Expr: ast.Expr("PATH_TEMPLATE.matches(formattedString)"),
			},
		},
	})

	// Getters for variable fields
	for _, v := range varNames {
		fieldName := lexicon.ToLowerCamel(v)
		classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
			Name:       lexicon.ToUpperCamel("get_" + v),
			ReturnType: ast.TypeString,
			Scope:      ast.Public,
			Body: []ast.Statement{
				&ast.ReturnStatement{Expr: ast.Expr(fieldName)},
			},
		})
	}

	// toString(), equals(), hashCode()
	classDef.Methods = append(classDef.Methods,
		&ast.MethodDefinition{
			Name:       "toString",
			ReturnType: ast.TypeString,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: composeToStringBody(varNames),
		},
		&ast.MethodDefinition{
			Name:       "equals",
			ReturnType: ast.TypeBoolean,
			Scope:      ast.Public,
			Parameters: []*ast.ParameterDefinition{
				{Name: "o", Type: ast.TypeObject},
			},
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				ast.StatementFrom("if (o == this) { return true; }"),
				ast.StatementFromF("if (o instanceof %s) {\n"+
					"  %s that = (%s) o;\n"+
					"  return %s;\n"+
					"}\n"+
					"return false;",
					res.ClassName, res.ClassName, res.ClassName,
					composeEqualsCondition(varNames),
				),
			},
		},
		&ast.MethodDefinition{
			Name:       "hashCode",
			ReturnType: ast.TypeInt,
			Scope:      ast.Public,
			Annotations: []*ast.AnnotationNode{
				ast.NewAnnotation(ast.ObjectType("Override", "java.lang")),
			},
			Body: []ast.Statement{
				&ast.ReturnStatement{
					Expr: ast.ExprF("Objects.hash(%s)", strings.Join(ofArgs, ", ")),
				},
			},
		},
	)

	return classDef
}

func composeParseBody(varNames []string) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("if (formattedString.isEmpty()) { return null; }"),
		ast.StatementFrom("Map<String, String> matchMap = PATH_TEMPLATE.validatedMatch(formattedString, \"parse: \" + formattedString);"),
	)
	var ofArgs []string
	for _, v := range varNames {
		ofArgs = append(ofArgs, fmt.Sprintf("matchMap.get(\"%s\")", v))
	}
	stmts = append(stmts, &ast.ReturnStatement{
		Expr: ast.ExprF("of(%s)", strings.Join(ofArgs, ", ")),
	})
	return stmts
}

func composeToStringBody(varNames []string) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("Map<String, String> map = new HashMap<>();"),
	)
	for _, v := range varNames {
		paramName := lexicon.ToLowerCamel(v)
		stmts = append(stmts, ast.StatementFromF("map.put(\"%s\", %s);", v, paramName))
	}
	stmts = append(stmts, &ast.ReturnStatement{
		Expr: ast.Expr("PATH_TEMPLATE.instantiate(map)"),
	})
	return stmts
}

func composeEqualsCondition(varNames []string) string {
	var conds []string
	for _, v := range varNames {
		paramName := lexicon.ToLowerCamel(v)
		conds = append(conds, fmt.Sprintf("Objects.equals(this.%s, that.%s)", paramName, paramName))
	}
	return strings.Join(conds, " && ")
}
