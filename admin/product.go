package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func List_Products(c *gin.Context) {

	var products []responsemodels.Product

	sql := `SELECT products.id, products.name AS name, categories.name AS category, 
	        products.description, products.image_url,products.sales ,products.price, products.stock,products.offer_price,products.popularity
	        FROM categories 
	        JOIN products ON categories.id = products.category_id ORDER BY products.id`

	database.Db.Raw(sql).Scan(&products)

	c.JSON(200, gin.H{
		"products": products,
	})
}

func Create_Product(c *gin.Context) {

	var product_Info models.Product_info

	if err := c.ShouldBindJSON(&product_Info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
		return
	}

	var count int64

	if err := database.Db.Model(&models.Category{}).Where("id = ?", product_Info.Category_Id).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
	}

	if count == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "invalid Category_id"})
		return
	}

	var count1 int64

	if err := database.Db.Model(&models.Product{}).Where("name = ?", product_Info.Name).Count(&count1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	if count1 != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Record exists with the given product name"})
		return
	}

	err := helper.Validate(product_Info)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	product := models.Product{
		Name:        product_Info.Name,
		Description: product_Info.Description,
		Image_Url:   product_Info.Image_Url,
		Category_Id: product_Info.Category_Id,
		Stock:       product_Info.Stock,
		Price:       product_Info.Price,
		Offer_Price: product_Info.Price,
	}

	if err := database.Db.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database operation failed",
		})
		return
	}

	var products responsemodels.Product

	sql := `SELECT products.id AS id, products.name AS name, 
	categories.name AS category, products.description, 
	products.image_url, products.price,products.offer_price, products.stock 
	FROM categories 
	JOIN products ON categories.id = products.category_id 
	WHERE products.id = ?`

	if err := database.Db.Raw(sql, product.ID).Scan(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "product created successfully",
		"products": products,
	})
}

func Update_Product(c *gin.Context) {

	var product models.Product

	id := c.Param("id")

	if err := database.Db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
		return
	}

	var count1 int64

	if err := database.Db.Model(&models.Category{}).Where("id = ?", product.Category_Id).Count(&count1).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
	}

	if count1 == 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "invalid Category_id"})
		return
	}

	err := helper.Validate(product)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := database.Db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database operation failed",
		})
		return
	}

	var products responsemodels.Product

	sql := `SELECT products.id AS id, products.name AS product_name, 
	categories.name AS category_name, products.description, 
	products.image_url, products.price, products.stock 
	FROM categories 
	JOIN products ON categories.id = products.category_id 
	WHERE products.id = ?`

	if err := database.Db.Raw(sql, id).Scan(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "product updated successfully",
		"products": products,
	})
}

func Delete_Product(c *gin.Context) {

	id := c.Param("id")
	var product models.Product

	if err := database.Db.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "product not found"})
		return
	}

	if err := database.Db.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete product",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "product deleted successfully",
	})
}
