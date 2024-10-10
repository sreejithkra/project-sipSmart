package user

import (
	"net/http"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/models"
	"sip_Smart/responsemodels"

	"github.com/gin-gonic/gin"
)

func Add_Tocart(c *gin.Context) {
	max_quantity := 10

	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	var cart_item models.Cart_Item
	if err := c.ShouldBindJSON(&cart_item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", user_id).First(&cart).Error; err != nil {
		cart = models.Cart{User_Id: user_id, Total_Price: 0, Final_Price: 0}
		database.Db.Create(&cart)
	}

	var totalQuantity int
	database.Db.Model(&models.Cart_Item{}).Where("cart_id = ?", cart.ID).Select("COALESCE(SUM(quantity), 0)").Scan(&totalQuantity)

	if totalQuantity+cart_item.Quantity > max_quantity {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Adding this item exceeds the maximum allowed quantity of 10 items in the cart"})
		return
	}

	var product models.Product
	if err := database.Db.Where("id = ?", cart_item.Product_Id).First(&product).Error; err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product not found"})
		return
	}

	if cart_item.Quantity > product.Stock {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Product stock not enough"})
		return
	}

	cart_item.Cart_Id = int(cart.ID)
	cart_item.Price = float32(product.Offer_Price)
	cart_item.Offer_Price = float32(product.Offer_Price)

	var existing_cart_item models.Cart_Item
	if err := database.Db.Where("cart_id = ? AND product_id = ?", cart.ID, cart_item.Product_Id).First(&existing_cart_item).Error; err == nil {
		existing_cart_item.Quantity += cart_item.Quantity

		if err := database.Db.Save(&existing_cart_item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update cart item"})
			return
		}
	} else {
		if err := database.Db.Create(&cart_item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add item to cart"})
			return
		}

	}
	cart.Total_Price += cart_item.Price * float32(cart_item.Quantity)

	cart.Final_Price = cart.Total_Price
	database.Db.Save(&cart)

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
}

func List_Cartitems(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", user_id).Preload("Items.Product").First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "cart is empty add products"})
		return
	}

	var cartResponse responsemodels.Cart
	cartResponse.ID = int(cart.ID)
	cartResponse.TotalPrice = cart.Total_Price
	cartResponse.Discount_price = cart.Final_Price

	for _, item := range cart.Items {
		cartItemResponse := responsemodels.Cart_Item{
			Product_Id:     item.Product_Id,
			Product_Name:   item.Product.Name,
			Quantity:       item.Quantity,
			Price:          item.Price,
			Discount_price: item.Offer_Price,
		}
		cartResponse.Items = append(cartResponse.Items, cartItemResponse)
	}

	c.JSON(http.StatusOK, gin.H{
		"cart": cartResponse,
	})
}

func Remove_Fromcart(c *gin.Context) {
	user_id := helper.Get_Userid(c)
	if user_id == 0 {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User ID not found from claims"})
		return
	}

	var cart_item models.Cart_Item
	if err := c.ShouldBindJSON(&cart_item); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON data"})
		return
	}

	var cart models.Cart
	if err := database.Db.Where("user_id = ?", uint(user_id)).First(&cart).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"message": "No cart found for the user"})
		return
	}

	quantity := cart_item.Quantity

	if err := database.Db.Where("cart_id = ? AND product_id = ?", cart.ID, cart_item.Product_Id).First(&cart_item).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Item not found in cart"})
		return
	}

	if cart_item.Quantity-quantity == 0 {
		database.Db.Delete(&cart_item)
	} else if cart_item.Quantity-quantity > 0 {
		cart_item.Quantity = cart_item.Quantity - quantity
		database.Db.Save(&cart_item)
	} else {
		c.JSON(400, gin.H{"error": "enter valid quatity to remove"})
		return
	}

	cart.Total_Price -= (cart_item.Price * float32(quantity))

	if cart.Coupon_code != "" {

		var coupon models.Coupon
		if err := database.Db.Where("code = ?", cart.Coupon_code).First(&coupon).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"message": "No coupons found"})
		}
		if cart.Total_Price < float32(coupon.Min_Purchase) {
			cart_item.Offer_Price = cart_item.Price
			database.Db.Save(&cart_item)
			cart.Final_Price = float32(cart.Total_Price)
			cart.Coupon_code = ""
			c.JSON(http.StatusOK, gin.H{"message": "offer coupon removed"})
		} else {
			cart.Final_Price -= (cart_item.Offer_Price * float32(quantity))
		}
	} else {
		cart.Final_Price = float32(cart.Total_Price)
	}

	database.Db.Save(&cart)

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart successfully"})
}
