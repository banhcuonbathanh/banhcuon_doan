package dish_grpc

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"google.golang.org/protobuf/types/known/emptypb"

	"english-ai-full/internal/proto_qr/dish"
	"english-ai-full/token"
	
)

type DishHandlerController struct {
	ctx        context.Context
	client     dish.DishServiceClient
	TokenMaker *token.JWTMaker
}

func NewDishHandler(client dish.DishServiceClient, secretKey string) *DishHandlerController {
	return &DishHandlerController{
		ctx:        context.Background(),
		client:     client,
		TokenMaker: token.NewJWTMaker(secretKey),
	}
}

func (h *DishHandlerController) CreateDish(w http.ResponseWriter, r *http.Request) {
	var dishReq CreateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&dishReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	log.Println("handler CreateDish before")
	createdDish, err := h.client.CreateDish(h.ctx, ToPBCreateDishRequest(dishReq))
	if err != nil {
		log.Println("handler CreateDish err ", err)
		http.Error(w, "error creating dish in handler", http.StatusInternalServerError)
		return
	}
	log.Println("handler CreateDish after")

	res := ToDishResFromPbDishResponse(createdDish)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(res)
}

func (h *DishHandlerController) GetDishDetail(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}

	dish, err := h.client.GetDishDetail(h.ctx, &dish.DishIdParam{Id: i})
	if err != nil {
		http.Error(w, "error getting dish", http.StatusInternalServerError)
		return
	}

	res := ToDishResFromPbDishResponse(dish)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *DishHandlerController) GetDishList(w http.ResponseWriter, r *http.Request) {
	dishes, err := h.client.GetDishList(h.ctx, &emptypb.Empty{})
	if err != nil {
		http.Error(w, "failed to fetch dishes: "+err.Error(), http.StatusInternalServerError)
		return
	}

	res := ToDishResListFromPbDishListResponse(dishes)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, "failed to encode response: "+err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *DishHandlerController) UpdateDish(w http.ResponseWriter, r *http.Request) {
	var dishReq UpdateDishRequest
	if err := json.NewDecoder(r.Body).Decode(&dishReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	updatedDish, err := h.client.UpdateDish(h.ctx, ToPBUpdateDishRequest(dishReq))
	if err != nil {
		http.Error(w, "error updating dish", http.StatusInternalServerError)
		return
	}

	res := ToDishResFromPbDishResponse(updatedDish)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(res)
}

func (h *DishHandlerController) DeleteDish(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	i, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		http.Error(w, "error parsing ID", http.StatusBadRequest)
		return
	}

	_, err = h.client.DeleteDish(h.ctx, &dish.DishIdParam{Id: i})
	if err != nil {
		http.Error(w, "error deleting dish", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}