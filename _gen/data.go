package main

import "strings"

const repoOrg = "dokimasia"

// Module describes a top-level vanity module exposed at go.dokimi.dev/<Name>.
// The repo defaults to github.com/<repoOrg>/<Name>; set Repo where the GitHub
// repository is named differently from the import path.
type Module struct {
	Name        string
	Repo        string // optional; GitHub repo name, defaults to Name
	Description string
	Public      bool
	Langs       []string // optional; defaults to []string{"go"}
	Subs        []Sub
}

// Sub is a nested module with its own go.mod, exposed at
// go.dokimi.dev/<Parent>/<Name>. Always inherits the parent's repo.
type Sub struct {
	Name        string
	Dir         string // optional; directory inside the parent repo, defaults to Name
	Description string
	Public      bool
	Langs       []string

	parent *Module // wired in main()
}

func (m Module) ImportPath() string { return "go.dokimi.dev/" + m.Name }
func (m Module) GoGet() string      { return "go get " + m.ImportPath() }

func (m Module) RepoName() string {
	if m.Repo == "" {
		return m.Name
	}
	return m.Repo
}

func (m Module) RepoURL() string { return "https://github.com/" + repoOrg + "/" + m.RepoName() }

func (m Module) Languages() []string {
	if len(m.Langs) == 0 {
		return []string{"go"}
	}
	return m.Langs
}

func (m Module) GoSource() string {
	return strings.Join([]string{
		m.ImportPath(),
		"    " + m.RepoURL(),
		"    " + m.RepoURL() + "/tree/main{/dir}",
		"    " + m.RepoURL() + "/blob/main{/dir}/{file}#L{line}",
	}, "\n")
}

func (s Sub) Path() string       { return s.parent.Name + "/" + s.Name }
func (s Sub) ImportPath() string { return "go.dokimi.dev/" + s.Path() }
func (s Sub) GoGet() string      { return "go get " + s.ImportPath() }
func (s Sub) ParentName() string { return s.parent.Name }
func (s Sub) ParentImport() string {
	return s.parent.ImportPath()
}

// RepoURL points at the subfolder on the parent repo — Go finds the nested
// go.mod there; the submodule's go-import meta tag still uses the parent repo
// as repo-root. Dir covers the case where that subfolder is not yet named
// after the module path suffix.
func (s Sub) RepoURL() string {
	dir := s.Dir
	if dir == "" {
		dir = s.Name
	}
	return s.parent.RepoURL() + "/tree/main/" + dir
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
	},
	{
		Name:        "eidos",
		Description: "Code generation across languages. A frontend parses source into a symbol graph that belongs to no language, plugins annotate that graph, a backend renders it. One run regenerates only what changed and produces the same bytes every time. Not released yet — the specification is complete and every module is a stub.",
		Public:      true,
		Subs: []Sub{
			{
				Name:        "core",
				Dir:         "eidos-core",
				Description: "The kernel: symbol model, projections, directives, plugins, workspace, engine, conformance and command kernels. Knows no language and takes no third-party dependencies.",
				Public:      true,
			},
			{
				Name:        "lang",
				Dir:         "eidos-lang",
				Description: "Tree-sitter binding layer and pinned grammars, shared by the tree-sitter satellites.",
				Public:      true,
			},
			{
				Name:        "lang-go",
				Dir:         "eidos-lang-go",
				Description: "Go language satellite.",
				Public:      true,
			},
			{
				Name:        "lang-java",
				Dir:         "eidos-lang-java",
				Description: "Java language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang-kotlin",
				Dir:         "eidos-lang-kotlin",
				Description: "Kotlin language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang-php",
				Dir:         "eidos-lang-php",
				Description: "PHP language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang-protobuf",
				Dir:         "eidos-lang-protobuf",
				Description: "Protobuf language satellite. Read-only by design — proto is parsed, never emitted.",
				Public:      true,
			},
			{
				Name:        "lang-rust",
				Dir:         "eidos-lang-rust",
				Description: "Rust language satellite. Not written yet.",
				Public:      true,
			},
			{
				Name:        "lang-typescript",
				Dir:         "eidos-lang-typescript",
				Description: "TypeScript language satellite.",
				Public:      true,
			},
			{
				Name:        "plugin-shape",
				Dir:         "eidos-plugin-shape",
				Description: "The classification catalog: shapes, mixins and contracts, written as specs first.",
				Public:      true,
			},
			{
				Name:        "reference",
				Dir:         "eidos-reference",
				Description: "The reference plugin ensemble, which doubles as the compatibility canary and the benchmark rig.",
				Public:      true,
			},
		},
	},
}
