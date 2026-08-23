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

// Package ast defines the Java abstract syntax tree for code generation.
package ast

import (
	"fmt"
	"strings"
)

// Scope represents Java member visibility.
type Scope string

const (
	// Public indicates public visibility.
	Public Scope = "public"
	// Protected indicates protected visibility.
	Protected Scope = "protected"
	// Private indicates private visibility.
	Private Scope = "private"
	// PackagePrivate indicates package-private (default) visibility.
	PackagePrivate Scope = ""
)

// Modifier represents Java modifiers.
type Modifier string

const (
	// Static indicates the static modifier.
	Static Modifier = "static"
	// Final indicates the final modifier.
	Final Modifier = "final"
	// Abstract indicates the abstract modifier.
	Abstract Modifier = "abstract"
	// Synchronized indicates the synchronized modifier.
	Synchronized Modifier = "synchronized"
	// Transient indicates the transient modifier.
	Transient Modifier = "transient"
	// Volatile indicates the volatile modifier.
	Volatile Modifier = "volatile"
	// Native indicates the native modifier.
	Native Modifier = "native"
	// Strictfp indicates the strictfp modifier.
	Strictfp Modifier = "strictfp"
	// Default indicates default interface method modifier.
	Default Modifier = "default"
)

// TypeKind defines the category of a TypeNode.
type TypeKind int

const (
	// KindPrimitive indicates a primitive type.
	KindPrimitive TypeKind = iota
	// KindObject indicates a reference or object type.
	KindObject
	// KindVoid indicates the void type.
	KindVoid
)

// TypeNode represents a Java type (primitive, class, interface, generic, array).
type TypeNode struct {
	Kind           TypeKind
	Name           string
	Package        string
	Generics       []*TypeNode
	IsArray        bool
	Wildcard       bool
	ExtendsBound   *TypeNode
	SuperBound     *TypeNode
	IsStaticImport bool
}

// Common primitive & object types.
var (
	TypeVoid    = &TypeNode{Kind: KindVoid, Name: "void"}
	TypeBoolean = &TypeNode{Kind: KindPrimitive, Name: "boolean"}
	TypeByte    = &TypeNode{Kind: KindPrimitive, Name: "byte"}
	TypeShort   = &TypeNode{Kind: KindPrimitive, Name: "short"}
	TypeInt     = &TypeNode{Kind: KindPrimitive, Name: "int"}
	TypeLong    = &TypeNode{Kind: KindPrimitive, Name: "long"}
	TypeFloat   = &TypeNode{Kind: KindPrimitive, Name: "float"}
	TypeDouble  = &TypeNode{Kind: KindPrimitive, Name: "double"}
	TypeChar    = &TypeNode{Kind: KindPrimitive, Name: "char"}

	TypeBoxedBoolean  = ObjectType("Boolean", "java.lang")
	TypeBoxedInteger  = ObjectType("Integer", "java.lang")
	TypeBoxedLong     = ObjectType("Long", "java.lang")
	TypeBoxedDouble   = ObjectType("Double", "java.lang")
	TypeBoxedFloat    = ObjectType("Float", "java.lang")
	TypeString        = ObjectType("String", "java.lang")
	TypeObject        = ObjectType("Object", "java.lang")
	TypeException     = ObjectType("Exception", "java.lang")
	TypeIOException   = ObjectType("IOException", "java.io")
	TypeCloseable     = ObjectType("Closeable", "java.io")
	TypeAutoCloseable = ObjectType("AutoCloseable", "java.lang")
	TypeList          = ObjectType("List", "java.util")
	TypeArrayList     = ObjectType("ArrayList", "java.util")
	TypeMap           = ObjectType("Map", "java.util")
	TypeHashMap       = ObjectType("HashMap", "java.util")
	TypeSet           = ObjectType("Set", "java.util")
	TypeHashSet       = ObjectType("HashSet", "java.util")
	TypeCollections   = ObjectType("Collections", "java.util")
	TypeObjects       = ObjectType("Objects", "java.util")
	TypeArrays        = ObjectType("Arrays", "java.util")
	TypeTimeUnit      = ObjectType("TimeUnit", "java.util.concurrent")
	TypeDuration      = ObjectType("Duration", "org.threeten.bp")
	TypeByteString    = ObjectType("ByteString", "com.google.protobuf")
	TypeGenerated     = ObjectType("Generated", "javax.annotation")
	TypeBetaApi       = ObjectType("BetaApi", "com.google.api.core")
	TypeInternalApi   = ObjectType("InternalApi", "com.google.api.core")
)

