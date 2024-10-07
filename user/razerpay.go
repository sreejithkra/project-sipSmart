package user

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/razorpay/razorpay-go"
)

var Temp_Order_Id uint

func CreaterazorpayOrder(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", uint(user_id)).Preload("Items.Product").First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Cart not found"})
		return
	}

	if len(cart.Items) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cart is empty"})
		return
	}

	for _, cart_item := range cart.Items {

		var product models.Product
		if err := database.Db.Where("id = ?", cart_item.Product_Id).First(&product).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
			return
		}

		if product.Stock < cart_item.Quantity {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient stock for product: " + product.Name})
			return
		}

	}

	client := razorpay.NewClient("rzp_test_fc8qTH1YMOywRn", "l0EMEeYVmuB10wu9Up2g5Rls")
	amount := int(cart.Discount_price * 100)
	data := map[string]interface{}{
		"amount":   amount,
		"currency": "INR",
		"receipt":  "receipt#" + strconv.Itoa(int(cart.ID)),
	}

	razor, err := client.Order.Create(data, nil)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Razorpay order creation failed", "details": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"message": "razorpay payment intiated successfully",
		"order":   razor,
	})
}

func PaymentCheck(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found"})
		return
	}

	type paymentResponse struct {
		RazorpayPaymentID string `json:"razorpay_payment_id"`
		RazorpayOrderID   string `json:"razorpay_order_id"`
		RazorpaySignature string `json:"razorpay_signature"`
	}

	var response paymentResponse

	if err := c.ShouldBindJSON(&response); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if VerifyRazorpaySignature(response.RazorpayOrderID, response.RazorpayPaymentID, response.RazorpaySignature) {
		c.JSON(http.StatusOK, gin.H{"status": "success", "message": "Payment verified successfully"})

		var cart models.Cart
		if err := database.Db.Preload("Items").Where("user_id = ?", user_id).First(&cart).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "item details not found"})
			return
		}

		order := models.Order{
			User_Id:        cart.User_Id,
			Total_Price:    cart.Total_Price,
			Discount_price: cart.Discount_price,
			Order_Status:   "Pending",
			Payment_Status: "Confirmed",
			Payment_Method: "Razorpay",
			Coupon_code:    cart.Coupon_code,
		}

		if err := database.Db.Create(&order).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create order"})
			return
		}

		for _, cart_item := range cart.Items {
			order_item := models.Order_Item{
				Order_Id:       int(order.ID),
				Product_Id:     cart_item.Product_Id,
				Quantity:       cart_item.Quantity,
				Price:          cart_item.Price,
				Discount_price: cart_item.Discount_price,
				Status:         "pending",
			}

			if err := database.Db.Create(&order_item).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add order item"})
				return
			}

			var product models.Product
			if err := database.Db.Where("id = ?", order_item.Product_Id).First(&product).Error; err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Product not found"})
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

		responseOrder := responsemodels.Order{
			Id:             int(order.ID),
			User_Id:        order.User_Id,
			Total_Price:    order.Total_Price,
			Discount_price: order.Discount_price,
			Payment_Status: order.Payment_Status,
			Order_Status:   "Pending",
			Payment_Method: order.Payment_Method,
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

	} else {
		c.JSON(http.StatusUnauthorized, gin.H{"status": "failure", "message": "Payment verification failed"})
	}
}

func VerifyRazorpaySignature(orderID, paymentID, razorpaySignature string) bool {

	secret := "l0EMEeYVmuB10wu9Up2g5Rls"
	message := orderID + "|" + paymentID

	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(message))
	expectedSignature := hex.EncodeToString(h.Sum(nil))

	return expectedSignature == razorpaySignature
}

func CORSMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	}
}
