package models

type Quote struct {
	ID     int    `db:"id" json:"id"`
	Quote  string `db:"quote" json:"quote"`
	Author string `db:"author" json:"author"`
}

type QuoteFilter struct {
	Author string `db:"author" json:"author"`
}
