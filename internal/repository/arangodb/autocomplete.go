package arangodb

import (
	"context"
	"fmt"
	"strings"

	driver "github.com/arangodb/go-driver"
	"github.com/dictyBase/modware-order/internal/repository"
)

// Names of the search assets created by the repository. The AQL statement
// autocompleteQuery references the same view and analyzers, so any change
// here must be mirrored there.
const (
	// searchViewName is the ArangoSearch view linked to the stock order
	// collection.
	searchViewName = "orders_search"
	// prefixAnalyzer is a norm analyzer for case-insensitive exact prefix
	// matching.
	prefixAnalyzer = "autocomplete_norm"
	// ngramAnalyzer is a pipeline analyzer (norm followed by ngram) for
	// fuzzy substring matching with typo tolerance.
	ngramAnalyzer = "autocomplete_ngram"
	// defaultAutocompleteLimit is the number of suggestions returned when
	// the caller passes a non-positive limit.
	defaultAutocompleteLimit = 5
	// ngramThreshold is the minimum n-gram similarity required by the
	// fuzzy stage. Calibrated against real data: exact and prefix matches
	// pass at 0.45 while low quality n-gram overlaps are rejected. The
	// value 0.3 admits junk that outranks real matches, 0.55 drops valid
	// short prefix matches like pine -> Pineapple.
	ngramThreshold = 0.45
)

// Suggestion field names returned by the repository. They mirror the
// field paths of the autocompleteQuery projection.
const (
	sugFieldPurchaseOrderNum = "purchase_order_num"
	sugFieldConsumerOrg      = "consumer_info.organization"
	sugFieldConsumerFirst    = "consumer_info.first_name"
	sugFieldConsumerLast     = "consumer_info.last_name"
	sugFieldPayerOrg         = "payer_info.organization"
	sugFieldPayerFirst       = "payer_info.first_name"
	sugFieldPayerLast        = "payer_info.last_name"
)

// autocompleteFields lists the order fields indexed for autocomplete
// search, in the same order as the view links.
var autocompleteFields = []string{
	sugFieldPurchaseOrderNum,
	sugFieldConsumerOrg,
	sugFieldConsumerFirst,
	sugFieldConsumerLast,
	sugFieldPayerOrg,
	sugFieldPayerFirst,
	sugFieldPayerLast,
}

// autocompleteQuery assembles one AQL statement that runs a prefix and a
// fuzzy match per field; each result row carries the field that matched.
// Prefix matches carry a large constant score so they rank above fuzzy
// ones.
var autocompleteQuery = buildAutocompleteQuery()

func buildAutocompleteQuery() string {
	branches := make([]string, 0, 2*len(autocompleteFields))
	letVars := make([]string, 0, 2*len(autocompleteFields))
	for idx, fieldPath := range autocompleteFields {
		preVar := fmt.Sprintf("p%d", idx)
		ngVar := fmt.Sprintf("n%d", idx)
		branches = append(branches,
			fmt.Sprintf(autocompletePrefixFieldQuery, preVar, fieldPath, prefixAnalyzer, fieldPath, fieldPath),
			fmt.Sprintf(autocompleteNgramFieldQuery, ngVar, fieldPath, ngramAnalyzer, fieldPath, fieldPath),
		)
		letVars = append(letVars, preVar, ngVar)
	}

	return strings.Join(branches, "\n") + "\n" +
		fmt.Sprintf(autocompleteMergeQuery, strings.Join(letVars, ", "))
}

// suggestionRow mirrors one row of the autocompleteQuery projection.
type suggestionRow struct {
	Key   string  `json:"k"`
	Field string  `json:"f"`
	Value string  `json:"v"`
	Score float64 `json:"s"`
}

func streamUTF8() *driver.ArangoSearchNGramStreamType {
	return new(driver.ArangoSearchNGramStreamUTF8)
}

