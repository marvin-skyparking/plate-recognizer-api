package minio

import (
	"context"
	"fmt"
	"io"
	"mime"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	minio "github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Client struct {
	c        *minio.Client
	endpoint string
	secure   bool
}

var instance *Client

func New() (*Client, error) {
	if instance != nil {
		return instance, nil
	}

	endpoint := os.Getenv("MINIO_ENDPOINT")
	access := os.Getenv("MINIO_ACCESS_KEY")
	secret := os.Getenv("MINIO_SECRET_KEY")
	bucket := os.Getenv("MINIO_BUCKET_IMAGE_LPR")
	useSSL := strings.ToLower(os.Getenv("MINIO_USE_SSL")) == "true"

	if endpoint == "" || access == "" || secret == "" || bucket == "" {
		return nil, fmt.Errorf("missing MinIO configuration; set MINIO_ENDPOINT, MINIO_ACCESS_KEY, MINIO_SECRET_KEY, MINIO_BUCKET_IMAGE_LPR")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(access, secret, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, err
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}
	if !exists {
		if err := client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, err
		}
	}

	instance = &Client{c: client, endpoint: endpoint, secure: useSSL}
	return instance, nil
}

// UploadFile uploads a file at local path to the given bucket and object name.
// It returns an accessible URL for the uploaded object.
func (m *Client) UploadFile(ctx context.Context, bucket, objectName, filePath string) (string, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer f.Close()

	stat, err := f.Stat()
	if err != nil {
		return "", err
	}

	contentType := mime.TypeByExtension(strings.ToLower(filepath.Ext(filePath)))
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	// Put the object
	_, err = m.c.PutObject(ctx, bucket, objectName, f, stat.Size(), minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}

	// Construct URL
	scheme := "https"
	if !m.secure {
		scheme = "http"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   m.endpoint,
		Path:   fmt.Sprintf("/%s/%s", bucket, objectName),
	}
	return u.String(), nil
}

// UploadReader uploads from an io.Reader with known size.
func (m *Client) UploadReader(ctx context.Context, bucket, objectName string, r io.Reader, size int64, contentType string) (string, error) {
	_, err := m.c.PutObject(ctx, bucket, objectName, r, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return "", err
	}
	scheme := "https"
	if !m.secure {
		scheme = "http"
	}
	u := url.URL{
		Scheme: scheme,
		Host:   m.endpoint,
		Path:   fmt.Sprintf("/%s/%s", bucket, objectName),
	}
	return u.String(), nil
}
