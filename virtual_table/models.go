package virtual_table

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type User struct {
	Id      primitive.ObjectID `json:"id" bson:"_id"`
	Name    string             `json:"name" bson:"name"`
	Picture string             `json:"picture" bson:"picture"`
}

type Character struct {
	Id      string `json:"id" bson:"id"`
	Table   string `json:"table" bson:"table"`
	Player  string `json:"player" bson:"player"`
	Name    string `json:"name" bson:"name"`
	Picture string `json:"picture" bson:"picture"`
}

type Message struct {
	Content string    `json:"content" bson:"content"`
	By      string    `json:"by" bson:"by"`
	At      time.Time `json:"at" bson:"at"`
}

type Discussion struct {
	Id         string    `json:"id" bson:"id"`
	Name       string    `json:"name" bson:"name"`
	Persistent bool      `json:"persistent" bson:"persistent"`
	Between    []string  `json:"between" bson:"between"`
	Messages   []Message `json:"messages" bson:"messages"`
}

type Table struct {
	Id          primitive.ObjectID `json:"id" bson:"_id"`
	Name        string             `json:"name" bson:"name"`
	Master      string             `json:"master" bson:"master"`
	Players     []string           `json:"players" bson:"players"`
	Characters  []Character        `json:"characters" bson:"characters"`
	Discussions []Discussion       `json:"discussions" bson:"discussions"`
}

type TableWithEvents struct {
	Table
	Events []Event `json:"events" bson:"events"`
}
