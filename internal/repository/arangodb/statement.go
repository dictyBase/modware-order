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
)
