package xlat

import (
	"net/http"

	"github.com/bradleyfalzon/ghinstallation"
	"github.com/google/go-github/v25/github"
)

func (t *Translator) appClient() (*github.Client, error) {
	tr, err := ghinstallation.NewAppsTransport(http.DefaultTransport, t.IntegrationID, []byte(t.APIKey))
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: tr}), nil
}

func (t *Translator) instClient(id int64) (*github.Client, error) {
	tr, err := ghinstallation.New(http.DefaultTransport, t.IntegrationID, int(id), []byte(t.APIKey))
	if err != nil {
		return nil, err
	}

	return github.NewClient(&http.Client{Transport: tr}), nil
}
