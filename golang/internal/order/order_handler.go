package order_grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"time"

	"english-ai-full/internal/proto_qr/order"
	"english-ai-full/logger"
	"english-ai-full/token"

	"github.com/go-chi/chi"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type OrderHandlerController struct {
	ctx        context.Context
	client     order.OrderServiceClient
	TokenMaker *token.JWTMaker
	logger     *logger.Logger
}

func NewOrderHandler(client order.OrderServiceClient, secretKey string) *OrderHandlerController {
	return &OrderHandlerController{
		ctx:        context.Background(),
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
		logger:     logger.NewLogger(),
	}
}

func (h *OrderHandlerController) CreateOrder(w http.ResponseWriter, r *http.Request) {
	var orderReq CreateOrderRequestType

	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	pbReq := ToPBCreateOrderRequest(orderReq)
	createdOrderResponse, err := h.client.CreateOrder(h.ctx, pbReq)
	if err != nil {
		h.logger.Error("Error creating order: " + err.Error())
		http.Error(w, "error creating order", http.StatusInternalServerError)
		return
	}

	res := ToOrderResFromPbOrderResponse(createdOrderResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *OrderHandlerController) GetOrderDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Fetching order detail for ID: %d", i))
	orderResponse, err := h.client.GetOrderDetail(h.ctx, &order.OrderIdParam{Id: i})
	if err != nil {
		h.logger.Error("Error fetching order detail: " + err.Error())
		http.Error(w, "error getting order", http.StatusInternalServerError)
		return
	}

	res := ToOrderResFromPbOrderResponse(orderResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *OrderHandlerController) GetOrders(w http.ResponseWriter, r *http.Request) {
	// Only accept GET requests
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse query parameters
	query := r.URL.Query()

	// Get page parameter with default value 1
	page := int32(1)
	if pageStr := query.Get("page"); pageStr != "" {
		if pageInt, err := strconv.ParseInt(pageStr, 10, 32); err == nil {
			page = int32(pageInt)
		}
	}

	// Get page_size parameter with default value 10
	pageSize := int32(10)
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
			pageSize = int32(pageSizeInt)
		}
	}

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}

	h.logger.Info("Fetching orders list")
	ordersResponse, err := h.client.GetOrders(h.ctx, &order.GetOrdersRequest{
		Page:     page,
		PageSize: pageSize,
	})

	fmt.Printf("golang/internal/order/order_handler.go ordersResponse %v\n", ordersResponse)
	if err != nil {
		h.logger.Error("Error fetching orders list: " + err.Error())
		http.Error(w, "failed to fetch orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert protobuf response to HTTP response
	res := ToOrderListResFromPbOrderListResponse(ordersResponse)

	// Set response headers and encode response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error("Error encoding response: " + err.Error())
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}
func (h *OrderHandlerController) UpdateOrder(w http.ResponseWriter, r *http.Request) {
	var orderReq UpdateOrderRequestType
	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Updating order: %d", orderReq.ID))
	updatedOrderResponse, err := h.client.UpdateOrder(h.ctx, ToPBUpdateOrderRequest(orderReq))
	if err != nil {
		h.logger.Error("Error updating order: " + err.Error())
		http.Error(w, "error updating order", http.StatusInternalServerError)
		return
	}

	res := ToOrderResFromPbOrderResponse(updatedOrderResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *OrderHandlerController) PayOrders(w http.ResponseWriter, r *http.Request) {
	var req PayOrdersRequestType
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	h.logger.Info("Processing order payment")
	paymentResponse, err := h.client.PayOrders(h.ctx, ToPBPayOrdersRequest(req))
	if err != nil {
		h.logger.Error("Error processing payment: " + err.Error())
		http.Error(w, "error processing payment", http.StatusInternalServerError)
		return
	}

	res := ToOrderListResFromPbOrderListResponse(paymentResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *OrderHandlerController) GetOrderProtoListDetail(w http.ResponseWriter, r *http.Request) {
	h.logger.Info("Fetching detailed order list")

	// Parse query parameters for pagination
	page, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || page < 1 {
		page = 1 // Default to first page if invalid
	}

	pageSize, err := strconv.Atoi(r.URL.Query().Get("page_size"))
	if err != nil || pageSize < 1 {
		pageSize = 10 // Default page size if invalid
	}

	// Create the request with pagination parameters
	req := &order.GetOrdersRequest{
		Page:     int32(page),
		PageSize: int32(pageSize),
	}

	// Call the service
	ordersResponse, err := h.client.GetOrderProtoListDetail(h.ctx, req)
	if err != nil {
		h.logger.Error("Error fetching detailed order list: " + err.Error())
		http.Error(w, "failed to fetch detailed orders: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert the response
	res := ToOrderDetailedListResponseFromProto(ordersResponse)
	// Send response
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		h.logger.Error("Error encoding response: " + err.Error())
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}
}

func ToPBPayOrdersRequest(req PayOrdersRequestType) *order.PayOrdersRequest {
	pbReq := &order.PayOrdersRequest{}
	if req.GuestID != nil {
		pbReq.Identifier = &order.PayOrdersRequest_GuestId{GuestId: *req.GuestID}
	} else if req.UserID != nil {
		pbReq.Identifier = &order.PayOrdersRequest_UserId{UserId: *req.UserID}
	}
	return pbReq
}

func ToPBDishOrderItems(items []OrderDish) []*order.DishOrderItem {
	pbItems := make([]*order.DishOrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &order.DishOrderItem{
			DishId:   item.DishID,
			Quantity: item.Quantity,
		}
	}
	return pbItems
}

func ToPBSetOrderItems(items []OrderSet) []*order.SetOrderItem {
	pbItems := make([]*order.SetOrderItem, len(items))
	for i, item := range items {
		pbItems[i] = &order.SetOrderItem{
			SetId:    item.SetID,
			Quantity: item.Quantity,
		}
	}
	return pbItems
}

func ToOrderResFromPbOrderResponse(pbRes *order.OrderResponse) OrderResponse {
	return OrderResponse{
		Data: ToOrderFromPbOrder(pbRes.Data),
	}
}

func ToOrderListResFromPbOrderListResponse(pbRes *order.OrderListResponse) *OrderListResponse {
	if pbRes == nil {
		return nil
	}

	// Initialize response with proper capacity
	orders := make([]OrderType, 0, len(pbRes.Data))

	// Convert each order
	for _, pbOrder := range pbRes.Data {
		if pbOrder != nil {
			orders = append(orders, ToOrderFromPbOrder(pbOrder))
		}
	}

	return &OrderListResponse{
		Data: orders,
		Pagination: PaginationInfo{
			CurrentPage: pbRes.GetPagination().GetCurrentPage(),
			TotalPages:  pbRes.GetPagination().GetTotalPages(),
			TotalItems:  pbRes.GetPagination().GetTotalItems(),
			PageSize:    pbRes.GetPagination().GetPageSize(),
		},
	}
}

func ToOrderDishesFromPbDishOrderItems(pbItems []*order.DishOrderItem) []OrderDish {
	if pbItems == nil {
		return nil
	}

	items := make([]OrderDish, 0, len(pbItems))
	for _, pbItem := range pbItems {
		if pbItem != nil {
			items = append(items, OrderDish{
				DishID:   pbItem.GetDishId(),
				Quantity: pbItem.GetQuantity(),
			})
		}
	}
	return items
}

func ToOrderSetsFromPbSetOrderItems(pbItems []*order.SetOrderItem) []OrderSet {
	if pbItems == nil {
		return nil
	}

	items := make([]OrderSet, 0, len(pbItems))
	for _, pbItem := range pbItems {
		if pbItem != nil {
			items = append(items, OrderSet{
				SetID:    pbItem.GetSetId(),
				Quantity: pbItem.GetQuantity(),
			})
		}
	}
	return items
}

func ToOrderDetailedDishesFromPbOrderDetailedDishes(pbDishes []*order.OrderDetailedDish) []OrderDetailedDish {
	dishes := make([]OrderDetailedDish, len(pbDishes))
	for i, pbDish := range pbDishes {
		dishes[i] = OrderDetailedDish{
			DishID:      pbDish.DishId,
			Quantity:    pbDish.Quantity,
			Name:        pbDish.Name,
			Price:       pbDish.Price,
			Description: pbDish.Description,
			Image:       pbDish.Image,
			Status:      pbDish.Status,
		}
	}
	return dishes
}

func ToOrderDetailedDishFromPbOrderDetailedDish(pbDish *order.OrderDetailedDish) OrderDetailedDish {
	if pbDish == nil {
		return OrderDetailedDish{}
	}

	return OrderDetailedDish{
		DishID:      pbDish.DishId,
		Quantity:    pbDish.Quantity,
		Name:        pbDish.Name,
		Price:       pbDish.Price,
		Description: pbDish.Description,
		Image:       pbDish.Image,
		Status:      pbDish.Status,
	}
}

func ToOrderSetDetailedFromPbOrderSetDetailed(pbSet *order.OrderSetDetailed) OrderSetDetailed {
	if pbSet == nil {
		return OrderSetDetailed{}
	}

	dishes := make([]OrderDetailedDish, len(pbSet.Dishes))
	for i, pbDish := range pbSet.Dishes {
		dishes[i] = ToOrderDetailedDishFromPbOrderDetailedDish(pbDish)
	}

	var createdAt, updatedAt time.Time
	if pbSet.CreatedAt != nil {
		createdAt = pbSet.CreatedAt.AsTime()
	}
	if pbSet.UpdatedAt != nil {
		updatedAt = pbSet.UpdatedAt.AsTime()
	}

	return OrderSetDetailed{
		ID:          pbSet.Id,
		Name:        pbSet.Name,
		Description: pbSet.Description,
		Dishes:      dishes,
		UserID:      pbSet.UserId,
		CreatedAt:   createdAt,
		UpdatedAt:   updatedAt,
		IsFavourite: pbSet.IsFavourite,
		LikeBy:      pbSet.LikeBy,
		IsPublic:    pbSet.IsPublic,
		Image:       pbSet.Image,
		Price:       pbSet.Price,
		Quantity:    pbSet.Quantity,
	}
}

// -------------

// Helper functions for conversion
func ToOrderSetsDetailedFromProto(pbSets []*order.OrderSetDetailed) []OrderSetDetailed {
	if pbSets == nil {
		return nil
	}

	sets := make([]OrderSetDetailed, len(pbSets))
	for i, pbSet := range pbSets {
		sets[i] = OrderSetDetailed{
			ID:          pbSet.Id,
			Name:        pbSet.Name,
			Description: pbSet.Description,
			Dishes:      ToOrderDetailedDishesFromProto(pbSet.Dishes),
			UserID:      pbSet.UserId,
			CreatedAt:   pbSet.CreatedAt.AsTime(),
			UpdatedAt:   pbSet.UpdatedAt.AsTime(),
			IsFavourite: pbSet.IsFavourite,
			LikeBy:      pbSet.LikeBy,
			IsPublic:    pbSet.IsPublic,
			Image:       pbSet.Image,
			Price:       pbSet.Price,
			Quantity:    pbSet.Quantity,
		}
	}
	return sets
}

func ToOrderDetailedDishesFromProto(pbDishes []*order.OrderDetailedDish) []OrderDetailedDish {
	if pbDishes == nil {
		return nil
	}

	dishes := make([]OrderDetailedDish, len(pbDishes))
	for i, pbDish := range pbDishes {
		dishes[i] = OrderDetailedDish{
			DishID:      pbDish.DishId,
			Quantity:    pbDish.Quantity,
			Name:        pbDish.Name,
			Price:       pbDish.Price,
			Description: pbDish.Description,
			Image:       pbDish.Image,
			Status:      pbDish.Status,
		}
	}
	return dishes
}

// -------------------

// Conversion functions
func ToPBCreateOrderRequest(req CreateOrderRequestType) *order.CreateOrderRequest {
	return &order.CreateOrderRequest{
		GuestId:        req.GuestID,
		UserId:         req.UserID,
		IsGuest:        req.IsGuest,
		TableNumber:    req.TableNumber,
		OrderHandlerId: req.OrderHandlerID,
		Status:         req.Status,
		CreatedAt:      timestamppb.New(req.CreatedAt),
		UpdatedAt:      timestamppb.New(req.UpdatedAt),
		TotalPrice:     req.TotalPrice,
		DishItems:      ToPBDishOrderItems(req.DishItems),
		SetItems:       ToPBSetOrderItems(req.SetItems),
		Topping:        req.Topping,
		TrackingOrder:  req.TrackingOrder,
		TakeAway:       req.TakeAway,
		ChiliNumber:    req.ChiliNumber,
		TableToken:     req.TableToken,
		OrderName:      req.OrderName, // Added new field
	}
}

func ToPBUpdateOrderRequest(req UpdateOrderRequestType) *order.UpdateOrderRequest {
	return &order.UpdateOrderRequest{
		Id:             req.ID,
		GuestId:        req.GuestID,
		UserId:         req.UserID,
		TableNumber:    req.TableNumber,
		OrderHandlerId: req.OrderHandlerID,
		Status:         req.Status,
		TotalPrice:     req.TotalPrice,
		DishItems:      ToPBDishOrderItems(req.DishItems),
		SetItems:       ToPBSetOrderItems(req.SetItems),
		IsGuest:        req.IsGuest,
		Topping:        req.Topping,
		TrackingOrder:  req.TrackingOrder,
		TakeAway:       req.TakeAway,
		ChiliNumber:    req.ChiliNumber,
		TableToken:     req.TableToken,
		OrderName:      req.OrderName, // Added new field
	}
}

func ToOrderFromPbOrder(pbOrder *order.Order) OrderType {
	if pbOrder == nil {
		return OrderType{}
	}

	var createdAt, updatedAt time.Time
	if pbOrder.CreatedAt != nil {
		createdAt = pbOrder.CreatedAt.AsTime()
	}
	if pbOrder.UpdatedAt != nil {
		updatedAt = pbOrder.UpdatedAt.AsTime()
	}

	return OrderType{
		ID:             pbOrder.GetId(),
		GuestID:        pbOrder.GetGuestId(),
		UserID:         pbOrder.GetUserId(),
		IsGuest:        pbOrder.GetIsGuest(),
		TableNumber:    pbOrder.GetTableNumber(),
		OrderHandlerID: pbOrder.GetOrderHandlerId(),
		Status:         pbOrder.GetStatus(),
		CreatedAt:      createdAt,
		UpdatedAt:      updatedAt,
		TotalPrice:     pbOrder.GetTotalPrice(),
		DishItems:      ToOrderDishesFromPbDishOrderItems(pbOrder.DishItems),
		SetItems:       ToOrderSetsFromPbSetOrderItems(pbOrder.SetItems),
		Topping:        pbOrder.GetTopping(),
		TrackingOrder:  pbOrder.GetTrackingOrder(),
		TakeAway:       pbOrder.GetTakeAway(),
		ChiliNumber:    pbOrder.GetChiliNumber(),
		TableToken:     pbOrder.GetTableToken(),
		OrderName:      pbOrder.GetOrderName(), // Added new field
	}
}

func ToOrderDetailedListResponseFromProto(pbRes *order.OrderDetailedListResponse) OrderDetailedListResponse {
	if pbRes == nil {
		return OrderDetailedListResponse{}
	}

	detailedResponses := make([]OrderDetailedResponse, len(pbRes.Data))
	for i, pbDetailedRes := range pbRes.Data {
		detailedResponses[i] = OrderDetailedResponse{
			ID:             pbDetailedRes.Id,
			GuestID:        pbDetailedRes.GuestId,
			UserID:         pbDetailedRes.UserId,
			TableNumber:    pbDetailedRes.TableNumber,
			OrderHandlerID: pbDetailedRes.OrderHandlerId,
			Status:         pbDetailedRes.Status,
			TotalPrice:     pbDetailedRes.TotalPrice,
			IsGuest:        pbDetailedRes.IsGuest,
			Topping:        pbDetailedRes.Topping,
			TrackingOrder:  pbDetailedRes.TrackingOrder,
			TakeAway:       pbDetailedRes.TakeAway,
			ChiliNumber:    pbDetailedRes.ChiliNumber,
			TableToken:     pbDetailedRes.TableToken,
			OrderName:      pbDetailedRes.OrderName, // Added new field
			DataSet:        ToOrderSetsDetailedFromProto(pbDetailedRes.DataSet),
			DataDish:       ToOrderDetailedDishesFromProto(pbDetailedRes.DataDish),
		}
	}

	return OrderDetailedListResponse{
		Data: detailedResponses,
		Pagination: PaginationInfo{
			CurrentPage: pbRes.Pagination.CurrentPage,
			TotalPages:  pbRes.Pagination.TotalPages,
			TotalItems:  pbRes.Pagination.TotalItems,
			PageSize:    pbRes.Pagination.PageSize,
		},
	}
}

func (h *OrderHandlerController) CreateOrder2(w http.ResponseWriter, r *http.Request) {
	var orderReq CreateOrderRequestType

	// Read the entire body first
	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Error reading request body: " + err.Error())
		http.Error(w, "error reading request body", http.StatusBadRequest)
		return
	}

	// Log the raw body for debugging
	h.logger.Info(fmt.Sprintf("Raw request body: %s", string(body)))

	// Decode the JSON
	if err := json.Unmarshal(body, &orderReq); err != nil {
		h.logger.Error("Error decoding request body: " + err.Error())
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	pbReq := ToPBCreateOrderRequest(orderReq)
	createdOrderResponse, err := h.client.CreateOrder(h.ctx, pbReq)
	if err != nil {
		h.logger.Error("Error creating order: " + err.Error())
		http.Error(w, "error creating order", http.StatusInternalServerError)
		return
	}

	res := ToOrderResFromPbOrderResponse(createdOrderResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
