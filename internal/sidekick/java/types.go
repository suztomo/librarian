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

	"github.com/googleapis/librarian/internal/sidekick/api"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/ast"
	"github.com/googleapis/librarian/internal/sidekick/java/engine/lexicon"
)

// Standard GAX packages.
const (
	PkgGaxRpc         = "com.google.api.gax.rpc"
	PkgGaxCore        = "com.google.api.gax.core"
	PkgGaxGrpc        = "com.google.api.gax.grpc"
	PkgGaxHttpJson    = "com.google.api.gax.httpjson"
	PkgGaxLongrunning = "com.google.api.gax.longrunning"
	PkgResourceNames  = "com.google.api.resourcenames"
	PkgProto          = "com.google.protobuf"
	PkgLongrunning    = "com.google.longrunning"
	PkgGrpc           = "io.grpc"
)

// Common GAX Types.
var (
	TypeUnaryCallable                        = ast.ObjectType("UnaryCallable", PkgGaxRpc)
	TypePagedCallable                        = ast.ObjectType("UnaryCallable", PkgGaxRpc)
	TypeServerStreamingCallable              = ast.ObjectType("ServerStreamingCallable", PkgGaxRpc)
	TypeBidiStreamingCallable                = ast.ObjectType("BidiStreamingCallable", PkgGaxRpc)
	TypeClientStreamingCallable              = ast.ObjectType("ClientStreamingCallable", PkgGaxRpc)
	TypeOperationCallable                    = ast.ObjectType("OperationCallable", PkgGaxRpc)
	TypeOperationFuture                      = ast.ObjectType("OperationFuture", PkgGaxLongrunning)
	TypeOperationsClient                     = ast.ObjectType("OperationsClient", PkgLongrunning)
	TypeOperationsStub                       = ast.ObjectType("OperationsStub", PkgLongrunning+".stub")
	TypeGrpcOperationsStub                   = ast.ObjectType("GrpcOperationsStub", PkgLongrunning+".stub")
	TypeHttpJsonOperationsStub               = ast.ObjectType("HttpJsonOperationsStub", PkgLongrunning+".stub")
	TypeUnaryCallSettings                    = ast.ObjectType("UnaryCallSettings", PkgGaxRpc)
	TypePagedCallSettings                    = ast.ObjectType("PagedCallSettings", PkgGaxRpc)
	TypeServerStreamingCallSettings          = ast.ObjectType("ServerStreamingCallSettings", PkgGaxRpc)
	TypeStreamingCallSettings                = ast.ObjectType("StreamingCallSettings", PkgGaxRpc)
	TypeOperationCallSettings                = ast.ObjectType("OperationCallSettings", PkgGaxRpc)
	TypeClientSettings                       = ast.ObjectType("ClientSettings", PkgGaxRpc)
	TypeStubSettings                         = ast.ObjectType("StubSettings", PkgGaxRpc)
	TypeClientContext                        = ast.ObjectType("ClientContext", PkgGaxRpc)
	TypeTransportChannelProvider             = ast.ObjectType("TransportChannelProvider", PkgGaxRpc)
	TypeCredentialsProvider                  = ast.ObjectType("CredentialsProvider", PkgGaxCore)
	TypeBackgroundResource                   = ast.ObjectType("BackgroundResource", PkgGaxRpc)
	TypeBackgroundResourceAggregation        = ast.ObjectType("BackgroundResourceAggregation", PkgGaxRpc)
	TypeApiStreamObserver                    = ast.ObjectType("ApiStreamObserver", PkgGaxRpc)
	TypeResponseObserver                     = ast.ObjectType("ResponseObserver", PkgGaxRpc)
	TypeClientStream                         = ast.ObjectType("ClientStream", PkgGaxRpc)
	TypeApiFunction                          = ast.ObjectType("ApiFunction", PkgGaxCore)
	TypeApiFuture                            = ast.ObjectType("ApiFuture", "com.google.api.core")
	TypeBetaApi                              = ast.ObjectType("BetaApi", "com.google.api.core")
	TypeInternalApi                          = ast.ObjectType("InternalApi", "com.google.api.core")
	TypeResourceName                         = ast.ObjectType("ResourceName", PkgResourceNames)
	TypePathTemplate                         = ast.ObjectType("PathTemplate", PkgGaxHttpJson)
	TypeMethodDescriptor                     = ast.ObjectType("MethodDescriptor", PkgGrpc)
	TypeGrpcCallSettings                     = ast.ObjectType("GrpcCallSettings", PkgGaxGrpc)
	TypeGrpcStubCallableFactory              = ast.ObjectType("GrpcStubCallableFactory", PkgGaxGrpc)
	TypeGrpcTransportChannel                 = ast.ObjectType("GrpcTransportChannel", PkgGaxGrpc)
	TypeInstantiatingGrpcChannelProvider     = ast.ObjectType("InstantiatingGrpcChannelProvider", PkgGaxGrpc)
	TypeHttpJsonCallSettings                 = ast.ObjectType("HttpJsonCallSettings", PkgGaxHttpJson)
	TypeHttpJsonStubCallableFactory          = ast.ObjectType("HttpJsonStubCallableFactory", PkgGaxHttpJson)
	TypeHttpJsonTransportChannel             = ast.ObjectType("HttpJsonTransportChannel", PkgGaxHttpJson)
	TypeInstantiatingHttpJsonChannelProvider = ast.ObjectType("InstantiatingHttpJsonChannelProvider", PkgGaxHttpJson)
	TypeFields                               = ast.ObjectType("Fields", PkgGaxHttpJson)
	TypeApiMethodDescriptor                  = ast.ObjectType("ApiMethodDescriptor", PkgGaxHttpJson)
	TypeProtoMessageRequestFormatter         = ast.ObjectType("ProtoMessageRequestFormatter", PkgGaxHttpJson)
	TypeProtoMessageResponseParser           = ast.ObjectType("ProtoMessageResponseParser", PkgGaxHttpJson)
	TypeProtoRestSerializer                  = ast.ObjectType("ProtoRestSerializer", PkgGaxHttpJson)
	TypeEmpty                                = ast.ObjectType("Empty", PkgProto)
	TypeOperation                            = ast.ObjectType("Operation", PkgLongrunning)
)

