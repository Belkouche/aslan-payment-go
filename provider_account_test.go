package aslan

import (
	"encoding/json"
	"testing"
)

func TestProviderAccountRequestEncoding(t *testing.T) {
	alias := "foreign"
	b, _ := json.Marshal(&CreateCheckoutSessionParams{Amount: 100, ProviderAccount: &alias})
	var m map[string]interface{}
	_ = json.Unmarshal(b, &m)
	if m["providerAccount"] != "foreign" {
		t.Fatalf("checkout: providerAccount not encoded: %s", b)
	}
	b, _ = json.Marshal(&CreatePaymentLinkParams{ProviderAccount: &alias})
	m = nil
	_ = json.Unmarshal(b, &m)
	if m["providerAccount"] != "foreign" {
		t.Fatalf("payment link: providerAccount not encoded: %s", b)
	}
	b, _ = json.Marshal(&CreateCheckoutSessionParams{Amount: 100})
	m = map[string]interface{}{}
	_ = json.Unmarshal(b, &m)
	if _, ok := m["providerAccount"]; ok {
		t.Fatalf("providerAccount must be omitted when unset: %s", b)
	}
}

func TestProviderAccountResponseDecoding(t *testing.T) {
	var tx Transaction
	if err := json.Unmarshal([]byte(`{"id":"t","providerAccount":"store-2"}`), &tx); err != nil || tx.ProviderAccount != "store-2" {
		t.Fatalf("transaction: %v %q", err, tx.ProviderAccount)
	}
}

func TestMapErrorCodeProviderAccount(t *testing.T) {
	for _, c := range []string{ErrProviderAccountNotFound, ErrProviderAccountInactive, ErrProviderAccountUnsupported, ErrProviderAccountRecurringDisabled} {
		if got := mapErrorCode(400, c); got != c {
			t.Fatalf("mapErrorCode(400, %q) = %q", c, got)
		}
	}
}
