package repository

import (
	"context"
	"cosmic-go/internal/domain"
	"log"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) AddBatch(ctx context.Context, b *domain.Batch) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("transaction rollback failed: %v", rbErr)
			}
		}
	}()

	err = r.insertProduct(ctx, b.Product.SKU, tx)
	if err != nil {
		return err
	}

	err = r.insertBatch(ctx, b, tx)
	if err != nil {
		return err
	}

	err = r.insertOrderLineAndAllocationsFromBatch(
		ctx,
		b,
		tx,
	)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) insertProduct(
	ctx context.Context,
	sku string,
	tx pgx.Tx,
) error {
	query := `
	INSERT INTO products (sku) 
	VALUES ($1);
	`
	_, err := tx.Exec(
		ctx,
		query,
		sku,
	)

	return err
}

func (r *PostgresRepository) insertBatch(
	ctx context.Context,
	b *domain.Batch,
	tx pgx.Tx,
) error {

	query := `
	INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
	VALUES ($1, $2, $3, $4);
	`
	_, err := tx.Exec(ctx,
		query,
		b.Reference,
		string(b.Product.SKU),
		b.PurchasedQuantity,
		b.ETA)

	return err
}

func (r *PostgresRepository) insertOrderLineAndAllocationsFromBatch(
	ctx context.Context,
	b *domain.Batch,
	tx pgx.Tx,
) error {
	for _, line := range b.Allocations {
		orderlineId, err := r.insertOrderLine(ctx, line.Product.SKU, line.Quantity, b.Reference, tx)
		if err != nil {
			return err
		}

		err = r.insertAllocation(ctx, b.Reference, orderlineId, tx)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) insertOrderLine(
	ctx context.Context,
	sku string,
	quantity int,
	orderID string,
	tx pgx.Tx,
) (int, error) {

	query := `
	INSERT INTO order_lines (product_sku, quantity, orderid)
	VALUES ($1, $2, $3)
	RETURNING id;
	`
	var orderlineId int
	err := tx.QueryRow(
		ctx,
		query,
		sku,
		quantity,
		orderID,
	).Scan(&orderlineId)

	return orderlineId, err
}

func (r *PostgresRepository) insertAllocation(
	ctx context.Context,
	ref string,
	orderlineId int,
	tx pgx.Tx,
) error {
	query := `
	INSERT INTO allocations (batch_reference, orderline_id)
	VALUES ($1, $2);
	`
	_, err := tx.Exec(ctx,
		query,
		ref,
		orderlineId)

	return err
}

func (r *PostgresRepository) GetBatchByReference(
	ctx context.Context,
	ref string,
) (*domain.Batch, error) {

	batch, err := r.getBatchByReference(ctx, ref)
	if err != nil {
		return nil, err
	}

	err = r.addOrderLinesAndAllocationsToBatch(ctx, batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (r *PostgresRepository) getBatchByReference(
	ctx context.Context,
	ref string,
) (*domain.Batch, error) {
	query := `
	SELECT b.reference, b.product_sku, b.purchased_quantity, b.eta
	FROM batches b
	WHERE b.reference = $1;
	`

	batch := domain.Batch{}
	err := r.db.QueryRow(
		ctx,
		query,
		ref,
	).Scan(
		&batch.Reference,
		&batch.Product.SKU,
		&batch.PurchasedQuantity,
		&batch.ETA,
	)

	return &batch, err
}

func (r *PostgresRepository) addOrderLinesAndAllocationsToBatch(
	ctx context.Context,
	b *domain.Batch,
) error {

	query := `
	SELECT o.quantity, o.product_sku, o.orderid
	FROM order_lines o
	JOIN allocations a ON a.orderline_id = o.id
	WHERE a.batch_reference = $1;
	`
	lines, err := r.db.Query(ctx, query, b.Reference)
	if err != nil {
		return err
	}
	defer lines.Close()

	for lines.Next() {
		var quantity int
		var productSku string
		var orderid string

		err = lines.Scan(&quantity, &productSku, &orderid)
		if err != nil {
			return err
		}

		err := b.Allocate(&domain.OrderLine{
			Product:  domain.Product{SKU: productSku},
			Quantity: quantity,
			OrderId:  orderid,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) ListBatches(
	ctx context.Context,
) ([]*domain.Batch, error) {
	query := `
	SELECT reference, product_sku, purchased_quantity, eta
	FROM batches;`

	rows, err := r.db.Query(context.Background(), query)
	if err != nil {
		return nil, err
	}

	var batches []*domain.Batch

	for rows.Next() {
		var batch domain.Batch

		err := rows.Scan(
			&batch.Reference,
			&batch.Product.SKU,
			&batch.PurchasedQuantity,
			&batch.ETA)

		batches = append(batches, &batch)

		if err != nil {
			return nil, err
		}
	}

	return batches, nil
}

func (r *PostgresRepository) GetBatchBySku(
	ctx context.Context,
	sku string,
) (*domain.Batch, error) {

	batch, err := r.getBatchBySku(ctx, sku)
	if err != nil {
		return nil, err
	}

	err = r.addOrderLinesAndAllocationsToBatch(ctx, batch)
	if err != nil {
		return nil, err
	}

	return batch, nil
}

func (r *PostgresRepository) getBatchBySku(
	ctx context.Context,
	sku string,
) (*domain.Batch, error) {

	query := `
	SELECT b.reference, b.product_sku, b.purchased_quantity, b.eta
	FROM batches b
	WHERE b.product_sku = $1;
	`
	batch := &domain.Batch{}
	err := r.db.QueryRow(
		ctx,
		query,
		sku,
	).Scan(
		&batch.Reference,
		&batch.Product.SKU,
		&batch.PurchasedQuantity,
		&batch.ETA,
	)
	if err != nil {
		return nil, err
	}

	return batch, err
}

func (r *PostgresRepository) UpdateBatch(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return err
	}

	defer func() {
		if err != nil {
			if rbErr := tx.Rollback(ctx); rbErr != nil {
				log.Printf("transaction rollback failed: %v", rbErr)
			}
		}
	}()

	err = r.updateBatch(ctx, updatedB, tx)
	if err != nil {
		return err
	}

	err = r.updateOrderLinesAndAllocationsFromBatch(ctx, existingB, updatedB, tx)
	if err != nil {
		return err
	}

	err = tx.Commit(ctx)
	if err != nil {
		return err
	}
	return nil
}

func (r *PostgresRepository) updateBatch(
	ctx context.Context,
	b *domain.Batch,
	tx pgx.Tx,
) error {
	query := `
	UPDATE batches
	SET purchased_quantity = $1, eta = $2
	WHERE reference = $3;
	`
	_, err := tx.Exec(ctx, query, b.PurchasedQuantity, b.ETA, b.Reference)
	return err
}

func (r *PostgresRepository) updateOrderLinesAndAllocationsFromBatch(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
	tx pgx.Tx,
) error {
	for _, al := range updatedB.Allocations {
		orderlineID, err := r.insertOrUpdateOrderLine(ctx, al, tx)
		if err != nil {
			return err
		}

		err = r.insertOrUpdateAllocations(ctx, existingB.Reference, orderlineID, tx)
		if err != nil {
			return err
		}
	}

	err := r.deleteDeallocatedOrderLines(ctx, existingB, updatedB, tx)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) insertOrUpdateOrderLine(
	ctx context.Context,
	al domain.OrderLine,
	tx pgx.Tx,
) (id int, err error) {
	query := `
	INSERT INTO order_lines (product_sku, quantity, orderid)
	VALUES ($1, $2, $3)
	ON CONFLICT (orderid) DO UPDATE
	SET product_sku = EXCLUDED.product_sku,
		quantity = EXCLUDED.quantity,
		orderid = EXCLUDED.orderid
	RETURNING id;
	`
	var orderlineID int
	err = tx.QueryRow(
		ctx,
		query,
		al.Product.SKU,
		al.Quantity,
		al.OrderId).
		Scan(&orderlineID)

	if err != nil {
		return 0, err
	}

	return orderlineID, nil
}

func (r *PostgresRepository) insertOrUpdateAllocations(
	ctx context.Context,
	batchRef string,
	orderlineID int,
	tx pgx.Tx,
) error {
	query := `
	INSERT INTO allocations (orderline_id, batch_reference)
	VALUES ($1, $2)
	ON CONFLICT (orderline_id) DO UPDATE
	SET batch_reference = EXCLUDED.batch_reference;
	`
	_, err := tx.Exec(ctx, query, orderlineID, batchRef)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) deleteDeallocatedOrderLines(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
	tx pgx.Tx,
) error {
	for _, allocatedOrderLine := range existingB.Allocations {
		if r.orderLineDeallocated(allocatedOrderLine, updatedB) {
			query := `
			DELETE FROM allocations
			WHERE orderline_id = $1
			AND batch_reference = $2;
			`
			_, err := tx.Exec(
				ctx,
				query,
				allocatedOrderLine.OrderId,
				updatedB.Reference,
			)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (r *PostgresRepository) orderLineDeallocated(
	al domain.OrderLine,
	updatedB *domain.Batch,
) bool {
	_, ok := updatedB.Allocations[al.Product.SKU]
	return ok
}
