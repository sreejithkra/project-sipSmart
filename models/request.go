package models

import "time"

type User_Otp struct {
	Email string `json:"email"`
	Otp   string `json:"otp"`
}

type Google_Response struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}

type Otp struct {
	Otp       string
	ExpiresAt time.Time
}

type User_Login struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Category_Info struct {
	Name        string `json:"name" validate:"required,max=15"`
	Description string `json:"description" validate:"required,max=100"`
	Image_Url   string `json:"image_url" validate:"required,max=255,url"`
}

type Product_info struct {
	Category_Id int     `json:"category_id" validate:"required"`
	Name        string  `json:"name" validate:"required,min=2,max=25"`
	Description string  `json:"description" validate:"required"`
	Price       float64 `json:"price" validate:"required,gt=0"`
	Stock       int     `json:"stock" validate:"required,gt=0"`
	Image_Url   string  `json:"image_url" validate:"required,url"`
}

type Address_info struct {
	Country     string `json:"country" validate:"required,max=15"`
	State       string `json:"state" validate:"required,max=15"`
	District    string `json:"district" validate:"required,max=15"`
	StreetName  string `json:"street_name" validate:"required,max=15"`
	Pincode     string `json:"pincode" validate:"required,numeric,len=6"`
	PhoneNumber string `json:"phone_number" validate:"required,numeric,len=10"`
}

type Addto_Cart struct {
	Product_Id int `json:"product_id" gorm:"not null" validate:"required"`
	Quantity   int `json:"quantity" gorm:"not null;default:1" validate:"required,gt=0"`
}

type Edit_profile struct {
	Name     string `json:"name" binding:"required" validate:"required,max=15"`
	Phone    string `json:"phone" binding:"required" validate:"required,len=10,numeric"`
	Email    string `json:"email" binding:"required" validate:"required,email"`
	Password string `json:"password" binding:"required" validate:"required,min=5"`
}

type Add_Review struct {
	Product_Id int    `json:"product_id" binding:"required"`
	Rating     int    `json:"rating" binding:"required" validate:"required,max=5,min=1"`
	Review     string `json:"review" binding:"required"`
}

type Order_status struct {
	Status string `json:"status" binding:"required,oneof=completed canceled pending"`
}

type Sales_Report struct {
	Start_Date string `json:"start_date" binding:"omitempty,datetime=2006-01-02"`
	End_Date   string `json:"end_date" binding:"omitempty,datetime=2006-01-02"`
}
