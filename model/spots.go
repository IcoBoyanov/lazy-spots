package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Spot struct {
	Lat float64 `json:"lat"`
	Lng float64 `json:"lng"`
}

type SpotList struct {
	Data []Spot `json:"data"`
}

const MovingStreamIdx = 0
const LatLngStreamIdx = 1

func NewSpotList(activities ...*ActivityStream) *SpotList {
	var result SpotList
	result.Data = make([]Spot, 0)
	for _, activity := range activities {
		for i := range activity.Streams[0].Data {

			if activity.Streams[MovingStreamIdx].Data[i].(bool) {
				// This is ugly, I know
				lat := activity.Streams[LatLngStreamIdx].Data[i].([]interface{})[0].(float64)
				lng := activity.Streams[LatLngStreamIdx].Data[i].([]interface{})[1].(float64)
				result.Data = append(result.Data, Spot{lat, lng})
			}
		}
	}
	return &result
}

func NewSpotListFromJSON(input io.Reader) (*SpotList, error) {
	var sl SpotList

	err := json.NewDecoder(input).Decode(&sl)
	if err != nil {
		return nil, fmt.Errorf("could not parse data: %v", err)
	}

	return &sl, nil
}

func (s *SpotList) Write(out io.Writer) error {
	return json.NewEncoder(out).Encode(s)
}

func (s *SpotList) Reader() io.Reader {
	content, _ := json.Marshal(s)
	return bytes.NewReader(content)
}
