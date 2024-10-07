package user

import (
	"fmt"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func Create_Order(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var cart models.Cart
	if err := database.Db.Preload("Items.Product").Where("user_id = ?", user_id).First(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "item details not found"})
		return
	}

	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	if cart.Discount_price > 1500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Order above Rs 1000 should not be allowed for COD"})
		return
	}

	order := models.Order{
		User_Id:        cart.User_Id,
		Total_Price:    cart.Total_Price,
		Discount_price: cart.Discount_price,
		Order_Status:   "Pending",
		Payment_Status: "Pending",
		Payment_Method: "COD",
		Coupon_code:    cart.Coupon_code,
	}

	for _, cart_item := range cart.Items {
		order_item := models.Order_Item{
			Product_Id:     cart_item.Product_Id,
			Quantity:       cart_item.Quantity,
			Price:          cart_item.Price,
			Discount_price: cart_item.Discount_price,
			Status:         "pending",
		}

		order.Items = append(order.Items, order_item)

		var product models.Product
		if err := database.Db.Where("id = ?", cart_item.Product_Id).First(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
			return
		}

		if product.Stock < cart_item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for product: " + product.Name})
			return
		}

		product.Stock -= order_item.Quantity
		product.Sales += order_item.Quantity
		helper.UpdateProductPopularity(&product)

		if err := database.Db.Save(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
			return
		}
	}

	if err := database.Db.Create(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
		return
	}

	var responseItems []responsemodels.Order_Item

	for _, item := range order.Items {
		orderItem := responsemodels.Order_Item{
			Order_Id:       item.Order_Id,
			Product_Id:     item.Product_Id,
			Product_Name:   item.Product.Name,
			Quantity:       item.Quantity,
			Price:          int(item.Price),
			Discount_price: item.Discount_price,
			Status:         item.Status,
		}
		responseItems = append(responseItems, orderItem)
	}

	responseOrder := responsemodels.Order{
		Id:             int(order.ID),
		User_Id:        order.User_Id,
		Total_Price:    order.Total_Price,
		Discount_price: order.Discount_price,
		Payment_Status: order.Payment_Status,
		Order_Status:   order.Order_Status,
		Payment_Method: order.Payment_Method,
		Items:          responseItems,
	}

	c.JSON(http.StatusOK, gin.H{"message": "Order created successfully", "order": responseOrder})

	if err := database.Db.Where("cart_id = ?", cart.ID).Delete(&models.Cart_Item{}).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to clear cart items"})
		return
	}

	cart.Total_Price = 0
	cart.Coupon_code = ""
	cart.Discount_price = 0

	if err := database.Db.Save(&cart).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}
}

func List_Orders(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var orders []models.Order
	if err := database.Db.Where("user_id = ?", user_id).Preload("Items.Product").Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	if len(orders) == 0 {
		c.JSON(http.StatusNotFound, gin.H{"error": "No orders found"})
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
				Discount_price: item.Discount_price,
				Status:         item.Status,
			}
			responseItems = append(responseItems, orderItem)
		}

		responseOrder := responsemodels.Order{
			Id:             int(order.ID),
			User_Id:        order.User_Id,
			Items:          responseItems,
			Total_Price:    order.Total_Price,
			Discount_price: order.Discount_price,
			Payment_Status: order.Payment_Status,
			Order_Status:   order.Order_Status,
			Payment_Method: order.Payment_Method,
		}
		responseOrders = append(responseOrders, responseOrder)
	}

	c.JSON(http.StatusOK, gin.H{"orders": responseOrders})
}

// func Cancel_Order(c *gin.Context) {

// 	user_id := helper.Get_Userid(c)
// 	if user_id == 0 {
// 		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
// 		return
// 	}

// 	order_id := c.Param("id")

// 	var order models.Order
// 	if err := database.Db.Where("id = ? AND user_id = ?", order_id, user_id).First(&order).Error; err != nil {
// 		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
// 		return
// 	}

// 	if order.Order_Status != "Pending" {
// 		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be canceled"})
// 		return
// 	}

// 	for _, item := range order.Items {
// 		var product models.Product
// 		if err := database.Db.Where("id = ?", item.Product_Id).First(&product).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
// 			return
// 		}

// 		product.Stock += item.Quantity
// 		if err := database.Db.Save(&product).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
// 			return
// 		}
// 	}

// 	order.Order_Status = "canceled"
// 	if err := database.Db.Save(&order).Error; err != nil {
// 		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
// 		return
// 	}

// 	if order.Payment_Status == "Confirmed" {

// 		var wallet models.Wallet
// 		if err := database.Db.Where("user_id = ?", order.User_Id).First(&wallet).Error; err != nil {

// 			wallet = models.Wallet{User_Id: order.User_Id, Balance: 0}
// 			if err := database.Db.Create(&wallet).Error; err != nil {
// 				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
// 				return
// 			}

// 		}

// 		wallet.Balance += order.Discount_price
// 		c.JSON(http.StatusOK, gin.H{
// 			"message": "refund processed",
// 			"refund":  order.Discount_price,
// 		})

// 		if err := database.Db.Save(&wallet).Error; err != nil {
// 			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet balance"})
// 			return
// 		}
// 	}

