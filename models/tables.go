package models

import (
	"time"

	"gorm.io/gorm"
)

type Admin_Credentials struct {
	Email    string `json:"email" binding:"required" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required"`
}

type Temp_User struct {
	Name     string `gorm:"size:15;index" json:"name" binding:"required" validate:"required,max=15"`
	Phone    string `gorm:"size:10;uniqueIndex" json:"phone" binding:"required" validate:"required,len=10,numeric"`
	Email    string `gorm:"size:50;uniqueIndex" json:"email" binding:"required" validate:"required,email"`
	Password string `gorm:"size:20" json:"password" binding:"required" validate:"required,min=8"`
}

type User struct {
	gorm.Model
	Name          string `gorm:"size:15;index;not null"`
	Phone         string `gorm:"uniqueIndex"`
	Email         string `gorm:"uniqueIndex;not null"`
	Password      string `gorm:"size:25"`
	Status        string `gorm:"default:'Active'"`
	Signup_Method string `gorm:"size:10;default:'manual'"`
}

type Category struct {
	gorm.Model
	Name        string `gorm:"size:15;not null"`
	Description string `gorm:"size:100;type:text"`
	Image_Url   string `gorm:"size:255"`
}

type Product struct {
	gorm.Model
	Category_Id    int     `gorm:"not null;index"`
	Name           string  `gorm:"size:255;index;not null"`
	Description    string  `gorm:"type:text"`
	Price          float64 `gorm:"type:decimal(10,2);not null"`
	Stock          int     `gorm:"type:decimal"`
	Image_Url      string  `gorm:"size:255"`
	Average_Rating float64 `gorm:"type:decimal(10,2);default:0"`
	Reviews        []Review
	Sales          int     `gorm:"type:decimal;default:0"`
	Views          int     `gorm:"type:decimal;default:0"`
	Popularity     float64 `gorm:"type:decimal(10,2);default:0"`
	Offer_Price    float64 `gorm:"type:decimal(10,2);not null"`

	// Relation
	Category Category `gorm:"foreignKey:Category_Id;references:ID"`
}

type Product_Offer struct {
	gorm.Model
	Product_Id int       `json:"product_id" binding:"required"  validate:"required"`
	Discount   float32   `json:"discount" binding:"required" validate:"required,gt=0,lte=100"`
	Status     bool      `json:"status" binding:"required" `
	ExpiryDate time.Time `json:"expiry_date" binding:"required" validate:"required"`

	// Relation
	Product Product `gorm:"foreignKey:Product_Id;references:ID"`
}

type Review struct {
	gorm.Model
	User_Id    int     `json:"user_id"`
	Product_Id int     `json:"product_id" binding:"required"`
	Rating     float32 `json:"rating" binding:"required" validate:"required,max=5,min=1"`
	Content    string  `json:"review" binding:"required"`

	// Relation
	User    User    `gorm:"foreignKey:User_Id;references:ID"`
	Product Product `gorm:"foreignKey:Product_Id;references:ID"`
}

type Address struct {
	gorm.Model
	User_Id  int    `json:"user_id" gorm:"not null"`
	Country  string `json:"country" gorm:"size:15" validate:"required,max=15"`
	State    string `json:"state" gorm:"size:15" validate:"required,max=15"`
	District string `json:"district" gorm:"size:15" validate:"required,max=15"`
	Street   string `json:"street_name" gorm:"size:15" validate:"required,max=15"`
	Pincode  string `json:"pincode" gorm:"type:decimal(6,0)" validate:"required,numeric,len=6"`
	Phone    string `json:"phone_number" gorm:"type:decimal(10,0)" validate:"required,numeric,len=10"`
	Default  bool   `json:"default" gorm:"not null"`

	// Relation
	User User `gorm:"foreignKey:User_Id;references:ID"`
}

type Cart struct {
	gorm.Model
	User_Id        int         `json:"user_id"`
	Items          []Cart_Item `json:"items"`
	Total_Price    float32     `json:"total_price"`
	Coupon_code    string      `json:"Coupon_code"`
	Discount_price float32     `json:"discount_price"`

	// Relation
	User User `gorm:"foreignKey:User_Id;references:ID"`
}

type Cart_Item struct {
	gorm.Model
	Cart_Id        int     `json:"cart_id"`
	Product_Id     int     `json:"product_id" validate:"required"`
	Quantity       int     `json:"quantity" validate:"required"`
	Price          float32 `json:"price"`
	Discount_price float32 `json:"discount_price"`

	// Relation
	Cart    Cart    `gorm:"foreignKey:Cart_Id;references:ID"`
	Product Product `gorm:"foreignKey:Product_Id;references:ID"`
}

type Order struct {
	gorm.Model
	User_Id        int          `json:"user_id"`
	Items          []Order_Item `json:"items"`
	Total_Price    float32      `json:"total_price"`
	Payment_Status string       `json:"payment_status"`
	Order_Status   string       `json:"order_status"`
	Payment_Method string       `json:"payment_method"`
	Discount_price float32      `json:"discount_price"`
	Coupon_code    string

	// Relation
	User User `gorm:"foreignKey:User_Id;references:ID"`
}

type Order_Item struct {
	gorm.Model
	Order_Id       int     `json:"order_id"`
	Product_Id     int     `json:"product_id"`
	Quantity       int     `json:"quantity"`
	Price          float32 `json:"price"`
	Discount_price float32 `json:"discount_price"`
	Status         string  `json:"status"`

	// Relation
	Order   Order   `gorm:"foreignKey:Order_Id;references:ID"`
	Product Product `gorm:"foreignKey:Product_Id;references:ID"`
}

type Coupon struct {
	gorm.Model
	Code         string    `json:"code" binding:"required,alphanum" validate:"required,min=3,max=20"`
	Discount     float32   `json:"discount" binding:"required" validate:"required,gt=0,lte=100"`
	Status       bool      `json:"status" binding:"required"`
	ExpiryDate   time.Time `json:"expiry_date" binding:"required" validate:"required"`
	Min_Purchase int       `json:"min_purchase" binding:"required" validate:"required,gt=0"`
	Max_Discount int       `json:"max_discount" binding:"required" validate:"required,gt=0"`
}

type Wallet struct {
	gorm.Model
	User_Id int
	Balance float32

	User         User          `gorm:"foreignKey:User_Id;references:ID"`
	Transactions []Transaction `gorm:"foreignKey:Wallet_Id"`
}

type Transaction struct {
	gorm.Model
	Wallet_Id int
	Amount    float32
	Type      string

	Wallet Wallet `gorm:"foreignKey:Wallet_Id;references:ID"`
}

type Wishlist_item struct {
	gorm.Model
	User_Id    int `json:"user_id"`
	Product_Id int `json:"product_id"`

	User    User    `gorm:"foreignKey:User_Id;references:ID"`
	Product Product `gorm:"foreignKey:Product_Id;references:ID"`
}
