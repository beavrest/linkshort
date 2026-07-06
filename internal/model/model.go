package model

type ShortenJSONRequest struct {
	URL string `json:"url"`
}

type ShortenJSONResponse struct {
	Result string `json:"result"`
}
