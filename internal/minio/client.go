package minio

import (
	"context"
	"fmt"
	"os"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	c              *minio.Client
	internalEP     string
	publicEndpoint string
	secure         bool
}

var instance *Client

func New() (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	internalEP := os.Getenv("MINIO_ENDPOINT")      // lpr-minio:9000
	publicEP := os.Getenv("MINIO_PUBLIC_ENDPOINT") // 192.168.7.200:9000
	access := os.Getenv("MINIO_ACCESS_KEY")
	secret := os.Getenv("MINIO_SECRET_KEY")
	useSSL := strings.ToLower(os.Getenv("MINIO_USE_SSL")) == "true"

	if internalEP == "" || publicEP == "" {
		return nil, fmt.Errorf("MINIO_ENDPOINT or MINIO_PUBLIC_ENDPOINT missing")
	}

	client, err := minio.New(internalEP, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	instance = &Client{
		c:              client,
		internalEP:     internalEP,
		publicEndpoint: publicEP,
		secure:         useSSL,
	}
	return instance, nil
}

func (m *Client) UploadFile(
	ctx context.Context,
	bucket, objectName, filePath string,
) (string, error) {

	_, err := m.c.FPutObject(
		ctx,
		bucket,
		objectName,
		filePath,
		minio.PutObjectOptions{},
	)
	if err != nil {
		return "", err
	}

	scheme := "http"
	if m.secure {
		scheme = "https"
	}

	return fmt.Sprintf(
		"%s://%s/%s/%s",
		scheme,
		m.publicEndpoint,
		bucket,
		objectName,
	), nil
}
