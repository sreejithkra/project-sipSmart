package responsemodels

import "time"

type Category struct {
	Id          int
	Name        string
	Description string
	Image_Url   string
}

type Product struct {
	Id          int
	Name        string
	Category    string
	Description string
	Image_Url   string
	Price       float64
	Stock       int
	Offer_Price float64
	Popularity  float64
	Sales       int
}

type Product_Details struct {
	Id             uint     `json:"id"`
	Name           string   `json:"name"`
	Description    string   `json:"description"`
	ImageURL       string   `json:"image_url"`
	Price          float64  `json:"price"`
	Stock          int      `json:"stock"`
	Average_Rating float64  `json:"average_rating"`
	Category       string   `json:"category_name"`
	Reviews        []Review `json:"reviews" gorm:"foreignKey:Product_Id"`
	Popularity     float64
}

type Review struct {
	ID         uint   `json:"id"`
	Product_Id uint   `json:"product_id"`
	Content    string `json:"content"`
	Rating     int    `json:"rating"`
}

type Address struct {
	Id          uint   `json:"id"`
	Country     string `json:"country" validate:"required"`
	State       string `json:"state" validate:"required"`
	District    string `json:"district" validate:"required"`
	StreetName  string `json:"street_name" validate:"required"`
	Pincode     string `json:"pincode" validate:"required"`
	PhoneNumber string `json:"phone_number" validate:"required"`
	Default     bool   `json:"default"`
}

type Cart struct {
	ID             int         `json:"id"`
	Items          []Cart_Item `json:"items"`
	TotalPrice     float32     `json:"total_price"`
	Discount_price float32     `json:"discount_price"`
}

type Cart_Item struct {
	Product_Id     int     `json:"product_id"`
	Product_Name   string  `json:"product_name"`
	Quantity       int     `json:"quantity"`
	Price          float32 `json:"price"`
	Discount_price float32 `json:"discount_price"`
}

type Order struct {
	Id             int
	User_Id        int          `json:"user_id"`
	Items          []Order_Item `json:"items"`
	Total_Price    float32      `json:"total_price"`
	Discount_price float32      `json:"discount_price"`
	Order_Status   string       `json:"order_status"`
	Payment_Status string       `json:"payment_status"`
	Payment_Method string       `json:"payment_method"`
}

type Order_Item struct {
	Order_Id       int     `json:"order_id"`
	Product_Id     int     `json:"product_id"`
	Product_Name   string  `json:"product_name"`
	Quantity       int     `json:"quantity"`
	Price          int     `json:"price"`
	Discount_price float32 `json:"discount_price"`
	Status         string  `json:"status"`
}

type Wishlist_item struct {
	Product_Id   int `json:"product_id"`
	Product_Name string
	Price        float32
}

type Wallet struct {
	User_Id int
	Balance float32

	Transactions []Transaction `gorm:"foreignKey:Wallet_Id"`
}

type Transaction struct {
	Amount float32
	Type   string
	Time   time.Time
}

type Best_Category struct {
	Id int
    CategoryName string
    SalesCount   int
}

type Best_Product struct {
	Id int
    ProductName string
	Category_id int
    SalesCount   int
}
