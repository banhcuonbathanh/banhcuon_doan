package set_qr

import (
	"context"
	"fmt"

	"time"

	"english-ai-full/internal/proto_qr/set"
	"english-ai-full/logger"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type SetRepository struct {
	db     *pgxpool.Pool
	logger *logger.Logger
}

func NewSetRepository(db *pgxpool.Pool) *SetRepository {
	return &SetRepository{
		db:     db,
		logger: logger.NewLogger(),
	}
}

func (sr *SetRepository) GetSetProtoList(ctx context.Context) ([]*set.SetProto, error) {
	sr.logger.Info("Fetching set list GetSetProtoList")
	query := `
		SELECT id, name, description, user_id, created_at, updated_at, is_favourite, like_by, is_public, image
		FROM sets
	`
	rows, err := sr.db.Query(ctx, query)
	if err != nil {
		sr.logger.Error("Error fetching sets: " + err.Error())
		return nil, fmt.Errorf("error fetching sets: %w", err)
	}
	defer rows.Close()

	var sets []*set.SetProto
	for rows.Next() {
		var s set.SetProto
		var createdAt, updatedAt time.Time
		var userID int32
		var likeBy []int64

		err := rows.Scan(
			&s.Id,
			&s.Name,
			&s.Description,
			&userID,
			&createdAt,
			&updatedAt,
			&s.IsFavourite,
			&likeBy,
			&s.IsPublic,
			&s.Image,
		)
		if err != nil {
			sr.logger.Error("Error scanning set: " + err.Error())
			return nil, fmt.Errorf("error scanning set: %w", err)
		}

		s.CreatedAt = timestamppb.New(createdAt)
		s.UpdatedAt = timestamppb.New(updatedAt)
		s.UserId = userID
		s.LikeBy = likeBy

		dishes, err := sr.GetDishesForSet(ctx, s.Id)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error fetching dishes for set %d: %s", s.Id, err.Error()))
			return nil, fmt.Errorf("error fetching dishes for set %d: %w", s.Id, err)
		}

		s.Dishes = dishes

		// Calculate price using the updated method
		price, err := sr.calculateSetPrice(ctx, dishes)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error calculating price for set %d: %s", s.Id, err.Error()))
			return nil, fmt.Errorf("error calculating price for set %d: %w", s.Id, err)
		}
		s.Price = price

		sets = append(sets, &s)
	}

	if err := rows.Err(); err != nil {
		sr.logger.Error("Error iterating over sets: " + err.Error())
		return nil, fmt.Errorf("error iterating over sets: %w", err)
	}

	sr.logger.Info(fmt.Sprintf("Successfully fetched %d sets", len(sets)))
	return sets, nil
}