// FieldTypeToJavaType converts an api.Field to an ast.TypeNode.
func FieldTypeToJavaType(f *api.Field) *TypeNodeWrapper {
	if f == nil {
		return &TypeNodeWrapper{Type: ast.TypeObject}
	}

	baseType := PrimitiveOrMessageType(f.Typez, f.MessageType, f.EnumType)

	if f.Map {
		// Maps in protobuf are repeated Message where message is key/value pair
		keyType := ast.TypeString
		valType := ast.TypeObject
		if f.MessageType != nil && len(f.MessageType.Fields) >= 2 {
			keyType = PrimitiveOrMessageType(f.MessageType.Fields[0].Typez, f.MessageType.Fields[0].MessageType, f.MessageType.Fields[0].EnumType)
			valType = PrimitiveOrMessageType(f.MessageType.Fields[1].Typez, f.MessageType.Fields[1].MessageType, f.MessageType.Fields[1].EnumType)
		}
		return &TypeNodeWrapper{
			Type:  ast.GenericType(ast.TypeMap, toBoxedType(keyType), toBoxedType(valType)),
			IsMap: true,
		}
	}

	if f.Repeated {
		return &TypeNodeWrapper{
			Type:        ast.GenericType(ast.TypeList, toBoxedType(baseType)),
			IsRepeated:  true,
			ElementType: baseType,
		}
	}

	return &TypeNodeWrapper{
		Type: baseType,
	}
}

// TypeNodeWrapper bundles an AST TypeNode with metadata.
type TypeNodeWrapper struct {
	Type        *ast.TypeNode
	IsRepeated  bool
	IsMap       bool
	ElementType *ast.TypeNode
}

// PrimitiveOrMessageType converts api.Typez / Message / Enum to ast.TypeNode.
func PrimitiveOrMessageType(typez api.Typez, msg *api.Message, enum *api.Enum) *ast.TypeNode {
	switch typez {
	case api.TypezString:
		return ast.TypeString
	case api.TypezBool:
		return ast.TypeBoolean
	case api.TypezBytes:
		return ast.TypeByteString
	case api.TypezInt32, api.TypezSint32, api.TypezSfixed32:
		return ast.TypeInt
	case api.TypezUint32, api.TypezFixed32:
		return ast.TypeInt
	case api.TypezInt64, api.TypezSint64, api.TypezSfixed64:
		return ast.TypeLong
	case api.TypezUint64, api.TypezFixed64:
		return ast.TypeLong
	case api.TypezFloat:
		return ast.TypeFloat
	case api.TypezDouble:
		return ast.TypeDouble
	case api.TypezMessage:
		if msg != nil {
			return MessageToJavaType(msg)
		}
		return ast.TypeObject
	case api.TypezEnum:
		if enum != nil {
			return EnumToJavaType(enum)
		}
		return ast.TypeObject
	default:
		return ast.TypeObject
	}
}

// MessageToJavaType converts an api.Message to its Java TypeNode.
func MessageToJavaType(m *api.Message) *ast.TypeNode {
	if m == nil {
		return ast.TypeObject
	}
	// Check well known types
	switch m.ID {
	case ".google.protobuf.Empty":
		return TypeEmpty
	case ".google.protobuf.FieldMask":
		return ast.ObjectType("FieldMask", PkgProto)
	case ".google.protobuf.Timestamp":
		return ast.ObjectType("Timestamp", PkgProto)
	case ".google.protobuf.Duration":
		return ast.ObjectType("Duration", PkgProto)
	case ".google.protobuf.Any":
		return ast.ObjectType("Any", PkgProto)
	case ".google.protobuf.Struct":
		return ast.ObjectType("Struct", PkgProto)
	case ".google.protobuf.Value":
		return ast.ObjectType("Value", PkgProto)
	case ".google.longrunning.Operation":
		return TypeOperation
	}

	pkg := lexicon.JavaPackageFromProto(m.Package)
	name := m.Name
	if m.Parent != nil {
		name = m.Parent.Name + "." + m.Name
	}
	return ast.ObjectType(name, pkg)
}

// EnumToJavaType converts an api.Enum to its Java TypeNode.
func EnumToJavaType(e *api.Enum) *ast.TypeNode {
	if e == nil {
		return ast.TypeObject
	}
	pkg := lexicon.JavaPackageFromProto(e.Package)
	name := e.Name
	if e.Parent != nil {
		name = e.Parent.Name + "." + e.Name
	}
	return ast.ObjectType(name, pkg)
}

func toBoxedType(t *ast.TypeNode) *ast.TypeNode {
	if t == nil {
		return ast.TypeObject
	}
	if t.Kind == ast.KindPrimitive {
		switch t.Name {
		case "int":
			return ast.TypeBoxedInteger
		case "long":
			return ast.TypeBoxedLong
		case "boolean":
			return ast.TypeBoxedBoolean
		case "double":
			return ast.TypeBoxedDouble
		case "float":
			return ast.TypeBoxedFloat
		}
	}
	return t
}

// FormatJavaDocComment converts markdown/doc comments into Javadoc formatted text.
func FormatJavaDocComment(doc string) string {
	doc = strings.TrimSpace(doc)
	if doc == "" {
		return ""
	}
	doc = lexicon.SanitizeComment(doc)
	return doc
}
