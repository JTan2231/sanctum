package utils

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func (pc *PineconeClient) AddCard(card Flashcard) (bool,error) {

	vec,err := MakeOpenAiEmbedRequest(card.Match)
	if err != nil {
		return false,fmt.Errorf("error making OpenAI Embed request: %v", err)
	}

	vectors := []*pinecone.Vector {
		{
			Id: 	   card.Uuid.String(),
			Values:    vec,
		},
	}

	n,err := pc.Index.UpsertVectors(pc.Ctx, vectors)
	if err != nil {
		return false,err
	}
	log.Printf("Vectors Upserted: %v", n)
	
	return true,nil
}

// TODO
func (pc *PineconeClient) RemoveCard(card Flashcard) (bool,error) {
	return true,nil
}

// TODO
func (pc *PineconeClient) FetchAnswer(card Flashcard) (score float64, grade string) {
	return 0,""
}

func (pc *PineconeClient) IndexMetrics() (IndexMetrics,error) {

	metrics,err := pc.Index.DescribeIndexStats(pc.Ctx)
	if err != nil {
		return IndexMetrics{},err
	}

	indexMetrics := IndexMetrics{
		VectorCount: int(metrics.TotalVectorCount),
		Dimension:   int(*metrics.Dimension),
	}	

	return indexMetrics,nil
}

func (indexMetric IndexMetrics) String() string {
	return fmt.Sprintf("Vector Count: %v \n Dimension: %v", indexMetric.VectorCount, indexMetric.Dimension)
}

func InitPineconeClient(indexName string) (*PineconeClient,error) {

	ctx := context.Background()

	API_KEY := os.Getenv("PINECONE_KEY")
	if API_KEY == "" {
		return nil, errors.New("API Key environment variable not found")
	}

	client, err := pinecone.NewClient(pinecone.NewClientParams{
		ApiKey: API_KEY,
	})
	if err != nil {
		return nil, err
	}

	idxModel, err := client.DescribeIndex(ctx, indexName)
	if err != nil {
		return nil,err
	}
	log.Printf("Connected to Pinecone Index: %v", idxModel.Host)

	index,err := client.Index(pinecone.NewIndexConnParams{
		Host: idxModel.Host,
		Namespace: "flashcards",
	})
	if err != nil {
		return nil, err
	}

	pc := &PineconeClient{
		Ctx :      ctx,
		Client :   client,
		Index :    index, 
	}

	return pc,nil
}