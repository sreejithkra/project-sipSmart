package admin

import (
	"fmt"
	"net/http"
	"sip_Smart/database"
	"sip_Smart/models"
	"sip_Smart/responsemodels"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-pdf/fpdf"
	"github.com/xuri/excelize/v2"
)

func Sales_Report(c *gin.Context) {
	reportType := c.Query("type")

	var startDate, endDate time.Time

	switch reportType {
	case "daily":
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)

	case "weekly":
		startDate = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "yearly":
		startDate = time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "custom":
		var req models.Sales_Report
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		startDate, _ = time.Parse("2006-01-02", req.Start_Date)
		endDate, _ = time.Parse("2006-01-02", req.End_Date)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	var salesCount int64
	if err := database.Db.Model(&models.Order{}).
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Count(&salesCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate order count"})
		return
	}

	var orders []models.Order
	if err := database.Db.Preload("Items.Product").
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	var totalSales float64
	productSalesMap := make(map[int]int)
	var bestSellingProductID int
	maxQuantitySold := 0

	for _, order := range orders {
		totalSales += float64(order.Discount_price)
		for _, item := range order.Items {
			productSalesMap[item.Product_Id] += item.Quantity
			if productSalesMap[item.Product_Id] > maxQuantitySold {
				maxQuantitySold = productSalesMap[item.Product_Id]
				bestSellingProductID = item.Product_Id
			}
		}
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
				Price:          int(item.Product.Offer_Price),
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
			Order_Status:   order.Order_Status,
			Payment_Status: order.Payment_Status,
			Payment_Method: order.Payment_Method,
		}
		responseOrders = append(responseOrders, responseOrder)
	}

	var bestSellingProduct models.Product
	if bestSellingProductID != 0 {
		database.Db.First(&bestSellingProduct, bestSellingProductID)
	}

	reportData := gin.H{
		"total_sales":          totalSales,
		"salesCount":           salesCount,
		"start_date":           startDate,
		"end_date":             endDate,
		"orders":               responseOrders,
		"best_selling_product": bestSellingProduct.Name,
	}

	c.JSON(http.StatusOK, reportData)

}

