# Sidekick Java Client Generator: Architecture & Design Decisions

## 1. Executive Summary

This document details the design decisions, architecture, and roadmap for the **Sidekick Java Generator** (`internal/sidekick/java`).

Historically, Google API client library generators (GAPIC generators) operated as `protoc` compiler plugins, ingesting raw `pluginpb.CodeGeneratorRequest` payloads containing `FileDescriptorProto` descriptors. While functional, the protoc-coupled approach poses significant limitations:
1. **Coupling to Protobuf**: Cannot generate clients from OpenAPI or Google Discovery doc specifications without awkward intermediaries.
2. **Re-parsing and Boilerplate**: Every language generator must independently re-parse custom options (AIP annotations, HTTP rules, method signatures, routing headers, LRO metadata).
3. **External Binary Overhead**: Requires compiling and invoking external protoc/plugin binaries.

The **Sidekick Model** replaces protoc plugins with a **unified semantic AST model** (`*api.API`) constructed by language-agnostic parsers (`internal/sidekick/parser`). The Java generator (`internal/sidekick/java`) generates complete, idiomatic Java client libraries (GAX-based) directly from this unified model.

Java serves as the initial vanguard for migrating other GAPIC generators (PHP, Ruby, Go, Python, TypeScript, etc.) into Sidekick.

---

## 2. Architecture Overview

```
                      +-----------------------------+
                      | Input Specifications        |
                      | (Protobuf, OpenAPI, Disco)  |
                      +--------------+--------------+
                                     |
                                     v
                      +-----------------------------+
                      | Sidekick Semantic Parser    |
                      | (internal/sidekick/parser)  |
                      +--------------+--------------+
                                     |
                                     v
                      +-----------------------------+
                      | Unified API Semantic Model  |
                      | (*api.API, *api.Service,    |
                      |  *api.Method, *api.Resource)|
                      +--------------+--------------+
                                     |
                +--------------------+--------------------+
                |                    |                    |
                v                    v                    v
     +--------------------+ +--------------------+ +--------------------+
     | Java Generator     | | PHP / Ruby / Go    | | Python / TS        |
     | (Sidekick Model)   | | Generators         | | Generators         |
     +----------+---------+ +--------------------+ +--------------------+
                |
                v
  +-----------------------------+
  | Pure Java AST & Composer    |
  | (Class, Method, Field, Type)|
  +-------------+---------------+
                |
                v
  +-----------------------------+
  | Source Writer & Formatter   |
  | (Imports, Javadoc, Java src)|
  +-------------+---------------+
                |
                v
  +-----------------------------+
  | Emitted Artifacts           |
  | - <Service>Client.java      |
  | - <Service>Settings.java    |
  | - stub/<Service>Stub.java   |
  | - stub/*StubSettings.java   |
  | - stub/Grpc*Stub.java       |
  | - stub/HttpJson*Stub.java   |
  | - <Resource>Name.java       |
  | - package-info.java         |
  | - Version.java              |
  | - gapic_metadata.json       |
  | - reflect-config.json       |
  +-----------------------------+
```

---

## 3. Java Generator Design Decisions

### 3.1 Pure AST vs Template-Based Generation
* **Decision**: We use a strongly-typed **Pure Java AST (`engine/ast`)** with a dedicated **AST Serializer (`engine/writer`)** rather than raw mustache/text templates.
* **Rationale**:
  - Java source code is strictly structured (packages, class hierarchies, imports, types, generic parameters, annotations, modifiers).
  - An AST eliminates the syntax and whitespace bugs common to string templating.
  - **Automated Import Management**: The AST writer collects all referenced types (`ast.TypeNode`), automatically resolves imports, avoids unneeded imports in the same package or `java.lang.*`, and avoids name collisions.
  - **Refactoring & Optimization**: AST nodes can be manipulated, inspected, and verified programmatically.

### 3.2 Annotation & Semantic Enrichment (`annotate.go`)
* **Decision**: Two-phase processing pipeline:
  1. `AnnotateModel(model *api.API, codec *Codec)` maps language-agnostic `api.API` entities into Java-specific semantic annotations (`ModelAnnotations`, `ServiceAnnotation`, `MethodAnnotation`, `ResourceAnnotation`).
  2. `ComposeAll(ann)` turns annotations into AST definitions (`ast.ClassDefinition`).
* **Rationale**:
  - Isolates Java naming rules (e.g., camelCase, PascalCase, reserved keyword escaping) and GAX idioms (e.g., `*Callable`, `*PagedResponse`, `*Settings`) from the raw API model.
  - Allows unit tests to verify semantic analysis (`annotate_test.go`) independently of AST composition (`composer_test.go`).

