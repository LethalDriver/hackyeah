package main

import (
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

func getBenefits(c *gin.Context, repo *BenefitRepository) {
	ctx := c.Request.Context()

	// Get query parameters
	category := c.Query("category")
	minPrice := c.Query("min_price")
	maxPrice := c.Query("max_price")
	search := c.Query("search")

	// Prepare filter options
	filter := bson.M{}
	if category != "" {
		if category != "" {
			filter["category"] = bson.M{"$regex": category, "$options": "i"}
		}
	}
	if minPrice != "" {
		minPriceFloat, err := strconv.ParseFloat(minPrice, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid min_price"})
			return
		}
		filter["price"] = bson.M{"$gte": minPriceFloat}
	}
	if maxPrice != "" {
		maxPriceFloat, err := strconv.ParseFloat(maxPrice, 64)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_price"})
			return
		}
		filter["price"] = bson.M{"$lte": maxPriceFloat}
	}
	if search != "" {
		filter["name"] = bson.M{"$regex": search, "$options": "i"}
	}

	// Get benefits based on filter
	benefits, err := repo.GetFilteredBenefits(ctx, filter)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, benefits)
}

func getBenefit(c *gin.Context, repo *BenefitRepository) {
	ctx := c.Request.Context()
	benefitIdString := c.Param("id")
	benefitId, err := primitive.ObjectIDFromHex(benefitIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	benefit, err := repo.GetBenefitByID(ctx, benefitId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, benefit)
}

func addBenefit(c *gin.Context, repo *BenefitRepository) {
	ctx := c.Request.Context()
	var benefit Benefit
	if err := c.ShouldBindJSON(&benefit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	savedBenefit, err := repo.AddBenefit(ctx, &benefit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, savedBenefit)
}

func deleteBenefit(c *gin.Context, repo *BenefitRepository) {
	ctx := c.Request.Context()
	benefitIdString := c.Param("id")
	benefitId, err := primitive.ObjectIDFromHex(benefitIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = repo.DeleteBenefit(ctx, benefitId)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Benefit deleted successfully"})
}

func updateBenefit(c *gin.Context, repo *BenefitRepository) {
	ctx := c.Request.Context()
	benefitIdString := c.Param("id")
	benefitId, err := primitive.ObjectIDFromHex(benefitIdString)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if _, err := repo.GetBenefitByID(ctx, benefitId); err != nil {
		if err == ErrBenefitNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var updatedBenefit Benefit
	if err := c.ShouldBindJSON(&updatedBenefit); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if updatedBenefit.Id != benefitId {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Benefit ID in URL does not match ID in request body"})
		return
	}

	savedBenefit, err := repo.UpdateBenefit(ctx, &updatedBenefit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, savedBenefit)
}

func getAllWallets(c *gin.Context, repo *WalletRepository) {
	ctx := c.Request.Context()
	wallets, err := repo.GetAllWallets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

func buyBenefit(c *gin.Context, benefitRepo *BenefitRepository, walletRepo *WalletRepository) {
	// Start a session
	session, err := benefitRepo.client.StartSession()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to start session"})
		return
	}
	defer session.EndSession(c.Request.Context())

	benefitId := c.Param("benefit_id")
	benefitObjectId, err := primitive.ObjectIDFromHex(benefitId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var req struct {
		UserID primitive.ObjectID `json:"user_id"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Define the transaction
	transaction := func(sessCtx mongo.SessionContext) (any, error) {
		// Perform all the steps inside the transaction

		// Step 1: Get Benefit by ID
		benefit, err := benefitRepo.GetBenefitByID(sessCtx, benefitObjectId)
		if err != nil {
			return nil, err
		}

		// Step 2: Get Wallet by User ID
		wallet, err := walletRepo.GetWalletByUserID(sessCtx, req.UserID)
		if err != nil {
			return nil, err
		}

		// Step 3: Check and update wallet balance
		if wallet.TokenBalance < benefit.Price {
			return nil, errors.New("insufficient funds")
		}
		wallet.TokenBalance -= benefit.Price
		_, err = walletRepo.UpdateWallet(sessCtx, wallet)
		if err != nil {
			return nil, err
		}

		// Step 4: Add purchased benefit
		ownedBenefit := OwnedBenefit{
			OwnerId:        req.UserID,
			BenefitId:      benefitObjectId,
			Purchased:      time.Now(),
			Content:        "TODO",
			ExpirationDate: benefit.ExpirationDate,
		}
		savedOwnedBenefit, err := benefitRepo.AddPurchasedBenefit(sessCtx, &ownedBenefit)
		if err != nil {
			return nil, err
		}

		return savedOwnedBenefit, nil
	}

	// Execute the transaction
	result, err := session.WithTransaction(c.Request.Context(), transaction)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func grantTokens(c *gin.Context, repo *WalletRepository) {
	ctx := c.Request.Context()
	var req struct {
		UserID primitive.ObjectID `json:"user_id"`
		Amount int                `json:"amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := repo.GetWalletByUserID(ctx, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet.TokenBalance += req.Amount
	_, err = repo.UpdateWallet(ctx, wallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}

func getWalletByUserID(c *gin.Context, repo *WalletRepository) {
	ctx := c.Request.Context()
	userId := c.Param("id")
	userIdObject, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := repo.GetWalletByUserID(ctx, userIdObject)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, wallet)
}
