package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/piquette/finance-go/datetime"
)

const alphaVantageAPIKey = "2G2R3SZ8BNV2EGAL"

// Alpha Vantage daily time series response struct
// Only the fields we need
// Example: https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=AAPL&apikey=demo

type alphaVantageDailyResponse struct {
	TimeSeries map[string]map[string]string `json:"Time Series (Daily)"`
}

// Fetch historical daily close price for a given ticker and date (YYYY-MM-DD)
func fetchStockDailyCloseAlphaVantage(ticker, date string) (float64, error) {
	url := fmt.Sprintf("https://www.alphavantage.co/query?function=TIME_SERIES_DAILY&symbol=%s&apikey=%s", ticker, alphaVantageAPIKey)
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

// Stub for fetching dividends (Alpha Vantage supports this in TIME_SERIES_DAILY_ADJUSTED)
func fetchStockDividendsAlphaVantage(ticker, date string) (float64, error) {
	// TODO: Parse "7. dividend amount" from the same API response
	return 0, nil
}

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

// Helper function to determine if amount is quantity or value, and extract currency
func parseAmount(amount string, hasOf bool) (float64, bool, string, error) {
	// Regex to extract currency symbol or code (e.g. $, €, £, ¥, USD, EUR, GBP, etc.)
	currencyRegex := regexp.MustCompile(`([\p{Sc}]|[A-Z]{3})`)
	currencyMatch := currencyRegex.FindString(amount)

	// Regex to extract the numeric part (supports decimals and minus)
	numRegex := regexp.MustCompile(`[-+]?[0-9]*\.?[0-9]+`)
	numMatch := numRegex.FindString(amount)

	if numMatch == "" {
		return 0, false, "", errors.New("No numeric value found in amount")
	}

	parsedAmount, err := strconv.ParseFloat(numMatch, 64)
	if err != nil {
		return 0, false, "", err
	}

	isValue := hasOf
	return parsedAmount, isValue, currencyMatch, nil
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
	url := fmt.Sprintf("https://api.frankfurter.app/%s?from=%s&to=%s", date, fromCurrency, toCurrency)
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
	typeParam := c.DefaultQuery("type", "stock") // default to stock
	if typeParam != "stock" && typeParam != "crypto" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid type parameter: must be 'stock' or 'crypto'"})
		return
	}

	amount := c.Param("amount")
	ticker := c.Param("ticker")
	buyDate := c.Param("buyDate")

	hasOf := strings.Contains(c.Request.URL.Path, "/of/")
	parsedAmount, isValue, currency, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	if typeParam == "stock" && isValue {
		// Value-based: fetch close price, convert value to USD if needed, calculate shares
		closePrice, err := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price", "details": err.Error()})
			return
		}
		stockCurrency := "USD" // Alpha Vantage returns USD for US stocks; for others, you may need to enhance this
		fxRate := 1.0
		if currency != "" && currency != stockCurrency {
			fxRate, err = getHistoricalFXRate(currency, stockCurrency, buyDate)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate", "details": err.Error()})
				return
			}
		}
		valueInStockCurrency := parsedAmount * fxRate
		numShares := valueInStockCurrency / closePrice
		c.JSON(http.StatusOK, gin.H{
			"message":       "Backtest result (value buy only)",
			"value":         parsedAmount,
			"currency":      currency,
			"ticker":        ticker,
			"buyDate":       buyDate,
			"closePrice":    closePrice,
			"fxRate":        fxRate,
			"stockCurrency": stockCurrency,
			"shares":        numShares,
			"type":          typeParam,
		})
		return
	}

	if typeParam == "stock" && !isValue {
		// Fetch historical close price for buyDate
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
		return
	}

	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy only)", "value": parsedAmount, "currency": currency, "ticker": ticker, "buyDate": buyDate, "type": typeParam})
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

	hasOf := strings.Contains(c.Request.URL.Path, "/of/")
	parsedAmount, isValue, currency, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	if typeParam == "stock" && isValue {
		// Value-based: fetch close prices, convert value to USD if needed, calculate shares, then value at sell date
		buyPrice, err1 := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		sellPrice, err2 := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price", "details": fmt.Sprintf("buy: %v, sell: %v", err1, err2)})
			return
		}
		stockCurrency := "USD"
		fxRateBuy := 1.0
		fxRateSell := 1.0
		if currency != "" && currency != stockCurrency {
			fxRateBuy, err1 = getHistoricalFXRate(currency, stockCurrency, buyDate)
			fxRateSell, err2 = getHistoricalFXRate(stockCurrency, currency, sellDate)
			if err1 != nil || err2 != nil {
				c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch FX rate", "details": fmt.Sprintf("buy: %v, sell: %v", err1, err2)})
				return
			}
		}
		valueInStockCurrency := parsedAmount * fxRateBuy
		numShares := valueInStockCurrency / buyPrice
		finalValueInStockCurrency := numShares * sellPrice
		finalValueInOriginalCurrency := finalValueInStockCurrency * fxRateSell
		c.JSON(http.StatusOK, gin.H{
			"message":                      "Backtest result (value buy/sell)",
			"value":                        parsedAmount,
			"currency":                     currency,
			"ticker":                       ticker,
			"buyDate":                      buyDate,
			"sellDate":                     sellDate,
			"buyPrice":                     buyPrice,
			"sellPrice":                    sellPrice,
			"fxRateBuy":                    fxRateBuy,
			"fxRateSell":                   fxRateSell,
			"stockCurrency":                stockCurrency,
			"shares":                       numShares,
			"finalValueInStockCurrency":    finalValueInStockCurrency,
			"finalValueInOriginalCurrency": finalValueInOriginalCurrency,
			"type":                         typeParam,
		})
		return
	}

	if typeParam == "stock" && !isValue {
		// Fetch historical close prices for buyDate and sellDate
		buyPrice, err1 := fetchStockDailyCloseAlphaVantage(ticker, buyDate)
		sellPrice, err2 := fetchStockDailyCloseAlphaVantage(ticker, sellDate)
		if err1 != nil || err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch stock price", "details": fmt.Sprintf("buy: %v, sell: %v", err1, err2)})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"message":   "Backtest result (quantity buy/sell)",
			"quantity":  parsedAmount,
			"ticker":    ticker,
			"buyDate":   buyDate,
			"sellDate":  sellDate,
			"buyPrice":  buyPrice,
			"sellPrice": sellPrice,
			"type":      typeParam,
		})
		return
	}

	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy/sell)", "value": parsedAmount, "currency": currency, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "type": typeParam})
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

	hasOf := strings.Contains(c.Request.URL.Path, "/of/")
	parsedAmount, isValue, currency, err := parseAmount(amount, hasOf)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid amount format"})
		return
	}

	if isValue {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (value buy/sell with DRIP)", "value": parsedAmount, "currency": currency, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "drip": true, "type": typeParam})
	} else {
		c.JSON(http.StatusOK, gin.H{"message": "Backtest result (quantity buy/sell with DRIP)", "quantity": parsedAmount, "ticker": ticker, "buyDate": buyDate, "sellDate": sellDate, "drip": true, "type": typeParam})
	}
}
