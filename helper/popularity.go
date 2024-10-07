package helper

import (
	"sip_Smart/database"
	"sip_Smart/models"
)

const salesWeight = 0.5
const viewsWeight = 0.3
const ratingWeight = 0.2

func CalculatePopularity(sales int, views int, rating float64) float64 {
	popularity := (salesWeight * float64(sales)) + (viewsWeight * float64(views)) + (ratingWeight * rating)
	return popularity
}

func UpdateProductPopularity(product *models.Product) {

	product.Popularity = CalculatePopularity(product.Sales, product.Views, product.Average_Rating)

	database.Db.Save(product)
}
