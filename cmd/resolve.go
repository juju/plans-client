// Copyright 2017 Canonical Ltd.
// Licensed under the GPLv3, see LICENCE file for details.

package cmd

import (
	"os"

	"github.com/juju/charm/v8"
	"github.com/juju/charmrepo/v6"
	"github.com/juju/errors"
	"gopkg.in/macaroon-bakery.v2/httpbakery"
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
	Resolve(*httpbakery.Client, string) (string, error)
	// Metrics returns a slice of metric names that the
	// charm collects.
	Metrics(*httpbakery.Client, string) ([]string, error)
}

// charmStoreResolver implements the charmResolver interface.
type charmStoreResolver struct {
	csURL string
}

// NewCharmStoreResolver creates a new charm store resolver.
func NewCharmStoreResolver() *charmStoreResolver {
	return &charmStoreResolver{
		csURL: defaultCharmStoreURL,
	}
}

// Resolve implements the charmResolver interface.
func (r *charmStoreResolver) Resolve(client *httpbakery.Client, charmURL string) (string, error) {
	repo := charmrepo.NewCharmStore(charmrepo.NewCharmStoreParams{
		URL:          r.csURL,
		BakeryClient: client,
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

func (r *charmStoreResolver) Metrics(client *httpbakery.Client, charmURL string) ([]string, error) {
	fail := func(err error) ([]string, error) {
		return []string{}, err
	}
	repo := charmrepo.NewCharmStore(charmrepo.NewCharmStoreParams{
		URL:          r.csURL,
		BakeryClient: client,
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
