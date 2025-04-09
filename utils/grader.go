package utils

import (
	"fmt"
	"math"
)

func Grade(pc *PineconeClient, cardId string, providedAnswer string) (float32, error) {
	actualAnswerEmbed, err := pc.FetchAnswer(cardId)
	if err != nil {
		return 0, err
	}

	providedAnswerEmbed, err := MakeOpenAIEmbedRequest(providedAnswer)
	if err != nil {
		return 0, fmt.Errorf("unable to embed provided answer: %v", err)
	}

	var numericGrade float32 = CosineSimilarity(actualAnswerEmbed, providedAnswerEmbed)

	return numericGrade, nil
}

func CosineSimilarity(a, b *[]float32) float32 {

	// Range [-1, 1]
	cosineSim := (DotProduct(a, b) / L2Norm(a, b))

	// Range [0, 1]
	normalizedSim := (cosineSim + 1) / 2

	return normalizedSim * 100
}

func DotProduct(a, b *[]float32) float32 {
	var ans float32
	for i := 0; i < len(*a); i += 1 {
		ans += ((*a)[i]) * ((*b)[i])
	}
	return ans
}

func L2Norm(a, b *[]float32) float32 {
	var aNorm, bNorm float32
	for i := 0; i < len(*a); i += 1 {
		aNorm += (*a)[i] * (*a)[i]
	}
	for i := 0; i < len(*b); i += 1 {
		bNorm += (*b)[i] * (*b)[i]
	}

	aNorm = float32(math.Sqrt(float64(aNorm)))
	bNorm = float32(math.Sqrt(float64(bNorm)))

	return aNorm * bNorm
}
