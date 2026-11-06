package aslan

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type capturedRequest struct {
	method string
	path   string
	query  url.Values
	body   map[string]interface{}
}

// newTestClient starts a server that records the request and replies with status/body.
func newTestClient(t *testing.T, status int, body string) (*Client, *capturedRequest) {
	t.Helper()
	got := &capturedRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.method = r.Method
		got.path = r.URL.Path
		got.query = r.URL.Query()
		if b, _ := io.ReadAll(r.Body); len(b) > 0 {
			if err := json.Unmarshal(b, &got.body); err != nil {
				t.Errorf("request body is not a JSON object: %s", b)
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, body)
	}))
	t.Cleanup(srv.Close)
	c, err := NewClient("sk_test_x", WithBaseURL(srv.URL), WithMaxRetries(0))
	if err != nil {
		t.Fatal(err)
	}
	return c, got
}

func assertKeys(t *testing.T, where string, m map[string]interface{}, keys ...string) {
	t.Helper()
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("%s: missing key %q in %v", where, k, m)
		}
	}
}

func TestCheckoutSessionCreateWireNames(t *testing.T) {
	c, got := newTestClient(t, 201, `{"id":"cs_1","token":"tok","url":"https://pay/x","expiresAt":"2026-11-06T10:00:00.000Z","amount":1000,"currency":"MAD","status":"open","openAmount":false,"metadata":{"order":"42","__cof":true}}`)
	s, err := c.CheckoutSessions.Create(context.Background(), &CreateCheckoutSessionParams{
		Amount:          1000,
		SuccessURL:      "https://m/ok",
		CancelURL:       "https://m/ko",
		Customer:        &SessionCustomer{Email: "a@b.ma", PhoneCountryCode: "+212"},
		BillingAddress:  &Address{Line1: "1 rue", PostalCode: "20000"},
		ShippingAddress: &ShippingAddress{Address: Address{PostalCode: "20000"}, RecipientName: "Ali"},
		LineItems:       []LineItem{{Name: "x", Quantity: 1, UnitPrice: 1000, ImageURL: "https://i/x.png", SellerID: "s1"}},
		Require3DS:      Bool(false),
	})
	if err != nil {
		t.Fatal(err)
	}
	b := got.body
	assertKeys(t, "checkout body", b, "amount", "successUrl", "cancelUrl", "customer", "billingAddress", "shippingAddress", "lineItems", "require3ds")
	assertKeys(t, "customer", b["customer"].(map[string]interface{}), "phoneCountryCode")
	assertKeys(t, "billingAddress", b["billingAddress"].(map[string]interface{}), "line1", "postalCode")
	assertKeys(t, "shippingAddress", b["shippingAddress"].(map[string]interface{}), "postalCode", "recipientName")
	assertKeys(t, "lineItem", b["lineItems"].([]interface{})[0].(map[string]interface{}), "unitPrice", "imageUrl", "sellerId")
	if b["require3ds"] != false {
		t.Errorf("require3ds = %v", b["require3ds"])
	}
	// Checkout responses are not wrapped in { data }.
	if s.ID != "cs_1" || s.ExpiresAt == "" || s.Metadata["order"] != "42" || s.Metadata["__cof"] != "true" {
		t.Errorf("unexpected session: %+v", s)
	}
}

func TestTransactionsRetrieveUnwrapsData(t *testing.T) {
	c, _ := newTestClient(t, 200, `{"data":{"id":"tx_1","amount":500,"currency":"MAD","status":"succeeded","checkoutSessionId":"cs_1","paidAt":"p","createdAt":"c","updatedAt":"u","metadata":{"__changeCard":true},"providerAccount":"default"}}`)
	tx, err := c.Transactions.Retrieve(context.Background(), "tx_1")
	if err != nil {
		t.Fatal(err)
	}
	if tx.ID != "tx_1" || tx.Amount != 500 || tx.CheckoutSessionID != "cs_1" || tx.PaidAt != "p" || tx.CreatedAt != "c" || tx.UpdatedAt != "u" || tx.ProviderAccount != "default" {
		t.Errorf("unexpected transaction: %+v", tx)
	}
}

