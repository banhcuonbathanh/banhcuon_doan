package qr_guests

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"time"

	"english-ai-full/internal/proto_qr/guest"
	"english-ai-full/logger"
	"english-ai-full/token"

	"google.golang.org/protobuf/types/known/timestamppb"
)

type GuestHandlerController struct {
	ctx        context.Context
	client     guest.GuestServiceClient
	TokenMaker *token.JWTMaker
	logger     *logger.Logger
}

func NewGuestHandler(client guest.GuestServiceClient, secretKey string) *GuestHandlerController {
	return &GuestHandlerController{
		ctx:        context.Background(),
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
		logger:     logger.NewLogger(),
	}
}

func (h *GuestHandlerController) GuestLogin(w http.ResponseWriter, r *http.Request) {
	// h.logger.Info("Guest login handler started")
	var loginReq GuestLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
		h.logger.Error(fmt.Sprintf("error decoding request body: %v", err))
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Create guest through gRPC service
	guestResponse, err := h.client.GuestLoginGRPC(h.ctx, ToPBGuestLoginRequest(loginReq))

	if err != nil {
		h.logger.Error(fmt.Sprintf("error during guest login: %v", err))
		http.Error(w, "error logging in guest", http.StatusInternalServerError)
		return
	}

	// Create access token (short-lived)
	accessToken, accessClaims, err := h.TokenMaker.CreateToken(
		guestResponse.Guest.Id,
		guestResponse.Guest.Name,
		"guest",
		15*time.Minute,
	)
	if err != nil {
		h.logger.Error(fmt.Sprintf("error creating access token: %v", err))
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	// Create refresh token (long-lived)
	refreshToken, refreshClaims, err := h.TokenMaker.CreateToken(
		guestResponse.Guest.Id,
		guestResponse.Guest.Name,
		"guest",
		24*time.Hour,
	)
	if err != nil {
		h.logger.Error(fmt.Sprintf("error creating refresh token: %v", err))
		http.Error(w, "error creating token", http.StatusInternalServerError)
		return
	}

	// Create session for refresh token
	session, err := h.client.GuestCreateSession(h.ctx, &guest.GuestSessionReq{
		Id:           refreshClaims.RegisteredClaims.ID,
		Name:         guestResponse.Guest.Name, // Using name as identifier for guests
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    timestamppb.New(refreshClaims.RegisteredClaims.ExpiresAt.Time),
	})
	if err != nil {
		h.logger.Error(fmt.Sprintf("error creating session: %v", err))
		http.Error(w, "error creating session", http.StatusInternalServerError)
		return
	}

	// Prepare response
	res := GuestLoginResponse{
		SessionID:             session.GetId(),
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		Guest:                 ToGuestInfoFromPB(guestResponse.Guest),
	}

	h.logger.Info("Guest logged in successfully. Guest ID:")

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

//---------------
// func (h *GuestHandlerController) GuestLogin(w http.ResponseWriter, r *http.Request) {
// 	log.Println("golang/quanqr/qr_guests/qr_guests_handler.go",)
// 	var loginReq GuestLoginRequest
// 	if err := json.NewDecoder(r.Body).Decode(&loginReq); err != nil {
// 		http.Error(w, "error decoding request body", http.StatusBadRequest)
// 		return
// 	}

// 	log.Println("Guest login attempt:", "Name:", loginReq.Name, "Table Number:", loginReq.TableNumber)

// 	response, err := h.client.GuestLoginGRPC(h.ctx, ToPBGuestLoginRequest(loginReq))
// 	if err != nil {
// 		log.Println("Error during guest login:", err)
// 		http.Error(w, "error logging in guest", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Println("Guest logged in successfully. Guest ID:", response.Guest.Id)

// 	res := ToGuestLoginResponseFromPB(response)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(res)
// }

func (h *GuestHandlerController) GuestLogout(w http.ResponseWriter, r *http.Request) {
	var logoutReq LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&logoutReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	log.Println("Guest logout attempt")

	_, err := h.client.GuestLogoutGRPC(h.ctx, ToPBLogoutRequest(logoutReq))
	if err != nil {
		log.Println("Error during guest logout:", err)
		http.Error(w, "error logging out guest", http.StatusInternalServerError)
		return
	}

	log.Println("Guest logged out successfully")
	w.WriteHeader(http.StatusNoContent)
}

func (h *GuestHandlerController) RefreshToken(w http.ResponseWriter, r *http.Request) {
	var refreshReq RefreshTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&refreshReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	log.Println("Token refresh attempt")

	response, err := h.client.GuestRefreshTokenGRPC(h.ctx, ToPBRefreshTokenRequest(refreshReq))
	if err != nil {
		log.Println("Error during token refresh:", err)
		http.Error(w, "error refreshing token", http.StatusInternalServerError)
		return
	}

	log.Println("Token refreshed successfully")

	res := ToRefreshTokenResponseFromPB(response)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

// func (h *GuestHandlerController) CreateOrders(w http.ResponseWriter, r *http.Request) {

// 	log.Print("golang/quanqr/qr_guests/qr_guests_handler.go")
// 	var orderReq CreateOrdersRequest
// 	if err := json.NewDecoder(r.Body).Decode(&orderReq); err != nil {
// 		http.Error(w, "error decoding request body", http.StatusBadRequest)
// 		return
// 	}

// 	log.Println("Create orders attempt:", "Number of items:", len(orderReq.Items))

// 	response, err := h.client.GuestCreateOrdersGRPC(h.ctx, ToPBCreateOrdersRequest(orderReq))
// 	if err != nil {
// 		if st, ok := status.FromError(err); ok {
// 			http.Error(w, st.Message(), http.StatusBadRequest)
// 		} else {
// 			log.Println("Error creating orders:", err)
// 			http.Error(w, "error creating orders", http.StatusInternalServerError)
// 		}
// 		return
// 	}

// 	log.Println("Orders created successfully. Number of orders:", len(response.Data))

// 	res := ToOrdersResponseFromPB(response)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusCreated)
// 	json.NewEncoder(w).Encode(res)
// }

// func (h *GuestHandlerController) GetOrders(w http.ResponseWriter, r *http.Request) {
// 	guestIDStr := chi.URLParam(r, "guestId")
// 	guestID, err := strconv.ParseInt(guestIDStr, 10, 64)
// 	if err != nil {
// 		http.Error(w, "invalid guest ID", http.StatusBadRequest)
// 		return
// 	}

// 	log.Println("Get orders attempt for guest ID:", guestID)

// 	response, err := h.client.GuestGetOrdersGRPC(h.ctx, &guest.GuestGetOrdersGRPCRequest{GuestId: guestID})
// 	if err != nil {
// 		log.Println("Error fetching orders:", err)
// 		http.Error(w, "error fetching orders", http.StatusInternalServerError)
// 		return
// 	}

// 	log.Println("Orders fetched successfully. Number of orders:", len(response.Orders))

// 	res := ToListOrdersResponseFromPB(response)
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(http.StatusOK)
// 	json.NewEncoder(w).Encode(res)
// }

// Helper functions for converting between protobuf and local types
func ToPBGuestLoginRequest(req GuestLoginRequest) *guest.GuestLoginRequest {
	return &guest.GuestLoginRequest{
		Name:        req.Name,
		TableNumber: req.TableNumber,
		Token:       req.Token,
	}
}

func ToGuestLoginResponseFromPB(resp *guest.GuestLoginResponse) GuestLoginResponse {
	return GuestLoginResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		Guest:        ToGuestInfoFromPB(resp.Guest),
	}
}

func ToGuestInfoFromPB(info *guest.GuestInfo) GuestInfo {
	return GuestInfo{
		ID:          info.Id,
		Name:        info.Name,
		Role:        info.Role,
		TableNumber: info.TableNumber,
		CreatedAt:   info.CreatedAt.AsTime(),
		UpdatedAt:   info.UpdatedAt.AsTime(),
	}
}

func ToPBLogoutRequest(req LogoutRequest) *guest.LogoutRequest {
	return &guest.LogoutRequest{
		RefreshToken: req.RefreshToken,
	}
}

func ToPBRefreshTokenRequest(req RefreshTokenRequest) *guest.RefreshTokenRequest {
	return &guest.RefreshTokenRequest{
		RefreshToken: req.RefreshToken,
	}
}

func ToRefreshTokenResponseFromPB(resp *guest.RefreshTokenResponse) RefreshTokenResponse {
	return RefreshTokenResponse{
		AccessToken:  resp.AccessToken,
		RefreshToken: resp.RefreshToken,
		Message:      resp.Message,
	}
}

// func ToPBCreateOrdersRequest(req CreateOrdersRequest) *guest.GuestCreateOrderRequest {
// 	items := make([]*guest.GuestCreateOrderItem, len(req.Items))
// 	for i, item := range req.Items {
// 		items[i] = &guest.GuestCreateOrderItem{
// 			DishId:   item.DishID,
// 			Quantity: item.Quantity,
// 			GuestId:  item.GuestID,
// 		}
// 	}
// 	return &guest.GuestCreateOrderRequest{
// 		Items: items,
// 	}
// }

// func ToOrdersResponseFromPB(resp *guest.OrdersResponse) OrdersResponse {
// 	orders := make([]Order, len(resp.Data))
// 	for i, order := range resp.Data {
// 		orders[i] = ToOrderFromPB(order)
// 	}
// 	return OrdersResponse{
// 		Data:    orders,
// 		Message: resp.Message,
// 	}
// }

// func ToOrderFromPB(order *guest.Order) Order {
// 	return Order{
// 		ID:          order.Id,
// 		GuestID:     order.GuestId,
// 		TableNumber: order.TableNumber,
// 		DishID:      order.DishId,
// 		Quantity:    order.Quantity,
// 		Status:      order.Status,
// 		CreatedAt:   order.CreatedAt.AsTime(),
// 		UpdatedAt:   order.UpdatedAt.AsTime(),
// 	}
// }

// func ToListOrdersResponseFromPB(resp *guest.ListOrdersResponse) ListOrdersResponse {
// 	orders := make([]Order, len(resp.Orders))
// 	for i, order := range resp.Orders {
// 		orders[i] = ToOrderFromPB(order)
// 	}
// 	return ListOrdersResponse{
// 		Orders:  orders,
// 		Message: resp.Message,
// 	}
// }
