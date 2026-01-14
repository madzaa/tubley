package config

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/madza/tubley/internal/database"
)

type ApiConfig struct {
	Db               database.Client
	JwtSecret        string
	Platform         string
	FilepathRoot     string
	AssetsRoot       string
	S3Bucket         string
	S3Region         string
	S3CfDistribution string
	Port             string
	S3Client         *s3.Client
}

type Thumbnail struct {
	Data      []byte
	MediaType string
}

var VideoThumbnails = map[uuid.UUID]Thumbnail{}

func NewApiConfig() *ApiConfig {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("unable to load .env file")
	}

	pathToDB := os.Getenv("DB_PATH")
	if pathToDB == "" {
		log.Fatal("DB_URL must be set")
	}

	db, err := database.NewClient(pathToDB)
	if err != nil {
		log.Fatalf("Couldn't connect to database: %v", err)
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable is not set")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM environment variable is not set")
	}

	filepathRoot := os.Getenv("FILEPATH_ROOT")
	if filepathRoot == "" {
		log.Fatal("FILEPATH_ROOT environment variable is not set")
	}

	assetsRoot := os.Getenv("ASSETS_ROOT")
	if assetsRoot == "" {
		log.Fatal("ASSETS_ROOT environment variable is not set")
	}

	s3Bucket := os.Getenv("S3_BUCKET")
	if s3Bucket == "" {
		log.Fatal("S3_BUCKET environment variable is not set")
	}

	s3Region := os.Getenv("S3_REGION")
	if s3Region == "" {
		log.Fatal("S3_REGION environment variable is not set")
	}

	s3CfDistribution := os.Getenv("S3_CF_DISTRO")
	if s3CfDistribution == "" {
		log.Fatal("S3_CF_DISTRO environment variable is not set")
	}

	port := os.Getenv("PORT")
	if port == "" {
		log.Fatal("PORT environment variable is not set")
	}

	s3Config, err := config.LoadDefaultConfig(context.Background(), config.WithRegion(s3Region))
	if err != nil {
		log.Fatal("Unable to get default AWS config")
	}

	s3Client := s3.NewFromConfig(s3Config)

	cfg := ApiConfig{
		Db:               db,
		JwtSecret:        jwtSecret,
		Platform:         platform,
		FilepathRoot:     filepathRoot,
		AssetsRoot:       assetsRoot,
		S3Bucket:         s3Bucket,
		S3Region:         s3Region,
		S3CfDistribution: s3CfDistribution,
		Port:             port,
		S3Client:         s3Client,
	}
	return &cfg
}
