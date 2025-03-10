package order

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
)

type Repository interface {
	Close()
	PutOrder(ctx context.Context, o Order) error
	GetOrdersForAccount(ctx context.Context, accountID string) ([]Order, error)
}

type postgresRepository struct {
	db *sql.DB
}

func NewPostgresReporsitory(url string) (Repository, error) {
	db, err := sql.Open("postgres", url)
	if err != nil {
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		return nil, err
	}
	return &postgresRepository{db}, nil
}

func (r *postgresRepository) Close() {
	r.db.Close()
}
func (r *postgresRepository) PutOrder(ctx context.Context, o Order) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer func() {
		if err != nil {
			tx.Rollback()
			return
		}
		err = tx.Commit()
	}()
	tx.ExecContext(
		ctx,
		"INSERT INTO orders(id, created_at, account_id, total_price) VALUES ($1, $2, $3, $4)",
		o.ID,
		o.CreatedAt,
		o.AccountID,
		o.TotalPrice,
	)
	if err != nil {
		return err
	}
	stmt, _ := tx.PrepareContext(ctx, pq.CopyIn("order_products", "order_id", "product_id", "quantity"))
	for _, p := range o.Products {
		_, err = stmt.ExecContext(ctx, o.ID, p.ID, p.Quantity)
		if err != nil {
			return err
		}
	}
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return err
	}
	err = stmt.Close()
	return err
}

func (r *postgresRepository) GetOrdersForAccount(ctx context.Context, accountId string) ([]Order, error) {
	rows, err := r.db.QueryContext(
		ctx,
		`SELECT
		o.id,
		o.created_at,
		o.account_id,
		o.total_price::money::numeric::float8,
		op.product_id,
		op.quantity
		FROM orders o JOIN order_products op ON(o.id = op.order_id)
		WHERE o.account_id = $1
		ORDER BY o.id`,
		accountId,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	orderMap := make(map[string]*Order)
	// orders := []Order{}
	// lastOrder := &Order{}
	// orderedProduct := &OrderedProduct{}
	// products := []OrderedProduct{}
	// var order Order
	for rows.Next() {
		var (
			id, accountID string
			createdAt     sql.NullTime
			totalPrice    sql.NullFloat64
			productID     string
			quantity      int
		)
		if err = rows.Scan(&id, &createdAt, &accountID, &totalPrice, &productID, &quantity); err != nil {
			return nil, err
		}
		order, exits := orderMap[id]
		if !exits {
			order = &Order{
				ID:         id,
				AccountID:  accountID,
				CreatedAt:  createdAt.Time,
				TotalPrice: totalPrice.Float64,
				Products:   []OrderedProduct{},
			}
			orderMap[id] = order
		}
		order.Products = append(order.Products, OrderedProduct{
			ID:       productID,
			Quantity: uint32(quantity),
		})
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	var orders []Order
	for _, order := range orderMap {
		orders = append(orders, *order)
	}
	return orders, nil

}
