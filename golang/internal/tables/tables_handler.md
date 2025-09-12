package tables_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/go-chi/chi"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"

	"english-ai-full/internal/proto_qr/table"
	"english-ai-full/logger"
)

type TablesHandlerController struct {
	ctx    context.Context
	client table.TableServiceClient
	logger *logger.Logger // Add this field
}

func NewTableHandler(client table.TableServiceClient) *TablesHandlerController {
	return &TablesHandlerController{
		ctx:    context.Background(),
		client: client,
		logger: logger.NewLogger(), // Initialize the logger
	}
}

func (h *TablesHandlerController) GetTableList(w http.ResponseWriter, r *http.Request) {
	log.Println("Handling GetTableList request")

	response, err := h.client.GetTableList(h.ctx, &emptypb.Empty{})
	if err != nil {
		log.Println("Error fetching table list:", err)
		http.Error(w, "error fetching table list", http.StatusInternalServerError)
		return
	}

	log.Println("Table list fetched successfully. Number of tables:", len(response.Data))

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *TablesHandlerController) GetTableDetail(w http.ResponseWriter, r *http.Request) {
	tableNumberStr := chi.URLParam(r, "tableNumber")
	tableNumber, err := strconv.ParseInt(tableNumberStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid table number", http.StatusBadRequest)
		return
	}

	log.Println("Handling GetTableDetail request for table number:", tableNumber)

	response, err := h.client.GetTableDetail(h.ctx, &table.TableNumberRequest{Number: int32(tableNumber)})
	if err != nil {
		log.Println("Error fetching table detail:", err)
		http.Error(w, "error fetching table detail", http.StatusInternalServerError)
		return
	}

	log.Println("Table detail fetched successfully for table number:", tableNumber)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *TablesHandlerController) CreateTable(w http.ResponseWriter, r *http.Request) {
	log.Print("Starting CreateTable handler - Path: golang/quanqr/tables/tables_handler.go")

	var createReq CreateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&createReq); err != nil {
		log.Printf("Error decoding request body: %v - Path: golang/quanqr/tables/tables_handler.go", err)
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	// Log the incoming request details
	h.logger.Info(fmt.Sprintf("Received CreateTable request - Number: %d, Capacity: %d, Status: %s",
		createReq.Number,
		createReq.Capacity,
		createReq.Status))

	log.Printf("Converting request to proto format - Table Number: %d", createReq.Number)

	protoReq := &table.CreateTableRequest{
		Number:   createReq.Number,
		Capacity: createReq.Capacity,
		Status:   table.TableStatus(table.TableStatus_value[string(createReq.Status)]),
	}

	log.Print("Calling gRPC CreateTable service")

	response, err := h.client.CreateTable(h.ctx, protoReq)
	if err != nil {
		log.Printf("Error from gRPC service: %v - Path: golang/quanqr/tables/tables_handler.go", err)
		if st, ok := status.FromError(err); ok {
			log.Printf("gRPC status error: %v", st.Message())
			http.Error(w, st.Message(), http.StatusBadRequest)
		} else {
			log.Printf("Internal server error: %v", err)
			http.Error(w, "error creating table", http.StatusInternalServerError)
		}
		return
	}

	log.Printf("Table created successfully - Number: %d, Response: %+v", createReq.Number, response)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding response: %v", err)
		http.Error(w, "error encoding response", http.StatusInternalServerError)
		return
	}

	log.Print("CreateTable handler completed successfully")
}

func (h *TablesHandlerController) UpdateTable(w http.ResponseWriter, r *http.Request) {
	var updateReq table.UpdateTableRequest
	if err := json.NewDecoder(r.Body).Decode(&updateReq); err != nil {
		http.Error(w, "error decoding request body", http.StatusBadRequest)
		return
	}

	log.Println("Handling UpdateTable request for table number:", updateReq.Number)

	protoReq := &table.UpdateTableRequest{
		Number:      updateReq.Number,
		ChangeToken: updateReq.ChangeToken,
		Capacity:    updateReq.Capacity,
		Status:      table.TableStatus(table.TableStatus_value[string(updateReq.Status)]),
	}

	response, err := h.client.UpdateTable(h.ctx, protoReq)
	if err != nil {
		if st, ok := status.FromError(err); ok {
			http.Error(w, st.Message(), http.StatusBadRequest)
		} else {
			log.Println("Error updating table:", err)
			http.Error(w, "error updating table", http.StatusInternalServerError)
		}
		return
	}

	log.Println("Table updated successfully for number:", updateReq.Number)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func (h *TablesHandlerController) DeleteTable(w http.ResponseWriter, r *http.Request) {
	tableNumberStr := chi.URLParam(r, "tableNumber")
	tableNumber, err := strconv.ParseInt(tableNumberStr, 10, 32)
	if err != nil {
		http.Error(w, "invalid table number", http.StatusBadRequest)
		return
	}

	log.Println("Handling DeleteTable request for table number:", tableNumber)

	response, err := h.client.DeleteTable(h.ctx, &table.TableNumberRequest{Number: int32(tableNumber)})
	if err != nil {
		log.Println("Error deleting table:", err)
		http.Error(w, "error deleting table", http.StatusInternalServerError)
		return
	}

	log.Println("Table deleted successfully for number:", tableNumber)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
