// Copyright 2015 Canonical Ltd. All rights reserved.

package commands

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/juju/errors"
	"gopkg.in/macaroon-bakery.v1/httpbakery"

	"github.com/juju/juju/api"
	"github.com/juju/juju/api/charms"
)

var (
	openWebBrowser = func(_ *url.URL) error { return nil }
)

type metricRegistrationPost struct {
	EnvironmentUUID string `json:"env-uuid"`
	CharmURL        string `json:"charm-url"`
	ServiceName     string `json:"service-name"`
}

var registerMeteredCharm = func(registrationURL string, state api.Connection, client *http.Client, charmURL string, serviceName, environmentUUID string) error {
	charmsClient := charms.NewClient(state)
	defer charmsClient.Close()
	metered, err := charmsClient.IsMetered(charmURL)
	if err != nil {
		return err
	}
	if metered {
		bakeryClient := httpbakery.Client{Client: client, VisitWebPage: httpbakery.OpenWebBrowser}
		credentials, err := registerMetrics(registrationURL, environmentUUID, charmURL, serviceName, &bakeryClient)
		if err != nil {
			logger.Infof("failed to register metrics: %v", err)
			return err
		}

		api, cerr := getMetricCredentialsAPI(state)
		if cerr != nil {
			logger.Infof("failed to get the metrics credentials setter: %v", cerr)
		}
		err = api.SetMetricCredentials(serviceName, credentials)
		if err != nil {
			logger.Infof("failed to set metric credentials: %v", err)
			return err
		}
		api.Close()
	}
	return nil
}

func registerMetrics(registrationURL, environmentUUID, charmURL, serviceName string, client *httpbakery.Client) ([]byte, error) {
	if registrationURL == "" {
		return nil, errors.Errorf("no metric registration url is specified")
	}
	registerURL, err := url.Parse(registrationURL)
	if err != nil {
		return nil, errors.Trace(err)
	}

	registrationPost := metricRegistrationPost{
		EnvironmentUUID: environmentUUID,
		CharmURL:        charmURL,
		ServiceName:     serviceName,
	}

	buff := &bytes.Buffer{}
	encoder := json.NewEncoder(buff)
	err = encoder.Encode(registrationPost)
	if err != nil {
		return nil, errors.Trace(err)
	}

	req, err := http.NewRequest("POST", registerURL.String(), nil)
	if err != nil {
		return nil, errors.Trace(err)
	}
	req.Header.Set("Content-Type", "application/json")

	response, err := client.DoWithBody(req, bytes.NewReader(buff.Bytes()))
	if err != nil {
		return nil, errors.Trace(err)
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return nil, errors.Errorf("failed to register metrics: http response is %d", response.StatusCode)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, errors.Annotatef(err, "failed to read the response")
	}
	return b, nil
}
