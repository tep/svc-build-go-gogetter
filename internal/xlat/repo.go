// Copyright 2019 Timothy E. Peoples
//
// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to
// deal in the Software without restriction, including without limitation the
// rights to use, copy, modify, merge, publish, distribute, sublicense, and/or
// sell copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
// FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS
// IN THE SOFTWARE.

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
