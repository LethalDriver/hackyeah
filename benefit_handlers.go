package main

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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
	ctx := c.Request.Context()
	benefitIdString := c.Param("benefit_id")
	benefitId, err := primitive.ObjectIDFromHex(benefitIdString)
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

	benefit, err := benefitRepo.GetBenefitByID(ctx, benefitId)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	wallet, err := walletRepo.GetWalletByUserID(ctx, req.UserID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if wallet.TokenBalance < benefit.Price {
		c.JSON(http.StatusBadRequest, gin.H{"error": "insufficient funds"})
		return
	}

	wallet.TokenBalance -= benefit.Price
	_, err = walletRepo.UpdateWallet(ctx, wallet)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	ownedBenefit := OwnedBenefit{
		OwnerId:        req.UserID,
		BenefitId:      benefitId,
		Purchased:      time.Now(),
		Content:        "TODO",
		ExpirationDate: benefit.ExpirationDate,
	}
	savedOwnedBenefit, err := benefitRepo.AddPurchasedBenefit(ctx, &ownedBenefit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
	}

	c.JSON(http.StatusOK, savedOwnedBenefit)
}

func grantTokens(c *gin.Context, repo *WalletRepository) {
	ctx := c.Request.Context()
	var req struct {
		UserID primitive.ObjectID `json:"user_id"`
		Amount float64            `json:"amount"`
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
