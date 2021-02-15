package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/IcoBoyanov/lazy-spots/repository"
	"github.com/IcoBoyanov/lazy-spots/repository/miniocli"
	"github.com/IcoBoyanov/lazy-spots/server"
	"github.com/IcoBoyanov/lazy-spots/strava"
	"github.com/julienschmidt/httprouter"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

const (
	MinioAccessKeyEnv = "MINIO_ACCESS_KEY"
	MinioSecretEnv    = "MINIO_SECRET"

	// Local minio instance endpoint
	Endpoint = "172.17.0.2:9000"
	UseSSL   = false

	ServerURL = "http://localhost"
)

var (
	minioAccessKey string
	minioSecret    string
	servePort      string
	repo           repository.Repository
	requestServer  *server.RequestServer
	logger         *log.Logger
)

func main() {

	minioAccessKey = os.Getenv(MinioAccessKeyEnv)
	minioSecret = os.Getenv(MinioSecretEnv)

	flag.StringVar(&servePort, "port", ":8888", "serve port")
	flag.Parse()

	// Initialize minio client object.
	minioClient, err := minio.New(Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(minioAccessKey, minioSecret, ""),
		Secure: UseSSL,
	})
	if err != nil {
		log.Fatalln(err)
	}
	repo = miniocli.New(log.New(log.Writer(), "storage: ", log.LstdFlags), minioClient)
	client, err := strava.NewStravaService(ServerURL + servePort + "/callback")
	if err != nil {
		fmt.Fprintf(os.Stderr, "could not create strava client: %v", err)
	}
	requestServer = server.NewRequestServer(repo, client)

	router := httprouter.New()
	router.GET("/", Home)
	router.GET("/login", requestServer.Login)
	router.GET("/callback", requestServer.Callback)
	router.GET("/athlete", requestServer.GetAthleteData)
	router.GET("/collect", requestServer.CollectAthleteActivities)
	router.GET("/places", requestServer.GetMapPlaces)
	router.ServeFiles("/static/*filepath", http.Dir("./web"))

	if err := http.ListenAndServe(servePort, router); err != nil {
		fmt.Fprintf(os.Stderr, "server is down: %v", err)
		os.Exit(1)
	}

}

func Home(w http.ResponseWriter, req *http.Request, _ httprouter.Params) {
	if !requestServer.Authenticated() {
		body := `<html><body><a href="/login">Login with Strava</a></body></html>`
		fmt.Fprintf(w, "%s", body)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}
	w.WriteHeader(http.StatusOK)

	html := `
	<html><body>
		<a href="/athlete">get Athlete data</a>
		</br>
		<a href="/collect">collect</a>
		</br>
		<a href="/places">places</a>	
		<a href="/map">go to map</a>	
	</body></html>
	`
	fmt.Fprintf(w, "%s", html)
}
