package arangodb

import (
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/repository"
	"github.com/stretchr/testify/require"
)

// testLastName repeats across fixtures for the last-name probe.
const testLastName = "Bell"

// autocompleteFixture describes one order inserted for autocomplete
// probes. Empty info fields keep the corresponding embedded profile null.
type autocompleteFixture struct {
	po         string
	first      string
	last       string
	org        string
	payerFirst string
	payerLast  string
	payerOrg   string
}

func newAutocompleteOrder(fixt autocompleteFixture) *order.NewOrder {
	attrs := &order.NewOrderAttributes{
		Courier:          "FedEx",
		CourierAccount:   "9912378999",
		Comments:         "autocompletion fixture",
		Payment:          "Credit card",
		PurchaseOrderNum: fixt.po,
		Status:           order.OrderStatus_IN_PREPARATION,
		Consumer:         fmt.Sprintf("%s@fixture.com", fixt.first),
		Payer:            fmt.Sprintf("%s@fixture.com", fixt.payerFirst),
		Purchaser:        "purchaser@fixture.com",
		Items:            []string{stockItemID},
	}
	if fixt.org != "" {
		attrs.ConsumerInfo = acUserInfo(fixt.first, fixt.last, fixt.org)
	}
	if fixt.payerOrg != "" {
		attrs.PayerInfo = acUserInfo(fixt.payerFirst, fixt.payerLast, fixt.payerOrg)
	}

	return &order.NewOrder{
		Data: &order.NewOrder_Data{
			Type:       orderDocType,
			Attributes: attrs,
		},
	}
}

func acUserInfo(first, last, org string) *order.UserInfo {
	return &order.UserInfo{
		FirstName:    first,
		LastName:     last,
		Organization: org,
		City:         "New York",
		State:        "NY",
		Country:      "USA",
	}
}

// displayTexts collects the display text of every suggestion.
func displayTexts(sugs []*repository.Suggestion) []string {
	out := make([]string, 0, len(sugs))
	for _, sug := range sugs {
		out = append(out, sug.DisplayText)
	}

	return out
}

// fieldOf returns the field name of the suggestion with the given
// display text, empty when absent.
func fieldOf(sugs []*repository.Suggestion, text string) string {
	for _, sug := range sugs {
		if sug.DisplayText == text {
			return sug.Field
		}
	}

	return ""
}

