// Copyright 2016 Canonical Ltd. All rights reserved.

package cmd

import (
	"net/http"
	"os"

	"github.com/juju/errors"
	"gopkg.in/juju/charm.v6-unstable"
	"gopkg.in/juju/charmrepo.v2-unstable"
	"gopkg.in/macaroon-bakery.v1/httpbakery"
)

var defaultCharmStoreURL = "http://api.jujucharms.com/charmstore"

func init() {
	if csURL := os.Getenv("CSURL"); csURL != "" {
		defaultCharmStoreURL = csURL
	}
}

// charmResolver interface defines the functionality to resolve a charm URL.
type charmResolver interface {
	// Resolve resolves the charm URL.
	Resolve(client *http.Client, charmURL string) (string, error)
	// Metrics returns a slice of metric names that the
	// charm collects.
	Metrics(client *http.Client, charmURL string) ([]string, error)
}

// charmStoreResolver implements the charmResolver interface.
type charmStoreResolver struct {
	csURL string
}

// NewcharmStoreResolver creates a new charm store resolver.
func NewCharmStoreResolver() *charmStoreResolver {
	return &charmStoreResolver{
		csURL: defaultCharmStoreURL,
	}
}

// Resolve implements the charmResolver interface.
func (r *charmStoreResolver) Resolve(client *http.Client, charmURL string) (string, error) {
	repo := charmrepo.NewCharmStore(charmrepo.NewCharmStoreParams{
		URL:          r.csURL,
		HTTPClient:   client,
		VisitWebPage: httpbakery.OpenWebBrowser,
	})

	curl, err := charm.ParseURL(charmURL)
	if err != nil {
		return "", errors.Annotate(err, "could not parse charm url")
	}
	// ignore local charm urls
	if curl.Schema == "local" {
		return charmURL, nil
	}
	resolvedURL, _, err := repo.Resolve(curl)
	if err != nil {
		return "", errors.Trace(err)
	}
	return resolvedURL.String(), nil
}

func (r *charmStoreResolver) Metrics(client *http.Client, charmURL string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return []string{}, err
	}
	repo := charmrepo.NewCharmStore(charmrepo.NewCharmStoreParams{
		URL:          r.csURL,
		HTTPClient:   client,
		VisitWebPage: httpbakery.OpenWebBrowser,
	})
	curl, err := charm.ParseURL(charmURL)
	if err != nil {
		return fail(errors.Annotate(err, "could not parse charm url"))
	}
	csClient := repo.Client()
	var result struct {
		CharmMetrics charm.Metrics
	}
	_, err = csClient.Meta(curl, &result)
	if err != nil {
		if errors.IsNotFound(err) {
			return fail(nil)
		}
		return fail(errors.Trace(err))
	}
	metrics := []string{}
	for key, _ := range result.CharmMetrics.Metrics {
		metrics = append(metrics, key)
	}
	return metrics, nil
}
