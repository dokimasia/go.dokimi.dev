# go.dokimi.dev

Vanity import host for the dokimi Go ecosystem. Serves the `go-import` meta
tags that let `go get go.dokimi.dev/<module>` resolve to the canonical sources
at [github.com/dokimasia](https://github.com/dokimasia).

Live at <https://go.dokimi.dev/>.

## Structure

```
.
├── index.html          generated — module index
├── 404.html            generated — styled fallback
├── <module>/index.html generated — per-module landing + meta tags
├── assets/
│   ├── site.css        hand-written stylesheet (Tailwind tokens, buildless)
│   └── site.js         click-to-copy for `go get` snippets
└── _gen/               site generator (excluded from Pages by Jekyll's
                        leading-underscore convention)
    ├── main.go
    ├── data.go         source of truth — modules list
    ├── go.mod
    └── templates/
```

The `_gen/` directory is the source of truth for everything generated. The
rendered `*.html` files are committed so GitHub Pages can serve them directly
with zero build pipeline.

## Adding or editing a module

1. Edit `_gen/data.go` — add a `Module{}` to the `modules` slice.
2. `cd _gen && go run .`
3. Commit both `_gen/data.go` and the regenerated HTML.

The generator runs `npx prettier` on its output by default. Pass `-no-fmt` to
skip if npx isn't available. It must be run from inside `_gen/`: the default
`-out` is `..`, resolved against the working directory.

```go
{
    Name:        "assert",
    Repo:        "assert-go",                // optional; defaults to Name
    Description: "Test assertions for Go: ...",
    Public:      true,                       // shown under Public modules
    Langs:       []string{"go", "rust"},     // optional; defaults to ["go"]
},
```

`Repo` covers the case where the GitHub repository is named differently from
the import path — `go.dokimi.dev/assert` lives in `github.com/dokimasia/assert-go`.

## Submodules (nested go.mod)

When a module has its own nested `go.mod` (e.g. `core/go.mod` declaring
`module go.dokimi.dev/eidos/core`), the vanity host serves a page at the
submodule path carrying the **parent's** `go-import` meta tag:

```html
<meta name="go-import"
      content="go.dokimi.dev/eidos git https://github.com/dokimasia/eidos" />
```

The import-prefix is the repo root, not the submodule path. Go treats the
prefix as the repo's root module and matches the trailing path segments
against nested `go.mod`s inside the clone; naming the submodule as the prefix
would send Go to the repo-root `go.mod`, which declares the parent module, and
the resolve fails on a path mismatch.

Because Go matches those trailing segments against directory names, the
`go.mod` for `go.dokimi.dev/eidos/core` has to sit at `core/` inside the repo.
`Sub.Dir` only changes the "view on github" link; it does not change where Go
looks.

```go
Subs: []Sub{
    {Name: "core", Description: "...", Public: true},
    {Name: "lang-go", Dir: "eidos-lang-go", Description: "...", Public: true},
},
```

> **Note on nested-module `go.mod`s.** A `replace ... => ../` paired with a
> `v0.0.0-00010101000000-...` pseudo-version works only for local dev — the
> `replace` is ignored by downstream consumers. For published submodules,
> require a real tagged version of the parent, or move the `replace` into a
> top-level `go.work` (gitignored) so the published `go.mod` is
> consumer-clean.

## Private modules

Set `Public: false`. The landing page surfaces a one-liner `GOPRIVATE` hint
since `go get` against a private repo needs:

```sh
go env -w GOPRIVATE=go.dokimi.dev/*
```

…plus git credentials that can reach the GitHub org.

## Local preview

```sh
python3 -m http.server 8000     # then open http://localhost:8000/
```

To test meta-tag resolution:

```sh
curl -sL 'http://localhost:8000/eidos/core/?go-get=1' | grep go-import
```

## Why buildless on the served side

GitHub Pages serves the repo as-is. No CI, no Actions, no deploy pipeline.
The generator is a developer-ergonomics layer; the production artifact is
plain HTML + CSS + ~50 lines of JS.