// 	c.JSON(http.StatusOK, gin.H{
// 		"message": "Order canceled successfully",
// 	})
// }

func Cancel_Order(c *gin.Context) {

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	order_id := c.Param("id")

	var order models.Order
	if err := database.Db.Where("id = ? AND user_id = ?", order_id, user_id).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	if order.Order_Status != "Pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Only pending orders can be canceled"})
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

	order.Order_Status = "canceled"
	if err := database.Db.Save(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to cancel order"})
		return
	}

	if order.Payment_Status == "Confirmed" {

		var wallet models.Wallet
		if err := database.Db.Where("user_id = ?", order.User_Id).First(&wallet).Error; err != nil {
			wallet = models.Wallet{User_Id: order.User_Id, Balance: 0}
			if err := database.Db.Create(&wallet).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
				return
			}
		}

		wallet.Balance += order.Discount_price
		if err := database.Db.Save(&wallet).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet balance"})
			return
		}

		transaction := models.Transaction{
			Wallet_Id: int(wallet.ID),
			Amount:    order.Discount_price,
			Type:      "credit",
		}

		if err := database.Db.Create(&transaction).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create refund transaction"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message": "Refund processed successfully",
			"refund":  order.Discount_price,
		})

		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Order canceled successfully",
	})
}

func Cancel_Orderitem(c *gin.Context) {

	var cancelRequest struct {
		OrderId   uint `json:"order_id" binding:"required"`
		ProductId uint `json:"product_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&cancelRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var order models.Order

	if err := database.Db.Preload("Items").Where("id = ?", cancelRequest.OrderId).First(&order).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "order item details not found"})
		return
	}

	var orderItem models.Order_Item
	if err := database.Db.Where("product_id = ? AND order_id = ?", cancelRequest.ProductId, cancelRequest.OrderId).First(&orderItem).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in the specified order"})
		return
	}

	if orderItem.Status != "pending" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "only pending items have this cancel option"})
		return
	}
	fmt.Println(orderItem.Status)

	orderItem.Status = "canceled"
	fmt.Println(orderItem)
	fmt.Println(orderItem.Status)

	if err := database.Db.Save(&orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update status to canceled"})
		return
	}

	var product models.Product
	if err := database.Db.Where("id = ?", cancelRequest.ProductId).First(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
		return
	}
	product.Stock += orderItem.Quantity
	if err := database.Db.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update product stock"})
		return
	}
	var deduction_amount float32

	order.Total_Price -= (orderItem.Price * float32(orderItem.Quantity))

	if order.Total_Price == 0 {
		order.Order_Status = "canceled"
	}

	if order.Coupon_code != "" {
		var coupon models.Coupon
		if err := database.Db.Where("code = ?", order.Coupon_code).First(&coupon).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "coupon details not found "})
			return
		}

		if order.Total_Price < float32(coupon.Min_Purchase) {
			for _, item := range order.Items {
				item.Discount_price = item.Price
				database.Db.Save(&item)
			}
			deduction_amount = order.Discount_price - order.Total_Price
			order.Discount_price = order.Total_Price
		} else {
			deduction_amount = (orderItem.Discount_price * float32(orderItem.Quantity))

			order.Discount_price -= deduction_amount

		}

	} else {
		deduction_amount = order.Discount_price - order.Total_Price
		order.Discount_price = order.Total_Price
	}
	database.Db.Save(&order)
	if order.Payment_Status == "Confirmed" {
		var wallet models.Wallet
		if err := database.Db.Where("user_id = ?", order.User_Id).First(&wallet).Error; err != nil {

			wallet = models.Wallet{User_Id: order.User_Id, Balance: 0}
			if err := database.Db.Create(&wallet).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create wallet"})
				return
			}
		}
		wallet.Balance += deduction_amount

		if err := database.Db.Save(&wallet).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update wallet balance"})
			return
		}

		transaction := models.Transaction{
			Wallet_Id: int(wallet.ID),
			Amount:    deduction_amount,
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

	c.JSON(http.StatusOK, gin.H{
		"message": "item canceled successfully",
	})
}

func Return_Product(c *gin.Context) {

	type ReturnProductRequest struct {
		OrderID   int `json:"order_id" binding:"required"`
		ProductID int `json:"product_id" binding:"required"`
	}

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var req ReturnProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request data"})
		return
	}

	var orderItem models.Order_Item

	sql := `SELECT order_items.id, order_items.order_id, products.id AS product_id, products.name AS name, products.price, orders.order_status AS order_status
	        FROM orders
	        JOIN order_items ON orders.id = order_items.order_id
	        JOIN products ON order_items.product_id = products.id
	        WHERE orders.user_id = ? AND orders.id = ? AND products.id = ? AND orders.order_status = 'completed'`

	if err := database.Db.Raw(sql, user_id, req.OrderID, req.ProductID).Scan(&orderItem).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found or not eligible for return"})
		return
	}

	orderItem.Status = "return_initiated"

	if err := database.Db.Model(&orderItem).Where("id = ?", orderItem.ID).Update("status", "return_initiated").Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found or not eligible for return"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "Product returned and refund initiated successfully",
	})
}
