package user

import (
	"fmt"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"
	"strconv"

	"github.com/jung-kurt/gofpdf"

	"github.com/gin-gonic/gin"
)

func GenerateInvoicePDF(c *gin.Context) {
	orderID := c.Query("id")

	var order models.Order
	if err := database.Db.Preload("Items.Product").Where("id = ?", orderID).First(&order).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "B", 16)

	pdf.Cell(40, 10, fmt.Sprintf("Invoice for Order #%d", order.ID))
	pdf.Ln(12)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, "Customer ID: "+strconv.Itoa(order.User_Id))
	pdf.Ln(8)
	pdf.Cell(40, 10, "Payment Status: "+order.Payment_Status)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Order Status: "+order.Order_Status)
	pdf.Ln(8)
	pdf.Cell(40, 10, "Payment Method: "+order.Payment_Method)
	pdf.Ln(8)
	if order.Coupon_code != "" {
		pdf.Cell(40, 10, "Coupon Code: "+order.Coupon_code)
		pdf.Ln(8)
	}
	pdf.Ln(12)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(60, 10, "Item")
	pdf.Cell(30, 10, "Quantity")
	pdf.Cell(30, 10, "Price")
	pdf.Cell(30, 10, "Total")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	for _, item := range order.Items {
		productName := item.Product.Name
		pdf.Cell(60, 10, productName)
		pdf.Cell(30, 10, strconv.Itoa(item.Quantity))
		pdf.Cell(30, 10, fmt.Sprintf("$%.2f", item.Price))
		totalPrice := float32(item.Quantity) * item.Price
		pdf.Cell(30, 10, fmt.Sprintf("$%.2f", totalPrice))
		pdf.Ln(8)
	}

	pdf.Ln(8)
	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(60, 10, "Subtotal:")
	pdf.Cell(30, 10, "")
	pdf.Cell(30, 10, "")
	pdf.Cell(30, 10, fmt.Sprintf("$%.2f", order.Total_Price))
	pdf.Ln(8)

	if order.Final_Price > 0 {
		pdf.Cell(60, 10, "Discount:")
		pdf.Cell(30, 10, "")
		pdf.Cell(30, 10, "")
		pdf.Cell(30, 10, fmt.Sprintf("-$%.2f", order.Total_Price-order.Final_Price))
		pdf.Ln(8)
	}

	pdf.Cell(60, 10, "Grand Total:")
	pdf.Cell(30, 10, "")
	pdf.Cell(30, 10, "")
	pdf.Cell(30, 10, fmt.Sprintf("$%.2f", order.Final_Price))

	c.Header("Content-Type", "application/pdf")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=invoice_%d.pdf", order.ID))
	err := pdf.Output(c.Writer)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}
}