func TestMissingDataEnvelopeIsAnError(t *testing.T) {
	c, _ := newTestClient(t, 200, `{"id":"tx_1"}`)
	if _, err := c.Transactions.Retrieve(context.Background(), "tx_1"); err == nil {
		t.Fatal("expected an error for a response without a data envelope")
	}
}

func TestTransactionsListQueryAndPagination(t *testing.T) {
	c, got := newTestClient(t, 200, `{"data":[{"id":"tx_1"},{"id":"tx_2"}],"pagination":{"total":42,"page":2,"pageSize":20,"totalPages":3}}`)
	res, err := c.Transactions.List(context.Background(), &ListTransactionsParams{
		Page: Int(2), PageSize: Int(20), SortBy: String("amount"), SortOrder: String("asc"),
	})
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{"page": "2", "pageSize": "20", "sortBy": "amount", "sortOrder": "asc"} {
		if got.query.Get(k) != want {
			t.Errorf("query %s = %q, want %q (query: %v)", k, got.query.Get(k), want, got.query)
		}
	}
	p := res.Pagination
	if len(res.Data) != 2 || p.Total != 42 || p.Page != 2 || p.PageSize != 20 || p.TotalPages != 3 {
		t.Errorf("unexpected list: %+v", res)
	}
}

func TestRefundsListFiltersByTransaction(t *testing.T) {
	c, got := newTestClient(t, 200, `{"data":[],"pagination":{"total":0,"page":1,"pageSize":20,"totalPages":1}}`)
	if _, err := c.Refunds.List(context.Background(), &ListRefundsParams{TransactionID: String("11111111-1111-1111-1111-111111111111"), PageSize: Int(5)}); err != nil {
		t.Fatal(err)
	}
	if got.query.Get("transactionId") != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("transactionId filter not sent: %v", got.query)
	}
	if got.query.Has("transaction_id") || got.query.Has("page_size") {
		t.Errorf("snake_case query keys sent (silently ignored by the API): %v", got.query)
	}
	if got.query.Get("pageSize") != "5" {
		t.Errorf("pageSize = %q", got.query.Get("pageSize"))
	}
}

func TestRefundsCreateWireAndUnwrap(t *testing.T) {
	c, got := newTestClient(t, 201, `{"data":{"id":"rf_1","transactionId":"tx_1","amount":100,"currency":"MAD","status":"pending","createdAt":"c"}}`)
	rf, err := c.Refunds.Create(context.Background(), &CreateRefundParams{TransactionID: "tx_1", Amount: 100})
	if err != nil {
		t.Fatal(err)
	}
	assertKeys(t, "refund body", got.body, "transactionId", "amount")
	if rf.ID != "rf_1" || rf.TransactionID != "tx_1" || rf.CreatedAt != "c" {
		t.Errorf("unexpected refund: %+v", rf)
	}
}

func TestCustomersListQuery(t *testing.T) {
	c, got := newTestClient(t, 200, `{"data":[{"id":"cu_1","createdAt":"c"}],"pagination":{"total":1,"page":1,"pageSize":10,"totalPages":1}}`)
	res, err := c.Customers.List(context.Background(), &ListCustomersParams{PageSize: Int(10), SortBy: String("email"), SortOrder: String("asc")})
	if err != nil {
		t.Fatal(err)
	}
	for k, want := range map[string]string{"pageSize": "10", "sortBy": "email", "sortOrder": "asc"} {
		if got.query.Get(k) != want {
			t.Errorf("query %s = %q, want %q", k, got.query.Get(k), want)
		}
	}
	if res.Data[0].CreatedAt != "c" || res.Pagination.PageSize != 10 {
		t.Errorf("unexpected list: %+v", res)
	}
}

func TestPaymentLinkWireAndUnwrap(t *testing.T) {
	c, got := newTestClient(t, 201, `{"data":{"id":"pl_1","url":"https://u","amount":100,"currency":"MAD","status":"active","maxPayments":3,"paymentCount":1,"expiresAt":"e","createdAt":"c"}}`)
	pl, err := c.PaymentLinks.Create(context.Background(), &CreatePaymentLinkParams{Amount: Int(100), MaxPayments: Int(3), ExpiresAt: String("2026-12-01T00:00:00Z")})
	if err != nil {
		t.Fatal(err)
	}
	assertKeys(t, "payment link body", got.body, "maxPayments", "expiresAt")
	if pl.ID != "pl_1" || pl.MaxPayments == nil || *pl.MaxPayments != 3 || pl.PaymentCount != 1 || pl.ExpiresAt != "e" || pl.CreatedAt != "c" {
		t.Errorf("unexpected link: %+v", pl)
	}
}