### 3.3 Dual Transport Support (gRPC + HTTP/JSON REST)
* **Decision**: Support gRPC, HTTP/JSON (REST), and hybrid dual transports out of the box via Google Api eXtensions (GAX).
* **Generated Classes**:
  - Base Stub: `stub/<Service>Stub.java` (abstract callable interface)
  - Base Stub Settings: `stub/<Service>StubSettings.java` (channel providers, retry configs, credentials)
  - gRPC Implementation: `stub/Grpc<Service>Stub.java` + `stub/Grpc<Service>CallableFactory.java`
  - REST Implementation: `stub/HttpJson<Service>Stub.java` + `stub/HttpJson<Service>CallableFactory.java`

### 3.4 Resource Names & Pattern Matching (`resource_name_composer.go`)
* **Decision**: Generates strongly-typed resource name classes implementing `com.google.api.resourcenames.ResourceName` with full single- and multi-pattern matching via `PathTemplate`, dedicated factory methods (`of`, `of<PatternVariant>Name`), format methods (`format`, `format<PatternVariant>Name`), parsing (`parse`, `isParsableFrom`), and equality helpers.

### 3.5 Cloud Native & GraalVM Readiness
* **Decision**:
  - Emits `gapic_metadata.json` for GAPIC tooling and client verification.
  - Emits `resources/META-INF/native-image/.../reflect-config.json` for GraalVM Ahead-Of-Time (AOT) native image compilation.

---

## 4. Multi-Language Extensibility Roadmap

Java is the first generator designed under this paradigm, establishing reusable principles for other languages:

| Language | Primary Target Architecture | Modeling Strategy | Transport Layer |
| :--- | :--- | :--- | :--- |
| **Java** (This PR) | Pure AST (`engine/ast`) + GAX | `*api.API` -> `AnnotateModel` -> AST -> Writer | gRPC + REST (GAX) |
| **PHP** | AST / Mustache + GAX PHP | `*api.API` -> `AnnotateModel` -> Class Composers | gRPC + REST (GAX-PHP) |
| **Ruby** | Template / AST + GAPIC Ruby | `*api.API` -> `AnnotateModel` -> Ruby Generators | gRPC + REST (gapic-common) |
| **Go** | AST (`go/ast`, `go/printer`) + GAPIC Go | `*api.API` -> `AnnotateModel` -> Go AST | gRPC + REST (gax-go) |
| **Python** | AST / Template + GAPIC Python | `*api.API` -> `AnnotateModel` -> Python Gen | gRPC + REST (google-api-core) |
| **TypeScript** | AST (`ts-morph` or AST) + GAX TS | `*api.API` -> `AnnotateModel` -> TS Gen | gRPC + REST (google-gax) |

### Shared Generator Core Principles:
1. **Unified Semantic Input**: All generators consume `*api.API` produced by `internal/sidekick/parser`.
2. **Lexicon Isolation**: Language keywords, casing rules, and comment sanitization reside in isolated `engine/lexicon` or `names.go`.
3. **Deterministic Output**: Consistent file layouts, licenses, and metadata.
4. **Transport Parity**: Native support for gRPC, REST, and dual-transport configurations.

---

## 5. Comparison: Protoc Plugin Model vs Sidekick Model

| Feature | Protoc Plugin Model (PR #7374) | Sidekick Model (This PR) |
| :--- | :--- | :--- |
| **Input Format** | Protobuf `CodeGeneratorRequest` only | Protobuf, OpenAPI, Google Discovery |
| **Parsing Overhead** | Re-parsed per language plugin | Parsed once into unified `*api.API` |
| **Tooling Dependency** | Requires external `protoc` & plugins | Pure Go native execution |
| **Extensibility** | Hard to add non-proto API sources | Trivially extensible via parser layer |
| **Testing & CI** | Heavyweight CLI invocations | Lightweight in-memory unit & golden tests |
| **Code Generation Strategy** | Monolithic plugin invocation | Modular AST / Composer architecture |

---

## 6. Verification and Quality Assurance

- **Unit Testing**:
  - `engine/lexicon`: Keyword collision prevention, camelCase/PascalCase conversion.
  - `engine/ast` & `engine/writer`: Structural AST validity, automated import resolution.
  - `names.go` & `types.go`: Canonical GAX type and method mapping.
  - `annotate.go` & `composer.go`: Full class and metadata generation for unary, paged, LRO, streaming RPCs, and resources.
  - `generate.go`: End-to-end client artifact generation.
- **Linter & Formatting**:
  - Complies with repository linters and Go 1.24 formatting standards.
