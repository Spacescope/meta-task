package models

// DBUnique return unique key in database
type DBUnique interface {
	Key() []string
}
