package strava

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"

	"github.com/IcoBoyanov/lazy-spots/model"
	"golang.org/x/oauth2"
)

const (
	StravaAPIEndpoint = "https://www.strava.com/api/v3/"
	StravaAuthURL     = "https://www.strava.com/oauth/authorize"
	StravaTokenURL    = "https://www.strava.com/oauth/token"
	ClientIDEnv       = "CLIENT_ID"
	ClientSecretEnv   = "CLIENT_SECRET"
)

type StravaService interface {
	// Client() *http.Client
	GetAuthURL() string
	Authenticate(context.Context, *url.URL) error
	IsTokenValid() bool
	GetAthleteData() (*model.Athlete, error)
	GetActivitySumamryList() (*model.ActivitySummaryList, error)
	GetRide(id string) (*model.ActivityStream, error)
}

type stravaService struct {
	client *http.Client
	config *oauth2.Config
	token  *oauth2.Token
	state  string
}

var configLock sync.Mutex
var stravaServiceInstance *stravaService

func NewStravaService(callbackURL string) (StravaService, error) {
	configLock.Lock()
	defer configLock.Unlock()

	if stravaServiceInstance == nil {
		return newConfig(callbackURL)
	}
	return stravaServiceInstance, nil
}

func newConfig(callbackURL string) (StravaService, error) {
	clientID, ok := os.LookupEnv(ClientIDEnv)
	if !ok {
		return nil, fmt.Errorf("missing value for '%s'", ClientIDEnv)
	}

	clientSecret, ok := os.LookupEnv(ClientSecretEnv)
	if !ok {
		return nil, fmt.Errorf("missing value for '%s'", ClientSecretEnv)
	}

	return &stravaService{
		config: &oauth2.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			Scopes:       []string{"activity:read"},
			RedirectURL:  callbackURL,
			Endpoint: oauth2.Endpoint{
				AuthURL:  StravaAuthURL,
				TokenURL: StravaTokenURL,
			},
		},
		state: "state", // random per each client?
	}, nil
}

func (s *stravaService) Authenticate(ctx context.Context, callbackURL *url.URL) error {
	state := callbackURL.Query().Get("state")
	if s.state != state {
		return fmt.Errorf("state does not match: expected '%s', but got '%s'", s.state, state)

	}
	code := callbackURL.Query().Get("code")
	if code == "" {
		return fmt.Errorf("code is missing")

	}

	configLock.Lock()
	defer configLock.Unlock()
	var err error
	s.token, err = s.config.Exchange(ctx, code)
	if err != nil {
		return fmt.Errorf("could not fetch token: %v", err)
	}
	s.client = s.config.Client(ctx, s.token)
	return nil
}

func (s *stravaService) GetAuthURL() string {
	return s.config.AuthCodeURL(s.state, oauth2.AccessTypeOffline)
}

func (s *stravaService) Client() *http.Client {
	return s.client
}

func (s *stravaService) IsTokenValid() bool {
	return s.token.Valid()
}

func (s *stravaService) GetAthleteData() (*model.Athlete, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/%s", StravaAPIEndpoint, "/athlete"))
	if err != nil {
		return nil, fmt.Errorf("could not get athlete data: %v", err)
	}
	defer resp.Body.Close()
	athlete, err := model.NewAthlete(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("could not parse body: %v", err)
	}
	return athlete, nil
}

func (s *stravaService) GetActivitySumamryList() (*model.ActivitySummaryList, error) {
	resp, err := s.client.Get(fmt.Sprintf("%s/%s", StravaAPIEndpoint, "/athlete/activities"))
	if err != nil {
		return nil, fmt.Errorf("could not get athlete's activity stream: %v", err)
	}
	defer resp.Body.Close()
	var open_buff bytes.Buffer
	open_buff.WriteString(`{"summary-list":`)
	var close_buff bytes.Buffer
	close_buff.WriteString(`}`)
	return model.NewActivitySummaryList(io.MultiReader(&open_buff, resp.Body, &close_buff))
}

func (s *stravaService) GetRide(id string) (*model.ActivityStream, error) {
	respMoving, err := s.client.Get(fmt.Sprintf("%s/activities/%s/streams?keys=%s&key_by_type=", StravaAPIEndpoint, id, "moving"))
	if err != nil {
		return nil, fmt.Errorf("could not get athlete's moving activity stream: %v", err)
	}
	defer respMoving.Body.Close()

	respLatLng, err := s.client.Get(fmt.Sprintf("%s/activities/%s/streams?keys=%s&key_by_type=", StravaAPIEndpoint, id, "latlng"))
	if err != nil {
		return nil, fmt.Errorf("could not get athlete's moving activity stream: %v", err)
	}
	defer respLatLng.Body.Close()

	var open_buff bytes.Buffer
	open_buff.WriteString(`{"streams":`)
	var close_buff bytes.Buffer
	close_buff.WriteString(`}`)

	moving, err := model.NewActivityStream(io.MultiReader(&open_buff, respMoving.Body, &close_buff))

	open_buff.WriteString(`{"streams":`)
	close_buff.WriteString(`}`)
	latlng, err := model.NewActivityStream(io.MultiReader(&open_buff, respLatLng.Body, &close_buff))
	return &model.ActivityStream{
		Streams: []model.StreamData{
			moving.Streams[0],
			latlng.Streams[0],
		},
	}, nil
}

func ActivityStreamURL(activity string, types []string) (string, error) {
	activityStreamURL, err := url.Parse(StravaAPIEndpoint + "activities/" + activity + "/streams")
	if err != nil {
		return "", fmt.Errorf("could not create activity url: %v", err)
	}

	query := activityStreamURL.Query()
	for _, t := range types {
		query.Add("keys", t)
	}
	activityStreamURL.RawQuery = query.Encode()
	return activityStreamURL.String(), nil
}

func ListActivitiesURL(per_page int, page int, before time.Time, after time.Time) (string, error) {
	// TODO fetch all pages
	activityListURL, err := url.Parse(StravaAPIEndpoint + "athlete/activities")
	if err != nil {
		return "", fmt.Errorf("could not create activity list url: %v", err)
	}

	query := activityListURL.Query()
	query.Set("per_page", fmt.Sprintf("%d", per_page))
	activityListURL.RawQuery = query.Encode()
	return activityListURL.String(), nil
}

func AthleteURL() (string, error) {
	athleteURL, err := url.Parse(StravaAPIEndpoint + "athlete")
	if err != nil {
		return "", fmt.Errorf("could not create activity list url: %v", err)
	}
	return athleteURL.String(), nil
}
