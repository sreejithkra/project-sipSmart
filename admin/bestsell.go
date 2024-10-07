package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"
	"sip_Smart/responsemodels"
	"sort"
	"time"

	"github.com/gin-gonic/gin"
)

func BestSellings(c *gin.Context) {

	filter := c.Query("filter")
	var startDate, endDate time.Time

	// Set time range based on filter (daily, weekly, yearly)
	switch filter {
	case "daily":
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)

	case "weekly":
		startDate = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "yearly":
		startDate = time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endDate = time.Now()

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid filter option"})
		return
	}

	var orders []models.Order
	if err := database.Db.Preload("Items.Product.Category").
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	productSalesMap := make(map[int]int)
	categorySalesMap := make(map[int]int)

	for _, order := range orders {
		for _, item := range order.Items {
			productSalesMap[item.Product_Id] += item.Quantity
			categorySalesMap[item.Product.Category_Id] += item.Quantity
		}
	}

	type ProductSales struct {
		ProductId int
		Sales     int
	}
	var productSalesList []ProductSales
	for productId, sales := range productSalesMap {
		productSalesList = append(productSalesList, ProductSales{ProductId: productId, Sales: sales})
	}

	sort.Slice(productSalesList, func(i, j int) bool {
		return productSalesList[i].Sales > productSalesList[j].Sales
	})

	topProducts := productSalesList
	if len(productSalesList) > 10 {
		topProducts = productSalesList[:10]
	}

	var topSellingProducts []responsemodels.Best_Product
	for _, productSales := range topProducts {
		var product models.Product
		if err := database.Db.Where("id = ?", productSales.ProductId).First(&product).Error; err == nil {
			product.Stock = productSales.Sales
			topSellingProducts = append(topSellingProducts, responsemodels.Best_Product{
				Id:          productSales.ProductId,
				ProductName: product.Name,
				Category_id: product.Category_Id,
				SalesCount:  (productSales.Sales) + 1,
			})
		}
	}

	type CategorySales struct {
		CategoryId int
		Sales      int
	}
	var categorySalesList []CategorySales
	for categoryId, sales := range categorySalesMap {
		categorySalesList = append(categorySalesList, CategorySales{CategoryId: categoryId, Sales: sales})
	}

	sort.Slice(categorySalesList, func(i, j int) bool {
		return categorySalesList[i].Sales > categorySalesList[j].Sales
	})

	topCategories := categorySalesList
	if len(categorySalesList) > 10 {
		topCategories = categorySalesList[:10]
	}

	var topSellingCategories []responsemodels.Best_Category
	for _, categorySales := range topCategories {
		var category models.Category
		if err := database.Db.Where("id = ?", categorySales.CategoryId).First(&category).Error; err == nil {
			topSellingCategories = append(topSellingCategories, responsemodels.Best_Category{
				Id:           categorySales.CategoryId,
				CategoryName: category.Name,
				SalesCount:   (categorySales.Sales) + 1,
			})
		}
	}

	c.JSON(200, gin.H{
		"top_selling_products":   topSellingProducts,
		"top_selling_categories": topSellingCategories,
	})
}
