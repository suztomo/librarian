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
	"slices"
	"strings"

	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

var variableRegex = regexp.MustCompile(`\{([a-zA-Z0-9_]+)(?:=[^}]+)?\}`)

type patternInfo struct {
	rawPattern   string
	constantName string
	methodSuffix string
	varNames     []string
}

// ComposeResourceNameClass generates <Resource>Name.java AST.
func ComposeResourceNameClass(res *ResourceAnnotation) *ast.ClassDefinition {
	if res == nil || res.ClassName == "" {
		return nil
	}

	patterns := res.Patterns
	if len(patterns) == 0 {
		defaultPattern := "projects/{project}/" + strings.ToLower(strings.TrimSuffix(res.ClassName, "Name")) + "s/{" + strings.ToLower(strings.TrimSuffix(res.ClassName, "Name")) + "}"
		patterns = []string{defaultPattern}
	}

	var parsedPatterns []patternInfo
	var allVarNames []string
	seenVars := make(map[string]bool)

	for i, pat := range patterns {
		matches := variableRegex.FindAllStringSubmatch(pat, -1)
		var varNames []string
		for _, m := range matches {
			if len(m) > 1 {
				varNames = append(varNames, m[1])
				if !seenVars[m[1]] {
					seenVars[m[1]] = true
					allVarNames = append(allVarNames, m[1])
				}
			}
		}
		if len(varNames) == 0 {
			varNames = []string{"project", strings.ToLower(strings.TrimSuffix(res.ClassName, "Name"))}
			for _, v := range varNames {
				if !seenVars[v] {
					seenVars[v] = true
					allVarNames = append(allVarNames, v)
				}
			}
		}

		constName := "PATH_TEMPLATE"
		suffix := res.ClassName
		if len(patterns) > 1 {
			constName = patternConstantName(pat, i)
			suffix = patternMethodSuffix(pat, i)
		}

		parsedPatterns = append(parsedPatterns, patternInfo{
			rawPattern:   pat,
			constantName: constName,
			methodSuffix: suffix,
			varNames:     varNames,
		})
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

	// Static PATH_TEMPLATE fields for each pattern
	for _, p := range parsedPatterns {
		classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
			Name:        p.constantName,
			Type:        ast.ObjectType("PathTemplate", "com.google.api.pathtemplate"),
			Scope:       ast.Private,
			Modifiers:   []ast.Modifier{ast.Static, ast.Final},
			Initializer: ast.ExprF("PathTemplate.createWithoutUrlEncoding(\"%s\")", p.rawPattern),
		})
	}

	// Member fields for all unique variables
	for _, v := range allVarNames {
		classDef.Fields = append(classDef.Fields, &ast.FieldDefinition{
			Name:      lexicon.ToLowerCamel(v),
			Type:      ast.TypeString,
			Scope:     ast.Private,
			Modifiers: []ast.Modifier{ast.Final},
		})
	}

	// Private Constructor taking all variables
	var ctorParams []*ast.ParameterDefinition
	var ctorStmts []ast.Statement
	for _, v := range allVarNames {
		paramName := lexicon.ToLowerCamel(v)
		ctorParams = append(ctorParams, &ast.ParameterDefinition{
			Name: paramName,
			Type: ast.TypeString,
		})
		ctorStmts = append(ctorStmts, ast.StatementFromF("this.%s = %s;", paramName, paramName))
	}
	classDef.Constructors = append(classDef.Constructors, &ast.ConstructorDefinition{
		Scope:      ast.Private,
		Parameters: ctorParams,
		Body:       ctorStmts,
	})

	// of(...) static factory method for the primary pattern
	primaryPattern := parsedPatterns[0]
	var primaryOfParams []*ast.ParameterDefinition
	var primaryConstructorArgs []string
	for _, v := range allVarNames {
		paramName := lexicon.ToLowerCamel(v)
		if slices.Contains(primaryPattern.varNames, v) {
			primaryConstructorArgs = append(primaryConstructorArgs, fmt.Sprintf("Objects.requireNonNull(%s)", paramName))
		} else {
			primaryConstructorArgs = append(primaryConstructorArgs, "null")
		}
	}
	for _, v := range primaryPattern.varNames {
		paramName := lexicon.ToLowerCamel(v)
		primaryOfParams = append(primaryOfParams, &ast.ParameterDefinition{
			Name: paramName,
			Type: ast.TypeString,
		})
	}
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "of",
		ReturnType: classType,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: primaryOfParams,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Expr: ast.ExprF("new %s(%s)", res.ClassName, strings.Join(primaryConstructorArgs, ", ")),
			},
		},
	})

	var primaryFormatArgs []string
	for _, v := range primaryPattern.varNames {
		primaryFormatArgs = append(primaryFormatArgs, lexicon.ToLowerCamel(v))
	}
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "format",
		ReturnType: ast.TypeString,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: primaryOfParams,
		Body: []ast.Statement{
			&ast.ReturnStatement{
				Expr: ast.ExprF("of(%s).toString()", strings.Join(primaryFormatArgs, ", ")),
			},
		},
	})

	// If multiple patterns, create of<Suffix> and format<Suffix> for each pattern
	if len(parsedPatterns) > 1 {
		for _, p := range parsedPatterns {
			var ofParams []*ast.ParameterDefinition
			var ctorArgs []string
			for _, v := range allVarNames {
				paramName := lexicon.ToLowerCamel(v)
				if slices.Contains(p.varNames, v) {
					ctorArgs = append(ctorArgs, fmt.Sprintf("Objects.requireNonNull(%s)", paramName))
				} else {
					ctorArgs = append(ctorArgs, "null")
				}
			}
			for _, v := range p.varNames {
				paramName := lexicon.ToLowerCamel(v)
				ofParams = append(ofParams, &ast.ParameterDefinition{
					Name: paramName,
					Type: ast.TypeString,
				})
			}
			ofMethodName := "of" + p.methodSuffix
			formatMethodName := "format" + p.methodSuffix

			var formatArgs []string
			for _, v := range p.varNames {
				formatArgs = append(formatArgs, lexicon.ToLowerCamel(v))
			}

			classDef.Methods = append(classDef.Methods,
				&ast.MethodDefinition{
					Name:       ofMethodName,
					ReturnType: classType,
					Scope:      ast.Public,
					Modifiers:  []ast.Modifier{ast.Static},
					Parameters: ofParams,
					Body: []ast.Statement{
						&ast.ReturnStatement{
							Expr: ast.ExprF("new %s(%s)", res.ClassName, strings.Join(ctorArgs, ", ")),
						},
					},
				},
				&ast.MethodDefinition{
					Name:       formatMethodName,
					ReturnType: ast.TypeString,
					Scope:      ast.Public,
					Modifiers:  []ast.Modifier{ast.Static},
					Parameters: ofParams,
					Body: []ast.Statement{
						&ast.ReturnStatement{
							Expr: ast.ExprF("%s(%s).toString()", ofMethodName, strings.Join(formatArgs, ", ")),
						},
					},
				},
			)
		}
	}

	// parse(...) static factory method
	classDef.Methods = append(classDef.Methods, &ast.MethodDefinition{
		Name:       "parse",
		ReturnType: classType,
		Scope:      ast.Public,
		Modifiers:  []ast.Modifier{ast.Static},
		Parameters: []*ast.ParameterDefinition{
			{Name: "formattedString", Type: ast.TypeString},
		},
		Body: composeMultiPatternParseBody(res.ClassName, parsedPatterns),
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
		Body: composeIsParsableFromBody(parsedPatterns),
	})

	// Getters for all variable fields
	for _, v := range allVarNames {
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
			Body: composeMultiPatternToStringBody(parsedPatterns, allVarNames),
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
					composeEqualsCondition(allVarNames),
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
					Expr: ast.ExprF("Objects.hash(%s)", strings.Join(formatVarNames(allVarNames), ", ")),
				},
			},
		},
	)

	return classDef
}

