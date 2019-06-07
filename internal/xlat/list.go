package xlat

import (
	"context"

	"github.com/google/go-github/v25/github"
)

type listCallback func(*github.ListOptions) (*github.Response, error)

func multipageList(lcb listCallback) error {
	var (
		resp *github.Response
		err  error
	)

	for lopts := (&github.ListOptions{PerPage: 100}); resp == nil || resp.NextPage != 0; lopts.Page = resp.NextPage {
		if resp, err = lcb(lopts); err != nil {
			return err
		}
	}

	return nil
}

func (t *Translator) listInstallations(ctx context.Context) ([]*github.Installation, error) {
	var out []*github.Installation

	client, err := t.appClient()
	if err != nil {
		return nil, err
	}

	lcb := func(lopts *github.ListOptions) (*github.Response, error) {
		inst, resp, err := client.Apps.ListInstallations(ctx, lopts)
		if err != nil {
			return nil, err
		}
		out = append(out, inst...)
		return resp, nil
	}

	if err := multipageList(lcb); err != nil {
		return nil, err
	}

	return out, nil
}

func (t *Translator) listRepos(ctx context.Context, id int64) ([]*github.Repository, error) {
	var out []*github.Repository

	client, err := t.instClient(id)
	if err != nil {
		return nil, err
	}

	lcb := func(lopts *github.ListOptions) (*github.Response, error) {
		repos, resp, err := client.Apps.ListRepos(ctx, lopts)
		if err != nil {
			return nil, err
		}
		out = append(out, repos...)
		return resp, nil
	}

	if err := multipageList(lcb); err != nil {
		return nil, err
	}

	return out, nil
}
