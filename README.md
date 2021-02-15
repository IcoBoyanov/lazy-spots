# Lazy Spots
A combination of a simple Strava API and minio client ehich will colect data from your recent activities and display the plaes where you most often stop for a rest.

## Prerequisite 
 - [Docker](https://www.docker.com/get-started)
 - [Go](https://golang.org/)
 - Strava API credentials [docs](https://developers.strava.com/)
 - Change all necessary values from `setup.sh`

## Setup
```sh
# run a local minio instance
docker run -p 9000:9000 minio/minio server /data
```

## Build and Run
```sh
go build

./setup.sh
lazy-spots
```

## Usage
`lazy-spots` export several endpoints:

| route | method | response | infog |
| --- | -- -| --- | --- |
|`/login` | GET | - | redirects to the strava authentication endpoint |
|`/athlete` | GET | [AthleteObject](https://developers.strava.com/docs/reference/#api-Athletes) | fetches your profile data from strava |
|`/collect` | GET | - | collects all strava activities in minio |
|`/static` | GET | static html page | render collected _lazy spots_ |