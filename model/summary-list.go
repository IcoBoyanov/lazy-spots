package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

// Types of activity streams collected
// https://developers.strava.com/docs/reference/#api-models-StreamSet
const ActivityStreamTypes string = "latlng,moving"

// UpperBoundSofia used to filter activities in this region
var SofiaBorder = struct {
	lowerLeftLat  float64
	lowerLeftLng  float64
	upperRightLat float64
	upperRightLng float64
}{42.656182, 23.102273, 42.753063, 23.572252}

type ActivitySummaryList struct {
	SumamryList []activitySummary `json:"summary-list"`
}

// ActivitySummary https://developers.strava.com/docs/reference/#api-models-SummaryActivity
type activitySummary struct {
	ID    int        `json:"id"`
	Name  string     `json:"name"`
	Start [2]float64 `json:"start_latlng"`
	End   [2]float64 `json:"end_latlng"`
}

// NewActivitySummaryList reads JSON data from an io.Reader and returns a filtered []ActivitySummary
func NewActivitySummaryList(input io.Reader) (*ActivitySummaryList, error) {
	var l ActivitySummaryList
	decoder := json.NewDecoder(input)

	err := decoder.Decode(&l)
	if err != nil {
		return nil, fmt.Errorf("could not parse activity list: %v", err)
	}
	filtered := filter(l, inBound)

	return &filtered, nil
}

func inBound(as activitySummary) bool {
	return as.End[0] > SofiaBorder.lowerLeftLat &&
		as.End[1] > SofiaBorder.lowerLeftLng &&
		as.End[0] < SofiaBorder.upperRightLat &&
		as.End[1] < SofiaBorder.upperRightLng
}

func filter(as ActivitySummaryList, test func(activitySummary) bool) (res ActivitySummaryList) {
	for _, s := range as.SumamryList {
		if test(s) {
			res.SumamryList = append(res.SumamryList, s)
		}
	}
	return res
}

func (as *ActivitySummaryList) Write(out io.Writer) error {
	return json.NewEncoder(out).Encode(as)
}

func (a *ActivitySummaryList) Reader() io.Reader {
	content, _ := json.Marshal(a)
	return bytes.NewReader(content)
}
