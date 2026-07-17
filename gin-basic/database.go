package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/go-sql-driver/mysql"
)

var ErrProductUnavailable = errors.New(
	"商品不存在、未上架或库存不足",
)

// openDatabase 创建并验证 MySQL 数据库连接池。
// 参数：ctx 用于控制连接验证的超时与取消。
// 返回值：连接成功时返回数据库连接池；创建或验证失败时返回错误。
func openDatabase(ctx context.Context) (*sql.DB, error) {
	config := mysql.Config{
		User:                 "root",
		Passwd:               os.Getenv("MYSQL_PASSWORD"),
		Net:                  "tcp",
		Addr:                 "127.0.0.1:3306",
		DBName:               "go_learning",
		ParseTime:            true,
		AllowNativePasswords: true,
		ClientFoundRows:      true,
	}
	db, err := sql.Open("mysql", config.FormatDSN())
	if err != nil {
		return nil, fmt.Errorf("创建数据库连接池失败: %w", err)
	}

	// 最多同时打开 20 条连接，避免请求过多时耗尽 MySQL 连接资源。
	db.SetMaxOpenConns(20)
	// 最多保留 5 条空闲连接，供后续请求直接复用。
	db.SetMaxIdleConns(5)
	// 每条连接最多复用 30 分钟，超过时间后由连接池逐步替换。
	db.SetConnMaxLifetime(30 * time.Minute)
	// 空闲连接连续 5 分钟未使用时允许连接池将其关闭。
	db.SetConnMaxIdleTime(5 * time.Minute)

	pingCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
	defer cancel()

	if err := db.PingContext(pingCtx); err != nil {
		db.Close()
		return nil, fmt.Errorf("连接MySQL失败: %w", err)
	}

	return db, nil
}

func queryProducts(ctx context.Context, db *sql.DB, filter ListProductsQuery) ([]Product, error) {
	query := `
		SELECT id, name, price_cents, stock, status
		FROM products
		where 1 = 1
	`
	args := []any{}
	if filter.Status != "" {
		query += ` and status = ?`
		args = append(args, filter.Status)
	}
	query += ` order by id asc limit ? offset ?`
	args = append(args, filter.PageSize, (filter.Page-1)*filter.PageSize)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("查询产品失败: %w", err)
	}
	defer rows.Close()

	products := make([]Product, 0)
	for rows.Next() {
		var product Product
		err := rows.Scan(&product.ID, &product.Name, &product.PriceCents, &product.Stock, &product.Status)
		if err != nil {
			return nil, fmt.Errorf("扫描产品失败: %w", err)
		}
		products = append(products, product)
	}

	err = rows.Err()
	if err != nil {
		return nil, fmt.Errorf("遍历产品失败: %w", err)
	}

	return products, nil
}

func queryProductByID(
	ctx context.Context,
	db *sql.DB,
	id int64,
) (Product, error) {
	query := `
		SELECT id, name, price_cents, stock, status
		FROM products
		where id = ?
	`

	row := db.QueryRowContext(ctx, query, id)

	var product Product
	err := row.Scan(&product.ID, &product.Name, &product.PriceCents, &product.Stock, &product.Status)
	if err != nil {
		return product, fmt.Errorf("扫描产品失败: %w", err)
	}

	return product, nil
}

func insertProduct(
	ctx context.Context,
	db *sql.DB,
	input CreateProductRequest,
) (Product, error) {
	query := `
		INSERT INTO products (name, price_cents, stock, status)
		VALUES (?, ?, ?, ?)
	`

	res, err := db.ExecContext(ctx, query, input.Name, input.PriceCents, input.Stock, input.Status)
	if err != nil {
		return Product{}, fmt.Errorf("插入产品失败: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return Product{}, fmt.Errorf("获取插入产品ID失败: %w", err)
	}

	product := Product{
		ID:         id,
		Name:       input.Name,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
	}

	return product, nil
}

func updatedProduct(
	ctx context.Context,
	db *sql.DB,
	id int64,
	input UpdateProductRequest,
) (Product, error) {
	query := `
		update products set name = ?, price_cents = ?, stock = ?, status = ? where id = ?
	`

	result, err := db.ExecContext(ctx, query, input.Name, input.PriceCents, input.Stock, input.Status, id)
	if err != nil {
		return Product{}, fmt.Errorf("更新产品失败: %w", err)
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return Product{}, fmt.Errorf("获取更新行数失败: %w", err)
	}

	if affected == 0 {
		return Product{}, sql.ErrNoRows
	}

	product := Product{
		ID:         id,
		Name:       input.Name,
		PriceCents: input.PriceCents,
		Stock:      input.Stock,
		Status:     input.Status,
	}

	return product, nil

}

func delProductByID(
	ctx context.Context,
	db *sql.DB,
	id int64,
) error {
	query := `
		delete from products where id = ?
	`

	result, err := db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("删除产品失败: %w", err)
	}

	affected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取删除行数失败: %w", err)
	}

	if affected == 0 {
		return sql.ErrNoRows
	}

	return nil
}

func sellProduct(
	ctx context.Context,
	db *sql.DB,
	id int64,
	quantity int,
) error {
	if quantity <= 0 {
		return errors.New("销量必须大于0")
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("开始事务失败: %w", err)
	}
	defer tx.Rollback()

	decreaseStockQuery := `
			update products set stock = stock - ? where id = ? and status = 'active' and stock >= ?
		`

	res, err := tx.ExecContext(ctx, decreaseStockQuery, quantity, id, quantity)
	if err != nil {
		return fmt.Errorf("更新库存失败: %w", err)
	}
	affected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("获取更新行数失败: %w", err)
	}
	if affected == 0 {
		return ErrProductUnavailable
	}

	const updateStatusQuery = `
		UPDATE products
		SET status = 'sold_out'
		WHERE id = ?
		  AND stock = 0
	`
	res, err = tx.ExecContext(ctx, updateStatusQuery, id)
	if err != nil {
		return fmt.Errorf("更新状态失败: %w", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("提交事务失败: %w", err)
	}

	return nil
}

func countProducts(
	ctx context.Context,
	db *sql.DB,
	filter ListProductsQuery,
) (int64, error) {
	query := `
		SELECT COUNT(*)
		FROM products
		WHERE 1 = 1
	`

	args := make([]any, 0)

	if filter.Status != "" {
		query += ` AND status = ?`
		args = append(args, filter.Status)
	}
	var total int64
	err := db.QueryRowContext(ctx, query, args...).Scan(&total)
	if err != nil {
		return 0, fmt.Errorf("查询产品总数失败: %w", err)
	}
	return total, nil
}
