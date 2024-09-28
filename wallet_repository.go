package main

import (
	"context"
	"errors"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrWalletNotFound = errors.New("wallet not found")

type Wallet struct {
	Id           primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	UserId       primitive.ObjectID `json:"user_id" bson:"user_id"`
	MoneyBalance float64            `json:"money_balance" bson:"money_balance"`
	TokenBalance int            `json:"token_balance" bson:"token_balance"`
}

type WalletRepository struct {
	collection *mongo.Collection
}

func NewWalletRepository(collection *mongo.Collection) (*WalletRepository, error) {
	return &WalletRepository{
		collection: collection,
	}, nil
}

func (r *WalletRepository) GetWalletByUserID(ctx context.Context, id primitive.ObjectID) (*Wallet, error) {
	var wallet Wallet
	err := r.collection.FindOne(ctx, bson.M{"user_id": id}).Decode(&wallet)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, ErrWalletNotFound
		}
		return nil, err
	}

	return &wallet, nil
}

func (r *WalletRepository) GetAllWallets(ctx context.Context) ([]Wallet, error) {
	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var wallets []Wallet
	for cursor.Next(ctx) {
		var wallet Wallet
		err := cursor.Decode(&wallet)
		if err != nil {
			return nil, err
		}
		wallets = append(wallets, wallet)
	}

	return wallets, nil
}

func (r *WalletRepository) UpdateWallet(ctx context.Context, wallet *Wallet) (*Wallet, error) {
	filter := bson.M{"user_id": wallet.UserId}
	update := bson.M{"$set": wallet}
	_, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return nil, err
	}

	return wallet, nil
}
