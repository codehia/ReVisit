package types

import "errors"

type Node struct {
	Name     string `json:"name"`
	Notes    string `json:"notes,omitempty"`
	Children []Node `json:"children,omitempty"`
}

type Card struct {
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Examples  string `json:"examples"`
	TradeOffs string `json:"tradeoffs"`
	CardType  string `json:"card_type"`
}

func (c Card) Validate() error {
	validCardTypes := map[string]bool{
		"definition":  true,
		"mechanism":   true,
		"tradeoff":    true,
		"application": true,
	}
	if c.Question == "" {
		return errors.New("missing question")
	} else if c.Answer == "" {
		return errors.New("missing answer")
	} else if !validCardTypes[c.CardType] {
		return errors.New("invalid card type: " + c.CardType)
	}
	return nil
}

type Response struct {
	Tag     string   `json:"tag"`
	TagPath []string `json:"tag_path"`
	Cards   []Card   `json:"cards"`
}

func (r Response) Validate() error {
	if r.Tag == "" {
		return errors.New("tag field is empty")
	} else if len(r.TagPath) == 0 {
		return errors.New("tag_path field is empty")
	} else if len(r.Cards) == 0 {
		return errors.New("cards field is empty")
	}
	for _, card := range r.Cards {
		if err := card.Validate(); err != nil {
			return err
		}
	}
	return nil
}