func patternConstantName(pattern string, index int) string {
	matches := variableRegex.FindAllStringSubmatch(pattern, -1)
	var parts []string
	for _, m := range matches {
		if len(m) > 1 {
			parts = append(parts, lexicon.ToUpperSnake(m[1]))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("PATH_TEMPLATE_%d", index)
	}
	return strings.Join(parts, "_") + "_PATH_TEMPLATE"
}

func patternMethodSuffix(pattern string, index int) string {
	matches := variableRegex.FindAllStringSubmatch(pattern, -1)
	var parts []string
	for _, m := range matches {
		if len(m) > 1 {
			parts = append(parts, lexicon.ToUpperCamel(m[1]))
		}
	}
	if len(parts) == 0 {
		return fmt.Sprintf("Pattern%dName", index)
	}
	return strings.Join(parts, "") + "Name"
}

func formatVarNames(varNames []string) []string {
	var res []string
	for _, v := range varNames {
		res = append(res, lexicon.ToLowerCamel(v))
	}
	return res
}

func composeMultiPatternParseBody(className string, patterns []patternInfo) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts,
		ast.StatementFrom("if (formattedString.isEmpty()) { return null; }"),
	)

	for _, p := range patterns {
		var ofArgs []string
		for _, v := range p.varNames {
			ofArgs = append(ofArgs, fmt.Sprintf("matchMap.get(\"%s\")", v))
		}
		targetOf := "of"
		if len(patterns) > 1 {
			targetOf = "of" + p.methodSuffix
		}
		stmts = append(stmts,
			ast.StatementFromF("if (%s.matches(formattedString)) {\n"+
				"  Map<String, String> matchMap = %s.validatedMatch(formattedString, \"%s.parse: \" + formattedString);\n"+
				"  return %s(%s);\n"+
				"}",
				p.constantName, p.constantName, className, targetOf, strings.Join(ofArgs, ", "),
			),
		)
	}

	stmts = append(stmts,
		ast.StatementFromF("throw new IllegalArgumentException(\"%s.parse: formattedString not in valid format: \" + formattedString);", className),
	)
	return stmts
}

