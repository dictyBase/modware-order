package arangodb

// ArangoDB document field names, shared by filter maps, bind parameters and
// statement builders.
const (
	fieldCourier      = "courier"
	fieldConsumerInfo = "consumer_info"
	fieldCreatedAt    = "created_at"
	fieldItems        = "items"
	fieldPayerInfo    = "payer_info"
	fieldPayment      = "payment"
	fieldStatus       = "status"
	fieldUpdatedAt    = "updated_at"
)

// FMap maps filters to database fields.
var FMap = map[string]string{
	fieldCreatedAt: fieldCreatedAt,
	"item":         fieldItems,
	fieldCourier:   fieldCourier,
	fieldPayment:   fieldPayment,
	fieldStatus:    fieldStatus,
}