func TestPaymentLinkQRFields(t *testing.T) {
	c, _ := newTestClient(t, 200, `{"data":{"dataUri":"data:image/png;base64,AA","mimeType":"image/png"}}`)
	qr, err := c.PaymentLinks.GetQR(context.Background(), "pl_1", nil)
	if err != nil {
		t.Fatal(err)
	}
	if qr.DataURI != "data:image/png;base64,AA" || qr.MimeType != "image/png" {
		t.Errorf("unexpected QR: %+v", qr)
	}
}

func asAslanError(t *testing.T, err error) *AslanError {
	t.Helper()
	var ae *AslanError
	if !errors.As(err, &ae) {
		t.Fatalf("expected *AslanError, got %T: %v", err, err)
	}
	return ae
}

func TestFlatErrorBody(t *testing.T) {
	c, _ := newTestClient(t, 400, `{"statusCode":400,"code":"PROVIDER_ACCOUNT_NOT_FOUND","error":"Bad Request","message":"Unknown provider account 'x'","details":{"alias":"x"}}`)
	_, err := c.CheckoutSessions.Create(context.Background(), &CreateCheckoutSessionParams{Amount: 100})
	ae := asAslanError(t, err)
	if ae.Code != ErrProviderAccountNotFound || ae.Message != "Unknown provider account 'x'" || ae.HTTPStatus != 400 || ae.Details["alias"] != "x" {
		t.Errorf("unexpected error: %+v", ae)
	}
}

func TestFlatErrorWithoutMessageUsesReason(t *testing.T) {
	c, _ := newTestClient(t, 404, `{"statusCode":404,"code":"NOT_FOUND","error":"Not Found"}`)
	_, err := c.Refunds.Retrieve(context.Background(), "rf_x")
	ae := asAslanError(t, err)
	if ae.Code != ErrNotFound || ae.Message != "Not Found" {
		t.Errorf("unexpected error: %+v", ae)
	}
}

func TestNestedErrorBody(t *testing.T) {
	c, _ := newTestClient(t, 400, `{"error":{"code":"VALIDATION_ERROR","message":"Invalid refund data","details":[{"field":"amount","message":"Required"}]}}`)
	_, err := c.Refunds.Create(context.Background(), &CreateRefundParams{TransactionID: "tx"})
	ae := asAslanError(t, err)
	if ae.Code != ErrValidation || ae.Message != "Invalid refund data" {
		t.Errorf("unexpected error: %+v", ae)
	}
	if list, ok := ae.Details["errors"].([]interface{}); !ok || len(list) != 1 {
		t.Errorf("array details not preserved: %+v", ae.Details)
	}
}

func TestNestedErrorObjectDetails(t *testing.T) {
	c, _ := newTestClient(t, 404, `{"error":{"code":"NOT_FOUND","message":"Transaction 'x' not found","details":{"entity":"Transaction","id":"x"}}}`)
	_, err := c.Transactions.Retrieve(context.Background(), "x")
	ae := asAslanError(t, err)
	if ae.Code != ErrNotFound || ae.Message != "Transaction 'x' not found" || ae.Details["entity"] != "Transaction" {
		t.Errorf("unexpected error: %+v", ae)
	}
}

func TestWebhookEventFields(t *testing.T) {
	var ev WebhookEvent
	if err := json.Unmarshal([]byte(`{"id":"evt_1","event":"payment.succeeded","createdAt":"c","data":{"transaction":{"id":"tx"}}}`), &ev); err != nil {
		t.Fatal(err)
	}
	if ev.Type != "payment.succeeded" || ev.CreatedAt != "c" || ev.Data["transaction"] == nil {
		t.Errorf("unexpected event: %+v", ev)
	}
}
