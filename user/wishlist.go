package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func AddToWishlist(c *gin.Context) {
	var wishlistItem models.Wishlist_item
	var product models.Product

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	productId := c.Param("id")

	if err := database.Db.Where("id = ?", productId).First(&product).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	if err := database.Db.Where("user_id = ? AND product_id = ?", user_id, productId).First(&wishlistItem).Error; err == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product already exists in the wishlist"})
		return
	}

	wishlistItem.User_Id = user_id
	wishlistItem.Product_Id = int(product.ID)

	if err := database.Db.Create(&wishlistItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not add product to wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product added to wishlist"})
}

func ListWishlist(c *gin.Context) {
	var wishlistItems []models.Wishlist_item
	user_id := helper.Get_Userid(c)

	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	if err := database.Db.Preload("Product").Where("user_id = ?", user_id).Find(&wishlistItems).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve wishlist items"})
		return
	}
	if len(wishlistItems) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "wishlist is empty"})
		return
	}
	var list_response responsemodels.Wishlist_item
	for _, item := range wishlistItems {
		list_response = responsemodels.Wishlist_item{
			Product_Id:   item.Product_Id,
			Product_Name: item.Product.Name,
			Price:        float32(item.Product.Offer_Price),
		}
	}

	c.JSON(http.StatusOK, gin.H{"wishlist_items": list_response})
}

func RemoveFromWishlist(c *gin.Context) {
	var wishlistItem models.Wishlist_item

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	productId := c.Param("id")

	if err := database.Db.Where("user_id = ? AND product_id = ?", user_id, productId).First(&wishlistItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found in wishlist"})
		return
	}

	if err := database.Db.Delete(&wishlistItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to remove product from wishlist"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product removed from wishlist"})
}
