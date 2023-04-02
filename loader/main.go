package main

import (
	"context"
	"fmt"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
	log "github.com/rs/zerolog/log"
)

func createRandDoc(ctx context.Context, client *firestore.Client, collectionName string, ch chan string) error {

	var data = make(map[string]interface{})
	id, _ := uuid.NewRandom()
	data["name"] = gofakeit.Name()
	data["SSN"] = gofakeit.SSN()

	_, err := client.Collection(collectionName).Doc(id.String()).Set(ctx, data)
	if err != nil {
		log.Error().Msgf("An error has occurred: %s", err)
	}
	log.Info().Msgf("id: %s", id.String())

	ch <- data["SSN"].(string)

	return err
}

var projectID = os.Getenv("PROJECT")
var collectionName = os.Getenv("COLLECTION")

func init() {
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()
	zerolog.LevelFieldName = "severity"
	zerolog.TimestampFieldName = "timestamp"
	zerolog.TimeFieldFormat = time.RFC3339Nano

	if collectionName == "" {
		collectionName = "authors"
	}

}

func main() {

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	client, err := firestore.NewClient(ctx, projectID)
	if err != nil {
		log.Error().Err(err)
	}

	ch := make(chan string)
	go WriteToFirestore(ctx, ch, client)

	for v := range ch {
		fmt.Println(v)
	}
}

func WriteToFirestore(ctx context.Context, ch chan string, client *firestore.Client) {
	defer close(ch) // explicit close

	limit := make(chan struct{}, 5)
	var wg sync.WaitGroup

	// これ自体を go routing化してchをcloseする
	for i := 0; i <= 10; i++ {
		wg.Add(1)
		go func() {
			limit <- struct{}{}
			defer wg.Done()

			createRandDoc(ctx, client, collectionName, ch)
			<-limit
		}()
	}
	wg.Wait()
}
