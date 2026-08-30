// Static site generator for go.dokimi.dev.
//
// Usage:
//	cd _gen && go run .
//
// The default -out is "..", resolved against the working directory, so the
// generator must be run from inside _gen/.
//
// Reads module data from data.go, expands the embedded templates, and writes
// index.html / 404.html plus <module>/index.html and <module>/<sub>/index.html
// for every entry. The generator is the source of truth — edit data.go (or
// the templates) and re-run.
package main

import (
	"embed"
	"flag"
	"html/template"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

//go:embed templates/*.html
var tmplFS embed.FS

func main() {
	out := flag.String("out", "..", "output directory (relative to _gen/)")
	noFmt := flag.Bool("no-fmt", false, "skip the prettier formatting pass")
	flag.Parse()

	root, err := filepath.Abs(*out)
	if err != nil {
		log.Fatal(err)
	}
	var written []string
	track := func(rel string) string {
		written = append(written, filepath.Join(root, rel))
		return rel
	}

	// Wire submodules back to their parent so template helpers can compute
	// their full path / repo URL without extra context.
	for i := range modules {
		for j := range modules[i].Subs {
			modules[i].Subs[j].parent = &modules[i]
		}
	}

	var pub, priv []Module
	for _, m := range modules {
		if m.Public {
			pub = append(pub, m)
		} else {
			priv = append(priv, m)
		}
	}

	render(root, track("index.html"), "index", indexPage{
		Title:   "go.dokimi.dev — Go modules",
		Public:  pub,
		Private: priv,
	})
	render(root, track("404.html"), "notfound", notFoundPage{
		Title: "404 — go.dokimi.dev",
	})

	for _, m := range modules {
		render(root, track(filepath.Join(m.Name, "index.html")), "module", modulePage{
			Title:    "go.dokimi.dev/" + m.Name,
			GoImport: m.ImportPath() + " git " + m.RepoURL(),
			GoSource: m.GoSource(),
			Module:   m,
		})
		for _, s := range m.Subs {
			// Emit the parent module's meta tag at every submodule path.
			// Go's resolver treats the meta-tag's import-prefix as the repo's
			// root module; the trailing URL segments are matched against
			// nested go.mod paths inside the repo. Emitting the submodule's
			// own path as import-prefix tricks Go into reading the repo-root
			// go.mod, which declares the parent module → path mismatch error.
			render(root, track(filepath.Join(m.Name, s.Name, "index.html")), "submodule", subPage{
				Title:    "go.dokimi.dev/" + m.Name + "/" + s.Name,
				GoImport: m.ImportPath() + " git " + m.RepoURL(),
				GoSource: m.GoSource(),
				Sub:      s,
			})
		}
	}

	log.Printf("wrote site to %s", root)

	if !*noFmt {
		if err := formatHTML(written); err != nil {
			log.Printf("prettier skipped: %v", err)
		}
	}
}

// formatHTML runs prettier over the rendered files via npx. Falls through with
// an error if npx isn't on PATH so the generator stays useful in barebones
// environments — the build is correct either way; only the formatting differs.
func formatHTML(files []string) error {
	if _, err := exec.LookPath("npx"); err != nil {
		return err
	}
	args := append([]string{"--yes", "--package=prettier@3", "prettier", "--write", "--log-level=warn"}, files...)
	cmd := exec.Command("npx", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func render(root, relPath, tmplName string, data any) {
	t := template.Must(template.New("").ParseFS(tmplFS,
		"templates/_partials.html",
		"templates/"+tmplName+".html",
	))
	full := filepath.Join(root, relPath)
	if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
		log.Fatal(err)
	}
	f, err := os.Create(full)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	if err := t.ExecuteTemplate(f, tmplName, data); err != nil {
		log.Fatal(err)
	}
}

// Page data types. Title/GoImport/GoSource are read by the head partial.

type indexPage struct {
	Title    string
	GoImport string // unused, satisfies head partial
	GoSource string
	Public   []Module
	Private  []Module
}

type modulePage struct {
	Title    string
	GoImport string
	GoSource string
	Module
}

type subPage struct {
	Title    string
	GoImport string
	GoSource string
	Sub
}

type notFoundPage struct {
	Title    string
	GoImport string
	GoSource string
}
