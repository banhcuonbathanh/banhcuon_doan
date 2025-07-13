package order_grpc

//
//import (
//	"context"
//	"database/sql"
//	"fmt"
//	"math"
//
//	"time"
//
//	"github.com/jackc/pgx/v4/pgxpool"
//	"google.golang.org/protobuf/types/known/timestamppb"
//
//	"english-ai-full/internal/proto_qr/order"
//	"english-ai-full/logger"
//)
//
//type OrderRepository struct {
//	db     *pgxpool.Pool
//	logger *logger.Logger
//}
//
//func NewOrderRepository(db *pgxpool.Pool) *OrderRepository {
//	return &OrderRepository{
//		db:     db,
//		logger: logger.NewLogger(),
//	}
//}
//
//func (or *OrderRepository) PayOrders(ctx context.Context, req *order.PayOrdersRequest) ([]*order.Order, error) {
//	or.logger.Info("Processing payment for orders")
//
//	var userIDFilter, guestIDFilter interface{}
//	if req.Identifier != nil {
//		switch v := req.Identifier.(type) {
//		case *order.PayOrdersRequest_UserId:
//			userIDFilter = v.UserId
//		case *order.PayOrdersRequest_GuestId:
//			guestIDFilter = v.GuestId
//		}
//	}
//
//	tx, err := or.db.Begin(ctx)
//	if err != nil {
//		return nil, fmt.Errorf("error starting transaction: %w", err)
//	}
//	defer tx.Rollback(ctx)
//
//	query := `
//        UPDATE orders
//        SET status = 'paid', updated_at = $1
//        WHERE (user_id = $2 OR $2 IS NULL)
//        AND (guest_id = $3 OR $3 IS NULL)
//        AND status = 'pending'
//        RETURNING id
//    `
//
//	rows, err := tx.Query(ctx, query, time.Now(), userIDFilter, guestIDFilter)
//	if err != nil {
//		return nil, fmt.Errorf("error updating orders: %w", err)
//	}
//	defer rows.Close()
//
//	var orderIDs []int64
//	for rows.Next() {
//		var orderID int64
//		if err := rows.Scan(&orderID); err != nil {
//			return nil, fmt.Errorf("error scanning order ID: %w", err)
//		}
//		orderIDs = append(orderIDs, orderID)
//	}
//
//	if err := tx.Commit(ctx); err != nil {
//		return nil, fmt.Errorf("error committing transaction: %w", err)
//	}
//
//	// Fetch updated orders
//	var orders []*order.Order
//	for _, orderID := range orderIDs {
//		order, err := or.GetOrderDetail(ctx, orderID)
//		if err != nil {
//			return nil, err
//		}
//		orders = append(orders, order)
//	}
//
//	return orders, nil
//}
//
//func (or *OrderRepository) GetOrderDishItems(ctx context.Context, orderID int64) ([]*order.DishOrderItem, error) {
//	query := `
//        SELECT dish_id, quantity
//        FROM order_dishes
//        WHERE order_id = $1
//    `
//	rows, err := or.db.Query(ctx, query, orderID)
//	if err != nil {
//		return nil, fmt.Errorf("error fetching order dish items: %w", err)
//	}
//	defer rows.Close()
//
//	var items []*order.DishOrderItem
//	for rows.Next() {
//		item := &order.DishOrderItem{}
//		if err := rows.Scan(&item.DishId, &item.Quantity); err != nil {
//			or.logger.Error(fmt.Sprintf("Error scanning order dish item: %s", err.Error()))
//			return nil, fmt.Errorf("error scanning order dish item: %w", err)
//		}
//		items = append(items, item)
//	}
//
//	if err = rows.Err(); err != nil {
//		or.logger.Error(fmt.Sprintf("Error iterating order dish items: %s", err.Error()))
//		return nil, fmt.Errorf("error iterating order dish items: %w", err)
//	}
//
//	return items, nil
//}
//
//func (or *OrderRepository) GetOrderSetItems(ctx context.Context, orderID int64) ([]*order.SetOrderItem, error) {
//	query := `
//        SELECT set_id, quantity
//        FROM order_sets
//        WHERE order_id = $1
//    `
//	rows, err := or.db.Query(ctx, query, orderID)
//	if err != nil {
//		or.logger.Error(fmt.Sprintf("Error fetching order set items: %s", err.Error()))
//		return nil, fmt.Errorf("error fetching order set items: %w", err)
//	}
//	defer rows.Close()
//
//	var items []*order.SetOrderItem
//	for rows.Next() {
//		item := &order.SetOrderItem{}
//		if err := rows.Scan(&item.SetId, &item.Quantity); err != nil {
//			or.logger.Error(fmt.Sprintf("Error scanning order set item: %s", err.Error()))
//			return nil, fmt.Errorf("error scanning order set item: %w", err)
//		}
//		items = append(items, item)
//	}
//
//	if err = rows.Err(); err != nil {
//		or.logger.Error(fmt.Sprintf("Error iterating order set items: %s", err.Error()))
//		return nil, fmt.Errorf("error iterating order set items: %w", err)
//	}
//
//	return items, nil
//}
//
//// ----------------------------------
//
//func (or *OrderRepository) GetOrderProtoListDetail(ctx context.Context, page, pageSize int32) (*order.OrderDetailedListResponse, error) {
//	// or.logger.Info("Fetching detailed order list with pagination")
//
//	// Get total count for pagination
//	countQuery := `SELECT COUNT(*) FROM orders`
//
//	var totalItems int64
//	err := or.db.QueryRow(ctx, countQuery).Scan(&totalItems)
//	if err != nil {
//		or.logger.Error("Error counting orders: " + err.Error())
//		return nil, fmt.Errorf("error counting orders: %w", err)
//	}
//
//	// Calculate pagination info
//	totalPages := int32(math.Ceil(float64(totalItems) / float64(pageSize)))
//	offset := (page - 1) * pageSize
//
//	// Main order query
//	query := `
//        SELECT
//            o.id,
//            o.guest_id,
//            o.user_id,
//            o.is_guest,
//            o.table_number,
//            o.order_handler_id,
//            COALESCE(o.status, 'Pending') as status,
//            o.total_price,
//            COALESCE(o.topping, '') as topping,
//            COALESCE(o.tracking_order, '') as tracking_order,
//            COALESCE(o.take_away, false) as take_away,
//            COALESCE(o.chili_number, 0) as chili_number,
//              o.table_token,
//            COALESCE(o.order_name, '') as order_name
//        FROM orders o
//        ORDER BY o.created_at DESC
//        LIMIT $1 OFFSET $2
//    `
//
//	rows, err := or.db.Query(ctx, query, pageSize, offset)
//	if err != nil {
//		or.logger.Error("Error fetching orders: " + err.Error())
//		return nil, fmt.Errorf("error fetching orders: %w", err)
//	}
//	defer rows.Close()
//
//	var detailedOrders []*order.OrderDetailedResponse
//	for rows.Next() {
//		var o order.OrderDetailedResponse
//
//		// Create nullable variables for fields that can be NULL
//		var (
//			guestId        sql.NullInt64
//			userId         sql.NullInt64
//			tableNumber    sql.NullInt64
//			orderHandlerId sql.NullInt64
//			totalPrice     sql.NullInt32
//			status         sql.NullString
//			topping        sql.NullString
//			trackingOrder  sql.NullString
//			chiliNumber    sql.NullInt64
//			orderName      sql.NullString
//		)
//
//		err := rows.Scan(
//			&o.Id,
//			&guestId,
//			&userId,
//			&o.IsGuest,
//			&tableNumber,
//			&orderHandlerId,
//			&status,
//			&totalPrice,
//			&topping,
//			&trackingOrder,
//			&o.TakeAway,
//			&chiliNumber,
//			&o.TableToken,
//			&orderName,
//		)
//		if err != nil {
//			or.logger.Error("Error scanning order: " + err.Error())
//			return nil, fmt.Errorf("error scanning order: %w", err)
//		}
//
//		// Handle NULL values
//		if guestId.Valid {
//			o.GuestId = guestId.Int64
//		}
//		if userId.Valid {
//			o.UserId = userId.Int64
//		}
//		if tableNumber.Valid {
//			o.TableNumber = tableNumber.Int64
//		}
//		if orderHandlerId.Valid {
//			o.OrderHandlerId = orderHandlerId.Int64
//		}
//		if totalPrice.Valid {
//			o.TotalPrice = totalPrice.Int32
//		}
//		if status.Valid {
//			o.Status = status.String
//		}
//		if topping.Valid {
//			o.Topping = topping.String
//		}
//		if trackingOrder.Valid {
//			o.TrackingOrder = trackingOrder.String
//		}
//		if chiliNumber.Valid {
//			o.ChiliNumber = chiliNumber.Int64
//		}
//		if orderName.Valid {
//			o.OrderName = orderName.String
//		}
//		// Fetch detailed dish items
//		dishQuery := `
//            SELECT
//                d.id,
//                doi.quantity,
//                d.name,
//                d.price,
//                d.description,
//                d.image,
//                d.status
//            FROM dish_order_items doi
//            JOIN dishes d ON doi.dish_id = d.id
//            WHERE doi.order_id = $1
//        `
//		dishRows, err := or.db.Query(ctx, dishQuery, o.Id)
//		if err != nil {
//			or.logger.Error("Error fetching dish details: " + err.Error())
//			return nil, fmt.Errorf("error fetching dish details: %w", err)
//		}
//		defer dishRows.Close()
//
//		var dishItems []*order.OrderDetailedDish
//		for dishRows.Next() {
//			var dish order.OrderDetailedDish
//			err := dishRows.Scan(
//				&dish.DishId,
//				&dish.Quantity,
//				&dish.Name,
//				&dish.Price,
//				&dish.Description,
//				&dish.Image,
//				&dish.Status,
//			)
//			if err != nil {
//				or.logger.Error("Error scanning dish detail: " + err.Error())
//				return nil, fmt.Errorf("error scanning dish detail: %w", err)
//			}
//			dishItems = append(dishItems, &dish)
//		}
//		o.DataDish = dishItems
//
//		// Fetch detailed set items
//		setQuery := `
//            SELECT
//                s.id,
//                s.name,
//                s.description,
//                s.user_id,
//                s.is_favourite,
//                s.is_public,
//                s.image,
//                s.price,
//                soi.quantity,
//                s.created_at,
//                s.updated_at
//            FROM set_order_items soi
//            JOIN sets s ON soi.set_id = s.id
//            WHERE soi.order_id = $1
//        `
//		setRows, err := or.db.Query(ctx, setQuery, o.Id)
//		if err != nil {
//			or.logger.Error("Error fetching set details: " + err.Error())
//			return nil, fmt.Errorf("error fetching set details: %w", err)
//		}
//		defer setRows.Close()
//
//		var setItems []*order.OrderSetDetailed
//		for setRows.Next() {
//			var set order.OrderSetDetailed
//			var createdAt, updatedAt time.Time
//			var userID sql.NullInt32
//
//			err := setRows.Scan(
//				&set.Id,
//				&set.Name,
//				&set.Description,
//				&userID,
//				&set.IsFavourite,
//				&set.IsPublic,
//				&set.Image,
//				&set.Price,
//				&set.Quantity,
//				&createdAt,
//				&updatedAt,
//			)
//			if err != nil {
//				or.logger.Error("Error scanning set detail: " + err.Error())
//				return nil, fmt.Errorf("error scanning set detail: %w", err)
//			}
//
//			if userID.Valid {
//				set.UserId = userID.Int32
//			}
//
//			set.CreatedAt = timestamppb.New(createdAt)
//			set.UpdatedAt = timestamppb.New(updatedAt)
//
//			// Fetch dishes for this set
//			setDishQuery := `
//                SELECT
//                    d.id,
//                    sd.quantity,
//                    d.name,
//                    d.price,
//                    d.description,
//                    d.image,
//                    d.status
//                FROM set_dishes sd
//                JOIN dishes d ON sd.dish_id = d.id
//                WHERE sd.set_id = $1
//            `
//			setDishRows, err := or.db.Query(ctx, setDishQuery, set.Id)
//			if err != nil {
//				or.logger.Error("Error fetching set dish details: " + err.Error())
//				return nil, fmt.Errorf("error fetching set dish details: %w", err)
//			}
//			defer setDishRows.Close()
//
//			var setDishes []*order.OrderDetailedDish
//			for setDishRows.Next() {
//				var dish order.OrderDetailedDish
//				err := setDishRows.Scan(
//					&dish.DishId,
//					&dish.Quantity,
//					&dish.Name,
//					&dish.Price,
//					&dish.Description,
//					&dish.Image,
//					&dish.Status,
//				)
//				if err != nil {
//					or.logger.Error("Error scanning set dish detail: " + err.Error())
//					return nil, fmt.Errorf("error scanning set dish detail: %w", err)
//				}
//				setDishes = append(setDishes, &dish)
//			}
//			set.Dishes = setDishes
//			setItems = append(setItems, &set)
//		}
//		o.DataSet = setItems
//
//		detailedOrders = append(detailedOrders, &o)
//	}
//
//	response := &order.OrderDetailedListResponse{
//		Data: detailedOrders,
//		Pagination: &order.PaginationInfo{
//			CurrentPage: page,
//			TotalPages:  totalPages,
//			TotalItems:  totalItems,
//			PageSize:    pageSize,
//		},
//	}
//
//	return response, nil
//}
//
//// ---------------------------
//
//func (or *OrderRepository) UpdateOrder(ctx context.Context, req *order.UpdateOrderRequest) (*order.Order, error) {
//	or.logger.Info(fmt.Sprintf("Updating order with ID: %d", req.Id))
//
//	tx, err := or.db.Begin(ctx)
//	if err != nil {
//		or.logger.Error("Error starting transaction: " + err.Error())
//		return nil, fmt.Errorf("error starting transaction: %w", err)
//	}
//	defer tx.Rollback(ctx)
//
//	query := `
//        UPDATE orders
//        SET guest_id = $2, user_id = $3, table_number = $4, order_handler_id = $5,
//            status = $6, updated_at = $7, total_price = $8, is_guest = $9,
//            topping = $10, tracking_order = $11, take_away = $12,
//            chili_number = $13, table_token = $14, order_name = $15
//        WHERE id = $1
//        RETURNING created_at, updated_at
//    `
//
//	var o order.Order
//	var createdAt, updatedAt time.Time
//
//	err = tx.QueryRow(ctx, query,
//		req.Id,
//		req.GuestId,
//		req.UserId,
//		req.TableNumber,
//		req.OrderHandlerId,
//		req.Status,
//		time.Now(),
//		req.TotalPrice,
//		req.IsGuest,
//		req.Topping,
//		req.TrackingOrder,
//		req.TakeAway,
//		req.ChiliNumber,
//		req.TableToken,
//		req.OrderName,
//	).Scan(&createdAt, &updatedAt)
//
//	if err != nil {
//		or.logger.Error(fmt.Sprintf("Error updating order: %s", err.Error()))
//		return nil, fmt.Errorf("error updating order: %w", err)
//	}
//
//	if err := tx.Commit(ctx); err != nil {
//		or.logger.Error("Error committing transaction: " + err.Error())
//		return nil, fmt.Errorf("error committing transaction: %w", err)
//	}
//
//	// Populate response
//	o.Id = req.Id
//	o.GuestId = req.GuestId
//	o.UserId = req.UserId
//	o.IsGuest = req.IsGuest
//	o.TableNumber = req.TableNumber
//	o.OrderHandlerId = req.OrderHandlerId
//	o.Status = req.Status
//	o.CreatedAt = timestamppb.New(createdAt)
//	o.UpdatedAt = timestamppb.New(updatedAt)
//	o.TotalPrice = req.TotalPrice
//	o.DishItems = req.DishItems
//	o.SetItems = req.SetItems
//	o.Topping = req.Topping
//	o.TrackingOrder = req.TrackingOrder
//	o.TakeAway = req.TakeAway
//	o.ChiliNumber = req.ChiliNumber
//	o.TableToken = req.TableToken
//	o.OrderName = req.OrderName
//
//	return &o, nil
//}
//
//func (or *OrderRepository) GetOrders(ctx context.Context, page, pageSize int32) ([]*order.Order, int64, error) {
//	or.logger.Info("Fetching orders with pagination")
//
//	// Get total count for pagination
//	countQuery := `SELECT COUNT(*) FROM orders`
//
//	var totalItems int64
//	err := or.db.QueryRow(ctx, countQuery).Scan(&totalItems)
//	if err != nil {
//		or.logger.Error("Error counting orders: " + err.Error())
//		return nil, 0, fmt.Errorf("error counting orders: %w", err)
//	}
//
//	// Calculate offset
//	offset := (page - 1) * pageSize
//
//	// Main order query
//	query := `
//        SELECT
//            o.id,
//            o.guest_id,
//            o.user_id,
//            o.is_guest,
//            o.table_number,
//            o.order_handler_id,
//            COALESCE(o.status, 'Pending') as status,
//            o.created_at,
//            o.updated_at,
//            o.total_price,
//            COALESCE(o.topping, '') as topping,
//            COALESCE(o.tracking_order, '') as tracking_order,
//            COALESCE(o.take_away, false) as take_away,
//            COALESCE(o.chili_number, 0) as chili_number,
//            o.table_token,
//            COALESCE(o.order_name, '') as order_name
//        FROM orders o
//        ORDER BY o.created_at DESC
//        LIMIT $1 OFFSET $2
//    `
//
//	rows, err := or.db.Query(ctx, query, pageSize, offset)
//	if err != nil {
//		or.logger.Error("Error fetching orders: " + err.Error())
//		return nil, 0, fmt.Errorf("error fetching orders: %w", err)
//	}
//	defer rows.Close()
//
//	var orders []*order.Order
//	for rows.Next() {
//		var o order.Order
//		var createdAt, updatedAt time.Time
//
//		// Create nullable variables for fields that can be NULL in the database
//		var (
//			guestId        sql.NullInt64
//			userId         sql.NullInt64
//			tableNumber    sql.NullInt64
//			orderHandlerId sql.NullInt64
//			totalPrice     sql.NullInt32
//			status         sql.NullString
//			topping        sql.NullString
//			trackingOrder  sql.NullString
//			chiliNumber    sql.NullInt64
//			orderName      sql.NullString
//		)
//		err = rows.Scan(
//			&o.Id,
//			&guestId,
//			&userId,
//			&o.IsGuest,
//			&tableNumber,
//			&orderHandlerId,
//			&status,
//			&createdAt,
//			&updatedAt,
//			&totalPrice,
//			&topping,
//			&trackingOrder,
//			&o.TakeAway,
//			&chiliNumber,
//			&o.TableToken,
//			&orderName,
//		)
//		if err != nil {
//			or.logger.Error("Error scanning order: " + err.Error())
//			return nil, 0, fmt.Errorf("error scanning order: %w", err)
//		}
//
//		// Convert nullable fields
//		o.GuestId = guestId.Int64
//		o.UserId = userId.Int64
//		if tableNumber.Valid {
//			o.TableNumber = tableNumber.Int64
//		}
//		o.OrderHandlerId = orderHandlerId.Int64
//		o.Status = status.String
//		o.TotalPrice = totalPrice.Int32
//		o.Topping = topping.String
//		o.TrackingOrder = trackingOrder.String
//		o.ChiliNumber = chiliNumber.Int64
//		if orderName.Valid {
//			o.OrderName = orderName.String
//		}
//
//		// Handle timestamps
//		o.CreatedAt = timestamppb.New(createdAt)
//		o.UpdatedAt = timestamppb.New(updatedAt)
//
//		orders = append(orders, &o)
//	}
//
//	return orders, totalItems, nil
//}
//
////--------------
//
//func (or *OrderRepository) GetOrderDetail(ctx context.Context, id int64) (*order.Order, error) {
//	// or.logger.Info(fmt.Sprintf("Fetching order detail for ID: %d", id))
//
//	query := `
//        SELECT
//            id, guest_id, user_id, is_guest, table_number, order_handler_id,
//            status, created_at, updated_at, total_price, topping, tracking_order,
//            take_away, chili_number, table_token, order_name
//        FROM orders
//        WHERE id = $1
//    `
//
//	var o order.Order
//	var createdAt, updatedAt time.Time
//	var orderName sql.NullString
//
//	err := or.db.QueryRow(ctx, query, id).Scan(
//		&o.Id,
//		&o.GuestId,
//		&o.UserId,
//		&o.IsGuest,
//		&o.TableNumber,
//		&o.OrderHandlerId,
//		&o.Status,
//		&createdAt,
//		&updatedAt,
//		&o.TotalPrice,
//		&o.Topping,
//		&o.TrackingOrder,
//		&o.TakeAway,
//		&o.ChiliNumber,
//		&o.TableToken,
//		&orderName,
//	)
//	if err != nil {
//		or.logger.Error(fmt.Sprintf("Error fetching order detail: %s", err.Error()))
//		return nil, fmt.Errorf("error fetching order detail: %w", err)
//	}
//
//	if orderName.Valid {
//		o.OrderName = orderName.String
//	}
//
//	o.CreatedAt = timestamppb.New(createdAt)
//	o.UpdatedAt = timestamppb.New(updatedAt)
//
//	return &o, nil
//}
//
//func (or *OrderRepository) getTrackingOrderInfo(ctx context.Context, tx *pgxpool.Pool, currentTime time.Time, isGuest bool, clientId int64) (string, error) {
//	startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
//	endOfDay := startOfDay.Add(24 * time.Hour)
//
//	or.logger.Info(fmt.Sprintf("golang/internal/order/order_repository.go getTrackingOrderInfo - Processing for clientId: %d, isGuest: %v, date range: %v to %v",
//		clientId, isGuest, startOfDay, endOfDay))
//
//	clientPosition, err := or.getClientPosition(ctx, tx, startOfDay, endOfDay, clientId, isGuest)
//	if err != nil {
//		return "", fmt.Errorf("error getting client position: %w", err)
//	}
//
//	orderCount, err := or.getClientOrderCount(ctx, tx, startOfDay, endOfDay, clientId, isGuest)
//	if err != nil {
//		return "", fmt.Errorf("error getting client order count: %w", err)
//	}
//
//	var clientType string
//	if isGuest {
//		clientType = "Guest"
//	} else {
//		clientType = "Client"
//	}
//
//	trackingOrder := fmt.Sprintf("%d%s %s - Order #%d",
//		clientPosition,
//		getOrdinalSuffix(clientPosition),
//		clientType,
//		orderCount)
//
//	return trackingOrder, nil
//}
//
//func getOrdinalSuffix(n int64) string {
//	if n%100 >= 11 && n%100 <= 13 {
//		return "th"
//	}
//	switch n % 10 {
//	case 1:
//		return "st"
//	case 2:
//		return "nd"
//	case 3:
//		return "rd"
//	default:
//		return "th"
//	}
//}
//
//func (or *OrderRepository) CreateOrder(ctx context.Context, req *order.CreateOrderRequest) (*order.Order, error) {
//	or.logger.Info(fmt.Sprintf("Creating new order: %+v", req))
//
//	tx, err := or.db.Begin(ctx)
//	if err != nil {
//		or.logger.Error("Error starting transaction: " + err.Error())
//		return nil, fmt.Errorf("error starting transaction: %w", err)
//	}
//	defer tx.Rollback(ctx)
//
//	now := time.Now()
//
//	// Get client ID based on whether it's a guest or user
//	clientId := req.GuestId
//	if !req.IsGuest {
//		clientId = req.UserId
//	}
//	// Get tracking order information
//	trackingOrder, err := or.getTrackingOrderInfo(ctx, or.db, now, req.IsGuest, clientId)
//	if err != nil {
//		or.logger.Error("Error getting tracking order info: " + err.Error())
//		return nil, fmt.Errorf("error getting tracking order info: %w", err)
//	}
//
//	// Format the order name
//
//	// Get client number for the order name
//	clientNumber, err := or.getClientNumberForDay(ctx, or.db, now, req.IsGuest, clientId)
//	if err != nil {
//		return nil, fmt.Errorf("error getting client number: %w", err)
//	}
//
//	var guestId, userId sql.NullInt64
//	if req.IsGuest {
//		guestId = sql.NullInt64{Int64: req.GuestId, Valid: true}
//		userId = sql.NullInt64{Valid: false}
//	} else {
//		userId = sql.NullInt64{Int64: req.UserId, Valid: true}
//		guestId = sql.NullInt64{Valid: false}
//	}
//
//	query := `
//        INSERT INTO orders (
//            guest_id, user_id, is_guest, table_number, order_handler_id,
//            status, created_at, updated_at, total_price, topping, tracking_order,
//            take_away, chili_number, table_token, order_name
//        )
//        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
//        RETURNING id, created_at, updated_at
//    `
//	or.logger.Info("golang/quanqr/order/order_repository.go CreateOrder - Executing order insertion")
//	var o order.Order
//	var createdAt, updatedAt time.Time
//
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
//		req.Topping,
//		trackingOrder, // Use the generated tracking order
//		req.TakeAway,
//		req.ChiliNumber,
//		req.TableToken,
//		req.OrderName,
//	).Scan(&o.Id, &createdAt, &updatedAt)
//
//	if err != nil {
//		or.logger.Error("Error creating order: " + err.Error())
//		return nil, fmt.Errorf("error creating order: %w", err)
//	}
//
//	// Insert dish items
//	for _, dish := range req.DishItems {
//		var exists bool
//		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM dishes WHERE id = $1)", dish.DishId).Scan(&exists)
//		if err != nil {
//			or.logger.Error(fmt.Sprintf("Error verifying dish existence: %s", err.Error()))
//			return nil, fmt.Errorf("error verifying dish existence: %w", err)
//		}
//		if !exists {
//			or.logger.Error(fmt.Sprintf("Dish with id %d does not exist", dish.DishId))
//			return nil, fmt.Errorf("dish with id %d does not exist", dish.DishId)
//		}
//
//		_, err = tx.Exec(ctx,
//			"INSERT INTO dish_order_items (order_id, dish_id, quantity) VALUES ($1, $2, $3)",
//			o.Id, dish.DishId, dish.Quantity)
//		if err != nil {
//			or.logger.Error(fmt.Sprintf("Error inserting order dish: %s", err.Error()))
//			return nil, fmt.Errorf("error inserting order dish: %w", err)
//		}
//	}
//
//	// Insert set items
//	for _, set := range req.SetItems {
//		var exists bool
//		err := tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM sets WHERE id = $1)", set.SetId).Scan(&exists)
//		if err != nil {
//			or.logger.Error(fmt.Sprintf("Error verifying set existence: %s", err.Error()))
//			return nil, fmt.Errorf("error verifying set existence: %w", err)
//		}
//		if !exists {
//			or.logger.Error(fmt.Sprintf("Set with id %d does not exist", set.SetId))
//			return nil, fmt.Errorf("set with id %d does not exist", set.SetId)
//		}
//
//		_, err = tx.Exec(ctx,
//			"INSERT INTO set_order_items (order_id, set_id, quantity) VALUES ($1, $2, $3)",
//			o.Id, set.SetId, set.Quantity)
//		if err != nil {
//			or.logger.Error(fmt.Sprintf("Error inserting order set: %s", err.Error()))
//			return nil, fmt.Errorf("error inserting order set: %w", err)
//		}
//	}
//
//	if err := tx.Commit(ctx); err != nil {
//		or.logger.Error("Error committing transaction: " + err.Error())
//		return nil, fmt.Errorf("error committing transaction: %w", err)
//	}
//
//	// Populate response
//	o.GuestId = req.GuestId
//	o.UserId = req.UserId
//	o.IsGuest = req.IsGuest
//	o.TableNumber = req.TableNumber
//	o.OrderHandlerId = req.OrderHandlerId
//	o.Status = req.Status
//	o.CreatedAt = timestamppb.New(createdAt)
//	o.UpdatedAt = timestamppb.New(updatedAt)
//	o.TotalPrice = req.TotalPrice
//	o.DishItems = req.DishItems
//	o.SetItems = req.SetItems
//	o.Topping = req.Topping
//	o.TrackingOrder = trackingOrder
//	o.TakeAway = req.TakeAway
//	o.ChiliNumber = req.ChiliNumber
//	o.TableToken = req.TableToken
//	o.OrderName = req.OrderName
//
//	return &o, nil
//}
//
//func (or *OrderRepository) getClientOrderCount(ctx context.Context, tx *pgxpool.Pool, startOfDay, endOfDay time.Time, clientId int64, isGuest bool) (int64, error) {
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientOrderCount - Starting for clientId: %d, isGuest: %v", clientId, isGuest))
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientOrderCount - Date range: %v to %v", startOfDay, endOfDay))
//	var query string
//	if isGuest {
//		query = `
//            SELECT COUNT(*) + 1
//            FROM orders
//            WHERE guest_id = $1
//            AND created_at >= $2 AND created_at < $3
//        `
//	} else {
//		query = `
//            SELECT COUNT(*) + 1
//            FROM orders
//            WHERE user_id = $1
//            AND created_at >= $2 AND created_at < $3
//        `
//	}
//
//	var count int64
//	err := tx.QueryRow(ctx, query, clientId, startOfDay, endOfDay).Scan(&count)
//	if err != nil {
//		if err.Error() == "no rows in result set" {
//			return 1, nil
//		}
//		return 0, fmt.Errorf("error getting order count: %w", err)
//	}
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientOrderCount - Retrieved count: %d", count))
//	return count, nil
//}
//
////
//
//func (or *OrderRepository) getClientPosition(ctx context.Context, tx *pgxpool.Pool, startOfDay, endOfDay time.Time, clientId int64, isGuest bool) (int64, error) {
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientPosition - Starting for clientId: %d, isGuest: %v", clientId, isGuest))
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientPosition - Date range: %v to %v", startOfDay, endOfDay))
//
//	// First, let's debug the virtual table
//	debugQuery := `
//        WITH ordered_clients AS (
//            SELECT
//                CASE
//                    WHEN guest_id IS NOT NULL THEN guest_id
//                    ELSE user_id
//                END as client_id,
//                MIN(created_at) as first_order_time,
//                DENSE_RANK() OVER (ORDER BY MIN(created_at)) as position
//            FROM orders
//            WHERE created_at >= $1 AND created_at < $2
//            GROUP BY
//                CASE
//                    WHEN guest_id IS NOT NULL THEN guest_id
//                    ELSE user_id
//                END
//        )
//        SELECT client_id, first_order_time, position
//        FROM ordered_clients
//        ORDER BY position;
//    `
//
//	// Execute debug query to see the virtual table
//	rows, err := tx.Query(ctx, debugQuery, startOfDay, endOfDay)
//	if err != nil {
//		or.logger.Error(fmt.Sprintf("Debug query error: %v", err))
//	} else {
//		defer rows.Close()
//		or.logger.Info("Virtual Table Content:")
//		or.logger.Info("| Client ID | First Order Time | Position |")
//		or.logger.Info("|-----------|-----------------|-----------|")
//
//		for rows.Next() {
//			var cid int64
//			var orderTime time.Time
//			var pos int64
//			if err := rows.Scan(&cid, &orderTime, &pos); err != nil {
//				or.logger.Error(fmt.Sprintf("Error scanning debug row: %v", err))
//				continue
//			}
//			or.logger.Info(fmt.Sprintf("| %9d | %s | %9d |", cid, orderTime.Format("15:04:05"), pos))
//		}
//	}
//
//	// Now execute the actual position query
//	query := `
//        WITH ordered_clients AS (
//            SELECT
//                CASE
//                    WHEN guest_id IS NOT NULL THEN guest_id
//                    ELSE user_id
//                END as client_id,
//                MIN(created_at) as first_order_time,
//                DENSE_RANK() OVER (ORDER BY MIN(created_at)) as position
//            FROM orders
//            WHERE created_at >= $1 AND created_at < $2
//            GROUP BY
//                CASE
//                    WHEN guest_id IS NOT NULL THEN guest_id
//                    ELSE user_id
//                END
//        )
//        SELECT position
//        FROM ordered_clients
//        WHERE client_id = $3
//        UNION ALL
//        SELECT COALESCE(
//            (SELECT MAX(position) + 1
//            FROM ordered_clients),
//            1
//        )
//        WHERE NOT EXISTS (
//            SELECT 1
//            FROM ordered_clients
//            WHERE client_id = $3
//        )
//        LIMIT 1
//    `
//
//	var position int64
//	err = tx.QueryRow(ctx, query, startOfDay, endOfDay, clientId).Scan(&position)
//	if err != nil {
//		return 0, fmt.Errorf("error getting client position: %w", err)
//	}
//
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientPosition - Retrieved position for client %d: %d", clientId, position))
//	return position, nil
//}
//
//func (or *OrderRepository) getClientNumberForDay(ctx context.Context, tx *pgxpool.Pool, currentTime time.Time, isGuest bool, clientId int64) (int64, error) {
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientNumberForDay - Starting for clientId: %d, isGuest: %v", clientId, isGuest))
//
//	startOfDay := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), 0, 0, 0, 0, time.UTC)
//	endOfDay := startOfDay.Add(24 * time.Hour)
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientNumberForDay - Date range: %v to %v", startOfDay, endOfDay))
//
//	var query string
//	var args []interface{}
//
//	if isGuest {
//		query = `
//            SELECT COALESCE(
//                (SELECT COUNT(*)
//                FROM orders
//                WHERE guest_id = $1
//                AND created_at >= $2
//                AND created_at < $3),
//                0
//            ) + 1
//        `
//		args = []interface{}{clientId, startOfDay, endOfDay}
//	} else {
//		query = `
//            SELECT COALESCE(
//                (SELECT COUNT(*)
//                FROM orders
//                WHERE user_id = $1
//                AND created_at >= $2
//                AND created_at < $3),
//                0
//            ) + 1
//        `
//		args = []interface{}{clientId, startOfDay, endOfDay}
//	}
//
//	var clientNumber int64
//	err := tx.QueryRow(ctx, query, args...).Scan(&clientNumber)
//	if err != nil {
//		if err.Error() == "no rows in result set" {
//			or.logger.Info("golang/quanqr/order/order_repository.go getClientNumberForDay - No existing client number found, returning 1")
//			return 1, nil
//		}
//		or.logger.Error(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientNumberForDay - Error getting client number: %v", err))
//		return 0, fmt.Errorf("error getting client number: %w", err)
//	}
//	or.logger.Info(fmt.Sprintf("golang/quanqr/order/order_repository.go getClientNumberForDay - clientNumber %v ", clientNumber))
//	return clientNumber, nil
//}
