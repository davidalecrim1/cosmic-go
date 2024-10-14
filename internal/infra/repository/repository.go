package repository

import (
	"context"
	"errors"
	"time"

	"cosmic-go/internal/application"
	"cosmic-go/internal/domain"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresRepository struct {
	db *pgxpool.Pool
}

func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

func (r *PostgresRepository) runWithTransaction(
	ctx context.Context,
	db *pgxpool.Pool,
	fn func(tx pgx.Tx) error,
) error {
	tx, err := db.Begin(ctx)
	if err != nil {
		return err
	}

	err = fn(tx)
	if err == nil {
		return tx.Commit(ctx)
	}

	rollbackErr := tx.Rollback(ctx)
	if rollbackErr != nil {
		return errors.Join(err, rollbackErr)
	}

	return err
}

func (r *PostgresRepository) AddBatch(
	ctx context.Context,
	b *domain.Batch,
) error {
	return r.runWithTransaction(ctx, r.db, func(tx pgx.Tx) error {
		err := r.insertProduct(ctx, b.Product.SKU, tx)
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

		return nil
	},
	)
}

func (r *PostgresRepository) insertProduct(
	ctx context.Context,
	sku string,
	tx pgx.Tx,
) error {
	query := `
	INSERT INTO products (sku) 
	VALUES ($1)
	ON CONFLICT (sku) DO NOTHING;
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
		b.GetETA())

	return err
}

func (r *PostgresRepository) insertOrderLineAndAllocationsFromBatch(
	ctx context.Context,
	b *domain.Batch,
	tx pgx.Tx,
) error {
	for _, line := range b.Allocations {
		orderlineId, err := r.insertOrderLine(
			ctx,
			line,
			tx,
		)
		if err != nil {
			return err
		}

		err = r.insertAllocation(
			ctx,
			b.Reference,
			orderlineId,
			tx,
		)
		if err != nil {
			return err
		}
	}
	return nil
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

	row := r.db.QueryRow(ctx, query, ref)
	return r.mapRowToBatch(row)
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
	rows, err := r.db.Query(ctx, query, b.Reference)
	if err != nil {
		return err
	}
	defer rows.Close()

	orderLines, err := r.mapRowsToOrderLines(rows)
	if err != nil {
		return err
	}

	for _, ol := range orderLines {
		err := b.Allocate(ol)
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

	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, err
	}

	return r.mapRowsToBatches(rows)
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

	row := r.db.QueryRow(
		ctx,
		query,
		sku,
	)

	return r.mapRowToBatch(row)
}

func (r *PostgresRepository) UpdateBatch(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
) error {
	return r.runWithTransaction(ctx, r.db, func(tx pgx.Tx) error {
		err := r.updateBatch(ctx, updatedB, tx)
		if err != nil {
			return err
		}

		err = r.updateOrderLinesAndAllocationsFromBatch(ctx, existingB, updatedB, tx)
		if err != nil {
			return err
		}

		return nil
	},
	)
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
	_, err := tx.Exec(ctx, query, b.PurchasedQuantity, b.GetETA(), b.Reference)
	return err
}

func (r *PostgresRepository) updateOrderLinesAndAllocationsFromBatch(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
	tx pgx.Tx,
) error {
	for _, al := range updatedB.Allocations {
		orderlineID, err := r.insertOrderLine(ctx, al, tx)
		if err != nil {
			return err
		}

		err = r.insertAllocation(ctx, existingB.Reference, orderlineID, tx)
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

func (r *PostgresRepository) insertOrderLine(
	ctx context.Context,
	al domain.OrderLine,
	tx pgx.Tx,
) (id int, err error) {
	query := `
	WITH existing AS (
		SELECT id FROM order_lines WHERE product_sku = $1 AND orderid = $3
	)
	INSERT INTO order_lines (product_sku, quantity, orderid)
	SELECT $1, $2, $3
	WHERE NOT EXISTS (SELECT 1 FROM existing)
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

func (r *PostgresRepository) insertAllocation(
	ctx context.Context,
	batchRef string,
	orderlineID int,
	tx pgx.Tx,
) error {
	query := `
	WITH existing AS (
		SELECT id FROM allocations WHERE orderline_id = $1 AND batch_reference = $2
	)

	INSERT INTO allocations (orderline_id, batch_reference)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM existing);
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
			WHERE orderline_id = (
				SELECT id FROM order_lines WHERE orderid = $1
			)
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
	_, ok := updatedB.Allocations[al.OrderId]
	return ok
}

func (r *PostgresRepository) mapRowsToBatches(
	rows pgx.Rows,
) ([]*domain.Batch, error) {
	var batches []*domain.Batch
	defer rows.Close()

	for rows.Next() {
		var reference string
		var sku string
		var quantity int
		var eta *time.Time

		if err := rows.Scan(
			&reference,
			&sku,
			&quantity,
			&eta,
		); err != nil {
			return nil, err
		}

		batch := domain.NewBatch(
			reference,
			domain.Product{SKU: sku},
			quantity,
			eta,
		)

		batches = append(batches, batch)
	}

	return batches, nil
}

func (r *PostgresRepository) mapRowToBatch(
	row pgx.Row,
) (*domain.Batch, error) {
	var reference string
	var sku string
	var quantity int
	var eta *time.Time

	err := row.Scan(
		&reference,
		&sku,
		&quantity,
		&eta,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, application.ErrBatchNotFound
	} else if err != nil {
		return nil, err
	}

	batch := domain.NewBatch(
		reference,
		domain.Product{SKU: sku},
		quantity,
		eta,
	)

	return batch, nil
}

func (r *PostgresRepository) mapRowsToOrderLines(
	rows pgx.Rows,
) ([]*domain.OrderLine, error) {
	var orderLines []*domain.OrderLine
	defer rows.Close()

	for rows.Next() {
		var quantity int
		var productSku string
		var orderid domain.OrderID

		err := rows.Scan(&quantity, &productSku, &orderid)
		if err != nil {
			return nil, err
		}

		orderLines = append(orderLines, &domain.OrderLine{
			Product:  domain.Product{SKU: productSku},
			Quantity: quantity,
			OrderId:  orderid,
		})
	}

	return orderLines, nil
}
