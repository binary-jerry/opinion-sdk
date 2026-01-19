package clob

import (
	"encoding/json"
	"testing"

	"github.com/shopspring/decimal"
)

func TestOrderSideString(t *testing.T) {
	tests := []struct {
		side     OrderSide
		expected string
	}{
		{OrderSideBuy, "BUY"},
		{OrderSideSell, "SELL"},
		{OrderSide(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		if got := tt.side.String(); got != tt.expected {
			t.Errorf("OrderSide(%d).String() = %s, want %s", tt.side, got, tt.expected)
		}
	}
}

func TestOrderSideConstants(t *testing.T) {
	if OrderSideBuy != 0 {
		t.Errorf("OrderSideBuy = %d, want 0", OrderSideBuy)
	}
	if OrderSideSell != 1 {
		t.Errorf("OrderSideSell = %d, want 1", OrderSideSell)
	}
}

func TestOrderTypeConstants(t *testing.T) {
	if OrderTypeMarket != "MARKET" {
		t.Errorf("OrderTypeMarket = %s, want MARKET", OrderTypeMarket)
	}
	if OrderTypeLimit != "LIMIT" {
		t.Errorf("OrderTypeLimit = %s, want LIMIT", OrderTypeLimit)
	}
}

func TestOrderStatusConstants(t *testing.T) {
	if OrderStatusPending != "pending" {
		t.Errorf("OrderStatusPending = %s, want pending", OrderStatusPending)
	}
	if OrderStatusFilled != "filled" {
		t.Errorf("OrderStatusFilled = %s, want filled", OrderStatusFilled)
	}
	if OrderStatusCancelled != "cancelled" {
		t.Errorf("OrderStatusCancelled = %s, want cancelled", OrderStatusCancelled)
	}
	if OrderStatusExpired != "expired" {
		t.Errorf("OrderStatusExpired = %s, want expired", OrderStatusExpired)
	}
}

func TestOrderJSON(t *testing.T) {
	order := &Order{
		ID:            "order-1",
		MarketID:      "market-1",
		TokenID:       "12345",
		Side:          "BUY",
		Type:          OrderTypeLimit,
		Price:         decimal.NewFromFloat(0.55),
		Size:          decimal.NewFromInt(100),
		FilledSize:    decimal.NewFromInt(50),
		RemainingSize: decimal.NewFromInt(50),
		Status:        OrderStatusPending,
		Maker:         "0x123",
	}

	data, err := json.Marshal(order)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Order
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != order.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, order.ID)
	}
	if !decoded.Price.Equal(order.Price) {
		t.Errorf("Price = %s, want %s", decoded.Price, order.Price)
	}
	if decoded.Status != order.Status {
		t.Errorf("Status = %s, want %s", decoded.Status, order.Status)
	}
}

func TestSignedOrderJSON(t *testing.T) {
	signedOrder := &SignedOrder{
		Salt:          12345,
		Maker:         "0x123",
		Signer:        "0x123",
		Taker:         "0x000",
		TokenId:       "67890",
		MakerAmount:   "1000000",
		TakerAmount:   "500000",
		Expiration:    "0",
		Nonce:         "1",
		FeeRateBps:    "0",
		Side:          "BUY",
		SignatureType: 0,
		Signature:     "0xabc123",
	}

	data, err := json.Marshal(signedOrder)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded SignedOrder
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Salt != signedOrder.Salt {
		t.Errorf("Salt = %d, want %d", decoded.Salt, signedOrder.Salt)
	}
	if decoded.Signature != signedOrder.Signature {
		t.Errorf("Signature = %s, want %s", decoded.Signature, signedOrder.Signature)
	}
}

func TestCreateOrderRequest(t *testing.T) {
	req := &CreateOrderRequest{
		MarketID:                "market-1",
		TokenID:                 "12345",
		Side:                    OrderSideBuy,
		Type:                    OrderTypeLimit,
		Price:                   decimal.NewFromFloat(0.55),
		MakerAmountInQuoteToken: decimal.NewFromInt(100),
		FeeRateBps:              0,
	}

	if req.Side != OrderSideBuy {
		t.Errorf("Side = %d, want BUY", req.Side)
	}
	if req.Type != OrderTypeLimit {
		t.Errorf("Type = %s, want LIMIT", req.Type)
	}
}

func TestTradeJSON(t *testing.T) {
	trade := &Trade{
		ID:        "trade-1",
		OrderID:   "order-1",
		MarketID:  "market-1",
		TokenID:   "12345",
		Side:      "BUY",
		Price:     decimal.NewFromFloat(0.55),
		Size:      decimal.NewFromInt(100),
		Fee:       decimal.NewFromFloat(0.5),
		Timestamp: "2024-01-01T00:00:00Z",
		TxHash:    "0xabc",
	}

	data, err := json.Marshal(trade)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Trade
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.ID != trade.ID {
		t.Errorf("ID = %s, want %s", decoded.ID, trade.ID)
	}
	if !decoded.Fee.Equal(trade.Fee) {
		t.Errorf("Fee = %s, want %s", decoded.Fee, trade.Fee)
	}
}

