package textrendering

import (
	"fmt"
	"testing"
)

func TestFlyW(t *testing.T) {

	// color := "black"
	// size := "medium"
	// family := "timesnewroman"
	// weight := "italic"

	text1 := NewText()

	for i := 0; i < 5; i++ {
		char := byte('a' + (i % 26)) // Cycle through 'a' to 'z'
		color := []string{"black", "red", "blue"}[i%3]
		family := []string{"timesnewroman", "arial", "wingdings"}[i%3]
		size := []string{"small", "medium", "large"}[i%3]
		weight := []string{"bold", "italic"}[i%2]
		text1.AddCharacters(char, 1, 2, color, weight, size, family)
		// text1.AddCharacters(char, 1, 2, color, weight, size, family)
	}

	fmt.Println(text1.Characters)
	// fmt.Println(text2.Characters)
	// fmt.Println(totalCharacterMemory(text1.Characters), totalCharacterMemory(text2.Characters))
}
