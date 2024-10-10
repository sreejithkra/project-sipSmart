package main

import (
	"sip_Smart/admin"
	"sip_Smart/database"
	"sip_Smart/helper"
	"sip_Smart/user"

	"github.com/gin-gonic/gin"
)

var c *gin.Context

func main() {

	router := gin.Default()

	database.Initialize(c)

	router.LoadHTMLGlob("templates/*")

	router.GET("/categories", admin.List_Categories)
	router.GET("/products", admin.List_Products)
	router.GET("/products/:id", user.Product_Details)

	router.POST("/user/signup", user.Signup_User)
	router.POST("/user/signup/validate_otp", helper.DeleteTempUserMiddleware(), helper.Validate_otp)

	router.GET("/auth/google", user.GoogleLogin)
	router.GET("/auth/google/callback", user.GoogleCallback)

	router.POST("/admin", admin.Admin_Login)
	authorized := router.Group("/admin")
	authorized.Use(admin.Adminauth_Middleware())
	{
		authorized.GET("/users", admin.List_Users)
		authorized.PATCH("/blockuser/:id", admin.Block_User)
		authorized.PATCH("/unblockuser/:id", admin.Unblock_User)

		authorized.POST("/create_category", admin.Create_Category)
		authorized.PUT("/update_category/:id", admin.Update_Category)
		authorized.DELETE("/delete_category/:id", admin.Delete_Category)

		authorized.POST("/create_product", admin.Create_Product)
		authorized.PUT("/update_product/:id", admin.Update_Product)
		authorized.DELETE("/delete_product/:id", admin.Delete_Product)

		authorized.POST("/create_coupon", admin.Create_Coupon)
		authorized.DELETE("/delete_coupon/:id", admin.Delete_Coupon)
		authorized.GET("/list_coupons", admin.List_Coupons)

		authorized.POST("/create_productoffer", admin.Create_Productoffer)

		authorized.GET("/orders", admin.Admin_ListOrders)
		authorized.PATCH("/edit_order/:id", admin.Admin_ChangeOrderStatus)
		authorized.PATCH("/cancel_order/:id", admin.Admin_CancelOrder)

		authorized.PATCH("/confirm_order/:id", admin.Confirm_Order)

		authorized.PATCH("/verify_return", admin.Return_item)

		authorized.GET("/sales", admin.Sales_Report)

		authorized.GET("/sales_pdf", admin.Salespdf)
		authorized.GET("/sales_excel", admin.SalesExcel)

		authorized.GET("/best_selling", admin.BestSellings)

	}

	router.POST("/user", user.User_Login)
	authuser := router.Group("/user")
	authuser.Use(user.Userauth_Middleware())
	{

		authuser.GET("/list_address", user.List_address)
		authuser.POST("/add_address", user.Add_address)
		authuser.PUT("/edit_address/:id", user.Edit_address)
		authuser.PUT("/change_default_address/:id", user.Change_default_address)

		authuser.DELETE("/delete_address/:id", user.Delete_address)

		authuser.POST("/add_tocart", user.Add_Tocart)
		authuser.GET("/list_cart", user.List_Cartitems)
		authuser.PUT("/edit_cart", user.Remove_Fromcart)

		authuser.POST("/add_towishlist/:id", user.AddToWishlist)
		authuser.GET("/list_wishlist", user.ListWishlist)
		authuser.PUT("/edit_wishlist/:id", user.RemoveFromWishlist)

		authuser.GET("/profile", user.User_Profile)
		authuser.PUT("/edit_profile", user.Edit_Profile)

		authuser.PATCH("/apply_coupon", user.Apply_Coupon)
		authuser.PATCH("/remove_coupon", user.Remove_Coupon)

		authuser.POST("/create_order/cod", user.Create_Order)
		authuser.GET("/list_orders", user.List_Orders)
		authuser.PATCH("/cancel_order/:id", user.Cancel_Order)
		authuser.PATCH("/cancel_orderitem", user.Cancel_Orderitem)
		authuser.PATCH("/return_orderitem", user.Return_Product)

		authuser.POST("/create_razorpayorder", user.CreaterazorpayOrder)
		authuser.POST("/verify_payment", user.PaymentCheck)

		authuser.POST("/add_review", user.Add_Review)

		authuser.GET("/filter", user.Category_Filter)

		authuser.GET("/wallet", user.Get_Wallet)

		authuser.GET("/invoicepdf", user.GenerateInvoicePDF)

	}

	router.Run(":8086")
}
