package repository

import (
	"context"
	"errors"
	"time"

	"cosmic-go/internal/domain"

	"github.com/jackc/pgx/v5"
)

var ErrProductNotFound = errors.New("product not found")

type PostgresRepository struct {
	tx pgx.Tx
}

func NewPostgresRepository(tx pgx.Tx) *PostgresRepository {
	return &PostgresRepository{tx: tx}
}

func (r *PostgresRepository) GetProduct(
	ctx context.Context,
	sku string,
) (*domain.Product, error) {
	batches, err := r.listBatches(ctx, sku)
	if err != nil {
		return nil, err
	}

	for _, b := range batches {
		err := r.addOrderLinesAndAllocationsToBatch(ctx, b)
		if err != nil {
			return nil, err
		}
	}

	row := r.getProduct(ctx, sku)
	return r.mapRowToProduct(row, batches)
}

func (r *PostgresRepository) getProduct(
	ctx context.Context,
	sku string,
) pgx.Row {
	query := `
	SELECT sku, version_id
	FROM products
	WHERE sku = $1;
	`

	row := r.tx.QueryRow(ctx, query, sku)
	return row
}

func (r *PostgresRepository) listBatches(
	ctx context.Context,
	sku string,
) ([]*domain.Batch, error) {
	query := `
	SELECT reference, product_sku, purchased_quantity, eta
	FROM batches
	WHERE product_sku = $1;
	`

	rows, err := r.tx.Query(
		ctx,
		query,
		sku,
	)
	if err != nil {
		return nil, err
	}

	return r.mapRowsToBatches(rows)
}

func (r *PostgresRepository) UpdateProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	err := r.updateProduct(ctx, p)
	if err != nil {
		return err
	}

	err = r.updateBatches(ctx, p.Batches)
	return err
}

func (r *PostgresRepository) updateProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	query := `
	UPDATE products
	SET version_id = $1
	WHERE sku = $2 AND
	version_id = (SELECT (version_id) FROM products WHERE sku = $2)
	`
	_, err := r.tx.Exec(
		ctx,
		query,
		p.VersionId,
		p.SKU,
	)
	return err
}

func (r *PostgresRepository) AddProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	err := r.insertProduct(ctx, p)
	if err != nil {
		return err
	}

	err = r.insertBatches(ctx, p.Batches)
	return err
}

func (r *PostgresRepository) insertProduct(
	ctx context.Context,
	p *domain.Product,
) error {
	query := `
	INSERT INTO products (sku, version_id)
	VALUES ($1, $2);
	`
	_, err := r.tx.Exec(
		ctx,
		query,
		p.SKU,
		p.VersionId,
	)

	return err
}

func (r *PostgresRepository) insertBatches(
	ctx context.Context,
	batches []*domain.Batch,
) error {
	for _, b := range batches {
		err := r.insertBatch(ctx, b)
		if err != nil {
			return err
		}

		err = r.insertOrderLineAndAllocationsFromBatch(
			ctx,
			b,
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) insertBatch(
	ctx context.Context,
	b *domain.Batch,
) error {
	query := `
	INSERT INTO batches (reference, product_sku, purchased_quantity, eta)
	VALUES ($1, $2, $3, $4);
	`
	_, err := r.tx.Exec(ctx,
		query,
		b.Reference,
		b.SKU,
		b.PurchasedQuantity,
		b.GetETA())

	return err
}

func (r *PostgresRepository) insertOrderLineAndAllocationsFromBatch(
	ctx context.Context,
	b *domain.Batch,
) error {
	for _, line := range b.Allocations {
		orderlineId, err := r.insertOrderLine(
			ctx,
			line,
		)
		if err != nil {
			return err
		}

		err = r.insertAllocation(
			ctx,
			b.Reference,
			orderlineId,
		)
		if err != nil {
			return err
		}
	}
	return nil
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
	rows, err := r.tx.Query(ctx, query, b.Reference)
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

func (r *PostgresRepository) updateBatches(
	ctx context.Context,
	batches []*domain.Batch,
) error {
	for _, b := range batches {
		err := r.updateBatch(ctx, b)
		if err != nil {
			return err
		}

		existingBatch, err := r.getBatchWithAllocations(ctx, b.Reference)
		if err != nil {
			return err
		}

		err = r.updateOrderLinesAndAllocationsFromBatch(ctx, existingBatch, b)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PostgresRepository) getBatchWithAllocations(ctx context.Context, batchRef string) (*domain.Batch, error) {
	existingBatch, err := r.getBatch(ctx, batchRef)
	if err != nil {
		return nil, err
	}

	err = r.addOrderLinesAndAllocationsToBatch(ctx, existingBatch)
	if err != nil {
		return nil, err
	}

	return existingBatch, nil
}

func (r *PostgresRepository) updateBatch(
	ctx context.Context,
	b *domain.Batch,
) error {
	query := `
	UPDATE batches
	SET purchased_quantity = $1, eta = $2
	WHERE reference = $3;
	`
	_, err := r.tx.Exec(ctx, query, b.PurchasedQuantity, b.GetETA(), b.Reference)
	return err
}

func (r *PostgresRepository) getBatch(
	ctx context.Context,
	ref string,
) (*domain.Batch, error) {
	query := `
	SELECT reference, product_sku, purchased_quantity, eta
	FROM batches
	WHERE reference = $1;
	`

	row := r.tx.QueryRow(ctx, query, ref)
	return r.mapRowToBatch(row)
}

func (r *PostgresRepository) updateOrderLinesAndAllocationsFromBatch(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
) error {
	for _, al := range updatedB.Allocations {
		orderlineID, err := r.insertOrderLine(ctx, al)
		if err != nil {
			return err
		}

		err = r.insertAllocation(ctx, existingB.Reference, orderlineID)
		if err != nil {
			return err
		}
	}

	err := r.deleteDeallocatedOrderLines(ctx, existingB, updatedB)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) insertOrderLine(
	ctx context.Context,
	al domain.OrderLine,
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
	err = r.tx.QueryRow(
		ctx,
		query,
		al.SKU,
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
) error {
	query := `
	WITH existing AS (
		SELECT id FROM allocations WHERE orderline_id = $1 AND batch_reference = $2
	)

	INSERT INTO allocations (orderline_id, batch_reference)
	SELECT $1, $2
	WHERE NOT EXISTS (SELECT 1 FROM existing);
	`
	_, err := r.tx.Exec(ctx, query, orderlineID, batchRef)
	if err != nil {
		return err
	}

	return nil
}

func (r *PostgresRepository) deleteDeallocatedOrderLines(
	ctx context.Context,
	existingB *domain.Batch,
	updatedB *domain.Batch,
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
			_, err := r.tx.Exec(
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
	return !ok
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
			sku,
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
	if err != nil {
		return nil, err
	}

	batch := domain.NewBatch(
		reference,
		sku,
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
		var sku string
		var orderid domain.OrderID

		err := rows.Scan(&quantity, &sku, &orderid)
		if err != nil {
			return nil, err
		}

		orderLines = append(orderLines, &domain.OrderLine{
			SKU:      sku,
			Quantity: quantity,
			OrderId:  orderid,
		})
	}

	return orderLines, nil
}

func (r *PostgresRepository) mapRowToProduct(
	row pgx.Row,
	batches []*domain.Batch,
) (*domain.Product, error) {
	var sku string
	var VersionId int

	err := row.Scan(
		&sku,
		&VersionId,
	)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrProductNotFound
	} else if err != nil {
		return nil, err
	}

	p := domain.NewProduct(
		sku,
		batches,
		VersionId,
	)

	return p, nil
}
