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
	"errors"
	"fmt"
	"path"
	"strconv"

	"toolman.org/base/log/v2"
	"toolman.org/svc/build/go/gogetter/internal/config"
)

type Translator struct {
	prefixes []string          // List of all configured pkg prefixes
	ownrpfx  map[string]string // Github repo owner -> Go package prefix
	repos    map[int64]*Repo   // Github repo id    -> *Repo
	gopkgs   map[string]int64  // Go package name   -> Github repo id

	*config.Config
}

func New(cfg *config.Config) (*Translator, error) {
	xlatr := &Translator{
		ownrpfx: make(map[string]string),
		repos:   make(map[int64]*Repo),
		gopkgs:  make(map[string]int64),

		Config: cfg,
	}

	pset := make(map[string]bool)

	for _, d := range cfg.Trans {
		for _, o := range d.Owners {
			if p, ok := xlatr.ownrpfx[o]; ok {
				return nil, fmt.Errorf("multiple prefix mappings for repo owner %q: %q and %q", o, p, d.Prefix)
			}
			xlatr.ownrpfx[o] = d.Prefix
			pset[d.Prefix] = true
		}
	}

	if len(xlatr.ownrpfx) == 0 {
		return nil, errors.New("no translator definitions")
	}

	var i int
	xlatr.prefixes = make([]string, len(pset))
	for p := range pset {
		xlatr.prefixes[i] = p
		i++
	}

	return xlatr, nil
}

func (t *Translator) Lookup(importPath string) *Repo {
	log.Infof("Lookup: %q", importPath)
	for name := trimVersion(importPath); name != "."; name = trimPackage(name) {
		log.Infof("name=%q", name)
		if id, ok := t.gopkgs[name]; ok {
			return t.repos[id]
		}
	}

	return nil
}

func (t *Translator) Dump() {
	for pkg, id := range t.gopkgs {
		log.Infof("%-45s %s", pkg, t.repos[id].goGetURL())
	}
}

func trimVersion(name string) string {
	front, back := path.Split(name)
	if len(back) > 1 && back[0] == 'v' {
		if _, err := strconv.Atoi(back[1:]); err == nil {
			name = path.Clean(front)
		}
	}
	return name
}

func trimPackage(name string) string {
	name, _ = path.Split(path.Clean(name))
	return path.Clean(name)
}
