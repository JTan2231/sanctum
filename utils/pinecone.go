package utils

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/pinecone-io/go-pinecone/v3/pinecone"
)

func (pc *PineconeClient) AddCard(card Flashcard) (bool,error) {

	var records = []*pinecone.IntegratedRecord {
		&pinecone.IntegratedRecord{
			"_id" : 	   card.Uuid.String(),
			"chunk_text" : card.Match,
		},
	}

	err := pc.Index.UpsertRecords(pc.Ctx, records)
	if err != nil {
		return false,err
	}
	
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