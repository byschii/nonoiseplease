package page

var FTSDOCATTRIBUTES = []string{"id", "category", "content"}

// currently Full text Search is provided by MeiliSearch
type FTSPageDoc struct {
	// ID is the primary key of the document.
	// It is used to identify the document and to perform search queries.
	// It is also used to update or delete a document.
	// It is also used to retrieve a document.
	ID       string   `json:"id"`
	Category []string `json:"category"`
	Content  string   `json:"content"`
}
