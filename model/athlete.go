package model

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
)

type Athlete struct {
	ID          int    `json:"id"`
	Name        string `json:"firstname"`
	Lastname    string `json:"lastname"`
	Sex         string `json:"sex,omitempty"`
	Profile     string `json:"profile_medium,omitempty"`
	Acitivities string `json:"activities,omitempty"`
}

func NewAthlete(r io.Reader) (*Athlete, error) {
	var a Athlete
	decoder := json.NewDecoder(r)

	err := decoder.Decode(&a)
	if err != nil {
		return nil, fmt.Errorf("could not parse athlete data: %v", err)
	}

	return &a, nil
}

func (a *Athlete) Write(out io.Writer) error {
	return json.NewEncoder(out).Encode(a)
}

func (a *Athlete) Reader() io.Reader {
	content, _ := json.Marshal(a)
	return bytes.NewReader(content)
}
