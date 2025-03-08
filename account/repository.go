package account

import "database/sql"

type Repository interface {
	Close()
	PutAccount
	GetAccountByID
	ListAccounts
}

type postgresRepository struct {
	db *sql.DB
}
