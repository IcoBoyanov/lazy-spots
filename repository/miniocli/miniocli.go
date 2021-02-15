package miniocli

import (
	"context"
	"fmt"
	"io"
	"log"

	"github.com/IcoBoyanov/lazy-spots/model"
	"github.com/IcoBoyanov/lazy-spots/repository"
	"github.com/minio/minio-go/v7"
)

const RidesBucketName = "rides"
const AthletesBucketName = "athletes"
const MapDataBucketName = "maps"

var (
	minioClient *minio.Client
	logger      *log.Logger
)

type MinioStorageClient struct{}

func New(logger *log.Logger, client *minio.Client) repository.Repository {
	minioClient = client
	logger.Printf("%#v\n", minioClient) // minioClient is now setup
	return &MinioStorageClient{}
}

func (m *MinioStorageClient) PostRide(ride string, data io.Reader) error {
	// Create a bucket at region 'us-east-1' with object locking enabled.
	err := minioClient.MakeBucket(context.Background(), RidesBucketName, minio.MakeBucketOptions{})
	if err == nil {
		fmt.Printf("Successfully created bucket '%s'.\n", RidesBucketName)
	}
	if err != nil {
		if err.(minio.ErrorResponse).StatusCode != 409 {
			fmt.Println(err)
			return err
		}
	}

	uploadInfo, err := minioClient.PutObject(context.Background(), RidesBucketName, ride, data, -1, minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
	return nil
}

func (m *MinioStorageClient) PostMapData(ride string, data io.Reader) error {
	// Create a bucket at region 'us-east-1' with object locking enabled.
	err := minioClient.MakeBucket(context.Background(), MapDataBucketName, minio.MakeBucketOptions{})
	if err == nil {
		fmt.Printf("Successfully created bucket '%s'.\n", MapDataBucketName)
	}
	if err != nil {
		if err.(minio.ErrorResponse).StatusCode != 409 {
			fmt.Println(err)
			return err
		}
	}

	uploadInfo, err := minioClient.PutObject(context.Background(), MapDataBucketName, ride, data, -1, minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
	return nil
}

func (m *MinioStorageClient) PostAthlete(athlete string, data io.Reader) error {
	// Create a bucket at region 'us-east-1' with object locking enabled.
	err := minioClient.MakeBucket(context.Background(), AthletesBucketName, minio.MakeBucketOptions{})
	if err == nil {
		fmt.Printf("Successfully created bucket '%s'.\n", AthletesBucketName)
	}
	if err != nil {
		if err.(minio.ErrorResponse).StatusCode != 409 {
			fmt.Println(err)
			return err
		}
	}

	uploadInfo, err := minioClient.PutObject(context.Background(), AthletesBucketName, athlete, data, -1, minio.PutObjectOptions{ContentType: "application/json"})
	if err != nil {
		fmt.Println(err)
		return err
	}
	fmt.Println("Successfully uploaded bytes: ", uploadInfo)
	return nil
}

func (m *MinioStorageClient) RemoveRide(ride string) error    { return nil }
func (m *MinioStorageClient) RemoveAthlete(ride string) error { return nil }

func (m *MinioStorageClient) GetRide(out io.Writer, ride string) (bool, error) {
	_, err := minioClient.StatObject(context.Background(), RidesBucketName, ride, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("could not stat object from minio: %v", err)
	}
	return true, m.writeObject(out, RidesBucketName, ride)
}

func (m *MinioStorageClient) GetAllMapPlaces() (*model.SpotList, error) {
	places := model.SpotList{}
	places.Data = make([]model.Spot, 0)
	objects := minioClient.ListObjects(context.Background(), MapDataBucketName, minio.ListObjectsOptions{})
	for o := range objects {
		if o.Key == "" {
			continue
		}
		data, err := m.getObject(MapDataBucketName, o.Key)
		if err != nil {
			fmt.Printf("could not get object from repo: %v", err)
			continue
		}

		sl, err := model.NewSpotListFromJSON(data)
		if err != nil {
			fmt.Printf("could not parse object: %v", err)
			continue
		}

		places.Data = append(places.Data, sl.Data...)
	}

	return &places, nil
}

func (m *MinioStorageClient) GetAthlete(out io.Writer, athlete string) (bool, error) {
	_, err := minioClient.StatObject(context.Background(), AthletesBucketName, athlete, minio.GetObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("could not stat object from minio: %v", err)
	}
	return true, m.writeObject(out, AthletesBucketName, athlete)
}

func (m *MinioStorageClient) writeObject(out io.Writer, bucket, object string) error {
	data, err := minioClient.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return err
	}

	if _, err = io.Copy(out, data); err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func (m *MinioStorageClient) getObject(bucket, object string) (*minio.Object, error) {
	data, err := minioClient.GetObject(context.Background(), bucket, object, minio.GetObjectOptions{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	return data, nil
}
