package aslan

// CheckoutSession represents a checkout session returned by the API.
type CheckoutSession struct {
	ID        string            `json:"id"`
	Token     string            `json:"token"`
	URL       string            `json:"url"`
	ExpiresAt string            `json:"expires_at"`
	Amount    int               `json:"amount"`
	Currency  string            `json:"currency"`
	Status    string            `json:"status"`
	Metadata  map[string]string `json:"metadata,omitempty"`
}

// CreateCheckoutSessionParams contains the parameters for creating a checkout session.
type CreateCheckoutSessionParams struct {
	// Amount in centimes (100 = 1.00 MAD). Min: 100, Max: 100,000,000
	Amount   int    `json:"amount"`
	Currency string `json:"currency,omitempty"`
	// Redirect URL on successful payment (must be HTTPS)
	SuccessURL string `json:"success_url"`
	// Redirect URL on cancellation (must be HTTPS)
	CancelURL       string           `json:"cancel_url"`
	Customer        *Customer        `json:"customer,omitempty"`
	BillingAddress  *Address         `json:"billing_address,omitempty"`
	ShippingAddress *ShippingAddress  `json:"shipping_address,omitempty"`
	LineItems       []LineItem       `json:"line_items,omitempty"`
	Metadata        map[string]string `json:"metadata,omitempty"`
	// IdempotencyKey is sent as the Idempotency-Key header if set.
	IdempotencyKey string `json:"-"`
}

// Customer contains optional customer information for a checkout session.
type Customer struct {
	Email            string `json:"email,omitempty"`
	Name             string `json:"name,omitempty"`
	Phone            string `json:"phone,omitempty"`
	PhoneCountryCode string `json:"phone_country_code,omitempty"`
}

// Address represents a billing or base address.
type Address struct {
	Line1      string `json:"line1,omitempty"`
	City       string `json:"city,omitempty"`
	PostalCode string `json:"postal_code,omitempty"`
	State      string `json:"state,omitempty"`
	// Country is an ISO 3166-1 alpha-2 country code.
	Country string `json:"country,omitempty"`
}

// ShippingAddress extends Address with a recipient name.
type ShippingAddress struct {
	Address
	RecipientName string `json:"recipient_name,omitempty"`
}

// LineItem represents a single item in a checkout session.
type LineItem struct {
	Name      string `json:"name"`
	Quantity  int    `json:"quantity"`
	UnitPrice int    `json:"unit_price"`
	Description string `json:"description,omitempty"`
	SKU       string `json:"sku,omitempty"`
	ImageURL  string `json:"image_url,omitempty"`
	Category  string `json:"category,omitempty"`
	SellerID  string `json:"seller_id,omitempty"`
}

// Transaction represents a payment transaction.
type Transaction struct {
	ID                string            `json:"id"`
	Amount            int               `json:"amount"`
	Currency          string            `json:"currency"`
	Status            string            `json:"status"`
	MerchantID        string            `json:"merchant_id"`
	CheckoutSessionID string            `json:"checkout_session_id,omitempty"`
	PaidAt            string            `json:"paid_at,omitempty"`
	CreatedAt         string            `json:"created_at"`
	UpdatedAt         string            `json:"updated_at"`
	Metadata          map[string]string `json:"metadata,omitempty"`
	Customer          *Customer         `json:"customer,omitempty"`
}

// ListTransactionsParams contains the query parameters for listing transactions.
type ListTransactionsParams struct {
	Page      *int    `json:"-"`
	PageSize  *int    `json:"-"`
	Status    *string `json:"-"`
	From      *string `json:"-"`
	To        *string `json:"-"`
	SortBy    *string `json:"-"`
	SortOrder *string `json:"-"`
}

// PaginatedResponse wraps a paginated list of items.
type PaginatedResponse[T any] struct {
	Data       []T        `json:"data"`
	Pagination Pagination `json:"pagination"`
}

// Pagination contains pagination metadata.
type Pagination struct {
	Page       int `json:"page"`
	PageSize   int `json:"page_size"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// PaymentLink represents a reusable payment link.
type PaymentLink struct {
	ID           string            `json:"id"`
	URL          string            `json:"url"`
	Amount       int               `json:"amount"`
	Currency     string            `json:"currency"`
	Description  string            `json:"description,omitempty"`
	Status       string            `json:"status"`
	MaxPayments  *int              `json:"max_payments,omitempty"`
	PaymentCount int               `json:"payment_count"`
	ExpiresAt    string            `json:"expires_at,omitempty"`
	CreatedAt    string            `json:"created_at"`
	Metadata     map[string]string `json:"metadata,omitempty"`
}

// CreatePaymentLinkParams contains the parameters for creating a payment link.
type CreatePaymentLinkParams struct {
	// Amount in centimes
	Amount      int               `json:"amount"`
	Currency    string            `json:"currency,omitempty"`
	Description *string           `json:"description,omitempty"`
	MaxPayments *int              `json:"max_payments,omitempty"`
	ExpiresAt   *string           `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
	// IdempotencyKey is sent as the Idempotency-Key header if set.
	IdempotencyKey string `json:"-"`
}

// ListPaymentLinksParams contains the query parameters for listing payment links.
type ListPaymentLinksParams struct {
	Page     *int    `json:"-"`
	PageSize *int    `json:"-"`
	Status   *string `json:"-"`
}

// Refund represents a refund on a transaction.
type Refund struct {
	ID            string `json:"id"`
	TransactionID string `json:"transaction_id"`
	Amount        int    `json:"amount"`
	Currency      string `json:"currency"`
	Status        string `json:"status"`
	Reason        string `json:"reason,omitempty"`
	CreatedAt     string `json:"created_at"`
}

// CreateRefundParams contains the parameters for creating a refund.
type CreateRefundParams struct {
	TransactionID string `json:"transaction_id"`
	// Refund amount in centimes
	Amount int    `json:"amount"`
	Reason string `json:"reason,omitempty"`
	// IdempotencyKey is sent as the Idempotency-Key header if set.
	IdempotencyKey string `json:"-"`
}

// UpdatePaymentLinkParams contains the parameters for updating a payment link.
type UpdatePaymentLinkParams struct {
	Description *string           `json:"description,omitempty"`
	Status      *string           `json:"status,omitempty"`
	MaxPayments *int              `json:"max_payments,omitempty"`
	ExpiresAt   *string           `json:"expires_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// QRCodeResult contains the generated QR code data.
type QRCodeResult struct {
	DataURI  string `json:"data_uri"`
	MimeType string `json:"mime_type"`
}

// QRCodeParams contains the query parameters for generating a QR code.
type QRCodeParams struct {
	Format string
	Size   int
}

// ListRefundsParams contains the query parameters for listing refunds.
type ListRefundsParams struct {
	Page          *int    `json:"-"`
	PageSize      *int    `json:"-"`
	Status        *string `json:"-"`
	TransactionID *string `json:"-"`
}

// WebhookEvent represents a parsed and verified webhook event.
type WebhookEvent struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Data      map[string]interface{} `json:"data"`
	CreatedAt string                 `json:"created_at"`
}

// Int returns a pointer to the given int value.
func Int(v int) *int {
	return &v
}

// String returns a pointer to the given string value.
func String(v string) *string {
	return &v
}
