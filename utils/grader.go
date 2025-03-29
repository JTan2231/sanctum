package utils

import (
	"fmt"
	"math"
)

func Grade(pc *PineconeClient, cardId string, providedAnswer string) (float32, string, error) {

	actualAnswerEmbed, err := pc.FetchAnswer(cardId)
	if err != nil {
		return 0, "", fmt.Errorf("unable to fetch actual answer of card: %v", err)
	}

	providedAnswerEmbed, err := MakeOpenAIEmbedRequest(providedAnswer)
	if err != nil {
		return 0,"",fmt.Errorf("unable to embed provided answer: %v", err)
	}

	// Need to normalize this since its not in range [0,1]
	var numericGrade float32 = CosineSimilarity(actualAnswerEmbed,providedAnswerEmbed)
	var letterGrade  string

	switch {

	case 90 <= numericGrade && numericGrade <= 100:
		letterGrade = "A"
	case 87 <= numericGrade && numericGrade <= 90:
		letterGrade = "A-"


	case 83 <= numericGrade && numericGrade <= 87:
		letterGrade = "B+"
	case 80 <= numericGrade && numericGrade <= 83:
		letterGrade = "B"
	case 77 <= numericGrade && numericGrade <= 80:
		letterGrade = "B-"


	case 73 <= numericGrade && numericGrade <= 77:
		letterGrade = "C+"
	case 70 <= numericGrade && numericGrade <= 73:
		letterGrade = "C"
	case 67 <= numericGrade && numericGrade <= 70:
		letterGrade = "C-"
	

	case 63 <= numericGrade && numericGrade <= 67:
		letterGrade = "D+"
	case 60 <= numericGrade && numericGrade <= 63:
		letterGrade = "D"
	case 50 <= numericGrade && numericGrade <= 60:
		letterGrade = "D-"


	default:
		letterGrade = "F"
	}

	return numericGrade,letterGrade,nil
}

func CosineSimilarity(a,b *[]float32) float32 {

	// Range [-1,1]
	cosineSim := (DotProduct(a,b) / L2Norm(a,b))
	
	// Range [0,1]
	normalizedSim := (cosineSim + 1) / 2
	
	return normalizedSim * 100
}

func DotProduct(a,b *[]float32) float32 {
	var ans float32
	for i := 0; i < len(*a); i += 1 {
		ans += ((*a)[i]) * ((*b)[i])
	}
	return ans
}

func L2Norm(a,b *[]float32) float32 {
	var aNorm,bNorm float32
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