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
)

// ModelAnnotations holds Java annotations for an entire api.API model.
type ModelAnnotations struct {
	Model           *api.API
	PackageName     string
	StubPackageName string
	Title           string
	Description     string
	Transport       string
	HasMetadata     bool
	HasVersionJava  bool
	Services        []*ServiceAnnotation
	Resources       []*ResourceAnnotation
	Codec           *Codec
}

// ServiceAnnotation holds Java annotations for an api.Service.
type ServiceAnnotation struct {
	Service             *api.Service
	Name                string
	ClientName          string
	SettingsName        string
	StubName            string
	StubSettingsName    string
	GrpcStubName        string
	GrpcFactoryName     string
	HttpJsonStubName    string
	HttpJsonFactoryName string
	DefaultHost         string
	DefaultPort         string
	DefaultEndpoint     string
	Scopes              []string
	Methods             []*MethodAnnotation
	HasLRO              bool
	HasPaged            bool
	HasStreaming        bool
	HasGrpc             bool
	HasHttpJson         bool
	OriginalJavaPackage string
}

// MethodAnnotation holds Java annotations for an api.Method.
type MethodAnnotation struct {
	Method                *api.Method
	Name                  string
	MethodName            string
	CallableName          string
	PagedCallableName     string
	OperationCallableName string
	RequestType           *ast.TypeNode
	ResponseType          *ast.TypeNode
	IsUnary               bool
	IsPaged               bool
	IsLRO                 bool
	IsServerStreaming     bool
	IsBidiStreaming       bool
	IsClientStreaming     bool
	Signatures            [][]*api.Field
	LroResponseType       *ast.TypeNode
	LroMetadataType       *ast.TypeNode
	PageItemType          *ast.TypeNode
	PageItemField         string
	PageTokenField        string
	NextPageTokenField    string
	Description           string
}

// ResourceAnnotation holds Java annotations for a Resource definition.
type ResourceAnnotation struct {
	Resource    *api.Resource
	Type        string
	ClassName   string
	Patterns    []string
	PackageName string
}

// AnnotateModel decorates an api.API model with Java AST and metadata annotations.
func AnnotateModel(model *api.API, codec *Codec) (*ModelAnnotations, error) {
	pkgName := JavaPackage(model, codec.PackageOverride)
	stubPkg := StubPackage(pkgName)

	transport := codec.Transport
	if transport == "" {
		transport = "grpc"
	}
	hasGrpc := transport == "grpc" || transport == "grpc+rest" || transport == "rest+grpc"
	hasHttpJson := transport == "rest" || transport == "grpc+rest" || transport == "rest+grpc"

	ann := &ModelAnnotations{
		Model:           model,
		PackageName:     pkgName,
		StubPackageName: stubPkg,
		Title:           model.Title,
		Description:     model.Description,
		Transport:       transport,
		HasMetadata:     codec.HasMetadata,
		HasVersionJava:  codec.GenerateVersionJava,
		Codec:           codec,
	}

	// Annotate Services
	for _, svc := range model.Services {
		svcAnn := annotateService(svc, model, ann, hasGrpc, hasHttpJson)
		ann.Services = append(ann.Services, svcAnn)
		svc.Codec = svcAnn
	}

	// Annotate Resources
	seenResources := make(map[string]bool)
	for _, res := range model.ResourceDefinitions {
		if res != nil && res.Type != "" && !seenResources[res.Type] {
			seenResources[res.Type] = true
			ann.Resources = append(ann.Resources, annotateResource(res, pkgName))
		}
	}
	for _, msg := range model.Messages {
		if msg.Resource != nil && msg.Resource.Type != "" && !seenResources[msg.Resource.Type] {
			seenResources[msg.Resource.Type] = true
			ann.Resources = append(ann.Resources, annotateResource(msg.Resource, pkgName))
		}
	}

	model.Codec = ann
	return ann, nil
}

func annotateService(svc *api.Service, model *api.API, modelAnn *ModelAnnotations, hasGrpc, hasHttpJson bool) *ServiceAnnotation {
	defaultHost := svc.DefaultHost
	defaultPort := "443"
	if defaultHost == "" && len(model.Services) > 0 {
		defaultHost = model.Services[0].DefaultHost
	}
	defaultEndpoint := defaultHost
	if defaultHost != "" && !strings.Contains(defaultHost, ":") {
		defaultEndpoint = defaultHost + ":" + defaultPort
	}

	svcAnn := &ServiceAnnotation{
		Service:             svc,
		Name:                svc.Name,
		ClientName:          ClientClassName(svc.Name),
		SettingsName:        SettingsClassName(svc.Name),
		StubName:            StubClassName(svc.Name),
		StubSettingsName:    StubSettingsClassName(svc.Name),
		GrpcStubName:        GrpcStubClassName(svc.Name),
		GrpcFactoryName:     GrpcCallableFactoryClassName(svc.Name),
		HttpJsonStubName:    HttpJsonStubClassName(svc.Name),
		HttpJsonFactoryName: HttpJsonCallableFactoryClassName(svc.Name),
		DefaultHost:         defaultHost,
		DefaultPort:         defaultPort,
		DefaultEndpoint:     defaultEndpoint,
		HasGrpc:             hasGrpc,
		HasHttpJson:         hasHttpJson,
		OriginalJavaPackage: modelAnn.PackageName,
	}

	for _, m := range svc.Methods {
		methodAnn := annotateMethod(m, modelAnn)
		svcAnn.Methods = append(svcAnn.Methods, methodAnn)
		m.Codec = methodAnn

		if methodAnn.IsLRO {
			svcAnn.HasLRO = true
		}
		if methodAnn.IsPaged {
			svcAnn.HasPaged = true
		}
		if methodAnn.IsServerStreaming || methodAnn.IsBidiStreaming || methodAnn.IsClientStreaming {
			svcAnn.HasStreaming = true
		}
	}

	return svcAnn
}