// ensureSearch creates the analyzers and the search view when they are
// absent. All operations are idempotent, so concurrent callers and
// multiple repository instances over the same database are safe.
func (ar *arangorepository) ensureSearch(ctx context.Context) error {
	hnd := ar.database.Handler()
	if _, _, err := hnd.EnsureCreatedAnalyzer(ctx, &driver.ArangoSearchAnalyzerDefinition{
		Name: prefixAnalyzer,
		Type: driver.ArangoSearchAnalyzerTypeNorm,
		Properties: driver.ArangoSearchAnalyzerProperties{
			Locale: "en.utf-8",
			Case:   driver.ArangoSearchCaseLower,
			Accent: new(false),
		},
	}); err != nil {
		return fmt.Errorf("error in ensuring analyzer %s %s", prefixAnalyzer, err)
	}
	ngram := driver.ArangoSearchAnalyzerProperties{
		Min:              new(int64(2)),
		Max:              new(int64(3)),
		PreserveOriginal: new(true),
		StreamType:       streamUTF8(),
	}
	if _, _, err := hnd.EnsureCreatedAnalyzer(ctx, &driver.ArangoSearchAnalyzerDefinition{
		Name: ngramAnalyzer,
		Type: driver.ArangoSearchAnalyzerTypePipeline,
		Properties: driver.ArangoSearchAnalyzerProperties{
			Pipeline: []driver.ArangoSearchAnalyzerPipeline{
				{
					Type: driver.ArangoSearchAnalyzerTypeNorm,
					Properties: driver.ArangoSearchAnalyzerProperties{
						Locale: "en.utf-8",
						Case:   driver.ArangoSearchCaseLower,
						Accent: new(false),
					},
				},
				{
					Type:       driver.ArangoSearchAnalyzerTypeNGram,
					Properties: ngram,
				},
			},
		},
		Features: []driver.ArangoSearchAnalyzerFeature{
			driver.ArangoSearchAnalyzerFeatureFrequency,
			driver.ArangoSearchAnalyzerFeatureNorm,
			driver.ArangoSearchAnalyzerFeaturePosition,
		},
	}); err != nil {
		return fmt.Errorf("error in ensuring analyzer %s %s", ngramAnalyzer, err)
	}

	return ar.ensureSearchView(ctx)
}

// ensureSearchView creates the autocomplete view with both analyzers
// linked to every indexed field.
func (ar *arangorepository) ensureSearchView(ctx context.Context) error {
	hnd := ar.database.Handler()
	exists, err := hnd.ViewExists(ctx, searchViewName)
	if err != nil {
		return fmt.Errorf("error in checking view %s %s", searchViewName, err)
	}
	if exists {
		return nil
	}
	anl := []string{prefixAnalyzer, ngramAnalyzer}
	links := driver.ArangoSearchLinks{
		ar.sorder.Name(): {
			Fields: driver.ArangoSearchFields{
				"purchase_order_num": {Analyzers: anl},
				"consumer_info": {
					Fields: driver.ArangoSearchFields{
						"organization": {Analyzers: anl},
						"first_name":   {Analyzers: anl},
						"last_name":    {Analyzers: anl},
					},
				},
				"payer_info": {
					Fields: driver.ArangoSearchFields{
						"organization": {Analyzers: anl},
						"first_name":   {Analyzers: anl},
						"last_name":    {Analyzers: anl},
					},
				},
			},
		},
	}
	if _, err := hnd.CreateArangoSearchView(ctx, searchViewName, &driver.ArangoSearchViewProperties{
		Links: links,
	}); err != nil && !driver.IsConflict(err) {
		return fmt.Errorf("error in creating view %s %s", searchViewName, err)
	}

	return nil
}

// toSuggestions converts raw query rows into repository suggestions.
func toSuggestions(rows []suggestionRow) []*repository.Suggestion {
	out := make([]*repository.Suggestion, 0, len(rows))
	for _, row := range rows {
		out = append(out, &repository.Suggestion{
			ID:          row.Key,
			Field:       row.Field,
			DisplayText: row.Value,
			Score:       row.Score,
		})
	}

	return out
}

// Autocomplete suggests orders for a partial search text. The search
// assets were created when the repository was constructed.
func (ar *arangorepository) Autocomplete(query string, limit int) ([]*repository.Suggestion, error) {
	if limit <= 0 {
		limit = defaultAutocompleteLimit
	}
	// STARTS_WITH compares the raw query text against the lowercased
	// document tokens, so the query must be lowercased here; the ngram
	// pipeline normalizes both sides but expects the same shape.
	query = strings.ToLower(strings.TrimSpace(query))
	result, err := ar.database.SearchRows(autocompleteQuery, map[string]any{
		"q":     query,
		"th":    ngramThreshold,
		"limit": limit,
	})
	if err != nil {
		return nil, fmt.Errorf("error in running autocomplete query %s", err)
	}
	defer result.Close() //nolint:errcheck
	rows := make([]suggestionRow, 0, limit)
	for result.Scan() {
		var row suggestionRow
		if err := result.Read(&row); err != nil {
			return nil, fmt.Errorf("error in reading autocomplete result %s", err)
		}
		rows = append(rows, row)
	}

	return toSuggestions(rows), nil
}
