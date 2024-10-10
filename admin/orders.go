package admin

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func Admin_ListOrders(c *gin.Context) {
	var orders []models.Order

	query := database.Db.Preload("Items.Product")

	if err := query.Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch orders"})
		return
	}

	var responseOrders []responsemodels.Order
	for _, order := range orders {
		var responseItems []responsemodels.Order_Item
		for _, item := range order.Items {
			orderItem := responsemodels.Order_Item{
				Order_Id:       item.Order_Id,
				Product_Id:     item.Product_Id,
				Product_Name:   item.Product.Name,
				Quantity:       item.Quantity,
				Price:          int(item.Price),
				Discount_price: item.Offer_price,
				Status:         item.Status,
			}
			responseItems = append(responseItems, orderItem)
		}

		responseOrder := responsemodels.Order{
			Id:             int(order.ID),
			User_Id:        order.User_Id,
			Items:          responseItems,
			Total_Price:    order.Total_Price,
			Order_Status:   order.Order_Status,
			Payment_Status: order.Payment_Status,
			Payment_Method: order.Payment_Method,
			Discount_price: order.Final_Price,
		}
		responseOrders = append(responseOrders, responseOrder)
	}

	c.JSON(http.StatusOK, gin.H{"orders": responseOrders})
}

func Admin_ChangeOrderStatus(c *gin.Context) {

	order_id := c.Param("id")

	var status models.Order_status

	if err := c.ShouldBindJSON(&status); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
		return
	}

	var order models.Order
	if err := database.Db.Preload("Items").Where("id = ?", order_id).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Order_Status != "confirmed" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is not confirmed "})
		return
	}

	if status.Status == "completed" {
		order.Order_Status = status.Status
		if order.Payment_Method == "COD" {
			order.Payment_Status = "Confirmed"
		}
		for _, item := range order.Items {
			item.Status = "completed"
			database.Db.Save(&item)
		}

	} else {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid status"})
		return
	}

	if err := database.Db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order status updated successfully",
	})
}

func Admin_CancelOrder(c *gin.Context) {
	order_id := c.Param("id")

	var order models.Order
	if err := database.Db.Where("id = ?", order_id).Preload("Items").First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Order_Status == "Canceled" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order is already canceled"})
		return
	}

	for _, item := range order.Items {
		var product models.Product
		if err := database.Db.Where("id = ?", item.Product_Id).First(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
			return
		}

		product.Stock += item.Quantity
		if err := database.Db.Save(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	order.Order_Status = "Canceled"
	if err := database.Db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order canceled successfully",
	})
}

func Return_item(c *gin.Context) {
	var item struct {
		OrderId   uint `json:"order_id" binding:"required"`
		ProductId uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderItem models.Order_Item
	if err := database.Db.Where("product_id = ? AND order_id = ?", item.ProductId, item.OrderId).First(&orderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in the specified order"})
		return
	}

	if orderItem.Status != "return_initiated" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid product return"})
		return
	}

	var product models.Product
	if err := database.Db.Where("id = ?", item.ProductId).First(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
		return
	}

	orderItem.Status = "return_confirmed"

	if err := database.Db.Model(&orderItem).Update("status", "return_confirmed").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status to return_confirmed"})
		return
	}

	product.Stock += orderItem.Quantity
	if err := database.Db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}

	var order models.Order
	if err := database.Db.Preload("Items").Where("id = ? ", item.OrderId).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	var wallet models.Wallet
	if err := database.Db.Where("user_id = ?", order.User_Id).First(&wallet).Error; err != nil {

		wallet = models.Wallet{User_Id: order.User_Id, Balance: 0}
		if err := database.Db.Create(&wallet).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
			return
		}
	}
	var deduction_amount float32
	var refund_amount float32


	updated_total := order.Total_Price - (orderItem.Price * float32(orderItem.Quantity))

	if order.Coupon_code != "" {
		var coupon models.Coupon
		if err := database.Db.Where("code = ?", order.Coupon_code).First(&coupon).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "coupon details not found "})
			return
		}

		if updated_total < float32(coupon.Min_Purchase) {
			for _, item := range order.Items {
				item.Offer_price = item.Price
				database.Db.Save(&item)
			}
			deduction_amount = order.Final_Price - updated_total
			order.Total_Price = updated_total
			order.Final_Price = updated_total

			refund_amount=deduction_amount
		} else {
			deduction_amount = (orderItem.Offer_price * float32(orderItem.Quantity))
			order.Total_Price = updated_total
			order.Final_Price -= deduction_amount

			refund_amount=deduction_amount
		}
	} else {
		deduction_amount = order.Total_Price - updated_total
		order.Total_Price = updated_total
		order.Final_Price=updated_total

		refund_amount=deduction_amount
	}
	database.Db.Save(&order)

	wallet.Balance += refund_amount

	if err := database.Db.Save(&wallet).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet balance"})
		return
	}

	transaction := models.Transaction{
		Wallet_Id: int(wallet.ID),
		Amount:    refund_amount,
		Type:      "credit",
	}

	if err := database.Db.Create(&transaction).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refund transaction"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":       "refund added to wallet",
		"refund_amount": deduction_amount,
	})
}

func Confirm_Order(c *gin.Context) {
	order_id := c.Param("id")

	var order models.Order

	if err := database.Db.Where("id = ?", order_id).Preload("Items").First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if len(order.Items) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No items found in this order"})
		return
	}
	order.Order_Status = "confirmed"
	database.Db.Save(&order)

	for _, item := range order.Items {
		if err := database.Db.Model(&item).Where("id = ?", item.ID).Update("status", "shipped").Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update order item status"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order confirmed and items marked as shipped",
	})
}
