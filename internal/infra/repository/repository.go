package repository

import (
	"context"
	"cosmic-go/internal/domain"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	initializeSchema(db)
	return &PostgresRepository{db: db}
}

func initializeSchema(db *pgxpool.Pool) {
	schema := `
	CREATE TABLE IF NOT EXISTS products (
		sku TEXT PRIMARY KEY
		);

	CREATE TABLE IF NOT EXISTS batches (
		reference TEXT PRIMARY KEY NOT NULL, 
		product_sku TEXT REFERENCES products(sku),
		purchased_quantity INT NOT NULL,
		eta DATE
		);

	CREATE TABLE IF NOT EXISTS order_lines (
		id SERIAL PRIMARY KEY NOT NULL, 
		product_sku TEXT REFERENCES products(sku),
		quantity INT NOT NULL,
		orderid TEXT
		);

	CREATE TABLE IF NOT EXISTS allocations (
		id SERIAL PRIMARY KEY NOT NULL, 
		orderline_id INT REFERENCES order_lines(id),
		batch_reference TEXT REFERENCES batches(reference)
		);
	`

	_, err := db.Exec(context.Background(), schema)
	if err != nil {
		log.Fatal("failed to initiliaze schema with error:", err)
	}
}

func (r *PostgresRepository) Add(b *domain.Batch) error {
	tx, err := r.db.Begin(context.Background())
	if err != nil {
		return err
	}

	query := `
	INSERT INTO products (sku) 
	VALUES ($1);
	`
	_, err = tx.Exec(context.Background(),
		query,
		string(b.Product.SKU))

	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	query = `
	INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
	VALUES ($1, $2, $3, $4);
	`

	_, err = tx.Exec(context.Background(),
		query,
		b.Reference,
		string(b.Product.SKU),
		b.PurchasedQuantity,
		b.ETA)

	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	var orderlineId int
	query = `
	INSERT INTO order_lines (product_sku, quantity)
	VALUES ($1, $2) RETURNING id;
	`
	err = tx.QueryRow(context.Background(),
		query,
		string(b.Product.SKU),
		b.PurchasedQuantity).Scan(&orderlineId)

	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	query = `
	INSERT INTO allocations (orderline_id, batch_reference)
	VALUES ($1, $2);
	`
	_, err = tx.Exec(context.Background(),
		query,
		orderlineId,
		b.Reference)

	if err != nil {
		tx.Rollback(context.Background())
		return err
	}

	tx.Commit(context.Background())
	return nil
}

func (r *PostgresRepository) Get(batchReference string) (*domain.Batch, error) {
	query := `
	SELECT b.reference, b.product_sku, b.purchased_quantity, b.eta
	FROM batches b
	WHERE b.reference = $1;
	`

	var batch domain.Batch
	err := r.db.QueryRow(context.Background(), query, batchReference).
		Scan(&batch.Reference, &batch.Product.SKU, &batch.PurchasedQuantity, &batch.ETA)

	if err != nil {
		return nil, err
	}

	query = `
	SELECT o.quantity, o.product_sku
	FROM order_lines o
	JOIN allocations a ON a.orderline_id = o.id
	WHERE a.batch_reference = $1;
	`
	lines, err := r.db.Query(context.Background(), query, batchReference)
	if err != nil {
		return nil, err
	}
	defer lines.Close()

	for lines.Next() {
		var quantity int
		var productSku string

		err = lines.Scan(&quantity, &productSku)
		if err != nil {
			return nil, err
		}

		batch.Allocate(&domain.OrderLine{
			Product:  domain.Product{SKU: domain.Reference(productSku)},
			Quantity: quantity})
	}

	return &batch, nil
}