func TestBalanceJSON(t *testing.T) {
	balance := &Balance{
		TokenID:   "usdt",
		Symbol:    "USDT",
		Available: decimal.NewFromInt(1000),
		Frozen:    decimal.NewFromInt(100),
		Total:     decimal.NewFromInt(1100),
	}

	data, err := json.Marshal(balance)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Balance
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.Symbol != balance.Symbol {
		t.Errorf("Symbol = %s, want %s", decoded.Symbol, balance.Symbol)
	}
	if !decoded.Available.Equal(balance.Available) {
		t.Errorf("Available = %s, want %s", decoded.Available, balance.Available)
	}
}

func TestPositionJSON(t *testing.T) {
	position := &Position{
		MarketID:     "market-1",
		TokenID:      "12345",
		Outcome:      "Yes",
		Size:         decimal.NewFromInt(100),
		AvgPrice:     decimal.NewFromFloat(0.50),
		CurrentPrice: decimal.NewFromFloat(0.60),
		PnL:          decimal.NewFromInt(10),
		PnLPercent:   decimal.NewFromFloat(20),
	}

	data, err := json.Marshal(position)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var decoded Position
	err = json.Unmarshal(data, &decoded)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if decoded.MarketID != position.MarketID {
		t.Errorf("MarketID = %s, want %s", decoded.MarketID, position.MarketID)
	}
	if !decoded.PnL.Equal(position.PnL) {
		t.Errorf("PnL = %s, want %s", decoded.PnL, position.PnL)
	}
}

func TestCancelOrderRequest(t *testing.T) {
	req := &CancelOrderRequest{
		OrderID: "order-123",
	}

	if req.OrderID != "order-123" {
		t.Errorf("OrderID = %s, want order-123", req.OrderID)
	}
}

func TestCancelOrdersRequest(t *testing.T) {
	req := &CancelOrdersRequest{
		OrderIDs: []string{"order-1", "order-2", "order-3"},
	}

	if len(req.OrderIDs) != 3 {
		t.Errorf("OrderIDs length = %d, want 3", len(req.OrderIDs))
	}
}

func TestCancelAllOrdersRequest(t *testing.T) {
	req := &CancelAllOrdersRequest{
		MarketID: "market-1",
		Side:     "BUY",
	}

	if req.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", req.MarketID)
	}
	if req.Side != "BUY" {
		t.Errorf("Side = %s, want BUY", req.Side)
	}
}

func TestSplitRequest(t *testing.T) {
	req := &SplitRequest{
		MarketID: "market-1",
		Amount:   decimal.NewFromInt(100),
	}

	if req.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", req.MarketID)
	}
	if !req.Amount.Equal(decimal.NewFromInt(100)) {
		t.Errorf("Amount = %s, want 100", req.Amount)
	}
}

func TestMergeRequest(t *testing.T) {
	req := &MergeRequest{
		MarketID: "market-1",
		Amount:   decimal.NewFromInt(50),
	}

	if req.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", req.MarketID)
	}
}

func TestRedeemRequest(t *testing.T) {
	req := &RedeemRequest{
		MarketID: "market-1",
	}

	if req.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", req.MarketID)
	}
}

func TestEnableTradingRequest(t *testing.T) {
	req := &EnableTradingRequest{
		QuoteTokenID: "usdt",
	}

	if req.QuoteTokenID != "usdt" {
		t.Errorf("QuoteTokenID = %s, want usdt", req.QuoteTokenID)
	}
}

func TestOrderListParams(t *testing.T) {
	params := &OrderListParams{
		MarketID: "market-1",
		TokenID:  "12345",
		Status:   OrderStatusPending,
		Side:     "BUY",
		Page:     1,
		Limit:    20,
	}

	if params.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", params.MarketID)
	}
	if params.Status != OrderStatusPending {
		t.Errorf("Status = %s, want pending", params.Status)
	}
}

func TestTradeListParams(t *testing.T) {
	params := &TradeListParams{
		MarketID: "market-1",
		TokenID:  "12345",
		Page:     1,
		Limit:    20,
	}

	if params.MarketID != "market-1" {
		t.Errorf("MarketID = %s, want market-1", params.MarketID)
	}
}

func TestTxResponse(t *testing.T) {
	resp := &TxResponse{
		APIResponse: APIResponse{Code: 0, Message: "success"},
	}
	resp.Result.TxHash = "0xabc123"
	resp.Result.Success = true

	if resp.Result.TxHash != "0xabc123" {
		t.Errorf("TxHash = %s, want 0xabc123", resp.Result.TxHash)
	}
	if !resp.Result.Success {
		t.Error("Success should be true")
	}
}
