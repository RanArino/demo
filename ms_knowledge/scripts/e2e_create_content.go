package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	knowledgev1 "demo/ms_knowledge/api/proto/v1"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
)

func main() {
	addr := os.Getenv("KNOWLEDGE_ADDR")
	if addr == "" {
		addr = "localhost:50052"
	}
	token := os.Getenv("CLERK_TEST_BEARER")
	if token == "" {
		log.Fatal("set CLERK_TEST_BEARER to a valid 'Bearer <jwt>'")
	}

	conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	c := knowledgev1.NewKnowledgeServiceClient(conn)

	ctx := metadata.NewOutgoingContext(context.Background(), metadata.Pairs("authorization", token))
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	sp, err := c.CreateSpace(ctx, &knowledgev1.CreateSpaceRequest{Title: "E2E Space", Description: "d"})
	if err != nil {
		log.Fatalf("CreateSpace: %v", err)
	}
	fmt.Println("space:", sp.GetId())

	resp, err := c.CreateUploadURL(ctx, &knowledgev1.CreateUploadURLRequest{
		SpaceId:    sp.GetId(),
		Filename:   "doc.pdf",
		MimeType:   "application/pdf",
		SizeBytes:  12345,
		Title:      "Doc",
		ObjectKind: knowledgev1.DownloadObjectKind_DOWNLOAD_OBJECT_KIND_ORIGINAL,
	})
	if err != nil {
		log.Fatalf("CreateUploadURL: %v", err)
	}
	fmt.Println("content:", resp.GetContentSource().GetId(), "owner:", resp.GetContentSource().GetOwnerId())
}
