package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	gogenai "github.com/google/generative-ai-go/genai"
	"github.com/rs/zerolog/log"
	"google.golang.org/api/option"
	"google.golang.org/genai"

	_ "github.com/lib/pq"
)

const (
	project = "imrenagi-gemini-experiment"
	region  = "us-central1"
)

func main() {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt)
	signal.Notify(ch, syscall.SIGTERM)
	go func() {
		oscall := <-ch
		log.Warn().Msgf("system call:%+v", oscall)
		cancel()
	}()

	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Project:  project,
		Location: region,
		Backend:  genai.BackendVertexAI,
	})
	if err != nil {
		log.Fatal().Err(err).Msgf("create genai client error")
	}

	log.Debug().
		Str("token", os.Getenv("GEMINI_API_KEY")).
		Msg("checking token")

	genAiClient, err := gogenai.NewClient(ctx,
		option.WithAPIKey(os.Getenv("GEMINI_API_KEY")))
	if err != nil {
		log.Fatal().Err(err).Msgf("create golang genai client error")

	}
	defer genAiClient.Close()

	em := genAiClient.EmbeddingModel(embeddingModelName)

	db := NewSQLx()

	vectorStore := NewVectorStore(db)

	srv := &Server{
		GenAIClient:    client,
		EmbeddingModel: em,
		DB:             db,
		VectorStore:    vectorStore,
	}
	srv.Start(ctx)
}
