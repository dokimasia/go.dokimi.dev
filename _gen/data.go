package main

import (
	"path"
	"strings"
)

const repoOrg = "dokimasia"

// Module describes a top-level module at go.dokimi.dev/<Name>.
type Module struct {
	Name string
	// Repo is the name of the GitHub repository when it differs from Name.
	// The repository of assert is assert-go.
	Repo string
	// Dir is the directory of the go.mod inside the repo when the module is
	// not at the root of the repo. The go.mod of conform is in the directory
	// conform of dokimi-go.
	Dir         string
	Description string
	Public      bool
	// Langs are the language tags that the site shows for the module. The
	// site shows go when Langs is empty.
	Langs []string
	Subs  []Sub
}

// Sub describes a nested module with its own go.mod in the repo of its parent,
// at go.dokimi.dev/<Parent>/<Name>.
type Sub struct {
	Name string
	// Dir is the directory of the go.mod inside the parent's repo when it
	// differs from Name.
	Dir         string
	Description string
	Public      bool
	Langs       []string

	parent *Module // main sets parent before it renders a page.
}

func (m Module) ImportPath() string { return "go.dokimi.dev/" + m.Name }
func (m Module) GoGet() string      { return "go get " + m.ImportPath() }

func (m Module) RepoName() string {
	if m.Repo == "" {
		return m.Name
	}
	return m.Repo
}

// repo returns the URL of the module's repo.
func (m Module) repo() string { return "https://github.com/" + repoOrg + "/" + m.RepoName() }

// RepoURL returns the URL of the module's directory on GitHub. For a module at
// the root of its repo, it returns the URL of the repo.
func (m Module) RepoURL() string {
	if m.Dir == "" {
		return m.repo()
	}
	return m.repo() + "/tree/main/" + m.Dir
}

func (m Module) Languages() []string {
	if len(m.Langs) == 0 {
		return []string{"go"}
	}
	return m.Langs
}

// GoImport returns the content of the go-import meta tag at the module's
// path. A module in a directory of its repo gets the directory as the fourth
// field of its tag. Go reads the fourth field from Go 1.25 on. The module's
// version tags then start with the directory. The tags of conform are
// conform/vX.Y.Z.
func (m Module) GoImport() string {
	if m.Dir == "" {
		return m.ImportPath() + " git " + m.repo()
	}
	return m.ImportPath() + " git " + m.repo() + " " + m.Dir
}

func (m Module) GoSource() string { return goSource(m.ImportPath(), m.repo(), m.Dir) }

// goSource returns the content of a go-source meta tag for the import path
// prefix. Its URL templates put the path after prefix below dir in the repo at
// repoURL.
func goSource(prefix, repoURL, dir string) string {
	base := path.Join("main", dir)
	return strings.Join([]string{
		prefix,
		"    " + repoURL,
		"    " + repoURL + "/tree/" + base + "{/dir}",
		"    " + repoURL + "/blob/" + base + "{/dir}/{file}#L{line}",
	}, "\n")
}

func (s Sub) Path() string       { return s.parent.Name + "/" + s.Name }
func (s Sub) ImportPath() string { return "go.dokimi.dev/" + s.Path() }
func (s Sub) GoGet() string      { return "go get " + s.ImportPath() }
func (s Sub) ParentName() string { return s.parent.Name }
func (s Sub) ParentImport() string {
	return s.parent.ImportPath()
}

// repoDir returns the directory of the module's go.mod inside the parent's
// repo.
func (s Sub) repoDir() string {
	if s.Dir == "" {
		return s.Name
	}
	return s.Dir
}

// RepoURL returns the URL of the module's directory on GitHub.
func (s Sub) RepoURL() string { return s.parent.repo() + "/tree/main/" + s.repoDir() }

// throughParent reports whether the module resolves through the go-import tag
// of its parent. Go takes the path after the tag's prefix as the module's
// directory inside the repo, so the parent must be at the root of its repo and
// the module's directory must be its path suffix.
func (s Sub) throughParent() bool { return s.parent.Dir == "" && s.repoDir() == s.Name }

