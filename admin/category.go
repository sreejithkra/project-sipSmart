package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func List_Categories(c *gin.Context) {

	var categories []responsemodels.Category

	result := database.Db.Model(&models.Category{}).
		Select("id, name, description, image_url").
		Find(&categories)

	if result.Error != nil {
		c.JSON(500, gin.H{"error": "Failed to retrieve categories"})
		return
	}

	c.JSON(200, gin.H{
		"catagories": categories,
	})
}

func Create_Category(c *gin.Context) {

	var category_Info models.Category_Info

	if err := c.ShouldBindJSON(&category_Info); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
		return
	}

	err := helper.Validate(category_Info)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	var count int64

	if err := database.Db.Model(&models.Category{}).Where("name = ?", category_Info.Name).Count(&count).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database operation failed"})
		return
	}
	if count != 0 {
		c.JSON(http.StatusConflict, gin.H{"error": "Record exists with the given category name"})
		return
	}

	category := models.Category{
		Name:        category_Info.Name,
		Description: category_Info.Description,
		Image_Url:   category_Info.Image_Url,
	}

	if err := database.Db.Create(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Database operation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message":  "Category created successfully",
		"category": category,
	})
}

func Update_Category(c *gin.Context) {

	// var category_Info models.Category_Info
	var category models.Category
	id := c.Param("id")

	if err := database.Db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := c.ShouldBindJSON(&category); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format", "details": err.Error()})
		return
	}

	err := helper.Validate(category)
	if err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if err := database.Db.Save(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "Database operation failed",
			"details": err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "Category updated successfully",
		"category": category,
	})
}

func Delete_Category(c *gin.Context) {

	id := c.Param("id")
	var category models.Category

	if err := database.Db.First(&category, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Category not found"})
		return
	}

	if err := database.Db.Delete(&category).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to delete category",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Category deleted successfully",
	})
}
