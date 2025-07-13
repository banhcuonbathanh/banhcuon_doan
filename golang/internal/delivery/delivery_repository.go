package delivery_grpc

//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"math"
//	"time"
//
//	"english-ai-full/logger"
//	"english-ai-full/quanqr/proto_qr/delivery"
//	"github.com/jackc/pgx/v4/pgxpool"
//	"google.golang.org/protobuf/types/known/timestamppb"
//)
//
//type Repository struct {
//	db     *pgxpool.Pool
//	logger *logger.Logger
//}
//
//func NewRepository(db *pgxpool.Pool) *Repository {
//	return &Repository{
//		db:     db,
//		logger: logger.NewLogger(),
//	}
//}
//
//func (dr *Repository) CreateDelivery(ctx context.Context, req *delivery.CreateDeliverRequest) (*delivery.DeliverResponse, error) {
//	dr.logger.Info(fmt.Sprintf("Creating new delivery: %+v", req))
//	tx, err := dr.db.Begin(ctx)
//	if err != nil {
//		dr.logger.Error("Error starting transaction: " + err.Error())
//		return nil, fmt.Errorf("error starting transaction: %w", err)
//	}
//	defer tx.Rollback(ctx)
//
//	query := `
//		INSERT INTO deliveries (
//			guest_id, user_id, is_guest, table_number, order_handler_id,
//			status, created_at, updated_at, total_price, bow_chili, bow_no_chili,
//			take_away, chili_number, table_token, client_name, delivery_address,
//			delivery_contact, delivery_notes, scheduled_time, delivery_fee,
//			delivery_status, estimated_delivery_time
//		)
//		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
//				$17, $18, $19, $20, $21, $22)
//		RETURNING id, created_at, updated_at
//	`
//
//	var d delivery.DeliverResponse
//	var createdAt, updatedAt time.Time
//	var guestId, userId sql.NullInt64
//
//	if req.IsGuest {
//		guestId = sql.NullInt64{Int64: req.GuestId, Valid: true}
//		userId = sql.NullInt64{Valid: false}
//	} else {
//		userId = sql.NullInt64{Int64: req.UserId, Valid: true}
//		guestId = sql.NullInt64{Valid: false}
//	}
//
//	now := time.Now()
//	err = tx.QueryRow(ctx, query,
//		guestId,
//		userId,
//		req.IsGuest,
//		req.TableNumber,
//		req.OrderHandlerId,
//		req.Status,
//		now,
//		now,
//		req.TotalPrice,
//		req.BowChili,
//		req.BowNoChili,
//		req.TakeAway,
//		req.ChiliNumber,
//		req.TableToken,
//		req.ClientName,
//		req.DeliveryAddress,
//		req.DeliveryContact,
//		req.DeliveryNotes,
//		req.ScheduledTime.AsTime(),
//		req.DeliveryFee,
//		req.DeliveryStatus,
//		req.EstimatedDeliveryTime.AsTime(),
//	).Scan(&d.Id, &createdAt, &updatedAt)
//
//	if err != nil {
//		dr.logger.Error("Error creating delivery: " + err.Error())
//		return nil, fmt.Errorf("error creating delivery: %w", err)
//	}
//
//	// Insert dish items
//	for _, dish := range req.DishItems {
//		var exists bool
//		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dishes WHERE id = $1)", dish.DishId).Scan(&exists)
//		if err != nil {
//			dr.logger.Error(fmt.Sprintf("Error verifying dish existence: %s", err.Error()))
//			return nil, fmt.Errorf("error verifying dish existence: %w", err)
//		}
//		if !exists {
//			dr.logger.Error(fmt.Sprintf("Dish with id %d does not exist", dish.DishId))
//			return nil, fmt.Errorf("dish with id %d does not exist", dish.DishId)
//		}
//
//		_, err = tx.Exec(ctx,
//			"INSERT INTO dish_delivery_items (delivery_id, dish_id, quantity) VALUES ($1, $2, $3)",
//			d.Id, dish.DishId, dish.Quantity)
//		if err != nil {
//			dr.logger.Error(fmt.Sprintf("Error inserting delivery dish: %s", err.Error()))
//			return nil, fmt.Errorf("error inserting delivery dish: %w", err)
//		}
//	}
//
//	if err := tx.Commit(ctx); err != nil {
//		dr.logger.Error("Error committing transaction: " + err.Error())
//		return nil, fmt.Errorf("error committing transaction: %w", err)
//	}
//
//	// Populate response
//	d.GuestId = req.GuestId
//	d.UserId = req.UserId
//	d.IsGuest = req.IsGuest
//	d.TableNumber = req.TableNumber
//	d.OrderHandlerId = req.OrderHandlerId
//	d.Status = req.Status
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//	d.TotalPrice = req.TotalPrice
//	d.DishItems = req.DishItems
//	d.BowChili = req.BowChili
//	d.BowNoChili = req.BowNoChili
//	d.TakeAway = req.TakeAway
//	d.ChiliNumber = req.ChiliNumber
//	d.TableToken = req.TableToken
//	d.ClientName = req.ClientName
//	d.DeliveryAddress = req.DeliveryAddress
//	d.DeliveryContact = req.DeliveryContact
//	d.DeliveryNotes = req.DeliveryNotes
//	d.ScheduledTime = req.ScheduledTime
//	d.DeliveryFee = req.DeliveryFee
//	d.DeliveryStatus = req.DeliveryStatus
//	d.EstimatedDeliveryTime = req.EstimatedDeliveryTime
//
//	return &d, nil
//}
//
//func (dr *Repository) GetDeliveriesList(ctx context.Context, req *delivery.GetDeliveriesRequest) (*delivery.DeliveryDetailedListResponse, error) {
//	countQuery := `SELECT COUNT(*) FROM deliveries`
//	dr.logger.Info("Fetching deliveries list handler repository")
//	var totalItems int64
//	err := dr.db.QueryRow(ctx, countQuery).Scan(&totalItems)
//	if err != nil {
//		dr.logger.Error("Error counting deliveries: " + err.Error())
//		return nil, fmt.Errorf("error counting deliveries: %w", err)
//	}
//
//	totalPages := int32(math.Ceil(float64(totalItems) / float64(req.PageSize)))
//	offset := (req.Page - 1) * req.PageSize
//
//	query := `
//		SELECT
//			d.id, d.guest_id, d.user_id, d.is_guest, d.table_number,
//			d.order_handler_id, d.status, d.created_at, d.updated_at,
//			d.total_price, d.bow_chili, d.bow_no_chili, d.take_away,
//			d.chili_number, d.table_token, d.client_name, d.delivery_address,
//			d.delivery_contact, d.delivery_notes, d.scheduled_time,
//			d.delivery_fee, d.delivery_status, d.estimated_delivery_time,
//			d.actual_delivery_time
//		FROM deliveries d
//		ORDER BY d.created_at DESC
//		LIMIT $1 OFFSET $2
//	`
//
//	rows, err := dr.db.Query(ctx, query, req.PageSize, offset)
//	if err != nil {
//		dr.logger.Error("Error fetching deliveries: " + err.Error())
//		return nil, fmt.Errorf("error fetching deliveries: %w", err)
//	}
//	defer rows.Close()
//
//	var deliveries []*delivery.DeliveryDetailedResponse
//	for rows.Next() {
//		var d delivery.DeliveryDetailedResponse
//		var createdAt, updatedAt time.Time
//		var guestIdNull sql.NullInt64
//		var scheduledTimeNull, estimatedTimeNull, actualTimeNull sql.NullTime
//
//		err := rows.Scan(
//			&d.Id, &guestIdNull, &d.UserId, &d.IsGuest, &d.TableNumber,
//			&d.OrderHandlerId, &d.Status, &createdAt, &updatedAt,
//			&d.TotalPrice, &d.BowChili, &d.BowNoChili, &d.TakeAway,
//			&d.ChiliNumber, &d.TableToken, &d.ClientName, &d.DeliveryAddress,
//			&d.DeliveryContact, &d.DeliveryNotes, &scheduledTimeNull,
//			&d.DeliveryFee, &d.DeliveryStatus, &estimatedTimeNull,
//			&actualTimeNull,
//		)
//		if err != nil {
//			dr.logger.Error("Error scanning delivery: " + err.Error())
//			return nil, fmt.Errorf("error scanning delivery: %w", err)
//		}
//
//		d.CreatedAt = timestamppb.New(createdAt)
//		d.UpdatedAt = timestamppb.New(updatedAt)
//
//		// Handle nullable fields
//		if guestIdNull.Valid {
//			d.GuestId = guestIdNull.Int64
//		}
//		if scheduledTimeNull.Valid {
//			d.ScheduledTime = timestamppb.New(scheduledTimeNull.Time)
//		}
//		if estimatedTimeNull.Valid {
//			d.EstimatedDeliveryTime = timestamppb.New(estimatedTimeNull.Time)
//		}
//		if actualTimeNull.Valid {
//			d.ActualDeliveryTime = timestamppb.New(actualTimeNull.Time)
//		}
//
//		// Fetch dish items
//		dishItems, err := dr.getDeliveryDishItems(ctx, d.Id)
//		if err != nil {
//			return nil, err
//		}
//		d.DishItems = dishItems
//
//		deliveries = append(deliveries, &d)
//	}
//
//	return &delivery.DeliveryDetailedListResponse{
//		Data: deliveries,
//		Pagination: &delivery.PaginationInfo{
//			CurrentPage: req.Page,
//			TotalPages:  totalPages,
//			TotalItems:  totalItems,
//			PageSize:    req.PageSize,
//		},
//	}, nil
//}
//
//// func (dr *Repository) GetDeliveriesList(ctx context.Context, req *delivery.GetDeliveriesRequest) (*delivery.DeliveryDetailedListResponse, error) {
//// 	countQuery := `SELECT COUNT(*) FROM deliveries`
//// 	dr.logger.Info("Fetching deliveries list hander repository")
//// 	var totalItems int64
//// 	err := dr.db.QueryRow(ctx, countQuery).Scan(&totalItems)
//// 	if err != nil {
//// 		dr.logger.Error("Error counting deliveries: " + err.Error())
//// 		return nil, fmt.Errorf("error counting deliveries: %w", err)
//// 	}
//// 	dr.logger.Info("Fetching deliveries list hander repository")
//// 	totalPages := int32(math.Ceil(float64(totalItems) / float64(req.PageSize)))
//// 	offset := (req.Page - 1) * req.PageSize
//
//// 	query := `
//// 		SELECT
//// 			d.id, d.guest_id, d.user_id, d.is_guest, d.table_number,
//// 			d.order_handler_id, d.status, d.created_at, d.updated_at,
//// 			d.total_price, d.bow_chili, d.bow_no_chili, d.take_away,
//// 			d.chili_number, d.table_token, d.client_name, d.delivery_address,
//// 			d.delivery_contact, d.delivery_notes, d.scheduled_time,
//// 			d.delivery_fee, d.delivery_status, d.estimated_delivery_time,
//// 			d.actual_delivery_time
//// 		FROM deliveries d
//// 		ORDER BY d.created_at DESC
//// 		LIMIT $1 OFFSET $2
//// 	`
//
//// 	rows, err := dr.db.Query(ctx, query, req.PageSize, offset)
//// 	if err != nil {
//// 		dr.logger.Error("Error fetching deliveries: " + err.Error())
//// 		return nil, fmt.Errorf("error fetching deliveries: %w", err)
//// 	}
//// 	defer rows.Close()
//// 	dr.logger.Info("Fetching deliveries list hander repository")
//// 	var deliveries []*delivery.DeliveryDetailedResponse
//// 	for rows.Next() {
//// 		var d delivery.DeliveryDetailedResponse
//// 		var createdAt, updatedAt time.Time
//// 		var scheduledTimeNull, estimatedTimeNull, actualTimeNull sql.NullTime
//
//// 		err := rows.Scan(
//// 			&d.Id, &d.GuestId, &d.UserId, &d.IsGuest, &d.TableNumber,
//// 			&d.OrderHandlerId, &d.Status, &createdAt, &updatedAt,
//// 			&d.TotalPrice, &d.BowChili, &d.BowNoChili, &d.TakeAway,
//// 			&d.ChiliNumber, &d.TableToken, &d.ClientName, &d.DeliveryAddress,
//// 			&d.DeliveryContact, &d.DeliveryNotes, &scheduledTimeNull,
//// 			&d.DeliveryFee, &d.DeliveryStatus, &estimatedTimeNull,
//// 			&actualTimeNull,
//// 		)
//// 		if err != nil {
//// 			dr.logger.Error("Error scanning delivery: " + err.Error())
//// 			return nil, fmt.Errorf("error scanning delivery: %w", err)
//// 		}
//// 		dr.logger.Info("Fetching deliveries list hander repository")
//// 		d.CreatedAt = timestamppb.New(createdAt)
//// 		d.UpdatedAt = timestamppb.New(updatedAt)
//
//// 		if scheduledTimeNull.Valid {
//// 			d.ScheduledTime = timestamppb.New(scheduledTimeNull.Time)
//// 		}
//// 		if estimatedTimeNull.Valid {
//// 			d.EstimatedDeliveryTime = timestamppb.New(estimatedTimeNull.Time)
//// 		}
//// 		if actualTimeNull.Valid {
//// 			d.ActualDeliveryTime = timestamppb.New(actualTimeNull.Time)
//// 		}
//
//// 		// Fetch dish items
//// 		dishItems, err := dr.getDeliveryDishItems(ctx, d.Id)
//// 		if err != nil {
//// 			return nil, err
//// 		}
//// 		d.DishItems = dishItems
//
//// 		deliveries = append(deliveries, &d)
//// 	}
//
//// 	return &delivery.DeliveryDetailedListResponse{
//// 		Data: deliveries,
//// 		Pagination: &delivery.PaginationInfo{
//// 			CurrentPage: req.Page,
//// 			TotalPages: totalPages,
//// 			TotalItems: totalItems,
//// 			PageSize:   req.PageSize,
//// 		},
//// 	}, nil
//// }
//
//func (dr *Repository) getDeliveryDishItems(ctx context.Context, deliveryId int64) ([]*delivery.DeliveryDetailedDish, error) {
//	query := `
//		SELECT
//			d.id, ddi.quantity, d.name, d.price,
//			d.description, d.image, d.status
//		FROM dish_delivery_items ddi
//		JOIN dishes d ON ddi.dish_id = d.id
//		WHERE ddi.delivery_id = $1
//	`
//
//	rows, err := dr.db.Query(ctx, query, deliveryId)
//	if err != nil {
//		dr.logger.Error("Error fetching delivery dish details: " + err.Error())
//		return nil, fmt.Errorf("error fetching delivery dish details: %w", err)
//	}
//	defer rows.Close()
//
//	var dishes []*delivery.DeliveryDetailedDish
//	for rows.Next() {
//		var dish delivery.DeliveryDetailedDish
//		err := rows.Scan(
//			&dish.DishId,
//			&dish.Quantity,
//			&dish.Name,
//			&dish.Price,
//			&dish.Description,
//			&dish.Image,
//			&dish.Status,
//		)
//		if err != nil {
//			dr.logger.Error("Error scanning delivery dish: " + err.Error())
//			return nil, fmt.Errorf("error scanning delivery dish: %w", err)
//		}
//		dishes = append(dishes, &dish)
//	}
//
//	return dishes, nil
//}
//
//func (dr *Repository) GetDeliveryById(ctx context.Context, id int64) (*delivery.DeliverResponse, error) {
//	query := `
//		SELECT
//			id, guest_id, user_id, is_guest, table_number, order_handler_id,
//			status, created_at, updated_at, total_price, bow_chili, bow_no_chili,
//			take_away, chili_number, table_token, client_name, delivery_address,
//			delivery_contact, delivery_notes, scheduled_time, delivery_fee,
//			delivery_status, estimated_delivery_time, actual_delivery_time
//		FROM deliveries
//		WHERE id = $1
//	`
//
//	var d delivery.DeliverResponse
//	var createdAt, updatedAt time.Time
//	var scheduledTimeNull, estimatedTimeNull, actualTimeNull sql.NullTime
//
//	err := dr.db.QueryRow(ctx, query, id).Scan(
//		&d.Id, &d.GuestId, &d.UserId, &d.IsGuest, &d.TableNumber,
//		&d.OrderHandlerId, &d.Status, &createdAt, &updatedAt,
//		&d.TotalPrice, &d.BowChili, &d.BowNoChili, &d.TakeAway,
//		&d.ChiliNumber, &d.TableToken, &d.ClientName, &d.DeliveryAddress,
//		&d.DeliveryContact, &d.DeliveryNotes, &scheduledTimeNull,
//		&d.DeliveryFee, &d.DeliveryStatus, &estimatedTimeNull,
//		&actualTimeNull,
//	)
//	if err != nil {
//		dr.logger.Error(fmt.Sprintf("Error fetching delivery detail: %s", err.Error()))
//		return nil, fmt.Errorf("error fetching delivery detail: %w", err)
//	}
//
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//
//	if scheduledTimeNull.Valid {
//		d.ScheduledTime = timestamppb.New(scheduledTimeNull.Time)
//	}
//	if estimatedTimeNull.Valid {
//		d.EstimatedDeliveryTime = timestamppb.New(estimatedTimeNull.Time)
//	}
//	if actualTimeNull.Valid {
//		d.ActualDeliveryTime = timestamppb.New(actualTimeNull.Time)
//	}
//
//	// Get dish items
//	dishItems, err := dr.getDishItems(ctx, id)
//	if err != nil {
//		return nil, err
//	}
//	d.DishItems = dishItems
//
//	return &d, nil
//}
//
//func (dr *Repository) getDishItems(ctx context.Context, deliveryId int64) ([]*delivery.DishDeliveryItem, error) {
//	query := `
//		SELECT dish_id, quantity
//		FROM dish_delivery_items
//		WHERE delivery_id = $1
//	`
//
//	rows, err := dr.db.Query(ctx, query, deliveryId)
//	if err != nil {
//		dr.logger.Error(fmt.Sprintf("Error fetching delivery dish items: %s", err.Error()))
//		return nil, fmt.Errorf("error fetching delivery dish items: %w", err)
//	}
//	defer rows.Close()
//
//	var items []*delivery.DishDeliveryItem
//	for rows.Next() {
//		item := &delivery.DishDeliveryItem{}
//		if err := rows.Scan(&item.DishId, &item.Quantity); err != nil {
//			dr.logger.Error(fmt.Sprintf("Error scanning dish item: %s", err.Error()))
//			return nil, fmt.Errorf("error scanning dish item: %w", err)
//		}
//		items = append(items, item)
//	}
//
//	if err = rows.Err(); err != nil {
//		dr.logger.Error(fmt.Sprintf("Error iterating dish items: %s", err.Error()))
//		return nil, fmt.Errorf("error iterating dish items: %w", err)
//	}
//
//	// If no items found, return empty slice instead of nil
//	if items == nil {
//		items = make([]*delivery.DishDeliveryItem, 0)
//	}
//
//	return items, nil
//}
//
//func (dr *Repository) GetDeliveryByClientName(ctx context.Context, clientName string) (*delivery.DeliveryDetailedListResponse, error) {
//	dr.logger.Info("Fetching deliveries repository GetDeliveryByClientName")
//	query := `
//        SELECT
//            id, guest_id, user_id, is_guest, table_number, order_handler_id,
//            status, created_at, updated_at, total_price, bow_chili, bow_no_chili,
//            take_away, chili_number, table_token, client_name, delivery_address,
//            delivery_contact, delivery_notes, scheduled_time, delivery_fee,
//            delivery_status, estimated_delivery_time, actual_delivery_time
//        FROM deliveries
//        WHERE client_name = $1
//        ORDER BY created_at DESC
//    `
//
//	rows, err := dr.db.Query(ctx, query, clientName)
//	if err != nil {
//		dr.logger.Error(fmt.Sprintf("Error querying deliveries: %s", err.Error()))
//		return nil, fmt.Errorf("error querying deliveries: %w", err)
//	}
//	defer rows.Close()
//
//	var deliveries []*delivery.DeliveryDetailedResponse
//
//	for rows.Next() {
//		var d delivery.DeliveryDetailedResponse
//		var createdAt, updatedAt time.Time
//
//		// Create nullable variables for fields that can be NULL
//		var (
//			nullGuestID        sql.NullInt64
//			nullUserID         sql.NullInt64
//			nullTableNumber    sql.NullInt64
//			nullOrderHandlerID sql.NullInt64
//			nullBowChili       sql.NullInt64
//			nullBowNoChili     sql.NullInt64
//			nullChiliNumber    sql.NullInt64
//			scheduledTimeNull  sql.NullTime
//			estimatedTimeNull  sql.NullTime
//			actualTimeNull     sql.NullTime
//			nullDeliveryFee    sql.NullInt32
//		)
//
//		err := rows.Scan(
//			&d.Id,
//			&nullGuestID,
//			&nullUserID,
//			&d.IsGuest,
//			&nullTableNumber,
//			&nullOrderHandlerID,
//			&d.Status,
//			&createdAt,
//			&updatedAt,
//			&d.TotalPrice,
//			&nullBowChili,
//			&nullBowNoChili,
//			&d.TakeAway,
//			&nullChiliNumber,
//			&d.TableToken,
//			&d.ClientName,
//			&d.DeliveryAddress,
//			&d.DeliveryContact,
//			&d.DeliveryNotes,
//			&scheduledTimeNull,
//			&nullDeliveryFee,
//			&d.DeliveryStatus,
//			&estimatedTimeNull,
//			&actualTimeNull,
//		)
//
//		if err != nil {
//			dr.logger.Error(fmt.Sprintf("Error scanning delivery row: %s", err.Error()))
//			return nil, fmt.Errorf("error scanning delivery row: %w", err)
//		}
//
//		// Convert NULL values to their appropriate types
//		if nullGuestID.Valid {
//			d.GuestId = nullGuestID.Int64
//		}
//		if nullUserID.Valid {
//			d.UserId = nullUserID.Int64
//		}
//		if nullTableNumber.Valid {
//			d.TableNumber = nullTableNumber.Int64
//		}
//		if nullOrderHandlerID.Valid {
//			d.OrderHandlerId = nullOrderHandlerID.Int64
//		}
//		if nullBowChili.Valid {
//			d.BowChili = nullBowChili.Int64
//		}
//		if nullBowNoChili.Valid {
//			d.BowNoChili = nullBowNoChili.Int64
//		}
//		if nullChiliNumber.Valid {
//			d.ChiliNumber = nullChiliNumber.Int64
//		}
//		if nullDeliveryFee.Valid {
//			d.DeliveryFee = nullDeliveryFee.Int32
//		}
//
//		// Convert timestamps
//		d.CreatedAt = timestamppb.New(createdAt)
//		d.UpdatedAt = timestamppb.New(updatedAt)
//
//		if scheduledTimeNull.Valid {
//			d.ScheduledTime = timestamppb.New(scheduledTimeNull.Time)
//		}
//		if estimatedTimeNull.Valid {
//			d.EstimatedDeliveryTime = timestamppb.New(estimatedTimeNull.Time)
//		}
//		if actualTimeNull.Valid {
//			d.ActualDeliveryTime = timestamppb.New(actualTimeNull.Time)
//		}
//
//		// Get detailed dish items
//		detailedDishItems, err := dr.getDetailedDishItems(ctx, d.Id)
//		if err != nil {
//			return nil, fmt.Errorf("error fetching detailed dish items: %w", err)
//		}
//		d.DishItems = detailedDishItems
//
//		deliveries = append(deliveries, &d)
//	}
//
//	if err = rows.Err(); err != nil {
//		dr.logger.Error(fmt.Sprintf("Error iterating over rows: %s", err.Error()))
//		return nil, fmt.Errorf("error iterating over rows: %w", err)
//	}
//
//	if len(deliveries) == 0 {
//		return nil, fmt.Errorf("no deliveries found for client: %s", clientName)
//	}
//
//	response := &delivery.DeliveryDetailedListResponse{
//		Data: deliveries,
//		Pagination: &delivery.PaginationInfo{
//			CurrentPage: 1,
//			TotalPages:  1,
//			TotalItems:  int64(len(deliveries)),
//			PageSize:    int32(len(deliveries)),
//		},
//	}
//
//	return response, nil
//}
//
//func (dr *Repository) UpdateDelivery(ctx context.Context, req *delivery.UpdateDeliverRequest) (*delivery.DeliverResponse, error) {
//	dr.logger.Info(fmt.Sprintf("Updating delivery: %+v", req))
//	tx, err := dr.db.Begin(ctx)
//	if err != nil {
//		dr.logger.Error("Error starting transaction: " + err.Error())
//		return nil, fmt.Errorf("error starting transaction: %w", err)
//	}
//	defer tx.Rollback(ctx)
//
//	query := `
//				UPDATE deliveries SET
//					guest_id = CASE WHEN $1 THEN $2 ELSE NULL END,
//					user_id = CASE WHEN NOT $1 THEN $3 ELSE NULL END,
//					is_guest = $1,
//					table_number = $4,
//					order_handler_id = $5,
//					status = $6,
//					updated_at = $7,
//					total_price = $8,
//					bow_chili = $9,
//					bow_no_chili = $10,
//					take_away = $11,
//					chili_number = $12,
//					table_token = $13,
//					client_name = $14,
//					delivery_address = $15,
//					delivery_contact = $16,
//					delivery_notes = $17,
//					scheduled_time = $18,
//					delivery_fee = $19,
//					delivery_status = $20,
//					estimated_delivery_time = $21,
//					actual_delivery_time = $22
//				WHERE id = $23
//				RETURNING id, created_at, updated_at
//			`
//
//	var createdAt, updatedAt time.Time
//	var d delivery.DeliverResponse
//
//	err = tx.QueryRow(ctx, query,
//		req.IsGuest,
//		req.GuestId,
//		req.UserId,
//		req.TableNumber,
//		req.OrderHandlerId,
//		req.Status,
//		time.Now(),
//		req.TotalPrice,
//		req.BowChili,
//		req.BowNoChili,
//		req.TakeAway,
//		req.ChiliNumber,
//		req.TableToken,
//		req.ClientName,
//		req.DeliveryAddress,
//		req.DeliveryContact,
//		req.DeliveryNotes,
//		req.ScheduledTime.AsTime(),
//		req.DeliveryFee,
//		req.DeliveryStatus,
//		req.EstimatedDeliveryTime.AsTime(),
//		req.ActualDeliveryTime.AsTime(),
//		req.Id,
//	).Scan(&d.Id, &createdAt, &updatedAt)
//
//	if err != nil {
//		dr.logger.Error("Error updating delivery: " + err.Error())
//		return nil, fmt.Errorf("error updating delivery: %w", err)
//	}
//
//	// Delete existing dish items
//	_, err = tx.Exec(ctx, "DELETE FROM dish_delivery_items WHERE delivery_id = $1", d.Id)
//	if err != nil {
//		dr.logger.Error("Error deleting existing dish items: " + err.Error())
//		return nil, fmt.Errorf("error deleting existing dish items: %w", err)
//	}
//
//	// Insert new dish items
//	for _, dish := range req.DishItems {
//		var exists bool
//		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dishes WHERE id = $1)", dish.DishId).Scan(&exists)
//		if err != nil {
//			dr.logger.Error(fmt.Sprintf("Error verifying dish existence: %s", err.Error()))
//			return nil, fmt.Errorf("error verifying dish existence: %w", err)
//		}
//		if !exists {
//			dr.logger.Error(fmt.Sprintf("Dish with id %d does not exist", dish.DishId))
//			return nil, fmt.Errorf("dish with id %d does not exist", dish.DishId)
//		}
//
//		_, err = tx.Exec(ctx,
//			"INSERT INTO dish_delivery_items (delivery_id, dish_id, quantity) VALUES ($1, $2, $3)",
//			d.Id, dish.DishId, dish.Quantity)
//		if err != nil {
//			dr.logger.Error(fmt.Sprintf("Error inserting delivery dish: %s", err.Error()))
//			return nil, fmt.Errorf("error inserting delivery dish: %w", err)
//		}
//	}
//
//	if err := tx.Commit(ctx); err != nil {
//		dr.logger.Error("Error committing transaction: " + err.Error())
//		return nil, fmt.Errorf("error committing transaction: %w", err)
//	}
//
//	// Populate response
//	d.GuestId = req.GuestId
//	d.UserId = req.UserId
//	d.IsGuest = req.IsGuest
//	d.TableNumber = req.TableNumber
//	d.OrderHandlerId = req.OrderHandlerId
//	d.Status = req.Status
//	d.CreatedAt = timestamppb.New(createdAt)
//	d.UpdatedAt = timestamppb.New(updatedAt)
//	d.TotalPrice = req.TotalPrice
//	d.DishItems = req.DishItems
//	d.BowChili = req.BowChili
//	d.BowNoChili = req.BowNoChili
//	d.TakeAway = req.TakeAway
//	d.ChiliNumber = req.ChiliNumber
//	d.TableToken = req.TableToken
//	d.ClientName = req.ClientName
//	d.DeliveryAddress = req.DeliveryAddress
//	d.DeliveryContact = req.DeliveryContact
//	d.DeliveryNotes = req.DeliveryNotes
//	d.ScheduledTime = req.ScheduledTime
//	d.DeliveryFee = req.DeliveryFee
//	d.DeliveryStatus = req.DeliveryStatus
//	d.EstimatedDeliveryTime = req.EstimatedDeliveryTime
//	d.ActualDeliveryTime = req.ActualDeliveryTime
//
//	return &d, nil
//}
//
//func (dr *Repository) DeleteDeliveryById(ctx context.Context, id int64) error {
//	query := `DELETE FROM deliveries WHERE id = $1`
//
//	result, err := dr.db.Exec(ctx, query, id)
//	if err != nil {
//		dr.logger.Error(fmt.Sprintf("Error deleting delivery: %s", err.Error()))
//		return fmt.Errorf("error deleting delivery: %w", err)
//	}
//
//	rowsAffected := result.RowsAffected()
//	if rowsAffected == 0 {
//		return fmt.Errorf("no delivery found with id: %d", id)
//	}
//
//	return nil
//}
//
//func (dr *Repository) getDetailedDishItems(ctx context.Context, deliveryID int64) ([]*delivery.DeliveryDetailedDish, error) {
//	// Updated table name from delivery_items to dish_delivery_items
//	query := `
//				SELECT
//					di.dish_id,
//					di.quantity,
//					d.name,
//					d.price,
//					d.description,
//					d.image,
//					d.status
//				FROM dish_delivery_items di  -- Changed from delivery_items to dish_delivery_items
//				JOIN dishes d ON di.dish_id = d.id
//				WHERE di.delivery_id = $1
//			`
//
//	rows, err := dr.db.Query(ctx, query, deliveryID)
//	if err != nil {
//		return nil, fmt.Errorf("error querying dish items: %w", err)
//	}
//	defer rows.Close()
//
//	var items []*delivery.DeliveryDetailedDish
//
//	for rows.Next() {
//		var item delivery.DeliveryDetailedDish
//		err := rows.Scan(
//			&item.DishId,
//			&item.Quantity,
//			&item.Name,
//			&item.Price,
//			&item.Description,
//			&item.Image,
//			&item.Status,
//		)
//		if err != nil {
//			return nil, fmt.Errorf("error scanning dish item: %w", err)
//		}
//		items = append(items, &item)
//	}
//
//	if err = rows.Err(); err != nil {
//		return nil, fmt.Errorf("error iterating over dish items: %w", err)
//	}
//
//	return items, nil
//}
