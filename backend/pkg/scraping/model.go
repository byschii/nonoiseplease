package scraping

type html struct {
	Body body `xml:"body"`
}
type body struct {
	Content string `xml:",innerxml"`
}

type ParsedPage struct {
	Title       string `json:"title"`
	TextContent string `json:"text_content"`
}
