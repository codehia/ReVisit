package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/codehia/goflash/internal/db/model"
	"github.com/codehia/goflash/internal/db/table"
	"github.com/codehia/goflash/internal/store"
	"github.com/go-jet/jet/v2/sqlite"
)

type Card struct {
	Question  string `json:"question"`
	Answer    string `json:"answer"`
	Examples  string `json:"examples"`
	TradeOffs string `json:"tradeoffs"`
	CardType  string `json:"card_type"`
}

type Response struct {
	Tag     string   `json:"tag"`
	TagPath []string `json:"tag_path"`
	Cards   []Card   `json:"cards"`
}

type Node struct {
	Name     string `json:"name"`
	Notes    string `json:"notes,omitempty"`
	Children []Node `json:"children,omitempty"`
}

func readJsonFile(path string) ([]Response, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, errors.New("error reading file")
	}
	var responses []Response
	error := json.Unmarshal(data, &responses)
	if error != nil {
		return nil, errors.New("error marshalling")
	}
	return responses, nil
}

func readHierarchyFile() ([]Node, error) {
	data, err := os.ReadFile("system-design-hierarchy.json")
	if err != nil {
		return nil, err
	}

	var root Node
	if err := json.Unmarshal(data, &root); err != nil {
		return nil, err
	}
	return root.Children, nil
}

func getTagRecord(tagName string, db *sql.DB) (model.Tags, error) {
	var foundTag model.Tags
	selectStmt := table.Tags.SELECT(table.Tags.AllColumns).WHERE(table.Tags.Name.EQ(sqlite.String(tagName)))

	selectErr := selectStmt.Query(db, &foundTag)
	if selectErr != nil {
		return foundTag, selectErr
	}
	return foundTag, nil
}

func saveTagsToDB(nodes []Node, db *sql.DB, parentID *string) {
	for _, node := range nodes {
		foundTag, err := getTagRecord(node.Name, db)
		fmt.Println(foundTag)
		if err != nil {
			tag := model.Tags{Name: node.Name, ParentID: parentID}
			stmt := table.Tags.INSERT(table.Tags.Name, table.Tags.ParentID).MODEL(tag).RETURNING(table.Tags.AllColumns)
			var inserted model.Tags
			err := stmt.Query(db, &inserted)
			if err != nil {
				log.Fatal(err)
			}
			saveTagsToDB(node.Children, db, inserted.ID)
		} else if foundTag.ParentID != parentID {
			table.Tags.UPDATE(table.Tags.ParentID).SET(parentID).WHERE(table.Tags.ID.EQ(sqlite.String(*foundTag.ID)))
		}
	}
}

func saveCardsToDB(ouputFilePath string, db *sql.DB) {
	responseData, error := readJsonFile(ouputFilePath)
	if error != nil {
		log.Fatal(error)
	}

	for _, response := range responseData {
		cards := response.Cards

		responseTags := []string{}
		responseTags = append(responseTags, response.Tag)
		responseTags = append(responseTags, response.TagPath...)

		var responseTagRecords []model.Tags
		for _, tag := range responseTags {
			foundTag, err := getTagRecord(tag, db)
			if err != nil {
				log.Fatal(err)
			}
			responseTagRecords = append(responseTagRecords, foundTag)
		}

		for _, card := range cards {
			selectStmt := table.Cards.SELECT(table.Cards.AllColumns).WHERE(table.Cards.Question.EQ(sqlite.String(card.Question)))
			var foundCard model.Cards
			selectErr := selectStmt.Query(db, &foundCard)
			if selectErr != nil {
				cardRecord := model.Cards{
					Question:  card.Question,
					Answer:    card.Answer,
					Examples:  card.Examples,
					Tradeoffs: card.TradeOffs,
					CardType:  card.CardType,
				}
				stmt := table.Cards.INSERT(table.Cards.Question, table.Cards.Answer, table.Cards.Examples,
					table.Cards.Tradeoffs,
					table.Cards.CardType).MODEL(cardRecord).RETURNING(table.Cards.AllColumns)

				var insertedCard model.Cards
				err := stmt.Query(db, &insertedCard)
				if err != nil {
					log.Fatal(err)
				}
				fmt.Println(insertedCard)
				for _, tag := range responseTagRecords {
					cardTagRecord := model.CardTags{
						CardID: *insertedCard.ID,
						TagID:  *tag.ID,
					}
					stmt := table.CardTags.INSERT(table.CardTags.CardID, table.CardTags.TagID).MODEL(cardTagRecord).RETURNING(table.CardTags.AllColumns)
					var insertedCardTag model.CardTags
					err := stmt.Query(db, &insertedCardTag)
					if err != nil {
						log.Fatal(err)
					}
					fmt.Println(insertedCardTag)
				}

			}
		}
	}
}

func main() {
	s, err := store.New("goflash.db")
	if err != nil {
		log.Fatal(err)
	}
	db := s.DB()

	defer db.Close()

	nodes, err := readHierarchyFile()
	if err != nil {
		log.Fatal(err)
	}
	saveTagsToDB(nodes, db, nil)
	saveCardsToDB("output.json", db)
}
