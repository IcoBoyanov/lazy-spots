package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type ActivityStream struct {
	Streams []StreamData `json:"streams"`
}

// Streams set one of {ditance, moving, latlng}
// https://developers.strava.com/docs/reference/#api-models-StreamSet
type StreamData struct {
	Type       string        `json:"type"`
	Resolution string        `json:"resolution"`
	Size       int           `json:"original_size"`
	Data       []interface{} `json:"data"`
}

func NewActivityStream(r io.Reader) (*ActivityStream, error) {
	var as ActivityStream
	decoder := json.NewDecoder(r)

	err := decoder.Decode(&as)
	if err != nil {
		return nil, fmt.Errorf("could not parse activity stream: %v", err)

	}
	return &as, nil
}

func (as *ActivityStream) Write(out io.Writer) error {
	return json.NewEncoder(out).Encode(as)
}

func (a *ActivityStream) Reader() io.Reader {
	content, _ := json.Marshal(a)
	return bytes.NewReader(content)
}
