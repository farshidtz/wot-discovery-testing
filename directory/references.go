package directory

const (
	MediaTypeJSON             = "application/json"
	MediaTypeJSONLD           = "application/ld+json"
	MediaTypeThingDescription = "application/td+json"
	MediaTypeMergePatch       = "application/merge-patch+json"
)

func mockedTD(id string) mapAny {
	var td = mapAny{
		"@context": "https://www.w3.org/2019/wot/td/v1",
		"title":    "example thing",
		"security": []string{"nosec_sc"},
		"securityDefinitions": mapAny{
			"nosec_sc": map[string]string{
				"scheme": "nosec",
			},
		},
	}
	if id != "" {
		td["id"] = id
	}
	return td
}

// Extractor:
// https://play.golang.org/p/2Kj-RxZ5nek

var tddAssertions = []string{
	"tdd-registrationinfo-expiry-purge",
	"tdd-registrationinfo-expiry-config",
	"tdd-reg-anonymous-td-identifier",
	"tdd-reg-anonymous-td-local-id",
	"tdd-https",
	"tdd-http-error-response",
	"tdd-reg-default-representation",
	"tdd-reg-additional-representation",
	"tdd-reg-crudl",
	"tdd-reg-create-body",
	"tdd-reg-create-contenttype",
	"tdd-reg-create-known-vs-anonymous",
	"tdd-reg-create-known-td",
	"tdd-reg-create-known-td-resp",
	"tdd-reg-create-anonymous-td",
	"tdd-reg-create-anonymous-td-resp",
	"tdd-reg-retrieve",
	"tdd-reg-retrieve-resp",
	"tdd-reg-update-types",
	"tdd-reg-update",
	"tdd-reg-update-contenttype",
	"tdd-reg-update-resp",
	"tdd-reg-update-partial",
	"tdd-reg-update-partial-mergepatch",
	"tdd-reg-update-partial-contenttype",
	"tdd-reg-update-partial-partialtd",
	"tdd-reg-update-partial-resp",
	"tdd-reg-delete",
	"tdd-reg-delete-resp",
	"tdd-reg-list-method",
	"tdd-reg-list-resp",
	"tdd-reg-list-http11-chunks",
	"tdd-reg-list-http2-frames",
	"tdd-reg-list-pagination",
	"tdd-reg-list-pagination-limit",
	"tdd-reg-list-pagination-header-nextlink",
	"tdd-reg-list-pagination-header-nextlink-attr",
	"tdd-reg-list-pagination-header-canonicallink",
	"tdd-reg-list-pagination-order-default",
	"tdd-reg-list-pagination-order",
	"tdd-reg-list-pagination-order-unsupported",
	"tdd-reg-list-pagination-order-nextlink",
	"tdd-validation-syntactic",
	"tdd-validation-jsonschema",
	"tdd-validation-result",
	"tdd-validation-response",
	"tdd-notification-sse",
	"tdd-notification-event-id",
	"tdd-notification-event-types",
	"tdd-notification-filter-type",
	"tdd-notification-data",
	"tdd-notification-data-tdid",
	"tdd-notification-data-create-full",
	"tdd-notification-data-update-diff",
	"tdd-notification-data-update-id",
	"tdd-notification-data-delete-diff",
	"tdd-notification-data-diff-unsupported",
	"tdd-search-jsonpath",
	"tdd-search-xpath",
	"tdd-search-sparql",
	"tdd-search-jsonpath-method",
	"tdd-search-jsonpath-parameter",
	"tdd-search-jsonpath-response",
	"tdd-search-xpath-method",
	"tdd-search-xpath-parameter",
	"tdd-search-xpath-response",
	"tdd-search-sparql-version",
	"tdd-search-sparql-method-get",
	"tdd-search-sparql-method-post",
	"tdd-search-sparql-resp",
	"tdd-search-sparql-federation",
	"tdd-search-sparql-federation-imp",
}
