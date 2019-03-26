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
