package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	storager2 "demo/ms_knowledge/internal/storage/r2"
)

func main() {
	var (
		bucket  string
		key     string
		ttl     time.Duration
		region  string
		usePath bool
	)

	flag.StringVar(&bucket, "bucket", os.Getenv("R2_BUCKET_SOURCE_NAME"), "R2 bucket name")
	flag.StringVar(&key, "key", "", "object key (e.g., integration-tests/xxx.pdf)")
	flag.DurationVar(&ttl, "ttl", 15*time.Minute, "presigned URL TTL (e.g., 15m)")
	flag.StringVar(&region, "region", os.Getenv("R2_REGION"), "R2 region (default 'auto')")
	flag.BoolVar(&usePath, "path-style", true, "use path-style addressing")
	flag.Parse()

	endpoint := os.Getenv("R2_ENDPOINT")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	if region == "" {
		region = "auto"
	}

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" || key == "" {
		log.Fatalf("missing required config: set R2_ENDPOINT, R2_ACCESS_KEY_ID, R2_SECRET_ACCESS_KEY, R2_BUCKET_SOURCE_NAME and pass --key")
	}

	ctx := context.Background()
	client, err := storager2.NewClient(ctx, storager2.Config{
		Endpoint:        endpoint,
		Region:          region,
		AccessKeyID:     accessKey,
		SecretAccessKey: secretKey,
		UsePathStyle:    usePath,
	})
	if err != nil {
		log.Fatalf("failed creating R2 client: %v", err)
	}

	url, err := client.GeneratePresignedDownloadURL(ctx, bucket, key, ttl)
	if err != nil {
		log.Fatalf("failed to presign: %v", err)
	}

	fmt.Println(url)
}
