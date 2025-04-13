package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

const LOGGING = false

const PINECONE_NAMESPACE = "sanctum-grading"

var (
	instance *PineconeClient
	once     sync.Once
	mu       sync.Mutex
)

func (pc *PineconeClient) AddCards(cards []Flashcard) (bool, error) {
	matches := []string{}
	for _, card := range cards {
		matches = append(matches, card.Match)
	}

	embeddings, err := MakeOpenAIEmbedRequest(matches)
	if err != nil {
		return false, fmt.Errorf("error making OpenAI Embed request: %v", err)
	}

	vectors := []*pinecone.Vector{}
	for i, embedding := range *embeddings {
		vectors = append(vectors, &pinecone.Vector{
			Id:     cards[i].Uuid,
			Values: &embedding.Embedding,
		})
	}

	n, err := pc.Index.UpsertVectors(pc.Ctx, vectors)
	if err != nil {
		return false, err
	}
	log.Printf("Vectors Upserted: %v", n)

	return true, nil
}

func (pc *PineconeClient) RemoveCard(cardId string) (bool, error) {
	err := pc.Index.DeleteVectorsById(pc.Ctx, []string{cardId})
	if err != nil {
		return false, err
	}
	return true, nil
}

func (pc *PineconeClient) FetchAnswer(cardId string) (*[]float32, error) {
	var answerEmbed *[]float32

	vectors, err := pc.Index.FetchVectors(pc.Ctx, []string{cardId})
	if err != nil {
		return nil, fmt.Errorf("unable to fetch vectors from pinecone: %v", err)
	}

	if len(vectors.Vectors) == 0 {
		return nil, fmt.Errorf("answer is unavailable, either vector with this id does not exist or this vector is in the process of being inserted")
	}

	answerEmbed = vectors.Vectors[cardId].Values

	return answerEmbed, nil
}

func (pc *PineconeClient) IndexMetrics() (IndexMetrics, error) {

	metrics, err := pc.Index.DescribeIndexStats(pc.Ctx)
	if err != nil {
		return IndexMetrics{}, err
	}

	indexMetrics := IndexMetrics{
		VectorCount: int(metrics.TotalVectorCount),
		Dimension:   int(*metrics.Dimension),
	}

	return indexMetrics, nil
}

func (indexMetric IndexMetrics) String() string {
	return fmt.Sprintf("Vector Count: %v \n Dimension: %v", indexMetric.VectorCount, indexMetric.Dimension)
}

func InitPineconeClient() (*PineconeClient, error) {
	ctx := context.Background()

	API_KEY := os.Getenv("PINECONE_API_KEY")
	if API_KEY == "" {
		return nil, errors.New("PINECONE_KEY environment variable not found")
	}

	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: API_KEY,
	})

	if err != nil {
		return nil, err
	}

	idxModel, err := client.DescribeIndex(ctx, PINECONE_NAMESPACE)
	if err != nil {
		return nil, err
	}

	if LOGGING {
		log.Printf("Connected to Pinecone Index: %v", idxModel.Host)
	}

	index, err := client.Index(pinecone.NewIndexConnParams{
		Host:      idxModel.Host,
		Namespace: "flashcards",
	})

	if err != nil {
		return nil, err
	}

	pc := &PineconeClient{
		Ctx:    ctx,
		Client: client,
		Index:  index,
	}

	return pc, nil
}

func GetPineconeClient() (*PineconeClient, error) {
	if instance != nil {
		return instance, nil
	}

	mu.Lock()
	defer mu.Unlock()

	if instance != nil {
		return instance, nil
	}

	var err error
	once.Do(func() {
		instance, err = InitPineconeClient()
	})

	if err != nil {
		return nil, err
	}

	return instance, nil
}
