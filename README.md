# Aslan Payment Go SDK

Official Go SDK for the [Aslan Payment API](https://aslanpay.ma).

## Requirements

- Go 1.21 or later
- No external dependencies (stdlib only)

## Installation

```bash
go get github.com/Belkouche/aslan-payment-go
```

## Quick Start

```go
package main

import (
    "context"
    "fmt"
    "log"

    aslan "github.com/Belkouche/aslan-payment-go"
)

func main() {
    client, err := aslan.NewClient("sk_test_xxx")
    if err != nil {
        log.Fatal(err)
    }

    ctx := context.Background()

    session, err := client.CheckoutSessions.Create(ctx, &aslan.CreateCheckoutSessionParams{
        Amount:     15000,
        Currency:   "MAD",
        SuccessURL: "https://example.com/success",
        CancelURL:  "https://example.com/cancel",
    })
    if err != nil {
        log.Fatal(err)
    }

    fmt.Printf("Checkout URL: %s\n", session.URL)
}
```

## Configuration

Use functional options to customize the client:

```go
client, err := aslan.NewClient("sk_test_xxx",
    aslan.WithBaseURL("https://staging.aslanpay.ma/pay"),
    aslan.WithTimeout(60 * time.Second),
    aslan.WithMaxRetries(3),
)
```

| Option | Default | Description |
|--------|---------|-------------|
| `WithBaseURL` | `https://api.aslanpay.ma/pay` | API base URL |
| `WithTimeout` | `30s` | HTTP request timeout |
| `WithMaxRetries` | `2` | Max retries on 5xx/network errors |

## Resources

### Checkout Sessions

```go
// Create a checkout session
session, err := client.CheckoutSessions.Create(ctx, &aslan.CreateCheckoutSessionParams{
    Amount:     15000,
    Currency:   "MAD",
    SuccessURL: "https://example.com/success",
    CancelURL:  "https://example.com/cancel",
    Customer: &aslan.Customer{
        Email: "customer@example.com",
        Name:  "John Doe",
    },
    Metadata: map[string]string{
        "order_id": "ord_123",
    },
    IdempotencyKey: "unique-key-123",
})

// Retrieve a checkout session
session, err := client.CheckoutSessions.Retrieve(ctx, "cs_xxx")
```

### Transactions

```go
// Retrieve a transaction
txn, err := client.Transactions.Retrieve(ctx, "txn_xxx")

// List transactions with filters
list, err := client.Transactions.List(ctx, &aslan.ListTransactionsParams{
    Page:      aslan.Int(1),
    PageSize:  aslan.Int(10),
    Status:    aslan.String("succeeded"),
    SortBy:    aslan.String("createdAt"),
    SortOrder: aslan.String("desc"),
})

for _, txn := range list.Data {
    fmt.Printf("Transaction %s: %d %s\n", txn.ID, txn.Amount, txn.Currency)
}
```

### Payment Links

```go
// Create a payment link
link, err := client.PaymentLinks.Create(ctx, &aslan.CreatePaymentLinkParams{
    Amount:      15000,
    Description: aslan.String("Order #123"),
    MaxPayments: aslan.Int(1),
})

// Retrieve a payment link
link, err := client.PaymentLinks.Retrieve(ctx, "pl_xxx")

// List payment links
list, err := client.PaymentLinks.List(ctx, &aslan.ListPaymentLinksParams{
    Page:   aslan.Int(1),
    Status: aslan.String("active"),
})
```

### Refunds

```go
// Create a refund
refund, err := client.Refunds.Create(ctx, &aslan.CreateRefundParams{
    TransactionID: "txn_xxx",
    Amount:        5000,
    Reason:        "customer_request",
})

// Retrieve a refund
refund, err := client.Refunds.Retrieve(ctx, "rf_xxx")
```

## Webhook Verification

Verify incoming webhook signatures using HMAC-SHA256:

```go
// With default tolerance (5 minutes)
event, err := aslan.VerifyWebhook("whsec_xxx", payload, signatureHeader)
if err != nil {
    // Signature invalid or timestamp too old
    log.Fatal(err)
}
fmt.Printf("Event type: %s\n", event.Type)

// With custom tolerance (in seconds)
event, err := aslan.VerifyWebhookWithTolerance("whsec_xxx", payload, signatureHeader, 600)
```

The signature header format is `t=<unix_timestamp>,v1=<hex_hmac>`.

## Error Handling

All errors are returned as `*aslan.AslanError`:

```go
session, err := client.CheckoutSessions.Create(ctx, params)
if err != nil {
    var aslanErr *aslan.AslanError
    if errors.As(err, &aslanErr) {
        fmt.Printf("Code: %s\n", aslanErr.Code)
        fmt.Printf("Message: %s\n", aslanErr.Message)
        fmt.Printf("HTTP Status: %d\n", aslanErr.HTTPStatus)
        fmt.Printf("Request ID: %s\n", aslanErr.RequestID)

        switch aslanErr.Code {
        case aslan.ErrAuthentication:
            // Invalid API key
        case aslan.ErrValidation:
            // Invalid parameters
        case aslan.ErrRateLimited:
            // Too many requests
        case aslan.ErrNotFound:
            // Resource not found
        }
    }
    return err
}
```

### Error Codes

| Code | Description |
|------|-------------|
| `AUTHENTICATION_ERROR` | Invalid or missing API key |
| `AUTHORIZATION_ERROR` | Insufficient permissions |
| `VALIDATION_ERROR` | Invalid request parameters |
| `NOT_FOUND` | Resource not found |
| `RATE_LIMITED` | Too many requests |
| `IDEMPOTENCY_ERROR` | Idempotency key conflict |
| `NETWORK_ERROR` | Network or timeout error |
| `SERVER_ERROR` | Server-side error (5xx) |

## Context Support

All API methods accept a `context.Context` for cancellation and deadlines:

```go
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()

session, err := client.CheckoutSessions.Create(ctx, params)
```

## Test Mode

Check if the client is using a test key:

```go
if client.IsTestMode() {
    fmt.Println("Running in test mode")
}
```

## Pointer Helpers

Use `aslan.Int()` and `aslan.String()` to create pointers for optional fields:

```go
params := &aslan.ListTransactionsParams{
    Page:     aslan.Int(1),
    PageSize: aslan.Int(25),
    Status:   aslan.String("succeeded"),
}
```

## Security

- Secret keys are validated on client creation (must start with `sk_live_` or `sk_test_`)
- Webhook signatures use HMAC-SHA256 with constant-time comparison
- Automatic retries with exponential backoff on 5xx and network errors
- Request timeouts via `context.Context`

## License

MIT License - see [LICENSE](LICENSE) for details.
