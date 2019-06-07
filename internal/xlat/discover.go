package xlat

import (
	"context"

	"github.com/google/go-github/v25/github"
	"toolman.org/base/log/v2"
)

func (t *Translator) Discover(ctx context.Context) error {
	inst, err := t.listInstallations(ctx)
	if err != nil {
		return err
	}

	for _, in := range inst {
		ownr := in.GetAccount().GetLogin()
		pfx, ok := t.ownrpfx[ownr]
		if !ok {
			log.Warningf("INST=%q not configured", ownr)
			continue
		}

		log.Infof("INST=%q PREFIX=%q", ownr, pfx)

		repos, err := t.listRepos(ctx, in.GetID())
		if err != nil {
			return err
		}

		for _, r := range repos {
			t.updateRepo(pfx, r)
		}
	}

	return nil
}

func (t *Translator) UpdateRepo(repo *github.Repository, del bool) {
	if del {
		t.deleteRepo(repo)
		return
	}

	ownr := repo.GetOwner().GetLogin()
	pfx, ok := t.ownrpfx[ownr]

	if !ok {
		log.Warningf("Repo owner not configured: %s", ownr)
		return
	}

	t.updateRepo(pfx, repo)
}

func (t *Translator) updateRepo(pfx string, repo *github.Repository) {
	if repo.GetLanguage() != "Go" {
		log.V(1).Infof("Rejecting non-go repo: %s", repo.GetFullName())
		return
	}

	id := repo.GetID()

	if tr := t.repos[id]; tr != nil {
		delete(t.gopkgs, tr.pkgpfx)
	}

	nr := newRepo(pfx, repo)

	t.repos[id] = nr
	t.gopkgs[nr.pkgpfx] = id
}

func (t *Translator) deleteRepo(repo *github.Repository) {
	id := repo.GetID()
	tr := t.repos[id]

	if tr == nil {
		return
	}

	delete(t.gopkgs, tr.pkgpfx)
	delete(t.repos, id)

	log.Infof("Deleted repo %s/%s", tr.owner, tr.name)
}

// XXX: Are these needed?
// func github2gopkg(prefix, ghname string) string {
// 	return path.Join(prefix, strings.Replace(strings.Replace(ghname, "-", "/", -1), "//", "-", -1))
// }

// func gopkg2github(goname string) string {
// 	return strings.Replace(strings.Replace(goname, "//", "-", -1), "/", "-", -1)
// }