func composeIsParsableFromBody(patterns []patternInfo) []ast.Statement {
	var conds []string
	for _, p := range patterns {
		conds = append(conds, fmt.Sprintf("%s.matches(formattedString)", p.constantName))
	}
	return []ast.Statement{
		&ast.ReturnStatement{
			Expr: ast.Expr(strings.Join(conds, " || ")),
		},
	}
}

func composeMultiPatternToStringBody(patterns []patternInfo, allVarNames []string) []ast.Statement {
	var stmts []ast.Statement
	stmts = append(stmts, ast.StatementFrom("Map<String, String> map = new HashMap<>();"))
	for _, v := range allVarNames {
		paramName := lexicon.ToLowerCamel(v)
		stmts = append(stmts, ast.StatementFromF("if (%s != null) { map.put(\"%s\", %s); }", paramName, v, paramName))
	}

	if len(patterns) == 1 {
		stmts = append(stmts, &ast.ReturnStatement{
			Expr: ast.ExprF("%s.instantiate(map)", patterns[0].constantName),
		})
		return stmts
	}

	for _, p := range patterns {
		var varChecks []string
		for _, v := range p.varNames {
			varChecks = append(varChecks, fmt.Sprintf("%s != null", lexicon.ToLowerCamel(v)))
		}
		stmts = append(stmts, ast.StatementFromF(
			"if (%s) {\n"+
				"  return %s.instantiate(map);\n"+
				"}",
			strings.Join(varChecks, " && "), p.constantName,
		))
	}

	stmts = append(stmts, &ast.ReturnStatement{
		Expr: ast.ExprF("%s.instantiate(map)", patterns[0].constantName),
	})
	return stmts
}

func composeEqualsCondition(varNames []string) string {
	var conds []string
	for _, v := range varNames {
		paramName := lexicon.ToLowerCamel(v)
		conds = append(conds, fmt.Sprintf("Objects.equals(this.%s, that.%s)", paramName, paramName))
	}
	if len(conds) == 0 {
		return "true"
	}
	return strings.Join(conds, " && ")
}
