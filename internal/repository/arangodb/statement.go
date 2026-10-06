package arangodb

const (
	orderIns = `
		INSERT {
			created_at: DATE_ISO8601(DATE_NOW()),
		    updated_at: DATE_ISO8601(DATE_NOW()),
			courier: @courier,
			courier_account: @courier_account,
			comments: @comments,
			payment: @payment,
			purchase_order_num: @purchase_order_num,
			status: @status,
			consumer: @consumer,
			payer: @payer,
			consumer_info: @consumer_info,
			payer_info: @payer_info,
			purchaser: @purchaser,
			items: @items
		} INTO @@stock_order_collection RETURN NEW
	`
	orderLoad = `
		INSERT {
			created_at: DATE_ISO8601(@created_at),
			updated_at: DATE_ISO8601(@updated_at),
			purchaser: @purchaser,
			items: @items
		} INTO @@stock_order_collection RETURN NEW
	`
	orderGet = `
		FOR sorder IN @@stock_order_collection
			FILTER sorder._key == @key
			RETURN sorder
	`
	orderUpd = `
		UPDATE { _key: @key }
			WITH { updated_at: DATE_ISO8601(DATE_NOW()), %s }
			IN @@stock_order_collection RETURN NEW
	`
	orderList = `
		FOR s IN %s
			SORT s.created_at DESC
			LIMIT %d
			RETURN s
	`
	orderListWithFilter = `
		FOR s IN %s
			SORT s.created_at DESC
			%s
			LIMIT %d
			RETURN s
`
	orderListWithCursor = `
		FOR s in %s
			FILTER s.created_at <= DATE_ISO8601(%d)
			SORT s.created_at DESC
			LIMIT %d
			RETURN s
	`
	orderListFilterWithCursor = `
		FOR s IN %s
			FILTER s.created_at <= DATE_ISO8601(%d)
			SORT s.created_at DESC
			%s
			LIMIT %d
			RETURN s
	`
)

// autocompletePrefixFieldQuery matches one field for exact prefixes.
// Template verbs: branch variable, field path, analyzer name, field
// label, field path again.
const autocompletePrefixFieldQuery = `LET %s = (
	FOR d IN orders_search
		SEARCH ANALYZER(STARTS_WITH(d.%s, @q), %q)
		LIMIT @limit
		RETURN { k: d._key, f: %q, v: d.%s, s: 1000 + BM25(d) }
)`

// autocompleteNgramFieldQuery matches one field for fuzzy n-gram
// similarity with the same projection as the prefix query.
const autocompleteNgramFieldQuery = `LET %s = (
	FOR d IN orders_search
		SEARCH NGRAM_MATCH(d.%s, @q, @th, %q)
		LIMIT @limit
		RETURN { k: d._key, f: %q, v: d.%s, s: BM25(d) }
)`

// autocompleteMergeQuery collects the per-field matches, keeps the best
// scoring match per order and returns the top entries. The field paths
// must stay in sync with the links of the view built in autocomplete.go.
const autocompleteMergeQuery = `LET hits = FLATTEN([%s])
LET best = (
	FOR x IN hits
		COLLECT key = x.k INTO grp = x
		LET top = FIRST(FOR m IN grp SORT m.s DESC, m.f ASC RETURN m)
		RETURN top
)
FOR x IN best
	SORT x.s DESC, x.k ASC
	LIMIT @limit
	RETURN x
`