// PrimitiveType creates a primitive type.
func PrimitiveType(name string) *TypeNode {
	return &TypeNode{Kind: KindPrimitive, Name: name}
}

// ObjectType creates an object/class type with a package.
func ObjectType(name, pkg string) *TypeNode {
	return &TypeNode{
		Kind:    KindObject,
		Name:    name,
		Package: pkg,
	}
}

// GenericType creates a parameterized generic type (e.g. List<String>).
func GenericType(base *TypeNode, typeArgs ...*TypeNode) *TypeNode {
	return &TypeNode{
		Kind:     base.Kind,
		Name:     base.Name,
		Package:  base.Package,
		Generics: typeArgs,
	}
}

// ArrayType creates an array type.
func ArrayType(element *TypeNode) *TypeNode {
	return &TypeNode{
		Kind:     element.Kind,
		Name:     element.Name,
		Package:  element.Package,
		Generics: element.Generics,
		IsArray:  true,
	}
}

// WildcardType creates a wildcard type (? extends Bound).
func WildcardType(extendsBound *TypeNode) *TypeNode {
	return &TypeNode{
		Kind:         KindObject,
		Wildcard:     true,
		ExtendsBound: extendsBound,
	}
}

// FullName returns the fully-qualified class name.
func (t *TypeNode) FullName() string {
	if t.Package == "" || t.Kind == KindPrimitive || t.Kind == KindVoid {
		return t.Name
	}
	return t.Package + "." + t.Name
}

// TypeString returns the Java source code representation of the type.
func (t *TypeNode) TypeString() string {
	if t.Wildcard {
		if t.ExtendsBound != nil {
			return "? extends " + t.ExtendsBound.TypeString()
		}
		if t.SuperBound != nil {
			return "? super " + t.SuperBound.TypeString()
		}
		return "?"
	}
	s := t.Name
	if len(t.Generics) > 0 {
		var genericStrs []string
		for _, g := range t.Generics {
			genericStrs = append(genericStrs, g.TypeString())
		}
		s += "<" + strings.Join(genericStrs, ", ") + ">"
	}
	if t.IsArray {
		s += "[]"
	}
	return s
}

// AnnotationNode represents a Java annotation.
type AnnotationNode struct {
	Type        *TypeNode
	SingleValue Expression
	Properties  map[string]Expression
}

// NewAnnotation creates an annotation node.
func NewAnnotation(annotationType *TypeNode) *AnnotationNode {
	return &AnnotationNode{
		Type:       annotationType,
		Properties: make(map[string]Expression),
	}
}

// WithValue sets a single value on the annotation: @Annotation(value).
func (a *AnnotationNode) WithValue(val Expression) *AnnotationNode {
	a.SingleValue = val
	return a
}

// WithProperty sets a key-value property on the annotation: @Annotation(key = value).
func (a *AnnotationNode) WithProperty(key string, val Expression) *AnnotationNode {
	if a.Properties == nil {
		a.Properties = make(map[string]Expression)
	}
	a.Properties[key] = val
	return a
}

// ClassKind defines class vs interface vs enum.
type ClassKind string

const (
	// ClassKindClass represents a standard Java class.
	ClassKindClass ClassKind = "class"
	// ClassKindInterface represents a Java interface.
	ClassKindInterface ClassKind = "interface"
	// ClassKindEnum represents a Java enum.
	ClassKindEnum ClassKind = "enum"
)

// ClassDefinition represents a top-level or inner Java class.
type ClassDefinition struct {
	Package       string
	Name          string
	Scope         Scope
	Modifiers     []Modifier
	Kind          ClassKind
	Extends       *TypeNode
	Implements    []*TypeNode
	Annotations   []*AnnotationNode
	Fields        []*FieldDefinition
	Constructors  []*ConstructorDefinition
	Methods       []*MethodDefinition
	InnerClasses  []*ClassDefinition
	JavaDoc       string
	ExtraImports  []string
	StaticImports []string
}

// FieldDefinition represents a class member field.
type FieldDefinition struct {
	Name        string
	Type        *TypeNode
	Scope       Scope
	Modifiers   []Modifier
	Initializer Expression
	Annotations []*AnnotationNode
	JavaDoc     string
}

// ParameterDefinition represents a method/constructor parameter.
type ParameterDefinition struct {
	Name        string
	Type        *TypeNode
	Final       bool
	Annotations []*AnnotationNode
}

