package main

import (
	"net/http"
	"payments-service/repository"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func getAllBenefits(c *gin.Context, repo *repository.BenefitRepository) {
	ctx := c.Request.Context()
	benefits, err := repo.GetAllBenefits(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, benefits)
}

func getBenefit(c *gin.Context, repo *repository.BenefitRepository) {
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

func addBenefit(c *gin.Context, repo *repository.BenefitRepository) {
	ctx := c.Request.Context()
	var benefit repository.Benefit
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

func getAllWallets(c *gin.Context, repo *repository.WalletRepository) {
	ctx := c.Request.Context()
	wallets, err := repo.GetAllWallets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

func buyBenefit(c *gin.Context, benefitRepo *repository.BenefitRepository, walletRepo *repository.WalletRepository) {
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

	ownedBenefit := repository.OwnedBenefit{
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
