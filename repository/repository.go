package repository

import (
	"io"

	"github.com/IcoBoyanov/lazy-spots/model"
)

type Repository interface {
	PostRide(ride string, data io.Reader) error
	PostMapData(ride string, data io.Reader) error
	GetAllMapPlaces() (*model.SpotList, error)
	PostAthlete(athlete string, data io.Reader) error
	RemoveRide(ride string) error
	RemoveAthlete(athlete string) error
	GetRide(io.Writer, string) (bool, error)
	GetAthlete(io.Writer, string) (bool, error)
}
