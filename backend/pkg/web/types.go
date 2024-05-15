package web

import (
	categories "be/pkg/categories"
	pagefts "be/pkg/page/fts"
	page "be/pkg/page/page"
)

type Url struct {
	Url string `json:"url"`
}

type Urls struct {
	Urls []string `json:"urls"`
}

type UrlWithHTML struct {
	Url            string `json:"url"`
	HTML           string `json:"html"`
	Title          string `json:"title"`
	ExtentionToken string `json:"extention_token"`
	UserId         string `json:"user_id"`
}

type PageResponse struct {
	Page       page.Page             `json:"page"`
	Categories []categories.Category `json:"categories"`
	FTSDoc     pagefts.FTSPageDoc    `json:"ftsdoc"`
}

type DeleteCategoryRequest struct {
	CategoryName string `json:"category_name"`
	PageID       string `json:"page_id"`
}

type DeletePageRequest struct {
	PageID string `json:"page_id"`
}

type PostPagemanageCategoryRequest struct {
	CategoryName string `json:"category_name"`
	PageID       string `json:"page_id"`
}

// set of infos usefull before search
type PreSearchInfoResponse struct {
	Categories []categories.Category `json:"categories"` // categories can be filtered
}

type SearchResponse struct {
	Pages []PageResponse `json:"pages"`
}