func annotateMethod(m *api.Method, modelAnn *ModelAnnotations) *MethodAnnotation {
	reqType := ast.TypeObject
	if m.InputType != nil {
		reqType = MessageToJavaType(m.InputType)
	}

	respType := ast.TypeObject
	if m.OutputType != nil {
		respType = MessageToJavaType(m.OutputType)
	} else if m.ReturnsEmpty {
		respType = TypeEmpty
	}

	isServerStreaming := m.ServerSideStreaming && !m.ClientSideStreaming
	isBidiStreaming := m.ServerSideStreaming && m.ClientSideStreaming
	isClientStreaming := m.ClientSideStreaming && !m.ServerSideStreaming
	isStreaming := isServerStreaming || isBidiStreaming || isClientStreaming

	isLRO := m.IsLRO || m.OperationInfo != nil
	var lroRespType, lroMetaType *ast.TypeNode
	if isLRO {
		lroRespType = TypeEmpty
		lroMetaType = ast.ObjectType("Any", PkgProto)
		if m.LongRunningResponseType != nil {
			lroRespType = MessageToJavaType(m.LongRunningResponseType)
		} else if m.OperationInfo != nil && m.OperationInfo.ResponseTypeID != "" && modelAnn.Model != nil {
			if target := modelAnn.Model.Message(m.OperationInfo.ResponseTypeID); target != nil {
				lroRespType = MessageToJavaType(target)
			}
		}
		if m.OperationInfo != nil && m.OperationInfo.MetadataTypeID != "" && modelAnn.Model != nil {
			if target := modelAnn.Model.Message(m.OperationInfo.MetadataTypeID); target != nil {
				lroMetaType = MessageToJavaType(target)
			}
		}
	}

	isPaged := m.IsList || (m.Pagination != nil && m.OutputType != nil && m.OutputType.Pagination != nil)
	var pageItemType *ast.TypeNode
	var pageItemField, pageTokenField, nextPageTokenField string
	if isPaged {
		pageTokenField = "page_token"
		nextPageTokenField = "next_page_token"
		if m.Pagination != nil {
			pageTokenField = m.Pagination.Name
		}
		if m.OutputType != nil && m.OutputType.Pagination != nil {
			if m.OutputType.Pagination.NextPageToken != nil {
				nextPageTokenField = m.OutputType.Pagination.NextPageToken.Name
			}
			if m.OutputType.Pagination.PageableItem != nil {
				pageItemField = m.OutputType.Pagination.PageableItem.Name
				wrapped := FieldTypeToJavaType(m.OutputType.Pagination.PageableItem)
				if wrapped.ElementType != nil {
					pageItemType = wrapped.ElementType
				} else {
					pageItemType = wrapped.Type
				}
			}
		}
		if pageItemType == nil {
			pageItemType = ast.TypeObject
		}
	}

	isUnary := !isStreaming && !isLRO && !isPaged

	// Method signatures
	var sigs [][]*api.Field
	for _, sig := range m.Signatures {
		if sig != nil && len(sig.Fields) > 0 {
			sigs = append(sigs, sig.Fields)
		}
	}

	return &MethodAnnotation{
		Method:                m,
		Name:                  m.Name,
		MethodName:            MethodName(m.Name),
		CallableName:          CallableMethodName(m.Name),
		PagedCallableName:     PagedCallableMethodName(m.Name),
		OperationCallableName: OperationCallableMethodName(m.Name),
		RequestType:           reqType,
		ResponseType:          respType,
		IsUnary:               isUnary,
		IsPaged:               isPaged,
		IsLRO:                 isLRO,
		IsServerStreaming:     isServerStreaming,
		IsBidiStreaming:       isBidiStreaming,
		IsClientStreaming:     isClientStreaming,
		Signatures:            sigs,
		LroResponseType:       lroRespType,
		LroMetadataType:       lroMetaType,
		PageItemType:          pageItemType,
		PageItemField:         pageItemField,
		PageTokenField:        pageTokenField,
		NextPageTokenField:    nextPageTokenField,
		Description:           FormatJavaDocComment(m.Documentation),
	}
}

func annotateResource(res *api.Resource, javaPkg string) *ResourceAnnotation {
	className := ResourceClassName(res.Type)
	var patterns []string
	for _, p := range res.Patterns {
		var parts []string
		for _, seg := range p {
			if seg.Literal != "" {
				parts = append(parts, seg.Literal)
			} else if seg.Variable != nil && len(seg.Variable.FieldPath) > 0 {
				parts = append(parts, "{"+strings.Join(seg.Variable.FieldPath, ".")+"}")
			}
		}
		if len(parts) > 0 {
			patterns = append(patterns, strings.Join(parts, "/"))
		}
	}
	return &ResourceAnnotation{
		Resource:    res,
		Type:        res.Type,
		ClassName:   className,
		Patterns:    patterns,
		PackageName: javaPkg,
	}
}