func (sr *SetRepository) GetSetProtoDetail(ctx context.Context, id int32) (*set.SetProto, error) {
	sr.logger.Info(fmt.Sprintf("Fetching set detail for ID: %d", id))
	query := `
		SELECT id, name, description, user_id, created_at, updated_at, is_favourite, like_by, is_public, image
		FROM sets 
		WHERE id = $1
	`
	var s set.SetProto
	var createdAt, updatedAt time.Time
	var userID int32
	var likeBy []int64
	err := sr.db.QueryRow(ctx, query, id).Scan(
		&s.Id,
		&s.Name,
		&s.Description,
		&userID,
		&createdAt,
		&updatedAt,
		&s.IsFavourite,
		&likeBy,
		&s.IsPublic,
		&s.Image,
	)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error fetching set detail for ID %d: %s", id, err.Error()))
		return nil, fmt.Errorf("error fetching set detail: %w", err)
	}
	s.CreatedAt = timestamppb.New(createdAt)
	s.UpdatedAt = timestamppb.New(updatedAt)
	s.UserId = userID
	s.LikeBy = likeBy

	s.Dishes, err = sr.GetDishesForSet(ctx, s.Id)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error fetching dishes for set %d: %s", s.Id, err.Error()))
		return nil, fmt.Errorf("error fetching dishes for set %d: %w", s.Id, err)
	}

	// Calculate price using the updated method
	price, err := sr.calculateSetPrice(ctx, s.Dishes)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error calculating price for set %d: %s", s.Id, err.Error()))
		return nil, fmt.Errorf("error calculating price for set %d: %w", s.Id, err)
	}
	s.Price = price

	sr.logger.Info(fmt.Sprintf("Successfully fetched set detail for ID: %d", id))
	return &s, nil
}
func (sr *SetRepository) CreateSetProto(ctx context.Context, req *set.CreateSetProtoRequest) (*set.SetProto, error) {
	sr.logger.Info(fmt.Sprintf("Creating new set:CreateSetProto repository %+v", req))
	tx, err := sr.db.Begin(ctx)
	if err != nil {
		sr.logger.Error("Error starting transaction: " + err.Error())
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	// Calculate price and store it
	calculatedPrice, err := sr.calculateSetPrice(ctx, req.Dishes)
	if err != nil {
		return nil, fmt.Errorf("error calculating set price: %w", err)
	}
	sr.logger.Info(fmt.Sprintf("Step 1 - Calculated price: %d", calculatedPrice))

	query := `
        INSERT INTO sets (name, description, user_id, created_at, updated_at, is_favourite, like_by, is_public, image, price)
        VALUES ($1, $2, $3, $4, $4, $5, $6, $7, $8, $9)
        RETURNING id, created_at, updated_at, price
    `
	var s set.SetProto
	var createdAt, updatedAt time.Time
	var dbPrice int32

	// Log the values being inserted
	sr.logger.Info(fmt.Sprintf("Step 2 - About to insert with price: %d", calculatedPrice))

	err = tx.QueryRow(ctx, query,
		req.Name,
		req.Description,
		req.UserId,
		time.Now(),
		false,
		[]int64{},
		req.IsPublic,
		req.Image,
		calculatedPrice,
	).Scan(&s.Id, &createdAt, &updatedAt, &dbPrice)
	if err != nil {
		sr.logger.Error("Error creating set: " + err.Error())
		return nil, fmt.Errorf("error creating set: %w", err)
	}

	// Log the values returned from the database
	sr.logger.Info(fmt.Sprintf("Step 3 - Retrieved from DB - ID: %d, Price: %d", s.Id, dbPrice))

	s.Name = req.Name
	s.Description = req.Description
	s.UserId = req.UserId
	s.CreatedAt = timestamppb.New(createdAt)
	s.UpdatedAt = timestamppb.New(updatedAt)
	s.IsFavourite = false
	s.LikeBy = []int64{}
	s.IsPublic = req.IsPublic
	s.Image = req.Image
	s.Price = dbPrice

	// Log the price after setting it in the struct
	sr.logger.Info(fmt.Sprintf("Step 4 - Set price in struct: %d", s.Price))

	for _, dish := range req.Dishes {
		_, err := tx.Exec(ctx, "INSERT INTO set_dishes (set_id, dish_id, quantity) VALUES ($1, $2, $3)",
			s.Id, dish.DishId, dish.Quantity)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error inserting set dish: %s", err.Error()))
			return nil, fmt.Errorf("error inserting set dish: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		sr.logger.Error("Error committing transaction: " + err.Error())
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	s.Dishes = req.Dishes

	// Log the final state of the struct before returning
	sr.logger.Info(fmt.Sprintf("Step 5 - Final state - ID: %d, Price: %d", s.Id, s.Price))

	return &s, nil
}
func (sr *SetRepository) UpdateSetProto(ctx context.Context, req *set.UpdateSetProtoRequest) (*set.SetProto, error) {
	sr.logger.Info(fmt.Sprintf("Updating set with ID: %d", req.Id))
	tx, err := sr.db.Begin(ctx)
	if err != nil {
		sr.logger.Error("Error starting transaction: " + err.Error())
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	query := `
		UPDATE sets
		SET name = $2, description = $3, updated_at = $4, is_public = $5, image = $6, price = $7
		WHERE id = $1
		RETURNING user_id, created_at, updated_at, is_favourite, like_by
	`
	var s set.SetProto
	var createdAt, updatedAt time.Time
	var userID int32
	var likeBy []int64
	price, err := sr.calculateSetPrice(ctx, req.Dishes)
	if err != nil {
		return nil, fmt.Errorf("error calculating set price: %w", err)
	}

	err = tx.QueryRow(ctx, query,
		req.Id,
		req.Name,
		req.Description,
		time.Now(),
		req.IsPublic,
		req.Image,
		price,
	).Scan(&userID, &createdAt, &updatedAt, &s.IsFavourite, &likeBy)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error updating set with ID %d: %s", req.Id, err.Error()))
		return nil, fmt.Errorf("error updating set: %w", err)
	}

	s.Id = req.Id
	s.Name = req.Name
	s.Description = req.Description
	s.UserId = userID
	s.CreatedAt = timestamppb.New(createdAt)
	s.UpdatedAt = timestamppb.New(updatedAt)
	s.LikeBy = likeBy
	s.IsPublic = req.IsPublic
	s.Image = req.Image
	s.Price = price

	_, err = tx.Exec(ctx, "DELETE FROM set_dishes WHERE set_id = $1", req.Id)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error deleting existing set dishes for set ID %d: %s", req.Id, err.Error()))
		return nil, fmt.Errorf("error deleting existing set dishes: %w", err)
	}

	for _, dish := range req.Dishes {
		_, err := tx.Exec(ctx, "INSERT INTO set_dishes (set_id, dish_id, quantity) VALUES ($1, $2, $3)",
			s.Id, dish.DishId, dish.Quantity)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error inserting updated set dish for set ID %d: %s", s.Id, err.Error()))
			return nil, fmt.Errorf("error inserting updated set dish: %w", err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		sr.logger.Error("Error committing transaction: " + err.Error())
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	s.Dishes = req.Dishes
	sr.logger.Info(fmt.Sprintf("Successfully updated set with ID: %d", s.Id))
	return &s, nil
}

func (sr *SetRepository) DeleteSetProto(ctx context.Context, id int32) (*set.SetProto, error) {
	sr.logger.Info(fmt.Sprintf("Deleting set with ID: %d", id))
	tx, err := sr.db.Begin(ctx)
	if err != nil {
		sr.logger.Error("Error starting transaction: " + err.Error())
		return nil, fmt.Errorf("error starting transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	_, err = tx.Exec(ctx, "DELETE FROM set_dishes WHERE set_id = $1", id)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error deleting set dishes for set ID %d: %s", id, err.Error()))
		return nil, fmt.Errorf("error deleting set dishes: %w", err)
	}

	query := `
		DELETE FROM sets
		WHERE id = $1
		RETURNING id, name, description, user_id, created_at, updated_at, is_favourite, like_by, is_public, image, price
	`
	var s set.SetProto
	var createdAt, updatedAt time.Time
	var userID int32
	var likeBy []int64
	err = tx.QueryRow(ctx, query, id).Scan(
		&s.Id,
		&s.Name,
		&s.Description,
		&userID,
		&createdAt,
		&updatedAt,
		&s.IsFavourite,
		&likeBy,
		&s.IsPublic,
		&s.Image,
		&s.Price,
	)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error deleting set with ID %d: %s", id, err.Error()))
		return nil, fmt.Errorf("error deleting set: %w", err)
	}
	s.CreatedAt = timestamppb.New(createdAt)
	s.UpdatedAt = timestamppb.New(updatedAt)
	s.UserId = userID
	s.LikeBy = likeBy

	if err := tx.Commit(ctx); err != nil {
		sr.logger.Error("Error committing transaction: " + err.Error())
		return nil, fmt.Errorf("error committing transaction: %w", err)
	}

	sr.logger.Info(fmt.Sprintf("Successfully deleted set with ID: %d", id))
	return &s, nil
}

func (sr *SetRepository) GetDishesForSet(ctx context.Context, setID int32) ([]*set.SetProtoDish, error) {
	query := `
		SELECT sd.dish_id, sd.quantity
		FROM set_dishes sd
		WHERE sd.set_id = $1
	`
	rows, err := sr.db.Query(ctx, query, setID)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error fetching dishes for set %d: %s", setID, err.Error()))
		return nil, fmt.Errorf("error fetching dishes for set: %w", err)
	}
	defer rows.Close()

	var dishes []*set.SetProtoDish
	for rows.Next() {
		var dish set.SetProtoDish
		err := rows.Scan(&dish.DishId, &dish.Quantity)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error scanning dish for set %d: %s", setID, err.Error()))
			return nil, fmt.Errorf("error scanning dish: %w", err)
		}
		dishes = append(dishes, &dish)
	}

	if err := rows.Err(); err != nil {
		sr.logger.Error(fmt.Sprintf("Error iterating over dishes for set %d: %s", setID, err.Error()))
		return nil, fmt.Errorf("error iterating over dishes: %w", err)
	}

	return dishes, nil
}

