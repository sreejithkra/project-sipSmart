package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func Product_Details(c *gin.Context) {

	id := c.Param("id")

	var product_details responsemodels.Product_Details

	var product models.Product

	productQuery :=
		`SELECT products.id, products.name, products.description, products.image_url, products.price, products.stock, products.popularity, products.average_rating, categories.name AS category
		FROM products
		JOIN categories ON products.category_id = categories.id
		WHERE products.id = ?`

	if err := database.Db.Raw(productQuery, id).Scan(&product_details).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "pProduct not found"})
		return
	}

	reviewQuery := `
		SELECT id,product_id,rating, content
		FROM reviews 
		WHERE product_id = ?
	`

	var reviews []responsemodels.Review
	if err := database.Db.Raw(reviewQuery, id).Scan(&reviews).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve reviews"})
		return
	}

	product_details.Reviews = reviews
	database.Db.Exec("UPDATE products SET views = views + 1 WHERE id = ?", id)

	if err := database.Db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	helper.UpdateProductPopularity(&product)

	c.JSON(200, gin.H{
		"product": product_details,
	})
}

func Add_Review(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var user_review models.Add_Review
	if err := c.ShouldBindJSON(&user_review); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	err := helper.Validate(user_review)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var product models.Product
	if err := database.Db.Where("id = ?", user_review.Product_Id).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	var reviewCount int64
	database.Db.Model(&models.Review{}).Where("user_id = ? AND product_id = ?", user_id, user_review.Product_Id).Count(&reviewCount)

	if reviewCount != 0 {
		c.JSON(400, gin.H{"error": "This user review already exists"})
		return
	}

	var orderCount int64
	database.Db.Model(&models.Order{}).
		Joins("JOIN order_items ON order_items.order_id = orders.id").
		Where("orders.user_id = ? AND order_items.product_id = ?", user_id, user_review.Product_Id).
		Count(&orderCount)

	if orderCount == 0 {
		c.JSON(http.StatusForbidden, gin.H{"error": "You can only review products you have purchased"})
		return
	}

	review := models.Review{
		User_Id:    user_id,
		Product_Id: user_review.Product_Id,
		Rating:     float32(user_review.Rating),
		Content:    user_review.Review,
	}

	if err := database.Db.Create(&review).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to submit review"})
		return
	}

	var totalReviews int64
	var totalRating int64
	database.Db.Model(&models.Review{}).Where("product_id = ?", product.ID).Count(&totalReviews)
	database.Db.Model(&models.Review{}).Where("product_id = ?", product.ID).Select("SUM(rating)").Scan(&totalRating)

	averageRating := float64(totalRating) / float64(totalReviews)
	product.Average_Rating = averageRating

	if err := database.Db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product rating"})
		return
	}

	helper.UpdateProductPopularity(&product)

	c.JSON(http.StatusOK, gin.H{
		"message":       "Review submitted successfully",
		"product_id":    product.ID,
		"averageRating": product.Average_Rating,
		"content":       review,
	})
}
