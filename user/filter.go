package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func SearchProducts(c *gin.Context) {

	sortBy := c.Query("sort_by")
	search := c.Query("search")

	var products []responsemodels.Product

	sql := `SELECT products.id, products.name AS name, categories.name AS category, 
	        products.description, products.image_url, products.price, products.stock, products.offer_price, products.popularity
	        FROM categories 
	        JOIN products ON categories.id = products.category_id`

	if search != "" {
		sql += " WHERE products.name ILIKE '%" + search + "%'"
	}

	switch sortBy {
	case "popularity":
		sql += " ORDER BY products.popularity DESC"
	case "price_asc":
		sql += " ORDER BY products.price ASC"
	case "price_desc":
		sql += " ORDER BY products.price DESC"
	case "ratings":
		sql += " ORDER BY products.average_rating DESC"
	case "a-z":
		sql += " ORDER BY products.name ASC"
	case "z-a":
		sql += " ORDER BY products.name DESC"
	default:
		sql += " ORDER BY products.created_at DESC"
	}

	database.Db.Raw(sql).Scan(&products)

	c.JSON(200, gin.H{
		"products": products,
	})
}

func Category_Filter(c *gin.Context) {
	categoryName := c.Query("category")

	if categoryName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Valid category name is required"})
		return
	}

	var products []responsemodels.Product

	sql := `SELECT products.id, products.name AS name, categories.name AS category, 
	        products.description, products.image_url, products.price, products.stock, products.offer_price, products.popularity
	        FROM categories 
	        JOIN products ON categories.id = products.category_id
	        WHERE categories.name = ?`

	if err := database.Db.Raw(sql, categoryName).Scan(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve products"})
		return
	}

	if len(products) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No products found for this category"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"products": products,
	})
}
