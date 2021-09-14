package object

type PayItem struct {
	Invoice     string `json:"invoice"`
	Price       string `json:"price"`
	Description string `json:"description"`
	Currency    string `json:"currency"`
}
