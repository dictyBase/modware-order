package arangodb

// FMap maps filters to database fields.
var FMap = map[string]string{
	"created_at": "created_at",
	"item":       "items",
	"courier":    "courier",
	"payment":    "payment",
	"status":     "status",
}
