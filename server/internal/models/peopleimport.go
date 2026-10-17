package models

// PeopleImportResult counts what a people import did: people created, people changed, and
// people whose row matched what is stored. It is not a table.
type PeopleImportResult struct {
	Created   int
	Updated   int
	Unchanged int
}
