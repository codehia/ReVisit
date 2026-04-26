package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Card struct {
	ID       string `json:"id"`
	Topic    string `json:"topic"`
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func loadCards(path string) ([]Card, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var cards []Card
	err = json.NewDecoder(f).Decode(&cards)
	return cards, err
}

func main() {
	cards, err := loadCards("cards.json")

	if err != nil {
		fmt.Println(err)
	}

	for _, card := range cards {
		fmt.Printf("TOPIC: %s\n", card.Topic)
		fmt.Printf("QUESTION: %s\n", card.Question)
		fmt.Printf("ANSWER: %s\n", card.Answer)
	}
}
