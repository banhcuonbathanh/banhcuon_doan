package dish_grpc

//
//import (
//	"context"
//	"database/sql"
//
//	"fmt"
//
//	"time"
//
//	"english-ai-full/quanqr/proto_qr/dish"
//
//	"google.golang.org/protobuf/types/known/timestamppb"
//)
//
//type DishRepository struct {
//	db *sql.DB
//}
//
//func NewDishRepository(db *sql.DB) *DishRepository {
//	return &DishRepository{
//		db: db,
//	}
//}
//
//func (dr *DishRepository) GetDishList(ctx context.Context) ([]*dish.Dish, error) {
//	query := `
//		SELECT id, name, price, description, image, status, created_at, updated_at
//		FROM dishes
//	`
//	rows, err := dr.db.Query(ctx, query)
//	if err != nil {
//		return nil, fmt.Errorf("error fetching dishes: %w", err)
//	}
//	defer rows.Close()
//
//	var dishes []*dish.Dish
//	for rows.Next() {
//		var d dish.Dish
//		var createdAt, updatedAt time.Time
//		err := rows.Scan(
//			&d.Id,
//			&d.Name,
//			&d.Price,
//			&d.Description,
//			&d.Image,
//			&d.Status,
//			&createdAt,
//			&updatedAt,
//		)
//		if err != nil {
//			return nil, fmt.Errorf("error scanning dish: %w", err)
//		}
//		d.CreatedAt = timestamppb.New(createdAt)
//		d.UpdatedAt = timestamppb.New(updatedAt)
//		dishes = append(dishes, &d)
//	}
//	if err := rows.Err(); err != nil {
//		return nil, fmt.Errorf("error iterating over dishes: %w", err)
//	}
//
//	return dishes, nil
//}
//
//func (dr *DishRepository) GetDishDetail(ctx context.Context, id int64) (*dish.Dish, error) {
//	query := `
//		SELECT id, name, price, description, image, status, created_at, updated_at
//		FROM dishes
//		WHERE id = $1
//	`
//	var d dish.Dish
//	var createdAt, updatedAt time.Time
//	err := dr.db.QueryRow(ctx, query, id).Scan(
//		&d.Id,
//		&d.Name,
//		&d.Price,
//		&d.Description,
//		&d.Image,
//		&d.Status,
//		&createdAt,
//		&updatedAt,
//	)
//	if err != nil {
//		return nil, fmt.Errorf("error fetching dish detail: %w", err)
//	}
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//	return &d, nil
//}
//
//func (dr *DishRepository) CreateDish(ctx context.Context, req *dish.CreateDishRequest) (*dish.Dish, error) {
//	query := `
//		INSERT INTO dishes (name, price, description, image, status, created_at, updated_at)
//		VALUES ($1, $2, $3, $4, $5, $6, $6)
//		RETURNING id, created_at, updated_at
//	`
//	var d dish.Dish
//	var createdAt, updatedAt time.Time
//	err := dr.db.QueryRow(ctx, query,
//		req.Name,
//		req.Price,
//		req.Description,
//		req.Image,
//		*req.Status, // Dereference the pointer here
//		time.Now(),
//	).Scan(&d.Id, &createdAt, &updatedAt)
//	if err != nil {
//		return nil, fmt.Errorf("error creating dish: %w", err)
//	}
//
//	d.Name = req.Name
//	d.Price = req.Price
//	d.Description = req.Description
//	d.Image = req.Image
//	d.Status = *req.Status // Dereference the pointer here
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//
//	return &d, nil
//}
//
//func (dr *DishRepository) UpdateDish(ctx context.Context, req *dish.UpdateDishRequest) (*dish.Dish, error) {
//	query := `
//		UPDATE dishes
//		SET name = $2, price = $3, description = $4, image = $5, status = $6, updated_at = $7
//		WHERE id = $1
//		RETURNING created_at, updated_at
//	`
//	var createdAt, updatedAt time.Time
//	err := dr.db.QueryRow(ctx, query,
//		req.Id,
//		req.Name,
//		req.Price,
//		req.Description,
//		req.Image,
//		*req.Status,
//		time.Now(),
//	).Scan(&createdAt, &updatedAt)
//	if err != nil {
//		return nil, fmt.Errorf("error updating dish: %w", err)
//	}
//
//	return &dish.Dish{
//		Id:          req.Id,
//		Name:        req.Name,
//		Price:       req.Price,
//		Description: req.Description,
//		Image:       req.Image,
//		Status:      *req.Status,
//		CreatedAt:   timestamppb.New(createdAt),
//		UpdatedAt:   timestamppb.New(updatedAt),
//	}, nil
//}
//
//func (dr *DishRepository) DeleteDish(ctx context.Context, id int64) (*dish.Dish, error) {
//	query := `
//		DELETE FROM dishes
//		WHERE id = $1
//		RETURNING id, name, price, description, image, status, created_at, updated_at
//	`
//	var d dish.Dish
//	var createdAt, updatedAt time.Time
//	err := dr.db.QueryRow(ctx, query, id).Scan(
//		&d.Id,
//		&d.Name,
//		&d.Price,
//		&d.Description,
//		&d.Image,
//		&d.Status,
//		&createdAt,
//		&updatedAt,
//	)
//	if err != nil {
//		return nil, fmt.Errorf("error deleting dish: %w", err)
//	}
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//	return &d, nil
//}