// ConstructorDefinition represents a class constructor.
type ConstructorDefinition struct {
	Scope       Scope
	Parameters  []*ParameterDefinition
	Throws      []*TypeNode
	Annotations []*AnnotationNode
	Body        []Statement
	JavaDoc     string
}

// MethodDefinition represents a class/interface member method.
type MethodDefinition struct {
	Name        string
	ReturnType  *TypeNode
	Scope       Scope
	Modifiers   []Modifier
	Parameters  []*ParameterDefinition
	Throws      []*TypeNode
	Annotations []*AnnotationNode
	Body        []Statement
	JavaDoc     string
}

// Statement represents a Java executable statement.
type Statement interface {
	statementNode()
	FormatStatement(indent string) string
}

// Expression represents a Java expression.
type Expression interface {
	expressionNode()
	FormatExpression() string
}

// RawStatement represents an arbitrary statement string.
type RawStatement struct {
	Code string
}

func (s *RawStatement) statementNode() {}

// FormatStatement formats the raw statement with indentation.
func (s *RawStatement) FormatStatement(indent string) string {
	if s.Code == "" {
		return ""
	}
	lines := strings.Split(s.Code, "\n")
	var formatted []string
	for _, l := range lines {
		if strings.TrimSpace(l) == "" {
			formatted = append(formatted, "")
		} else {
			formatted = append(formatted, indent+l)
		}
	}
	return strings.Join(formatted, "\n")
}

// StatementFrom creates a RawStatement from string.
func StatementFrom(code string) Statement {
	return &RawStatement{Code: code}
}

// StatementFromF creates a formatted RawStatement.
func StatementFromF(format string, args ...any) Statement {
	return &RawStatement{Code: fmt.Sprintf(format, args...)}
}

// ReturnStatement represents `return expr;`.
type ReturnStatement struct {
	Expr Expression
}

func (s *ReturnStatement) statementNode() {}

// FormatStatement formats the return statement.
func (s *ReturnStatement) FormatStatement(indent string) string {
	if s.Expr == nil {
		return indent + "return;"
	}
	return indent + "return " + s.Expr.FormatExpression() + ";"
}

// VariableDeclarationStatement represents `[final] Type varName = init;`.
type VariableDeclarationStatement struct {
	Type        *TypeNode
	Name        string
	Initializer Expression
	Final       bool
}

func (s *VariableDeclarationStatement) statementNode() {}

// FormatStatement formats the variable declaration statement.
func (s *VariableDeclarationStatement) FormatStatement(indent string) string {
	res := indent
	if s.Final {
		res += "final "
	}
	res += s.Type.TypeString() + " " + s.Name
	if s.Initializer != nil {
		res += " = " + s.Initializer.FormatExpression()
	}
	return res + ";"
}

// RawExpression represents an arbitrary expression string.
type RawExpression struct {
	Code string
}

func (e *RawExpression) expressionNode() {}

// FormatExpression returns the raw code expression.
func (e *RawExpression) FormatExpression() string { return e.Code }

// Expr creates a RawExpression from string.
func Expr(code string) Expression {
	return &RawExpression{Code: code}
}

// ExprF creates a formatted RawExpression.
func ExprF(format string, args ...any) Expression {
	return &RawExpression{Code: fmt.Sprintf(format, args...)}
}

// StringLiteralExpr creates a quoted Java string literal.
func StringLiteralExpr(val string) Expression {
	escaped := strings.ReplaceAll(val, `\`, `\\`)
	escaped = strings.ReplaceAll(escaped, `"`, `\"`)
	escaped = strings.ReplaceAll(escaped, "\n", `\n`)
	return &RawExpression{Code: `"` + escaped + `"`}
}

// MethodInvocationExpr represents `target.method(args...)`.
type MethodInvocationExpr struct {
	Target Expression
	Method string
	Args   []Expression
}

func (e *MethodInvocationExpr) expressionNode() {}

// FormatExpression formats the method invocation expression.
func (e *MethodInvocationExpr) FormatExpression() string {
	var argStrs []string
	for _, a := range e.Args {
		argStrs = append(argStrs, a.FormatExpression())
	}
	argsJoin := strings.Join(argStrs, ", ")
	if e.Target == nil {
		return fmt.Sprintf("%s(%s)", e.Method, argsJoin)
	}
	return fmt.Sprintf("%s.%s(%s)", e.Target.FormatExpression(), e.Method, argsJoin)
}