func (sr *SetRepository) calculateSetPrice(ctx context.Context, dishes []*set.SetProtoDish) (int32, error) {
	sr.logger.Info(fmt.Sprintf("Calculating price for dishes: %+v", dishes))

	if len(dishes) == 0 {
		sr.logger.Info("No dishes provided, returning price 0")
		return 0, nil
	}

	dishIDs := make([]int32, len(dishes))
	dishQuantityMap := make(map[int32]int32)
	for i, dish := range dishes {
		dishIDs[i] = dish.DishId
		dishQuantityMap[dish.DishId] = dish.Quantity
	}

	sr.logger.Info(fmt.Sprintf("DishIDs to query: %v", dishIDs))
	sr.logger.Info(fmt.Sprintf("Quantity map: %v", dishQuantityMap))

	query := `
        SELECT id, price 
        FROM dishes 
        WHERE id = ANY($1)
    `
	rows, err := sr.db.Query(ctx, query, dishIDs)
	if err != nil {
		sr.logger.Error("Error fetching dish prices: " + err.Error())
		return 0, fmt.Errorf("error fetching dish prices: %w", err)
	}
	defer rows.Close()

	var totalPrice int32
	for rows.Next() {
		var dishID int32
		var price int32
		if err := rows.Scan(&dishID, &price); err != nil {
			sr.logger.Error("Error scanning dish price: " + err.Error())
			return 0, fmt.Errorf("error scanning dish price: %w", err)
		}
		quantity := dishQuantityMap[dishID]
		totalPrice += int32(quantity) * price
		sr.logger.Info(fmt.Sprintf("DishID: %d, Price: %d, Quantity: %d, Running Total: %d",
			dishID, price, quantity, totalPrice))
	}

	if err := rows.Err(); err != nil {
		sr.logger.Error("Error iterating over dish prices: " + err.Error())
		return 0, fmt.Errorf("error iterating over dish prices: %w", err)
	}

	sr.logger.Info(fmt.Sprintf("Final calculated price: %d", totalPrice))
	return totalPrice, nil
}

