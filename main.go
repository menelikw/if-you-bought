package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/datetime"
)

// Environment variables
var (
	alphaVantageAPIKey  = getEnv("ALPHA_VANTAGE_API_KEY", "2G2R3SZ8BNV2EGAL")
	alphaVantageBaseURL = getEnv("ALPHA_VANTAGE_BASE_URL", "https://www.alphavantage.co")
	frankfurterBaseURL  = getEnv("FRANKFURTER_BASE_URL", "https://api.frankfurter.app")
	serverPort          = getEnv("PORT", "8080")
	ginMode             = getEnv("GIN_MODE", "debug")
)

// Helper function to get environment variables with defaults
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// Alpha Vantage daily time series response struct
// Only the fields we need
// Example: https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=AAPL&apikey=demo

type alphaVantageDailyResponse struct {
	TimeSeries map[string]map[string]string `json:"Time Series (Daily)"`
}

// Alpha Vantage daily adjusted time series response struct (includes dividends)
type alphaVantageDailyAdjustedResponse struct {
	TimeSeries map[string]map[string]string `json:"Time Series (Daily)"`
}

// Dividend data structure
type dividendData struct {
	Date   string  `json:"date"`
	Amount float64 `json:"amount"`
}

// Fetch historical daily close price for a given ticker and date (YYYY-MM-DD)
func fetchStockDailyCloseAlphaVantage(ticker, date string) (float64, error) {
	url := fmt.Sprintf("%s/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", alphaVantageBaseURL, ticker, alphaVantageAPIKey)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	// Read the response body for debugging
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, err
	}

	// Debug: print the first 500 characters of the response
	fmt.Printf("Alpha Vantage response (first 500 chars): %s\n", string(body[:min(500, len(body))]))

	var result alphaVantageDailyResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return 0, fmt.Errorf("JSON unmarshal error: %v", err)
	}

	if result.TimeSeries == nil {
		return 0, fmt.Errorf("No time series data returned from Alpha Vantage")
	}

	dayData, ok := result.TimeSeries[date]
	if !ok {
		return 0, fmt.Errorf("No data for date %s", date)
	}

	closeStr, ok := dayData["4. close"]
	if !ok {
		return 0, fmt.Errorf("No close price for date %s", date)
	}

	closeVal, err := strconv.ParseFloat(closeStr, 64)
	if err != nil {
		return 0, err
	}
	return closeVal, nil
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// Fetch historical dividends for a given ticker and date range
func fetchStockDividendsAlphaVantage(ticker, startDate, endDate string) ([]dividendData, error) {
	// Note: Alpha Vantage TIME_SERIES_DAILY_ADJUSTED is premium, so we'll use a free alternative
	// For now, we'll simulate dividend data based on typical dividend yields
	// In production, you'd use a paid API or alternative data source

	var dividends []dividendData

	// Simulate quarterly dividends for demonstration
	// In reality, you'd fetch this from a dividend API
	// Common dividend dates: March, June, September, December

	// For AAPL, typical dividend is around $0.24 per share quarterly
	// This is a simplified simulation - real implementation would use actual dividend data

	// Parse dates to check if they fall in dividend periods
	start, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return nil, err
	}

	end, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return nil, err
	}

	// Simulate dividends for the date range
	// This is a simplified approach - real implementation would use actual dividend data
	dividendAmount := 0.24 // Typical AAPL quarterly dividend

	// Check if the date range includes dividend periods
	// This is a simplified simulation
	if ticker == "AAPL" {
		// Simulate quarterly dividends
		dividendDates := []string{"2025-03-15", "2025-06-15", "2025-09-15", "2025-12-15"}

		for _, divDate := range dividendDates {
			divTime, err := time.Parse("2006-01-02", divDate)
			if err != nil {
				continue
			}

			// Check if dividend date falls within our range
			if (divTime.After(start) || divTime.Equal(start)) && (divTime.Before(end) || divTime.Equal(end)) {
				dividends = append(dividends, dividendData{
					Date:   divDate,
					Amount: dividendAmount,
				})
			}
		}
	}

	return dividends, nil
}

// Calculate DRIP reinvestment
func calculateDRIP(shares float64, dividends []dividendData, stockPrice float64) (float64, []dividendData) {
	totalReinvestedShares := 0.0
	reinvestedDividends := []dividendData{}

	for _, dividend := range dividends {
		// Calculate dividend payment for current shares
		dividendPayment := shares * dividend.Amount

		// Calculate additional shares from dividend reinvestment
		additionalShares := dividendPayment / stockPrice

		if additionalShares > 0 {
			totalReinvestedShares += additionalShares
			reinvestedDividends = append(reinvestedDividends, dividendData{
				Date:   dividend.Date,
				Amount: dividendPayment,
			})
		}
	}

	return totalReinvestedShares, reinvestedDividends
}

