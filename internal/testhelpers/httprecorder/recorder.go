package httprecorder

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/dnaeon/go-vcr/cassette"
	"github.com/dnaeon/go-vcr/recorder"
	"github.com/scaleway/scaleway-sdk-go/scw"
)

const (
	envUpdateTestdata  = "UPDATE_TESTDATA"
	envDisableTestdata = "DISABLE_TESTDATA"
)

func UpdateTestdata() bool {
	return os.Getenv(envUpdateTestdata) == "true"
}

func DisableTestData() bool {
	return os.Getenv(envDisableTestdata) == "true"
}

func CreateRecordedScwClient(cassetteName string) (*scw.Client, *recorder.Recorder, error) {
	recorderMode := recorder.ModeReplaying
	if UpdateTestdata() {
		recorderMode = recorder.ModeRecording
	}
	if DisableTestData() {
		recorderMode = recorder.ModeDisabled
	}

	r, err := recorder.NewAsMode(fmt.Sprintf("testdata/%s", cassetteName), recorderMode, nil)
	if err != nil {
		return nil, nil, err
	}

	// Do not record secret keys
	r.AddFilter(func(i *cassette.Interaction) error {
		delete(i.Request.Headers, "x-auth-token")
		delete(i.Request.Headers, "X-Auth-Token")

		// panics if the secret key is found elsewhere
		if UpdateTestdata() && !DisableTestData() {
			if i != nil && strings.Contains(fmt.Sprintf("%v", *i), os.Getenv(scw.ScwSecretKeyEnv)) {
				panic(errors.New("found secret key in cassette"))
			}
		}

		return nil
	})

	// Create new http.Client where transport is the recorder
	httpClient := &http.Client{Transport: r}

	var client *scw.Client

	if UpdateTestdata() || DisableTestData() {
		// When updating the recoreded test requests, we need the access key and secret key.
		client, err = scw.NewClient(
			scw.WithHTTPClient(httpClient),
			scw.WithEnv(),
			scw.WithDefaultRegion(scw.RegionFrPar),
			scw.WithDefaultZone(scw.ZoneFrPar1),
		)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// No need for auth when using cassette
		client, err = scw.NewClient(
			scw.WithHTTPClient(httpClient),
			scw.WithDefaultRegion(scw.RegionFrPar),
			scw.WithDefaultZone(scw.ZoneFrPar1),
		)
		if err != nil {
			return nil, nil, err
		}
	}
	return client, r, nil
}
