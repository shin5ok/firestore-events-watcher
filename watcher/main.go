package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// listenChanges listens to a query, returning the list of document changes.
func listenChanges(ctx context.Context, w io.Writer, projectID, collection string, timeout int) error {
	// projectID := "project-id"
	to := time.Duration(timeout)
	ctx, cancel := context.WithTimeout(ctx, to*time.Second)
	defer cancel()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		return fmt.Errorf("firestore.NewClient: %v", err)
	}
	defer client.Close()
	defer func() {
		fmt.Println("timeout exceeded", timeout)
	}()

	it := client.Collection(collection).Where("kind", "==", "item").Snapshots(ctx)
	for {
		snap, err := it.Next()
		// DeadlineExceeded will be returned when ctx is cancelled.
		if status.Code(err) == codes.DeadlineExceeded {
			return nil
		}
		if err != nil {
			return fmt.Errorf("Snapshots.Next: %v", err)
		}
		if snap != nil {
			for _, change := range snap.Changes {
				switch change.Kind {
				case firestore.DocumentAdded:
					fmt.Fprintf(w, "New item: %v\n", change.Doc.Data())
				case firestore.DocumentModified:
					fmt.Fprintf(w, "Modified item: %v\n", change.Doc.Data())
				case firestore.DocumentRemoved:
					fmt.Fprintf(w, "Removed item: %v\n", change.Doc.Data())
				}
			}
		}
	}
}

func main() {
	projectID := flag.String("project", "", "")
	collection := flag.String("collection", "", "")
	timeout := flag.Int("timeout", 30, "")
	flag.Parse()

	ctx := context.Background()
	w := os.Stdout
	listenChanges(ctx, w, *projectID, *collection, *timeout)
}
