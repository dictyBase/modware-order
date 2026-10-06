package service

import (
	"context"
	"errors"
	"testing"

	"github.com/dictyBase/go-genproto/dictybaseapis/order"
	"github.com/dictyBase/modware-order/internal/repository"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// acParams builds an autocomplete request with the given query and limit.
func acParams(query string, limit int64) *order.AutocompleteParameters {
	return &order.AutocompleteParameters{
		Data: &order.AutocompleteParameters_Data{
			Type: orderResourceType,
			Attributes: &order.AutocompleteAttributes{
				Query: query,
				Limit: limit,
			},
		},
	}
}

func TestAutocompleteOrderMapsRepositorySuggestions(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	repo := &stubRepository{sugs: []*repository.Suggestion{
		{ID: "8123", Field: "consumer_info.organization",
			DisplayText: "Apple Inc", Score: 1000},
		{ID: "8124", Field: "purchase_order_num",
			DisplayText: "PO-2026-0012", Score: 1.5},
	}}
	svc := newTestService(repo, &stubPublisher{})
	coll, err := svc.AutocompleteOrder(context.Background(), acParams("app", 3))
	req.NoError(err)
	req.Len(coll.Data, 2, "should map every suggestion")
	req.Equal("8123", coll.Data[0].Id, "should carry the order id")
	req.Equal("consumer_info.organization", coll.Data[0].Field, "should carry the field name")
	req.Equal("Apple Inc", coll.Data[0].DisplayText, "should carry the display text")
	req.InDelta(1000, coll.Data[0].Score, 0.001, "should carry the score")
	req.Equal(int64(2), coll.Meta.Total, "total should match the row count")
	req.Equal("app", repo.acQuery, "query should pass through unchanged")
	req.Equal(3, repo.acLimit, "limit should pass through unchanged")
}

func TestAutocompleteOrderRejectsShortQuery(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	_, err := svc.AutocompleteOrder(context.Background(), acParams("ap", 5))
	req.Error(err, "query below the 3 character minimum must be rejected")
	req.Equal(codes.InvalidArgument, status.Code(err))
}

func TestAutocompleteOrderRejectsMissingData(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	_, err := svc.AutocompleteOrder(context.Background(), &order.AutocompleteParameters{})
	req.Error(err, "missing data must be rejected")
	req.Equal(codes.InvalidArgument, status.Code(err))
}

func TestAutocompleteOrderRejectsMissingAttributes(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	_, err := svc.AutocompleteOrder(context.Background(), &order.AutocompleteParameters{
		Data: &order.AutocompleteParameters_Data{Type: orderResourceType},
	})
	req.Error(err, "missing attributes must be rejected")
	req.Equal(codes.InvalidArgument, status.Code(err))
}

func TestAutocompleteOrderPropagatesRepositoryError(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	svc := newTestService(&stubRepository{err: errors.New("db down")}, &stubPublisher{})
	_, err := svc.AutocompleteOrder(context.Background(), acParams("app", 5))
	req.Error(err, "repository errors must propagate")
	req.Equal(codes.Internal, status.Code(err))
}

func TestAutocompleteOrderReturnsEmptyCollection(t *testing.T) {
	t.Parallel()
	req := require.New(t)
	svc := newTestService(&stubRepository{}, &stubPublisher{})
	coll, err := svc.AutocompleteOrder(context.Background(), acParams("zzzz", 5))
	req.NoError(err)
	req.Empty(coll.Data, "no matches should produce no rows")
	req.Equal(int64(0), coll.Meta.Total, "total should be zero")
}
