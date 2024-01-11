package main

type Market struct {
	Albums []Cover `json:"cover"`
}

type Album struct {
	Cover
	Songs []Song `json:"songs"`
}

type Cover struct {
	Id       int    `json:"id"`
	Name     string `json:"name"`
	ImageUrl string `json:"image_url"`
}

type Song struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}
