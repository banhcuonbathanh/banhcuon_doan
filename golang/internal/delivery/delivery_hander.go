package delivery_grpc

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"english-ai-full/internal/proto_qr/delivery"
	"english-ai-full/logger"
	"english-ai-full/token"

	"github.com/go-chi/chi"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type DeliveryHandlerController struct {
	ctx        context.Context
	client     delivery.DeliveryServiceClient
	TokenMaker *token.JWTMaker
	logger     *logger.Logger
}

func NewDeliveryHandler(client delivery.DeliveryServiceClient, secretKey string) *DeliveryHandlerController {
	return &DeliveryHandlerController{
		ctx:        context.Background(),
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
		logger:     logger.NewLogger(),
	}
}

func (h *DeliveryHandlerController) GetDeliveryDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Fetching delivery detail for ID: %d", i))
	deliveryResponse, err := h.client.GetDeliveryDetailById(h.ctx, &delivery.DeliveryIdParam{Id: i})
	if err != nil {
		h.logger.Error("Error fetching delivery detail: " + err.Error())
		http.Error(w, "error getting delivery", http.StatusInternalServerError)
		return
	}

	res := ToDeliveryResFromPbDeliveryResponse(deliveryResponse)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *DeliveryHandlerController) GetDeliveriesListDetail(w http.ResponseWriter, r *http.Request) {
	// Parse query parameters
	query := r.URL.Query()

	page := int32(1)
	if pageStr := query.Get("page"); pageStr != "" {
		if pageInt, err := strconv.ParseInt(pageStr, 10, 32); err == nil {
			page = int32(pageInt)
		}
	}

	pageSize := int32(10)
	if pageSizeStr := query.Get("page_size"); pageSizeStr != "" {
		if pageSizeInt, err := strconv.ParseInt(pageSizeStr, 10, 32); err == nil {
			pageSize = int32(pageSizeInt)
		}
	}

	h.logger.Info("Fetching deliveries list hander")
	deliveriesResponse, err := h.client.GetDeliveriesListDetail(h.ctx, &delivery.GetDeliveriesRequest{
		Page:     page,
		PageSize: pageSize,
	})

	if err != nil {
		h.logger.Error("Error fetching deliveries list: hander" + err.Error())
		http.Error(w, "failed to fetch deliveries:hander "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToDeliveryDetailedListResponseFromProto(deliveriesResponse)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *DeliveryHandlerController) UpdateDelivery(w http.ResponseWriter, r *http.Request) {
	var deliveryReq UpdateDeliveryRequest
	if err := json.NewDecoder(r.Body).Decode(&deliveryReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Updating delivery: %d", deliveryReq.ID))
	updatedDeliveryResponse, err := h.client.UpdateDelivery(h.ctx, ToPBUpdateDeliveryRequest(deliveryReq))
	if err != nil {
		h.logger.Error("Error updating delivery: " + err.Error())
		http.Error(w, "error updating delivery", http.StatusInternalServerError)
		return
	}

	res := ToDeliveryResFromPbDeliveryResponse(updatedDeliveryResponse)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

func (h *DeliveryHandlerController) DeleteDelivery(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Deleting delivery with ID: %d", i))
	_, err = h.client.DeleteDeliveryDetailById(h.ctx, &delivery.DeliveryIdParam{Id: i})
	if err != nil {
		h.logger.Error("Error deleting delivery: " + err.Error())
		http.Error(w, "error deleting delivery", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Converter functions
func ToPBCreateDeliveryRequest(req CreateDeliveryRequest) *delivery.CreateDeliverRequest {
	return &delivery.CreateDeliverRequest{
		GuestId:         req.GuestID,
		UserId:          req.UserID,
		IsGuest:         req.IsGuest,
		TableNumber:     req.TableNumber,
		OrderHandlerId:  req.OrderHandlerID,
		Status:          req.Status,
		TotalPrice:      req.TotalPrice,
		DishItems:       ToPBDishDeliveryItems(req.DishItems),
		OrderId:         req.OrderID,
		BowChili:        req.BowChili,
		BowNoChili:      req.BowNoChili,
		TakeAway:        req.TakeAway,
		ChiliNumber:     req.ChiliNumber,
		TableToken:      req.TableToken,
		ClientName:      req.ClientName,
		DeliveryAddress: req.DeliveryAddress,
		DeliveryContact: req.DeliveryContact,
		DeliveryNotes:   req.DeliveryNotes,
		ScheduledTime:   timestamppb.New(req.ScheduledTime),
		DeliveryFee:     req.DeliveryFee,
		DeliveryStatus:  req.DeliveryStatus,
	}
}

func ToPBUpdateDeliveryRequest(req UpdateDeliveryRequest) *delivery.UpdateDeliverRequest {
	return &delivery.UpdateDeliverRequest{
		Id:                    req.ID,
		Status:                req.Status,
		DeliveryStatus:        req.DeliveryStatus,
		DeliveryNotes:         req.DeliveryNotes,
		EstimatedDeliveryTime: timestamppb.New(req.EstimatedDeliveryTime),
		ActualDeliveryTime:    timestamppb.New(req.ActualDeliveryTime),
	}
}

func ToPBDishDeliveryItems(items []DishDeliveryItem) []*delivery.DishDeliveryItem {
	pbItems := make([]*delivery.DishDeliveryItem, len(items))
	for i, item := range items {
		pbItems[i] = &delivery.DishDeliveryItem{
			DishId:   item.DishID,
			Quantity: item.Quantity,
		}
	}
	return pbItems
}

func ToDeliveryResFromPbDeliveryResponse(pbRes *delivery.DeliverResponse) DeliveryResponse {
	if pbRes == nil {
		return DeliveryResponse{}
	}

	return DeliveryResponse{
		ID:              pbRes.Id,
		GuestID:         pbRes.GuestId,
		UserID:          pbRes.UserId,
		IsGuest:         pbRes.IsGuest,
		TableNumber:     pbRes.TableNumber,
		OrderHandlerID:  pbRes.OrderHandlerId,
		Status:          pbRes.Status,
		CreatedAt:       pbRes.CreatedAt.AsTime(),
		UpdatedAt:       pbRes.UpdatedAt.AsTime(),
		TotalPrice:      pbRes.TotalPrice,
		DishItems:       ToDeliveryDishItemsFromProto(pbRes.DishItems),
		OrderID:         pbRes.OrderId,
		BowChili:        pbRes.BowChili,
		BowNoChili:      pbRes.BowNoChili,
		TakeAway:        pbRes.TakeAway,
		ChiliNumber:     pbRes.ChiliNumber,
		TableToken:      pbRes.TableToken,
		ClientName:      pbRes.ClientName,
		DeliveryStatus:  pbRes.DeliveryStatus,
		DeliveryAddress: pbRes.DeliveryAddress,
		DeliveryContact: pbRes.DeliveryContact,
		DeliveryNotes:   pbRes.DeliveryNotes,
		DeliveryFee:     pbRes.DeliveryFee,
	}
}

func ToDeliveryDishItemsFromProto(pbItems []*delivery.DishDeliveryItem) []DishDeliveryItem {
	if pbItems == nil {
		return nil
	}

	items := make([]DishDeliveryItem, len(pbItems))
	for i, pbItem := range pbItems {
		items[i] = DishDeliveryItem{
			DishID:   pbItem.DishId,
			Quantity: pbItem.Quantity,
		}
	}
	return items
}

// -----------------------

func (h *DeliveryHandlerController) GetDeliveryByClientName(w http.ResponseWriter, r *http.Request) {
	clientName := chi.URLParam(r, "name")
	if clientName == "" {
		http.Error(w, "client name is required", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf("Fetching delivery for client: hander layer  %s", clientName))
	deliveryResponse, err := h.client.GetDeliveryDetailByClientName(h.ctx, &delivery.DeliveryClientNameParam{Name: clientName})
	if err != nil {
		h.logger.Error("Error fetching delivery by client name: " + err.Error())
		http.Error(w, "error getting delivery", http.StatusInternalServerError)
		return
	}

	// Convert the response to the appropriate type
	res := ToDeliveryDetailedListResponseFromProto(deliveryResponse)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(res)
}

// ToDeliveryDetailedListResponseFromProto Update the converter function to handle the detailed list response
func ToDeliveryDetailedListResponseFromProto(pbRes *delivery.DeliveryDetailedListResponse) DeliveryDetailedListResponse {
	if pbRes == nil {
		return DeliveryDetailedListResponse{}
	}

	detailedResponses := make([]DeliveryDetailedResponse, len(pbRes.Data))
	for i, pbDetailedRes := range pbRes.Data {
		detailedResponses[i] = DeliveryDetailedResponse{
			ID:                    pbDetailedRes.Id,
			GuestID:               pbDetailedRes.GuestId,
			UserID:                pbDetailedRes.UserId,
			TableNumber:           pbDetailedRes.TableNumber,
			OrderHandlerID:        pbDetailedRes.OrderHandlerId,
			Status:                pbDetailedRes.Status,
			TotalPrice:            pbDetailedRes.TotalPrice,
			DataDish:              ToDeliveryDetailedDishFromProto(pbDetailedRes.DishItems),
			IsGuest:               pbDetailedRes.IsGuest,
			BowChili:              pbDetailedRes.BowChili,
			BowNoChili:            pbDetailedRes.BowNoChili,
			TakeAway:              pbDetailedRes.TakeAway,
			ChiliNumber:           pbDetailedRes.ChiliNumber,
			TableToken:            pbDetailedRes.TableToken,
			ClientName:            pbDetailedRes.ClientName,
			DeliveryStatus:        pbDetailedRes.DeliveryStatus,
			DeliveryAddress:       pbDetailedRes.DeliveryAddress,
			DeliveryContact:       pbDetailedRes.DeliveryContact,
			DeliveryNotes:         pbDetailedRes.DeliveryNotes,
			EstimatedDeliveryTime: pbDetailedRes.EstimatedDeliveryTime.AsTime(),
		}
	}

	return DeliveryDetailedListResponse{
		Data: detailedResponses,
		Pagination: PaginationInfo{
			CurrentPage: pbRes.Pagination.CurrentPage,
			TotalPages:  pbRes.Pagination.TotalPages,
			TotalItems:  pbRes.Pagination.TotalItems,
			PageSize:    pbRes.Pagination.PageSize,
		},
	}
}

// ToDeliveryDetailedDishFromProto function to convert detailed dish items
func ToDeliveryDetailedDishFromProto(pbDishes []*delivery.DeliveryDetailedDish) []DeliveryDetailedDish {
	if pbDishes == nil {
		return nil
	}

	dishes := make([]DeliveryDetailedDish, len(pbDishes))
	for i, pbDish := range pbDishes {
		dishes[i] = DeliveryDetailedDish{
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

func (h *DeliveryHandlerController) CreateDelivery3(w http.ResponseWriter, r *http.Request) {
	var deliveryReq CreateDeliveryRequest

	body, err := io.ReadAll(r.Body)
	if err != nil {
		h.logger.Error("Error reading request body: " + err.Error())
		http.Error(w, "error reading request body", http.StatusBadRequest)
		return
	}

	h.logger.Info(fmt.Sprintf(" golang/internal/delivery/delivery_hander.go Raw request CreateDelivery3: %s", string(body)))
	if err := json.Unmarshal(body, &deliveryReq); err != nil {
		h.logger.Error("Error decoding request body: " + err.Error())
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}
	pbReq := ToPBCreateDeliveryRequest(deliveryReq)
	createdDeliveryResponse, err := h.client.CreateOrder(h.ctx, pbReq)
	if err != nil {
		h.logger.Error("Error creating delivery: " + err.Error())
		http.Error(w, "error creating delivery", http.StatusInternalServerError)
		return
	}

	res := ToDeliveryResFromPbDeliveryResponse(createdDeliveryResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *DeliveryHandlerController) CreateDelivery(w http.ResponseWriter, r *http.Request) {
	var deliveryReq CreateDeliveryRequest

	// First decode the request body
	if err := json.NewDecoder(r.Body).Decode(&deliveryReq); err != nil {
		h.logger.Error("Error decoding request body: " + err.Error())
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Log after decoding to see the actual data
	h.logger.Info(fmt.Sprintf("Creating new delivery:222222 %+v", deliveryReq))

	pbReq := ToPBCreateDeliveryRequest(deliveryReq)
	createdDeliveryResponse, err := h.client.CreateOrder(h.ctx, pbReq)
	if err != nil {
		h.logger.Error("Error creating delivery: " + err.Error())
		http.Error(w, "error creating delivery", http.StatusInternalServerError)
		return
	}

	res := ToDeliveryResFromPbDeliveryResponse(createdDeliveryResponse)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}
