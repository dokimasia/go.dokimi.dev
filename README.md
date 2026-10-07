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

A module with its own nested `go.mod` gets a page at its path. Set `Dir` when
the `go.mod` is in a directory other than the path suffix:

```go
Subs: []Sub{
    {Name: "lint", Description: "...", Public: true},
    {Name: "lang/go", Dir: "ergon-lang-go", Description: "...", Public: true},
},
```

Without `Dir`, the `go.mod` of `go.dokimi.dev/assert/lint` is in `lint/`. The
page serves the parent's `go-import` tag:

```html
<meta name="go-import"
      content="go.dokimi.dev/assert git https://github.com/dokimasia/assert-go" />
```

Go takes the path after the tag's prefix as the module's directory in the
repository. The module's version tags are `lint/vX.Y.Z`.

With `Dir`, the page serves the module's own path as the prefix and `Dir` as
the fourth field, which Go reads from Go 1.25 on:

```html
<meta name="go-import"
      content="go.dokimi.dev/ergon/lang/go git https://github.com/dokimasia/ergon ergon-lang-go" />
```

The module's version tags are `ergon-lang-go/vX.Y.Z`. The `go-source` tag
states the same prefix, because pkg.go.dev rejects a page whose two tags state
different prefixes.

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
