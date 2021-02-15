package server

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/IcoBoyanov/lazy-spots/model"
	"github.com/IcoBoyanov/lazy-spots/repository"
	"github.com/IcoBoyanov/lazy-spots/strava"
	"github.com/julienschmidt/httprouter"
)

const CallbackTimeout = 5 * time.Second

const HomeRoute = "/"

// type StravaRequestURL interface {
// 	ActivityStreamURL(string, []string) (string, error)
// 	ListActivitiesURL(max int, page int, before time.Time, after time.Time) (string, error)
// 	AthleteURL() (string, error)
// }

type RequestServer struct {
	strava strava.StravaService
	repo   repository.Repository

	athleteID string
	// logger        *log.Logger
}

func NewRequestServer(repo repository.Repository, strava strava.StravaService) *RequestServer {
	return &RequestServer{
		strava: strava,
		repo:   repo,
	}
}

func (rh *RequestServer) Authenticated() bool {
	return rh.strava.IsTokenValid()
}

// func (rh *RequestServer) Client() *http.Client { return rh.strava.Client() }

func (rh *RequestServer) Login(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	authURL := rh.strava.GetAuthURL()
	http.Redirect(w, req, authURL, http.StatusTemporaryRedirect)
}

func (rh *RequestServer) Callback(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	err := rh.strava.Authenticate(req.Context(), req.URL)
	if err != nil {
		return
	}
	http.Redirect(w, req, HomeRoute, http.StatusTemporaryRedirect)
}

func (rh *RequestServer) GetAthleteData(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if rh.athleteID != "" {
		if ok, err := rh.repo.GetAthlete(w, rh.athleteID); err == nil && ok {
			return
		}
	}

	var athlete *model.Athlete
	athlete, err := rh.strava.GetAthleteData()
	if err != nil {
		fmt.Fprintf(w, "something went wrong: %v", err)
	}
	rh.athleteID = strconv.Itoa(athlete.ID)
	rh.repo.PostAthlete(rh.athleteID, athlete.Reader())
	athlete.Write(w)
}

func (rh *RequestServer) CollectAthleteActivities(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	sl, err := rh.strava.GetActivitySumamryList()
	if err != nil {
		fmt.Fprintf(w, "something went wrong: %v", err)
	}
	var activityID string
	for _, sum := range sl.SumamryList {
		fmt.Fprintf(w, "\nfetching activity '%s'..", sum.Name)
		activityID = strconv.Itoa(sum.ID)
		stream, err := rh.strava.GetRide(activityID)
		if err != nil {
			fmt.Fprintf(w, "\nerror fetching activity '%s': %v", sum.Name, err)
		}
		err = rh.repo.PostRide(activityID, stream.Reader())
		if err != nil {
			fmt.Fprintf(w, "\nerror storing activity '%s': %v\n\n", sum.Name, err)
			continue
		}

		// Post spots
		sl := model.NewSpotList(stream)
		err = rh.repo.PostMapData(activityID, sl.Reader())
		if err != nil {
			fmt.Fprintf(w, "\n could not store activity '%s' palces: %v\n\n", sum.Name, err)
			continue
		}
		fmt.Fprintf(w, "\ncompleted fetching activity '%s' \n\n", sum.Name)
	}

	fmt.Fprintf(w, "Ready colelcting activities")
}

func (rh *RequestServer) GetMapPlaces(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	// if req.Method == "OPTIONS" {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json")
	// 	return
	// }
	spots, err := rh.repo.GetAllMapPlaces()
	if err != nil {
		fmt.Fprint(w, "failed fetching map places: %v", err)
	}
	spots.Write(w)
}

func (rh *RequestServer) LoadMap(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	http.FileServer(http.Dir("./web"))
	return
}
