package main

import (
	"encoding/json"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type BenefitCategory int

const (
	Transport BenefitCategory = iota
	Health
	Entertainment
	Culture
)

type Benefit struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	Name           string             `json:"name" bson:"name"`
	Category       BenefitCategory           `json:"category" bson:"category"`
	Description    string             `json:"description" bson:"description"`
	ImageUrl       string             `json:"imageUrl" bson:"imageUrl"`
	Price          float64            `json:"price" bson:"price"`
	InStock        int                `json:"inStock" bson:"inStock"`
	ExpirationDate time.Time          `json:"expirationDate" bson:"expirationDate"`
}

type OwnedBenefit struct {
	Id             primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	OwnerId        primitive.ObjectID `json:"ownerId" bson:"ownerId"`
	BenefitId      primitive.ObjectID `json:"benefitId" bson:"benefitId"`
	Purchased      time.Time          `json:"purchased" bson:"purchased"`
	Content        string             `json:"content" bson:"content"`
	ExpirationDate time.Time          `json:"expirationDate" bson:"expirationDate"`
}

// String values for the enum
func (c BenefitCategory) String() string {
	return [...]string{"Transport", "Health", "Entertainment", "Culture"}[c]
}

// UnmarshalJSON customizes the unmarshalling of the Category field from JSON.
func (c *BenefitCategory) UnmarshalJSON(data []byte) error {
	var str string
	if err := json.Unmarshal(data, &str); err != nil {
		return err
	}
	switch str {
	case "Transport":
		*c = Transport
	case "Health":
		*c = Health
	case "Entertainment":
		*c = Entertainment
	case "Culture":
		*c = Culture
	default:
		return fmt.Errorf("unknown category: %s", str)
	}

	return nil
}

// Implement MarshalJSON for custom JSON serialization
func (c BenefitCategory) MarshalJSON() ([]byte, error) {
	return json.Marshal(c.String())
}
