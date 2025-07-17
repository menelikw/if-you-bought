package main

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/datetime"
)

func main() {
	r := gin.Default()

	// Quantity-based routes
	r.GET("/:amount/:ticker/on/:buyDate", handleAmountBuy)
	r.GET("/:amount/:ticker/on/:buyDate/and-sold-on/:sellDate", handleAmountBuySell)
	r.GET("/:amount/:ticker/on/:buyDate/and-sold-on/:sellDate/with-drip", handleAmountBuySellDrip)

	// Value-based routes
	r.GET("/:amount/of/:ticker/on/:buyDate", handleAmountBuy)
	r.GET("/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate", handleAmountBuySell)
	r.GET("/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate/with-drip", handleAmountBuySellDrip)

	r.Run(":8080")
}

// Utility function stubs
// Fetch historical stock prices and dividends
func fetchStockHistory(ticker string, start, end *datetime.Datetime) (interface{}, error) {
	// TODO: Implement using glebarez/yahoo-finance or another library for historical prices
	return nil, nil
}

// Fetch historical crypto prices from CoinGecko
func fetchCryptoHistory(coinID string, fromUnix, toUnix int64) ([][2]float64, error) {
	// TODO: Implement using CoinGecko API
	return nil, nil
}

// Helper function to determine if amount is quantity or value
func parseAmount(amount string, hasOf bool) (float64, bool, error) {
	// Remove currency symbols and commas
	cleanAmount := strings.ReplaceAll(amount, "$", "")
	cleanAmount = strings.ReplaceAll(cleanAmount, ",", "")
	cleanAmount = strings.TrimSpace(cleanAmount)

	// Try to parse as float
	parsedAmount, err := strconv.ParseFloat(cleanAmount, 64)
	if err != nil {
		return 0, false, err
	}

	// If path contains "of", it's a value (currency), otherwise it's a quantity
	isValue := hasOf

	return parsedAmount, isValue, nil
}

// Handler stubs
func handleAmountBuy(c *gin.Context) {
	typeParam := c.DefaultQuery("type", "stock") // default to stock
	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")

	// Check if path contains "of" to determine if it's value or quantity
	hasOf := strings.Contains(c.Request.URL.Path, "/of/")

	parsedAmount, isValue, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	// TODO: Use typeParam to call correct fetch function
	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy only)", "value": parsedAmount, "ticker": ticker, "buyDate": buyDate, "type": typeParam})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (quantity buy only)", "quantity": parsedAmount, "ticker": ticker, "buyDate": buyDate, "type": typeParam})
	}
}

func handleAmountBuySell(c *gin.Context) {
	typeParam := c.DefaultQuery("type", "stock")
	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")
	sellDate := c.Param("sellDate")

	// Check if path contains "of" to determine if it's value or quantity
	hasOf := strings.Contains(c.Request.URL.Path, "/of/")

	parsedAmount, isValue, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	// TODO: Use typeParam to call correct fetch function
	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy/sell)", "value": parsedAmount, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "type": typeParam})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (quantity buy/sell)", "quantity": parsedAmount, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "type": typeParam})
	}
}

func handleAmountBuySellDrip(c *gin.Context) {
	typeParam := c.DefaultQuery("type", "stock")
	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")
	sellDate := c.Param("sellDate")

	// Check if path contains "of" to determine if it's value or quantity
	hasOf := strings.Contains(c.Request.URL.Path, "/of/")

	parsedAmount, isValue, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	// TODO: Use typeParam to call correct fetch function and simulate DRIP for stocks
	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy/sell with DRIP)", "value": parsedAmount, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "drip": true, "type": typeParam})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (quantity buy/sell with DRIP)", "quantity": parsedAmount, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "drip": true, "type": typeParam})
	}
}
