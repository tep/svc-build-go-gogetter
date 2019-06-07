package xlat

import (
	"fmt"
	"io"
	"path"
	"strings"

	"github.com/google/go-github/v25/github"
)

type Repo struct {
	id      int64  // Github repository id
	owner   string // Github repository owner name (either user or org)
	name    string // Github repository name
	pkgpfx  string // Go package prefix corresponding to the repository root
	private bool   // Private repo flag
	htmlurl string // HTML URL for source browsers
	puburl  string // Clone URL for public repos
	privurl string // Clone URL for private repos
}

func newRepo(pfx string, gr *github.Repository) *Repo {
	nam := gr.GetName()

	// Translates pfx="example.com" and nam="one-two-buckle--my--shoe"
	// into "example.com/one/two/buckle-my-shoe"
	pkg := path.Join(pfx, strings.Replace(strings.Replace(nam, "-", "/", -1), "//", "-", -1))

	return &Repo{
		id:      gr.GetID(),
		owner:   gr.GetOwner().GetLogin(),
		name:    nam,
		pkgpfx:  pkg,
		private: gr.GetPrivate(),
		htmlurl: gr.GetHTMLURL(),
		puburl:  gr.GetCloneURL(),
		privurl: strings.Replace(gr.GetGitURL(), "git://", "ssh://git@", 1),
	}
}

const (
	importTag = `<meta name="go-import" content="%s git %s">` + "\r\n"
	sourceTag = `<meta name="go-source" content="%[1]s %[2]s %[2]s/tree/master{/dir} %[2]s/blob/master/{/dir}/{file}#L{line}">` + "\r\n"
)

func (r *Repo) WriteImportTags(w io.Writer) {
	// fmt.Fprintf(w, `<meta name="go-import" content="%s git %s">%s`, r.pkgpfx, r.goGetURL(), "\r\n")
	fmt.Fprintf(w, importTag, r.pkgpfx, r.goGetURL())
	fmt.Fprintf(w, sourceTag, r.pkgpfx, r.htmlurl)
}

func (r *Repo) goGetURL() string {
	if r.private {
		return r.privurl
	}
	return r.puburl
}