func main() {
	// Set Gin mode from environment
	gin.SetMode(ginMode)

	r := gin.Default()

	// Quantity-based routes
	r.GET("/:amount/:ticker/on/:buyDate", handleAmountBuy)
	r.GET("/:amount/:ticker/on/:buyDate/and-sold-on/:sellDate", handleAmountBuySell)
	r.GET("/:amount/:ticker/on/:buyDate/and-sold-on/:sellDate/with-drip", handleAmountBuySellDrip)

	// Value-based routes
	r.GET("/:amount/of/:ticker/on/:buyDate", handleAmountBuy)
	r.GET("/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate", handleAmountBuySell)
	r.GET("/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate/with-drip", handleAmountBuySellDrip)

	// Start server with configured port
	r.Run(":" + serverPort)
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

// Helper function to determine if amount is quantity or value, and extract currency
func parseAmount(amount string) (float64, string, bool) {
	// Regex to extract currency symbol or code (e.g. $, €, £, ¥, USD, EUR, GBP, etc.)
	currencyRegex := regexp.MustCompile(`([\p{Sc}]|[A-Z]{3})`)
	currencyMatch := currencyRegex.FindString(amount)

	// Regex to extract the numeric part (supports decimals and minus)
	numRegex := regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
	numMatch := numRegex.FindString(amount)

	if numMatch == "" {
		return 0, "", false
	}

	parsedAmount, err := strconv.ParseFloat(numMatch, 64)
	if err != nil {
		return 0, "", false
	}

	isValue := currencyMatch != ""
	return parsedAmount, currencyMatch, isValue
}

// Frankfurter exchange rate response struct
type frankfurterResponse struct {
	Amount float64            `json:"amount"`
	Base   string             `json:"base"`
	Date   string             `json:"date"`
	Rates  map[string]float64 `json:"rates"`
}

// Fetch historical FX rates using Frankfurter (free, no API key required)
func getHistoricalFXRate(fromCurrency, toCurrency, date string) (float64, error) {
	// Frankfurter format: https://api.frankfurter.app/2020-01-01?from=EUR&to=USD
	url := fmt.Sprintf("%s/%s?from=%s&to=%s", frankfurterBaseURL, date, fromCurrency, toCurrency)
	resp, err := http.Get(url)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var result frankfurterResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return 0, err
	}

	if result.Rates == nil {
		return 0, fmt.Errorf("No rates returned from Frankfurter")
	}

	rate, ok := result.Rates[toCurrency]
	if !ok {
		return 0, fmt.Errorf("No rate found for %s to %s on %s", fromCurrency, toCurrency, date)
	}

	return rate, nil
}

// Handler stubs
func handleAmountBuy(c *gin.Context) {
	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")
	typeParam := c.DefaultQuery("type", "stock")

	// Parse amount and detect if it's value-based
	parsedAmount, currency, isValue := parseAmount(amount)
	if parsedAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	if isValue {
		// Value-based investment
		// Get FX rate for buy date
		fxRate, err := getHistoricalFXRate(currency, "USD", buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate", "details": err.Error()})
			return
		}

		// Get stock price
		closePrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price", "details": err.Error()})
			return
		}

		// Calculate shares bought
		shares := (parsedAmount * fxRate) / closePrice

		c.JSON(http.StatusOK, gin.H{
			"message":       "Backtest result (value buy only)",
			"value":         parsedAmount,
			"currency":      currency,
			"ticker":        ticker,
			"buyDate":       buyDate,
			"closePrice":    closePrice,
			"shares":        shares,
			"stockCurrency": "USD",
			"fxRate":        fxRate,
			"type":          typeParam,
		})
	} else {
		// Quantity-based investment
		closePrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price", "details": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"message":    "Backtest result (quantity buy only)",
			"quantity":   parsedAmount,
			"ticker":     ticker,
			"buyDate":    buyDate,
			"closePrice": closePrice,
			"type":       typeParam,
		})
	}
}