func (sr *SetRepository) GetSetProtoListDetail(ctx context.Context) ([]*set.SetProtoDetailedResponse, error) {
	sr.logger.Info("Fetching detailed set list GetSetProtoListDetail")

	// First, fetch all sets
	query := `
        SELECT id, name, description, user_id, created_at, updated_at, is_favourite, like_by, is_public, image, price
        FROM sets
    `
	rows, err := sr.db.Query(ctx, query)
	if err != nil {
		sr.logger.Error("Error fetching sets: " + err.Error())
		return nil, fmt.Errorf("error fetching sets: %w", err)
	}
	defer rows.Close()

	var sets []*set.SetProtoDetailedResponse
	for rows.Next() {
		var s set.SetProtoDetailedResponse
		var createdAt, updatedAt time.Time
		var userID int32
		var likeBy []int64

		err := rows.Scan(
			&s.Id,
			&s.Name,
			&s.Description,
			&userID,
			&createdAt,
			&updatedAt,
			&s.IsFavourite,
			&likeBy,
			&s.IsPublic,
			&s.Image,
			&s.Price,
		)
		if err != nil {
			sr.logger.Error("Error scanning set: " + err.Error())
			return nil, fmt.Errorf("error scanning set: %w", err)
		}

		s.CreatedAt = timestamppb.New(createdAt)
		s.UpdatedAt = timestamppb.New(updatedAt)
		s.UserId = userID
		s.LikeBy = likeBy

		// Fetch detailed dishes for this set
		dishes, err := sr.GetDetailedDishesForSet(ctx, s.Id)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error fetching detailed dishes for set %d: %s", s.Id, err.Error()))
			return nil, fmt.Errorf("error fetching detailed dishes for set %d: %w", s.Id, err)
		}

		s.Dishes = dishes
		sets = append(sets, &s)
	}

	if err := rows.Err(); err != nil {
		sr.logger.Error("Error iterating over sets: " + err.Error())
		return nil, fmt.Errorf("error iterating over sets: %w", err)
	}

	sr.logger.Info(fmt.Sprintf("Successfully fetched %d detailed sets", len(sets)))
	return sets, nil
}

// New helper function to get detailed dishes for a set
func (sr *SetRepository) GetDetailedDishesForSet(ctx context.Context, setID int32) ([]*set.SetProtoDishDetailed, error) {
	query := `
        SELECT d.id, sd.quantity, d.name, d.price, d.description, d.image, d.status
        FROM set_dishes sd
        JOIN dishes d ON sd.dish_id = d.id
        WHERE sd.set_id = $1
    `
	rows, err := sr.db.Query(ctx, query, setID)
	if err != nil {
		sr.logger.Error(fmt.Sprintf("Error fetching detailed dishes for set %d: %s", setID, err.Error()))
		return nil, fmt.Errorf("error fetching detailed dishes for set: %w", err)
	}
	defer rows.Close()

	var dishes []*set.SetProtoDishDetailed
	for rows.Next() {
		var dish set.SetProtoDishDetailed
		err := rows.Scan(
			&dish.DishId,
			&dish.Quantity,
			&dish.Name,
			&dish.Price,
			&dish.Description,
			&dish.Image,
			&dish.Status,
		)
		if err != nil {
			sr.logger.Error(fmt.Sprintf("Error scanning detailed dish for set %d: %s", setID, err.Error()))
			return nil, fmt.Errorf("error scanning detailed dish: %w", err)
		}
		dishes = append(dishes, &dish)
	}

	if err := rows.Err(); err != nil {
		sr.logger.Error(fmt.Sprintf("Error iterating over detailed dishes for set %d: %s", setID, err.Error()))
		return nil, fmt.Errorf("error iterating over detailed dishes: %w", err)
	}

	return dishes, nil
}