func Salespdf(c *gin.Context) {
	reportType := c.Query("type")
	var startDate, endDate time.Time

	switch reportType {
	case "daily":
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)

	case "weekly":
		startDate = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "yearly":
		startDate = time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "custom":
		var req models.Sales_Report
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		startDate, _ = time.Parse("2006-01-02", req.Start_Date)
		endDate, _ = time.Parse("2006-01-02", req.End_Date)

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	var salesCount int64
	if err := database.Db.Model(&models.Order{}).
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Count(&salesCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate order count"})
		return
	}

	var orders []models.Order
	if err := database.Db.Preload("Items.Product").
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	var totalSales float64
	productSalesMap := make(map[int]int)
	var bestSellingProductID int
	maxQuantitySold := 0

	for _, order := range orders {
		totalSales += float64(order.Discount_price)
		for _, item := range order.Items {
			productSalesMap[item.Product_Id] += item.Quantity
			if productSalesMap[item.Product_Id] > maxQuantitySold {
				maxQuantitySold = productSalesMap[item.Product_Id]
				bestSellingProductID = item.Product_Id
			}
		}
	}

	var bestSellingProduct models.Product
	if bestSellingProductID != 0 {
		database.Db.First(&bestSellingProduct, bestSellingProductID)
	}

	pdf := fpdf.New("P", "mm", "A4", "")
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.Cell(190, 10, "Sales Report")
	pdf.Ln(10)

	pdf.SetFont("Arial", "", 12)
	pdf.Cell(40, 10, fmt.Sprintf("Report Type: %s", reportType))
	pdf.Ln(7)
	pdf.Cell(40, 10, fmt.Sprintf("Start Date: %s", startDate.Format("2006-01-02")))
	pdf.Ln(7)
	pdf.Cell(40, 10, fmt.Sprintf("End Date: %s", endDate.Format("2006-01-02")))
	pdf.Ln(7)
	pdf.Cell(40, 10, fmt.Sprintf("Total Sales: %.2f Rs", totalSales))
	pdf.Ln(7)
	pdf.Cell(40, 10, fmt.Sprintf("Total Sales Count: %d", salesCount))
	pdf.Ln(7)
	pdf.Cell(40, 10, fmt.Sprintf("Best Selling Product: %s", bestSellingProduct.Name))
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.Cell(190, 10, "Orders:")
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(20, 7, "Order ID", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 7, "User ID", "1", 0, "C", false, 0, "")
	// pdf.CellFormat(30, 7, "Total Price (Rs)", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, "Final Price (Rs)", "1", 0, "C", false, 0, "")

	pdf.CellFormat(60, 7, "Product", "1", 0, "C", false, 0, "")
	pdf.CellFormat(20, 7, "Quantity", "1", 0, "C", false, 0, "")
	pdf.CellFormat(30, 7, "Price (Rs)", "1", 0, "C", false, 0, "")
	pdf.Ln(7)

	pdf.SetFont("Arial", "", 10)
	for _, order := range orders {
		if len(order.Items) > 0 {
			item := order.Items[0]
			pdf.CellFormat(20, 7, fmt.Sprintf("%d", order.ID), "1", 0, "C", false, 0, "")
			pdf.CellFormat(20, 7, fmt.Sprintf("%d", order.User_Id), "1", 0, "C", false, 0, "")
			// pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", order.Total_Price), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", order.Discount_price), "1", 0, "C", false, 0, "")

			pdf.CellFormat(60, 7, item.Product.Name, "1", 0, "C", false, 0, "")
			pdf.CellFormat(20, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", item.Product.Offer_Price), "1", 0, "C", false, 0, "")
			pdf.Ln(7)
		}

		for i := 1; i < len(order.Items); i++ {
			item := order.Items[i]
			pdf.CellFormat(70, 7, "", "0", 0, "", false, 0, "")
			pdf.CellFormat(60, 7, item.Product.Name, "1", 0, "C", false, 0, "")
			pdf.CellFormat(20, 7, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
			pdf.CellFormat(30, 7, fmt.Sprintf("%.2f", item.Product.Offer_Price), "1", 0, "C", false, 0, "")
			pdf.Ln(7)
		}

		pdf.Ln(5)
	}

	err := pdf.OutputFileAndClose("sales_report.pdf")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate PDF"})
		return
	}

	c.File("sales_report.pdf")
}

func SalesExcel(c *gin.Context) {
	reportType := c.Query("type")
	var startDate, endDate time.Time

	switch reportType {
	case "daily":
		startDate = time.Now().Truncate(24 * time.Hour)
		endDate = startDate.Add(24 * time.Hour)

	case "weekly":
		startDate = time.Now().AddDate(0, 0, -7).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "yearly":
		startDate = time.Now().AddDate(-1, 0, 0).Truncate(24 * time.Hour)
		endDate = time.Now()

	case "custom":
		var req models.Sales_Report
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var err error
		startDate, err = time.Parse("2006-01-02", req.Start_Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid start date format"})
			return
		}
		endDate, err = time.Parse("2006-01-02", req.End_Date)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid end date format"})
			return
		}

	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid report type"})
		return
	}

	var salesCount int64
	if err := database.Db.Model(&models.Order{}).
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Count(&salesCount).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to calculate order count"})
		return
	}

	var orders []models.Order
	if err := database.Db.Preload("Items.Product").
		Where("created_at BETWEEN ? AND ? AND order_status = ?", startDate, endDate, "completed").
		Find(&orders).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve orders"})
		return
	}

	if salesCount == 0 {
		c.JSON(http.StatusNotFound, gin.H{"message": "No sales found for the selected period"})
		return
	}

	var totalSales float64
	productSalesMap := make(map[int]int)
	var bestSellingProductID int
	maxQuantitySold := 0

	for _, order := range orders {
		totalSales += float64(order.Discount_price)
		for _, item := range order.Items {
			productSalesMap[item.Product_Id] += item.Quantity
			if productSalesMap[item.Product_Id] > maxQuantitySold {
				maxQuantitySold = productSalesMap[item.Product_Id]
				bestSellingProductID = item.Product_Id
			}
		}
	}

	var bestSellingProduct models.Product
	if bestSellingProductID != 0 {
		if err := database.Db.First(&bestSellingProduct, bestSellingProductID).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve best-selling product"})
			return
		}
	}

	f := excelize.NewFile()

	sheetName := "Sales Report"
	index, _ := f.NewSheet(sheetName)

	headers := []string{"Report Type", "Start Date", "End Date", "Total Sales (Rs)", "Total Sales Count", "Best Selling Product"}
	for i, header := range headers {
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune(65+i)), 1), header) // Column A, B, C, etc.
	}

	reportMetadata := []interface{}{reportType, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"), totalSales, salesCount, bestSellingProduct.Name}
	for i, value := range reportMetadata {
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune(65+i)), 2), value)
	}

	orderHeaders := []string{"Order ID", "User ID", "Total Price (Rs)", "Final Price (Rs)", "Product", "Quantity", "Price (Rs)"}
	for i, header := range orderHeaders {
		f.SetCellValue(sheetName, fmt.Sprintf("%s%d", string(rune(65+i)), 4), header) // Row 4 for order headers
	}

	row := 5
	for _, order := range orders {
		if len(order.Items) > 0 {
			item := order.Items[0]
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), order.ID)
			f.SetCellValue(sheetName, fmt.Sprintf("B%d", row), order.User_Id)
			f.SetCellValue(sheetName, fmt.Sprintf("C%d", row), order.Total_Price)
			f.SetCellValue(sheetName, fmt.Sprintf("D%d", row), order.Discount_price)
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.Product.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.Quantity)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.Product.Offer_Price)
			row++
		}

		for i := 1; i < len(order.Items); i++ {
			item := order.Items[i]
			f.SetCellValue(sheetName, fmt.Sprintf("A%d", row), "")
			f.SetCellValue(sheetName, fmt.Sprintf("E%d", row), item.Product.Name)
			f.SetCellValue(sheetName, fmt.Sprintf("F%d", row), item.Quantity)
			f.SetCellValue(sheetName, fmt.Sprintf("G%d", row), item.Product.Offer_Price)
			row++
		}
	}

	f.SetActiveSheet(index)

	excelFilePath := fmt.Sprintf("sales_report_%s.xlsx", time.Now().Format("20060102150405"))
	if err := f.SaveAs(excelFilePath); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate Excel report"})
		return
	}

	c.File(excelFilePath)
}
