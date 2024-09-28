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

type ErrorResponse struct {
	Error string `json:"error"`
}

// @Summary Get benefits
// @Description Retrieves benefits based on query parameters
// @Tags benefits
// @Accept json
// @Produce json
// @Param category query string false "Category of the benefit"
// @Param min_price query string false "Minimum price of the benefit"
// @Param max_price query string false "Maximum price of the benefit"
// @Param search query string false "Search term"
// @Success 200 {array} Benefit "List of benefits"
// @Failure 400 {object} ErrorResponse "Invalid query parameters"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /benefits [get]
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

// @Summary Get a single benefit
// @Description Retrieves a single benefit by its ID
// @Tags benefits
// @Accept json
// @Produce json
// @Param id path string true "Benefit ID"
// @Success 200 {object} Benefit "Single benefit"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 404 {object} ErrorResponse "Benefit not found"
// @Router /benefits/{id} [get]
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

// @Summary Add a new benefit
// @Description Adds a new benefit to the database
// @Tags benefits
// @Accept json
// @Produce json
// @Param benefit body Benefit true "Benefit to add"
// @Success 201 {object} Benefit "Benefit created"
// @Failure 400 {object} ErrorResponse "Invalid benefit format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /benefits [post]
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

// @Summary Delete a benefit
// @Description Deletes a benefit by its ID
// @Tags benefits
// @Accept json
// @Produce json
// @Param id path string true "Benefit ID"
// @Success 200 {object} object "Benefit deleted successfully"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /benefits/{id} [delete]
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

// @Summary Update a benefit
// @Description Updates a benefit by its ID
// @Tags benefits
// @Accept json
// @Produce json
// @Param id path string true "Benefit ID"
// @Param benefit body Benefit true "Updated benefit information"
// @Success 200 {object} Benefit "Benefit updated successfully"
// @Failure 400 {object} ErrorResponse "Invalid ID or benefit format"
// @Failure 404 {object} ErrorResponse "Benefit not found"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /benefits/{id} [put]
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

// @Summary Get all wallets
// @Description Retrieves all wallets
// @Tags wallets
// @Accept json
// @Produce json
// @Success 200 {array} Wallet "List of wallets"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /wallets [get]
func getAllWallets(c *gin.Context, repo *WalletRepository) {
	ctx := c.Request.Context()
	wallets, err := repo.GetAllWallets(ctx)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, wallets)
}

// @Summary Buy a benefit
// @Description Buys a benefit for a user, deducting from their wallet
// @Tags benefits
// @Accept json
// @Produce json
// @Param benefit_id path string true "Benefit ID"
// @Param user_id body string true "User ID"
// @Success 200 {object} OwnedBenefit "Purchased benefit"
// @Failure 400 {object} ErrorResponse "Invalid ID format or insufficient funds"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /benefits/{benefit_id}/buy [post]
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

// @Summary Grant tokens to a wallet
// @Description Grants tokens to a user's wallet
// @Tags wallets
// @Accept json
// @Produce json
// @Param user_id body string true "User ID"
// @Param amount body int true "Amount of tokens to grant"
// @Success 200 {object} Wallet "Updated wallet"
// @Failure 400 {object} ErrorResponse "Invalid request format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /wallets/grant [post]
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

// @Summary Get wallet by user ID
// @Description Retrieves a wallet by the user's ID
// @Tags wallets
// @Accept json
// @Produce json
// @Param id path string true "User ID"
// @Success 200 {object} Wallet "User's wallet"
// @Failure 400 {object} ErrorResponse "Invalid ID format"
// @Failure 500 {object} ErrorResponse "Internal server error"
// @Router /wallets/{id} [get]
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