func TestAutocomplete(t *testing.T) { //nolint:funlen,paralleltest // sequential: owns the collection
	req := require.New(t)
	connP := getConnectParams()
	repo, err := NewOrderRepo(connP, collection)
	req.NoErrorf(err, "expect no error, received %s", err)
	req.NoError(repo.ClearOrders(), "failed to clear orders at start")
	defer func() { _ = repo.ClearOrders() }()
	fixtures := []autocompleteFixture{
		{po: "PO-2026-0012", first: "Anna", last: testLastName, org: "Apple Inc",
			payerFirst: "Peter", payerLast: "Payne", payerOrg: "Pineapple Labs"},
		{po: "PO-2026-0341", first: "Bob", last: "Stone", org: "Apparel Corp",
			payerFirst: "Rita", payerLast: "Cole", payerOrg: "Pinnacle Trust"},
		{po: "PO-2025-0918", first: "Carla", last: "Dee", org: "Banana Metrics",
			payerFirst: "Carla", payerLast: "Dee", payerOrg: "Banana Metrics"},
		{po: "PO-2026-0007"},
		{po: "PO-2026-0500", first: "Dana", last: testLastName, org: "Piedmont Supply",
			payerFirst: "Dana", payerLast: testLastName, payerOrg: "Piedmont Supply"},
	}
	for _, fixt := range fixtures {
		_, err := repo.AddOrder(newAutocompleteOrder(fixt))
		req.NoErrorf(err, "expect no error, received %s", err)
	}

	// Prime the search assets; the first call creates the analyzers and
	// the view. The links need to commit before queries see the fixtures.
	var pollErr error
	var pollLen int
	req.Eventually(func() bool {
		sugs, err := repo.Autocomplete("PO-20", 5)
		pollLen = len(sugs)
		if err != nil {
			pollErr = err

			return false
		}
		return pollLen >= 3
	}, 15*time.Second, 500*time.Millisecond,
		"view should return indexed orders, last=%d last error: %s", pollLen, pollErr)

	// Prefix match on the consumer organization.
	sugs, err := repo.Autocomplete("app", 5)
	req.NoError(err)
	req.Contains(displayTexts(sugs), "Apple Inc", "should match Apple Inc")
	req.Contains(displayTexts(sugs), "Apparel Corp", "should match Apparel Corp")
	req.NotContains(displayTexts(sugs), "Banana Metrics", "should not match unrelated org")
	req.Equal(sugFieldConsumerOrg, fieldOf(sugs, "Apple Inc"),
		"should report the consumer organization field")

	// Prefix match on the payer organization.
	sugs, err = repo.Autocomplete("pine", 5)
	req.NoError(err)
	req.Contains(displayTexts(sugs), "Pineapple Labs", "should match payer org")
	req.Equal(sugFieldPayerOrg, fieldOf(sugs, "Pineapple Labs"),
		"should report the payer organization field")

	// Exact first name match without junk from similar names.
	sugs, err = repo.Autocomplete("anna", 5)
	req.NoError(err)
	req.Contains(displayTexts(sugs), "Anna", "should match Anna")
	req.NotContains(displayTexts(sugs), "Dana", "should not match Dana for anna")
	req.Equal(sugFieldConsumerFirst, fieldOf(sugs, "Anna"),
		"should report the consumer first name field")

	// Typo tolerated by the fuzzy stage.
	sugs, err = repo.Autocomplete("aple", 5)
	req.NoError(err)
	req.Contains(displayTexts(sugs), "Apple Inc", "should match Apple Inc with typo")
	req.Equal(sugFieldConsumerOrg, fieldOf(sugs, "Apple Inc"),
		"typo match should come from the organization field")

	// Prefix of a mid-document value.
	sugs, err = repo.Autocomplete("bana", 5)
	req.NoError(err)
	req.Contains(displayTexts(sugs), "Banana Metrics", "should match Banana Metrics")

	// Purchase order prefix across all purchase orders.
	sugs, err = repo.Autocomplete("po-20", 5)
	req.NoError(err)
	poCount := 0
	for _, sug := range sugs {
		if strings.HasPrefix(sug.DisplayText, "PO-") {
			poCount++
		}
	}
	req.GreaterOrEqual(poCount, 3, "should match purchase orders")

	// Last name matches both orders with the same last name.
	sugs, err = repo.Autocomplete("bell", 5)
	req.NoError(err)
	req.Len(sugs, 2, "should match Anna Bell and Dana Bell")
	req.Equal(sugFieldConsumerLast, sugs[0].Field,
		"consumer last name should outrank the payer last name")

	// No match returns an empty result.
	sugs, err = repo.Autocomplete("zzzz", 5)
	req.NoError(err)
	req.Empty(sugs, "should return no suggestions")

	// The limit caps the result size.
	sugs, err = repo.Autocomplete("po-20", 2)
	req.NoError(err)
	req.Len(sugs, 2, "should return at most the requested limit")

	// Zero limit falls back to the default limit.
	sugs, err = repo.Autocomplete("po-20", 0)
	req.NoError(err)
	req.GreaterOrEqual(len(sugs), 3, "should use the default limit")

	// Sparse order matches on its purchase order first.
	sugs, err = repo.Autocomplete("PO-2026-0007", 5)
	req.NoError(err)
	req.NotEmpty(sugs, "should match the purchase order")
	req.Equal(sugFieldPurchaseOrderNum, sugs[0].Field, "should report the PO field")
	req.Equal("PO-2026-0007", sugs[0].DisplayText, "should report the PO number")
}