func handleAmountBuySell(c *gin.Context) {
	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")
	sellDate := c.Param("sellDate")
	typeParam := c.DefaultQuery("type", "stock")

	// Parse amount and detect if it's value-based
	parsedAmount, currency, isValue := parseAmount(amount)
	if parsedAmount == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	if isValue {
		// Value-based investment
		// Get FX rate for buy date
		fxRateBuy, err := getHistoricalFXRate(currency, "USD", buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate for buy date", "details": err.Error()})
			return
		}

		// Get FX rate for sell date
		fxRateSell, err := getHistoricalFXRate("USD", currency, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate for sell date", "details": err.Error()})
			return
		}

		// Get stock prices
		buyPrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buy price", "details": err.Error()})
			return
		}

		sellPrice, err := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sell price", "details": err.Error()})
			return
		}

		// Convert investment value to USD
		investmentUSD := parsedAmount * fxRateBuy

		// Calculate shares bought
		shares := investmentUSD / buyPrice

		// Calculate final value in USD
		finalValueUSD := shares * sellPrice

		// Convert back to original currency
		finalValueInOriginalCurrency := finalValueUSD * fxRateSell

		c.JSON(http.StatusOK, gin.H{
			"message":                      "Backtest result (value buy/sell)",
			"value":                        parsedAmount,
			"currency":                     currency,
			"ticker":                       ticker,
			"buyDate":                      buyDate,
			"sellDate":                     sellDate,
			"buyPrice":                     buyPrice,
			"sellPrice":                    sellPrice,
			"shares":                       shares,
			"stockCurrency":                "USD",
			"finalValueUSD":                finalValueUSD,
			"finalValueInOriginalCurrency": finalValueInOriginalCurrency,
			"fxRateBuy":                    fxRateBuy,
			"fxRateSell":                   fxRateSell,
			"type":                         typeParam,
		})
	} else {
		// Quantity-based investment
		buyPrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buy price", "details": err.Error()})
			return
		}

		sellPrice, err := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sell price", "details": err.Error()})
			return
		}

		finalValue := parsedAmount * sellPrice

		c.JSON(http.StatusOK, gin.H{
			"message":    "Backtest result (quantity buy/sell)",
			"quantity":   parsedAmount,
			"ticker":     ticker,
			"buyDate":    buyDate,
			"sellDate":   sellDate,
			"buyPrice":   buyPrice,
			"sellPrice":  sellPrice,
			"finalValue": finalValue,
			"type":       typeParam,
		})
	}
}

func handleAmountBuySellDrip(c *gin.Context) {
	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")
	sellDate := c.Param("sellDate")
	typeParam := c.DefaultQuery("type", "stock")

	// Parse amount and detect if it's value-based
	parsedAmount, currency, isValue := parseAmount(amount)

	if isValue {
		// Value-based investment with DRIP
		// Get FX rate for buy date
		fxRateBuy, err := getHistoricalFXRate(currency, "USD", buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate for buy date", "details": err.Error()})
			return
		}

		// Get FX rate for sell date
		fxRateSell, err := getHistoricalFXRate("USD", currency, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate for sell date", "details": err.Error()})
			return
		}

		// Get stock prices
		buyPrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buy price", "details": err.Error()})
			return
		}

		sellPrice, err := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sell price", "details": err.Error()})
			return
		}

		// Convert investment value to USD
		investmentUSD := parsedAmount * fxRateBuy

		// Calculate initial shares
		initialShares := investmentUSD / buyPrice

		// Fetch dividends for the period
		dividends, err := fetchStockDividendsAlphaVantage(ticker, buyDate, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dividends", "details": err.Error()})
			return
		}

		// Calculate DRIP reinvestment
		reinvestedShares, reinvestedDividends := calculateDRIP(initialShares, dividends, buyPrice)

		// Total shares after DRIP
		totalShares := initialShares + reinvestedShares

		// Calculate final value in USD
		finalValueUSD := totalShares * sellPrice

		// Convert back to original currency
		finalValueInOriginalCurrency := finalValueUSD * fxRateSell

		c.JSON(http.StatusOK, gin.H{
			"message":                      "Backtest result (value buy/sell with DRIP)",
			"value":                        parsedAmount,
			"currency":                     currency,
			"ticker":                       ticker,
			"buyDate":                      buyDate,
			"sellDate":                     sellDate,
			"buyPrice":                     buyPrice,
			"sellPrice":                    sellPrice,
			"initialShares":                initialShares,
			"reinvestedShares":             reinvestedShares,
			"totalShares":                  totalShares,
			"dividends":                    reinvestedDividends,
			"finalValueUSD":                finalValueUSD,
			"finalValueInOriginalCurrency": finalValueInOriginalCurrency,
			"fxRateBuy":                    fxRateBuy,
			"fxRateSell":                   fxRateSell,
			"drip":                         true,
			"type":                         typeParam,
		})
	} else {
		// Quantity-based investment with DRIP
		// Get stock prices
		buyPrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch buy price", "details": err.Error()})
			return
		}

		sellPrice, err := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch sell price", "details": err.Error()})
			return
		}

		// Fetch dividends for the period
		dividends, err := fetchStockDividendsAlphaVantage(ticker, buyDate, sellDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch dividends", "details": err.Error()})
			return
		}

		// Calculate DRIP reinvestment
		reinvestedShares, reinvestedDividends := calculateDRIP(parsedAmount, dividends, buyPrice)

		// Total shares after DRIP
		totalShares := parsedAmount + reinvestedShares

		// Calculate final value
		finalValue := totalShares * sellPrice

		c.JSON(http.StatusOK, gin.H{
			"message":          "Backtest result (quantity buy/sell with DRIP)",
			"quantity":         parsedAmount,
			"ticker":           ticker,
			"buyDate":          buyDate,
			"sellDate":         sellDate,
			"buyPrice":         buyPrice,
			"sellPrice":        sellPrice,
			"reinvestedShares": reinvestedShares,
			"totalShares":      totalShares,
			"dividends":        reinvestedDividends,
			"finalValue":       finalValue,
			"drip":             true,
			"type":             typeParam,
		})
	}
}
