package repository

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrBenefitNotFound = errors.New("benefit not found")

type BenefitRepository struct {
	benefitCollection          *mongo.Collection
	purchasedBenefitCollection *mongo.Collection
}

func NewBenefitRepository(benefitCollection *mongo.Collection, purchasedBenefitCollection *mongo.Collection) (*BenefitRepository, error) {
	return &BenefitRepository{
		benefitCollection:          benefitCollection,
		purchasedBenefitCollection: purchasedBenefitCollection,
	}, nil
}

func (r *BenefitRepository) GetAllBenefits(ctx context.Context) ([]Benefit, error) {
	cursor, err := r.benefitCollection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var benefits []Benefit
	for cursor.Next(ctx) {
		var benefit Benefit
		err := cursor.Decode(&benefit)
		if err != nil {
			return nil, err
		}
		benefits = append(benefits, benefit)
	}

	return benefits, nil
}

func (r *BenefitRepository) GetBenefitByID(ctx context.Context, id primitive.ObjectID) (*Benefit, error) {
	var benefit Benefit
	err := r.benefitCollection.FindOne(context.Background(), bson.M{"_id": id}).Decode(&benefit)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrBenefitNotFound
		}
		return nil, err
	}

	return &benefit, nil
}

func (r *BenefitRepository) AddBenefit(ctx context.Context, benefit *Benefit) (*Benefit, error) {
	result, err := r.benefitCollection.InsertOne(ctx, benefit)
	if err != nil {
		return nil, err
	}
	generatedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("failed to convert InsertedID to ObjectID")
	}
	benefit.Id = generatedID
	return benefit, nil
}

func (r *BenefitRepository) UpdateBenefit(ctx context.Context, benefit *Benefit) (*Benefit, error) {
	filter := bson.M{"_id": benefit.Id}
	update := bson.M{"$set": benefit}
	_, err := r.benefitCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return benefit, nil
}

func (r *BenefitRepository) DeleteBenefit(ctx context.Context, id primitive.ObjectID) error {
	filter := bson.M{"_id": id}
	_, err := r.benefitCollection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

func (r *BenefitRepository) GetOwnedBenefits(ctx context.Context, userID primitive.ObjectID) ([]OwnedBenefit, error) {
	cursor, err := r.purchasedBenefitCollection.Find(ctx, bson.M{"owner_id": userID})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var ownedBenefits []OwnedBenefit
	for cursor.Next(ctx) {
		var ownedBenefit OwnedBenefit
		err := cursor.Decode(&ownedBenefit)
		if err != nil {
			return nil, err
		}
		ownedBenefits = append(ownedBenefits, ownedBenefit)
	}

	return ownedBenefits, nil
}

func (r *BenefitRepository) AddPurchasedBenefit(ctx context.Context, ownedBenefit *OwnedBenefit) (*OwnedBenefit, error) {
	result, err := r.purchasedBenefitCollection.InsertOne(ctx, ownedBenefit)
	if err != nil {
		return nil, err
	}
	generatedID, ok := result.InsertedID.(primitive.ObjectID)
	if !ok {
		return nil, errors.New("failed to convert InsertedID to ObjectID")
	}
	ownedBenefit.Id = generatedID
	return ownedBenefit, nil
}
