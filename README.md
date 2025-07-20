# If You Bought - Investment Backtesting API

A powerful REST API for backtesting stock and cryptocurrency investments with natural language URLs. Discover what your investments would be worth today if you had bought them in the past.

## 🚀 Features

- **Natural Language URLs**: Intuitive API endpoints that read like sentences
- **Stock & Crypto Support**: Backtest both stocks and cryptocurrencies
- **Currency Conversion**: Support for multiple currencies (USD, EUR, GBP, JPY, etc.)
- **DRIP Support**: Dividend Reinvestment Plan calculations
- **Real Market Data**: Integration with Alpha Vantage and Frankfurter APIs
- **Flexible Investment Types**: Specify investments by quantity or value

## 📋 Table of Contents

- [Quick Start](#quick-start)
- [API Endpoints](#api-endpoints)
- [Examples](#examples)
- [Installation](#installation)
- [Configuration](#configuration)
- [Testing](#testing)
- [Contributing](#contributing)
- [License](#license)

## 🏃‍♂️ Quick Start

### Start the Server

```bash
# Clone the repository
git clone https://github.com/menelikw/if-you-bought.git
cd if-you-bought

# Install dependencies
go mod tidy

# Run the server
go run main.go
```

The API will be available at `http://localhost:8080`

### Example Request

```bash
# What if I bought $1000 worth of Apple stock on January 1, 2020?
curl "http://localhost:8080/1000USD/of/AAPL/on/2020-01-01?type=stock"
```

## 📡 API Endpoints

### Base URL
```
http://localhost:8080
```

### URL Structure
The API uses natural language URLs that are easy to understand:

```
/:amount/:ticker/on/:buyDate
/:amount/of/:ticker/on/:buyDate
/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate
/:amount/of/:ticker/on/:buyDate/and-sold-on/:sellDate/with-drip
```

### Parameters

| Parameter | Type | Description | Example |
|-----------|------|-------------|---------|
| `amount` | string | Investment amount (quantity or value with currency) | `10`, `1000USD`, `500EUR` |
| `ticker` | string | Stock or crypto symbol | `AAPL`, `BTC`, `TSLA` |
| `buyDate` | string | Purchase date (YYYY-MM-DD) | `2020-01-01` |
| `sellDate` | string | Sale date (YYYY-MM-DD) | `2025-07-18` |
| `type` | string | Asset type (`stock` or `crypto`) | `stock` (default) |

### Investment Types

#### Quantity-Based Investment
Specify the number of shares/coins to buy:
```
/10/AAPL/on/2020-01-01
```
*"What if I bought 10 shares of Apple on January 1, 2020?"*

#### Value-Based Investment
Specify the dollar amount to invest:
```
/1000/of/AAPL/on/2020-01-01
```
*"What if I invested $1000 in Apple on January 1, 2020?"*

#### Value-Based with Currency
Specify amount in different currencies:
```
/1000EUR/of/AAPL/on/2020-01-01
```
*"What if I invested €1000 in Apple on January 1, 2020?"*

## 📊 Examples

### Stock Examples

#### 1. Quantity Investment
```bash
curl "http://localhost:8080/10/AAPL/on/2020-01-01?type=stock"
```

**Response:**
```json
{
  "message": "If you bought 10 shares of AAPL on 2020-01-01",
  "ticker": "AAPL",
  "quantity": 10,
  "buyDate": "2020-01-01",
  "closePrice": 75.09,
  "type": "stock",
  "currentValue": 2111.80,
  "totalReturn": 1812.71,
  "percentageReturn": 241.4
}
```

#### 2. Value Investment
```bash
curl "http://localhost:8080/1000/of/AAPL/on/2020-01-01?type=stock"
```

**Response:**
```json
{
  "message": "If you invested $1000 in AAPL on 2020-01-01",
  "ticker": "AAPL",
  "value": 1000,
  "buyDate": "2020-01-01",
  "closePrice": 75.09,
  "shares": 13.32,
  "type": "stock",
  "currentValue": 2813.12,
  "totalReturn": 1813.12,
  "percentageReturn": 181.3
}
```

#### 3. Currency Conversion
```bash
curl "http://localhost:8080/1000EUR/of/AAPL/on/2020-01-01?type=stock"
```

**Response:**
```json
{
  "message": "If you invested €1000 in AAPL on 2020-01-01",
  "ticker": "AAPL",
  "value": 1000,
  "currency": "EUR",
  "buyDate": "2020-01-01",
  "closePrice": 75.09,
  "fxRate": 1.0815,
  "shares": 12.31,
  "type": "stock",
  "currentValue": 2597.89,
  "totalReturn": 1597.89,
  "percentageReturn": 159.8
}
```

#### 4. Buy and Sell
```bash
curl "http://localhost:8080/1000/of/AAPL/on/2020-01-01/and-sold-on/2025-07-18?type=stock"
```

**Response:**
```json
{
  "message": "If you invested $1000 in AAPL on 2020-01-01 and sold on 2025-07-18",
  "ticker": "AAPL",
  "value": 1000,
  "buyDate": "2020-01-01",
  "sellDate": "2025-07-18",
  "buyPrice": 75.09,
  "sellPrice": 211.18,
  "shares": 13.32,
  "type": "stock",
  "finalValue": 2813.12,
  "totalReturn": 1813.12,
  "percentageReturn": 181.3
}
```

#### 5. DRIP (Dividend Reinvestment)
```bash
curl "http://localhost:8080/1000/of/AAPL/on/2020-01-01/and-sold-on/2025-07-18/with-drip?type=stock"
```

**Response:**
```json
{
  "message": "If you invested $1000 in AAPL on 2020-01-01 and sold on 2025-07-18 with DRIP",
  "ticker": "AAPL",
  "value": 1000,
  "buyDate": "2020-01-01",
  "sellDate": "2025-07-18",
  "buyPrice": 75.09,
  "sellPrice": 211.18,
  "initialShares": 13.32,
  "reinvestedShares": 2.15,
  "totalShares": 15.47,
  "drip": true,
  "type": "stock",
  "finalValue": 3267.89,
  "totalReturn": 2267.89,
  "percentageReturn": 226.8,
  "dividends": [
    {
      "date": "2020-06-15",
      "amount": 3.19
    }
  ]
}
```

### Crypto Examples

#### 1. Bitcoin Investment
```bash
curl "http://localhost:8080/1000/of/BTC/on/2020-01-01?type=crypto"
```

#### 2. Ethereum Investment
```bash
curl "http://localhost:8080/500EUR/of/ETH/on/2021-01-01?type=crypto"
```

## 🛠️ Installation

### Prerequisites

- Go 1.19 or higher
- Git

### Setup

1. **Clone the repository**
   ```bash
   git clone https://github.com/menelikw/if-you-bought.git
   cd if-you-bought
   ```

2. **Install dependencies**
   ```bash
   go mod tidy
   ```

3. **Configure environment variables** (optional)
   ```bash
   # Copy the example environment file
   cp env.example .env
   
   # Edit .env to add your API keys
   # ALPHA_VANTAGE_API_KEY=your_api_key_here
   ```

4. **Run the server**
   ```bash
   go run main.go
   ```

The API will be available at `http://localhost:8080`

## ⚙️ Configuration

### Environment Variables

The application uses environment variables for configuration. Copy `env.example` to `.env` and modify as needed:

```bash
cp env.example .env
```

#### Required Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `ALPHA_VANTAGE_API_KEY` | Alpha Vantage API key | `2G2R3SZ8BNV2EGAL` | No (uses demo key) |
| `ALPHA_VANTAGE_BASE_URL` | Alpha Vantage API base URL | `https://www.alphavantage.co` | No |
| `FRANKFURTER_BASE_URL` | Frankfurter API base URL | `https://api.frankfurter.app` | No |
| `PORT` | Server port | `8080` | No |
| `GIN_MODE` | Gin mode (`debug`/`release`) | `debug` | No |

#### API Keys

The application uses free APIs by default, but you can enhance it with paid API keys:

- **Alpha Vantage**: For stock data (free tier: 25 requests/day)
  - Get your free API key: https://www.alphavantage.co/support/#api-key
- **Frankfurter**: For currency conversion (free, no key required)

#### Example Configuration

```bash
# Production configuration
export ALPHA_VANTAGE_API_KEY=your_paid_api_key_here
export GIN_MODE=release
export PORT=3000

# Development configuration
export ALPHA_VANTAGE_API_KEY=your_dev_api_key_here
export GIN_MODE=debug
export PORT=8080
```

## 🧪 Testing

### Run All Tests
```bash
go test -v
```

### Run Specific Tests
```bash
# Unit tests (no external dependencies)
go test -v -run "TestParseAmount|TestURLRoutingNoAPI"

# Integration tests (with external APIs)
go test -v -run "WithMocks"
```

### Test Coverage
```bash
go test -cover
```

## 📈 Supported Assets

### Stocks
- All major US stocks (AAPL, GOOGL, MSFT, TSLA, etc.)
- International stocks (limited by API availability)

### Cryptocurrencies
- Bitcoin (BTC)
- Ethereum (ETH)
- And other major cryptocurrencies

### Currencies
- USD (US Dollar)
- EUR (Euro)
- GBP (British Pound)
- JPY (Japanese Yen)
- And more via Frankfurter API

## 🔧 Development

### Project Structure
```
ifyoubought/
├── main.go          # Main application file
├── main_test.go     # Test suite
├── go.mod           # Go module file
├── go.sum           # Go module checksums
└── README.md        # This file
```

### Adding New Features

1. **New Asset Types**: Modify the handlers to support new asset types
2. **New Data Sources**: Add new API integrations
3. **New Calculations**: Extend the calculation logic

### Code Style

- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Write tests for new features

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Write tests for new features
- Update documentation
- Follow existing code style
- Handle errors gracefully
- Add appropriate logging

## 📝 API Response Format

All API responses follow a consistent JSON format:

```json
{
  "message": "Human-readable description",
  "ticker": "Asset symbol",
  "buyDate": "YYYY-MM-DD",
  "sellDate": "YYYY-MM-DD", // (if applicable)
  "type": "stock|crypto",
  "currentValue": 1234.56,
  "totalReturn": 234.56,
  "percentageReturn": 23.4
}
```

### Error Responses

```json
{
  "error": "Error message",
  "details": "Additional error details"
}
```

## 🚨 Rate Limits

- **Alpha Vantage**: 25 requests/day (free tier)
- **Frankfurter**: No rate limits (free API)

For production use, consider:
- Upgrading to paid API plans
- Implementing caching
- Using multiple API keys

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Alpha Vantage](https://www.alphavantage.co/) for stock data
- [Frankfurter](https://www.frankfurter.app/) for currency conversion
- [Gin](https://github.com/gin-gonic/gin) for the web framework
- [Testify](https://github.com/stretchr/testify) for testing utilities

## 📞 Support

If you have questions or need help:

1. Check the [Issues](https://github.com/menelikw/if-you-bought/issues) page
2. Create a new issue with detailed information
3. Include your request URL and expected behavior

## 🔮 Roadmap

- [ ] Add more cryptocurrency support
- [ ] Implement historical dividend data
- [ ] Add portfolio backtesting
- [ ] Create web interface
- [ ] Add more data sources
- [ ] Implement caching
- [ ] Add authentication
- [ ] Create mobile app

---

**Happy Backtesting! 📈** 