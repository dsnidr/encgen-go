package example

//go:generate go run ../cmd/encgen/. --name Parcel
type Parcel struct {
	ID       string   `json:"id"`
	Name     string   `json:"name"`
	Items    []*Item  `json:"items" enc:"batch"`
	Metadata []string `json:"metadata" enc:"batch"`
	Tags     []*Tag   `json:"tags"`
}

type Item struct {
	SKU    string  `json:"sku"`
	Name   string  `json:"name"`
	Weight float64 `json:"weight"` // in kgs
}

type Tag struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