// GoImport returns the content of the go-import meta tag at the module's path.
// A module that resolves through its parent gets the parent's tag. Any other
// module gets a tag with its own path as the prefix and its directory as the
// fourth field. Go reads the fourth field from Go 1.25 on. The module's
// version tags then start with the directory.
func (s Sub) GoImport() string {
	if s.throughParent() {
		return s.parent.GoImport()
	}
	return s.ImportPath() + " git " + s.parent.repo() + " " + s.repoDir()
}

// GoSource returns the content of the go-source meta tag at the module's path.
// Its prefix is the prefix of [Sub.GoImport], because pkg.go.dev rejects a page
// whose two tags state different prefixes.
func (s Sub) GoSource() string {
	if s.throughParent() {
		return s.parent.GoSource()
	}
	return goSource(s.ImportPath(), s.parent.repo(), s.repoDir())
}

func (s Sub) Languages() []string {
	if len(s.Langs) == 0 {
		return []string{"go"}
	}
	return s.Langs
}

// modules is the source of truth for the site. Edit this slice and re-run
// `cd _gen && go run .`.
var modules = []Module{
	{
		Name:        "assert",
		Repo:        "assert-go",
		Description: "Test assertions for Go, defined by a language-neutral standard and held to it on every run. Every assertion takes its contract last, and that contract is the first line of the failure.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "lint",
				Description: "The analyzer assertlint reports a check that a test writes by hand and that an assertion of go.dokimi.dev/assert states. It suggests the rewrite where the rewrite keeps the check's meaning.",
				Public:      true,
			},
			{
				Name:        "lint/golangci",
				Description: "The module plugin assertlint of golangci-lint. golangci-lint custom builds a binary with the analyzer of go.dokimi.dev/assert/lint, and a .golangci.yml enables it.",
				Public:      true,
			},
		},
	},
	{
		Name:        "conform",
		Repo:        "dokimi-go",
		Dir:         "conform",
		Description: "The module contains the runtime of the conformance suites that dokimi generates for Go. It proves each check of a suite on its planted defect, and applies the checks to each implementation that a test passes.",
		Public:      true,
	},
	{
		Name:        "dokimi",
		Description: "The command dokimi generates the tests and the test helpers of the declarations of a workspace. It reads Go packages and TypeScript sources with eidos.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "corpus/go",
				Description: "The module contains the Go corpus of dokimi. Each package declares the types of one fixture beside the tests and the test helpers that dokimi generates for them.",
				Public:      true,
			},
		},
	},
	{
		Name:        "eidos",
		Description: "Code generation across languages. A frontend parses source into a symbol graph that belongs to no language, plugins annotate that graph, a backend renders it. One run regenerates only what changed and produces the same bytes every time.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "cli",
				Dir:         "eidos-cli",
				Description: "The module contains the kernel commands that every binary built with eidos offers: run, plan, explain, prune, doctor, watch, version and schema. A binary runs them through Main or mounts them on a command line of its own.",
				Public:      true,
			},
			{
				Name:        "conformance",
				Dir:         "eidos-conformance",
				Description: "The module checks every language satellite against one corpus: a feature inventory, a tree for each language and one set of expectations. Its package matrix renders the support matrix of the satellites.",
				Public:      true,
			},
			{
				Name:        "core",
				Dir:         "eidos-core",
				Description: "The kernel: symbol model, projections, directives, plugins, workspace, engine, conformance and command kernels. Knows no language and takes no third-party dependencies.",
				Public:      true,
			},
			{
				Name:        "declarative",
				Dir:         "eidos-declarative",
				Description: "The module loads declarative plugins. A declarative plugin is a directory with a manifest and template trees. It runs as a generator without Go code.",
				Public:      true,
			},
			{
				Name:        "lang",
				Dir:         "eidos-lang",
				Description: "Tree-sitter binding layer and pinned grammars, shared by the tree-sitter satellites.",
				Public:      true,
			},
			{
				Name:        "lang/go",
				Dir:         "eidos-lang-go",
				Description: "Go language satellite.",
				Public:      true,
			},
			{
				Name:        "lang/graphql",
				Dir:         "eidos-lang-graphql",
				Description: "The module is the GraphQL satellite of eidos. Its frontend loads schemas and executable documents into the node graph. Its backend renders emit declarations as GraphQL documents.",
				Public:      true,
			},
			{
				Name:        "lang/java",
				Dir:         "eidos-lang-java",
				Description: "Java language satellite.",
				Public:      true,
			},
			{
				Name:        "lang/kotlin",
				Dir:         "eidos-lang-kotlin",
				Description: "Kotlin language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang/php",
				Dir:         "eidos-lang-php",
				Description: "PHP language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang/protobuf",
				Dir:         "eidos-lang-protobuf",
				Description: "Protobuf language satellite. Read-only by design — proto is parsed, never emitted.",
				Public:      true,
			},
			{
				Name:        "lang/python",
				Dir:         "eidos-lang-python",
				Description: "The module is the Python satellite of eidos. Its frontend loads modules and stub files into the node graph. Its backend renders emit declarations as Python modules and stub files.",
				Public:      true,
			},
			{
				Name:        "lang/rust",
				Dir:         "eidos-lang-rust",
				Description: "Rust language satellite.",
				Public:      true,
			},
			{
				Name:        "lang/typescript",
				Dir:         "eidos-lang-typescript",
				Description: "TypeScript language satellite.",
				Public:      true,
			},
			{
				Name:        "plugin/shape",
				Dir:         "eidos-plugin-shape",
				Description: "The shape catalog classifies callables by shape, mixin and contract. It stamps each classification as a fact. A directive at the declaration corrects a wrong classification.",
				Public:      true,
			},
			{
				Name:        "plugin/shape/tools",
				Dir:         "eidos-plugin-shape/tools",
				Description: "The module generates the registry of the shape catalog from its YAML specs through an eidos workspace.",
				Public:      true,
			},
			{
				Name:        "reference",
				Dir:         "eidos-reference",
				Description: "The reference plugin ensemble, which doubles as the compatibility canary and the benchmark rig.",
				Public:      true,
			},
			{
				Name:        "sdk",
				Dir:         "eidos-sdk",
				Description: "The plugin-facing facade. Re-exports the kernel's supported surface as aliases and wrappers generated from a curated package list, and defines nothing of its own. Plugin code imports SDK paths alone.",
				Public:      true,
			},
		},
	},
	{
		Name:        "ergon",
		Description: "The command ergon sets up a repository with code in one or more of eleven languages. It writes the common files, the GitHub files and the files of each language. The command ergon init check reports each of those files that is missing, edited by hand or outdated. The root module contains the binary and the composition root.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "core",
				Dir:         "ergon-core",
				Description: "The module declares the toolchains and the languages of ergon, the catalog that registers them, and the role interfaces that each command selects. Its packages import only the standard library and each other.",
				Public:      true,
			},
			{
				Name:        "lang/bash",
				Dir:         "ergon-lang-bash",
				Description: "The module declares Bash and its toolchain, the bash command. Its package baseline renders the files that ergon init writes for Bash.",
				Public:      true,
			},
			{
				Name:        "lang/csharp",
				Dir:         "ergon-lang-csharp",
				Description: "The module declares C# and its toolchain, the dotnet command. Its package baseline renders the files that ergon init writes for C#.",
				Public:      true,
			},
			{
				Name:        "lang/go",
				Dir:         "ergon-lang-go",
				Description: "The module declares Go and its toolchain, the go command. Its package baseline renders the files that ergon init writes for Go.",
				Public:      true,
			},
			{
				Name:        "lang/java",
				Dir:         "ergon-lang-java",
				Description: "The module declares Java and the jvm toolchain, which builds Java and Kotlin projects with Gradle. Its package baseline renders the files that ergon init writes for Java, and the files that Java and Kotlin share.",
				Public:      true,
			},
			{
				Name:        "lang/javascript",
				Dir:         "ergon-lang-javascript",
				Description: "The module declares JavaScript and the js toolchain, which builds JavaScript and TypeScript packages with npm, pnpm or bun. Its package baseline renders the files that ergon init writes for JavaScript, and the files that JavaScript and TypeScript share.",
				Public:      true,
			},
			{
				Name:        "lang/kotlin",
				Dir:         "ergon-lang-kotlin",
				Description: "The module declares Kotlin, which the jvm toolchain of the Java module builds. Its package baseline renders the files that ergon init writes for Kotlin.",
				Public:      true,
			},
			{
				Name:        "lang/php",
				Dir:         "ergon-lang-php",
				Description: "The module declares PHP and its toolchain, the php command with Composer. Its package baseline renders the files that ergon init writes for PHP.",
				Public:      true,
			},
			{
				Name:        "lang/python",
				Dir:         "ergon-lang-python",
				Description: "The module declares Python and its toolchain, uv. Its package baseline renders the files that ergon init writes for Python.",
				Public:      true,
			},
			{
				Name:        "lang/rust",
				Dir:         "ergon-lang-rust",
				Description: "The module declares Rust and its toolchain, Cargo. Its package baseline renders the files that ergon init writes for Rust.",
				Public:      true,
			},
			{
				Name:        "lang/terraform",
				Dir:         "ergon-lang-terraform",
				Description: "The module declares Terraform and its toolchain, the terraform command. Its package baseline renders the files that ergon init writes for Terraform.",
				Public:      true,
			},
			{
				Name:        "lang/typescript",
				Dir:         "ergon-lang-typescript",
				Description: "The module declares TypeScript, which the js toolchain of the JavaScript module builds. Its package baseline renders the files that ergon init writes for TypeScript.",
				Public:      true,
			},
			{
				Name:        "service",
				Dir:         "ergon-service",
				Description: "The module contains the side of each ergon command that is the same for every language. Its package baseline sets up a repository with the files of ergon init and keeps them at the baseline of the installed ergon.",
				Public:      true,
			},
		},
	},
	{
		Name:        "fault",
		Repo:        "dokimi-go",
		Dir:         "fault",
		Description: "The module contains the fault strategies of the doubles that dokimi generates for Go. A strategy selects the calls of a double that fail, by count, by probability, in a window of time or by a predicate of the call.",
		Public:      true,
	},
	{
		Name:        "lint",
		Repo:        "lint-go",
		Description: "The command dokimi-lint-go runs the analyzers errorprefix and skipexpiry over Go packages, as go vet runs its analyzers. errorprefix reports an error text that does not start with the name of its package, and skipexpiry reports a skipped test whose expiry has passed. The target lint-go of ergon runs it in every module.",
		Public:      true,
	},
	{
		Name:        "mutate",
		Repo:        "mutate-go",
		Description: "The command dokimi-mutate-go measures how well the tests of a Go package detect faults. It runs the tests against small changes to the package's code and reports each change that no test detects. The function mutate.Check runs the same engine from a test. A language-neutral standard defines the changes.",
		Public:      true,
	},
	{
		Name:        "stub",
		Repo:        "dokimi-go",
		Dir:         "stub",
		Description: "The module contains the runtime of the doubles that dokimi generates for Go. A double records each call, fails the calls that the faults of a test select, and returns the results that the test configures.",
		Public:      true,
	},
	{
		Name:        "techne",
		Description: "Code intelligence and type-checked refactoring, exposed as tools. Every answer carries the evidence behind it, so a caller can tell a resolved fact from a syntactic guess. The root module holds the binary and the composition root.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "core",
				Dir:         "techne-core",
				Description: "The language-agnostic half: the vocabulary every other module speaks, the ports an engine implements, and the services that drive them. Declares nothing itself; the packages beneath it do.",
				Public:      true,
			},
			{
				Name:        "lang",
				Dir:         "techne-lang",
				Description: "Declares what a language is, independently of which engine answers questions about it. Carries the tree-sitter engine, which serves any language with a grammar at the syntactic tier.",
				Public:      true,
			},
			{
				Name:        "lang/c",
				Dir:         "techne-lang-c",
				Description: "The module declares the C language with its tree-sitter grammar and clangd as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/csharp",
				Dir:         "techne-lang-csharp",
				Description: "The module declares the C# language with its tree-sitter grammar and csharp-ls as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/go",
				Dir:         "techne-lang-go",
				Description: "Go satellite: everything true of Go and nothing true of any other language — its declaration, its queries, and the engines only Go can use.",
				Public:      true,
			},
			{
				Name:        "lang/java",
				Dir:         "techne-lang-java",
				Description: "The module declares the Java language with its tree-sitter grammar and jdtls as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/javascript",
				Dir:         "techne-lang-javascript",
				Description: "The module declares the JavaScript language with its tree-sitter grammar and typescript-language-server as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/kotlin",
				Dir:         "techne-lang-kotlin",
				Description: "Kotlin satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang/mock",
				Dir:         "techne-lang-mock",
				Description: "The module declares a line-oriented mock language. Tests register it to exercise each tool of techne at a tier that they choose.",
				Public:      true,
			},
			{
				Name:        "lang/php",
				Dir:         "techne-lang-php",
				Description: "PHP satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang/protobuf",
				Dir:         "techne-lang-protobuf",
				Description: "Protobuf satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang/python",
				Dir:         "techne-lang-python",
				Description: "The module declares the Python language with its tree-sitter grammar and pyright as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/ruby",
				Dir:         "techne-lang-ruby",
				Description: "The module declares the Ruby language with its tree-sitter grammar and ruby-lsp as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/rust",
				Dir:         "techne-lang-rust",
				Description: "The module declares the Rust language with its tree-sitter grammar and rust-analyzer as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/scala",
				Dir:         "techne-lang-scala",
				Description: "The module declares the Scala language with its tree-sitter grammar and Metals as its language server.",
				Public:      true,
			},
			{
				Name:        "lang/typescript",
				Dir:         "techne-lang-typescript",
				Description: "The module declares the TypeScript language with its tree-sitter grammars and typescript-language-server as its language server.",
				Public:      true,
			},
			{
				Name:        "presenter",
				Dir:         "techne-presenter",
				Description: "Carries a tool call between a transport and the tool that serves it. Holds no domain knowledge, so one presenter serves every tool and transport pair.",
				Public:      true,
			},
			{
				Name:        "service",
				Dir:         "techne-service",
				Description: "The module contains the services that call the ports of core. Its package query serves the read roles. Its package change plans, gates and writes a change.",
				Public:      true,
			},
			{
				Name:        "tool",
				Dir:         "techne-tool",
				Description: "The module defines the tools that techne offers an agent. A tool returns its answer as JSON for a program and as text for a model.",
				Public:      true,
			},
		},
	},
	{
		Name:        "treesitter",
		Dir:         "treesitter",
		Description: "Reports what a source file declares, in one vocabulary, for every language a grammar exists for. The vocabulary is smaller than any grammar's node set and larger than the weakest grammar can tell apart, so a Java class and a Go struct both answer as one shape.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "c",
				Dir:         "treesitter-c",
				Description: "Reads C with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "csharp",
				Dir:         "treesitter-csharp",
				Description: "Reads C# with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "go",
				Dir:         "treesitter-go",
				Description: "Reads Go with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "java",
				Dir:         "treesitter-java",
				Description: "Reads Java with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "javascript",
				Dir:         "treesitter-javascript",
				Description: "Reads JavaScript with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "python",
				Dir:         "treesitter-python",
				Description: "Reads Python with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "ruby",
				Dir:         "treesitter-ruby",
				Description: "Reads Ruby with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "rust",
				Dir:         "treesitter-rust",
				Description: "Reads Rust with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "scala",
				Dir:         "treesitter-scala",
				Description: "Reads Scala with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
			{
				Name:        "tools/measure",
				Description: "The command measure compares the symbols that the library finds in a real repository with the symbols that the parser of the language finds. It reports the recall and the disagreements by kind.",
				Public:      true,
			},
			{
				Name:        "typescript",
				Dir:         "treesitter-typescript",
				Description: "Reads TypeScript with tree-sitter. The grammar is cgo, so importing this needs a C toolchain.",
				Public:      true,
			},
		},
	},
}
