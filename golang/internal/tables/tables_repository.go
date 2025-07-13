package tables_test

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
	"google.golang.org/protobuf/types/known/timestamppb"

	"english-ai-full/internal/proto_qr/table"
	"english-ai-full/logger" // Add this import
	"english-ai-full/token"
)

type TableRepository struct {
	db       *pgxpool.Pool
	jwtMaker *token.JWTMaker
	logger   *logger.Logger // Add this field
}

func NewTableRepository(db *pgxpool.Pool, secretKey string) *TableRepository {
	return &TableRepository{
		db:       db,
		jwtMaker: token.NewJWTMaker(secretKey),
		logger:   logger.NewLogger(), // Initialize the logger
	}
}

func (tr *TableRepository) GetTableList(ctx context.Context) ([]*table.Table, error) {
	tr.logger.Info("golang/quanqr/tables/tables_repository.go:GetTableList - Fetching all tables")

	query := `
		SELECT number, capacity, status, token, created_at, updated_at
		FROM tables
	`
	rows, err := tr.db.Query(ctx, query)
	if err != nil {
		tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableList - Error fetching tables: %v", err))
		return nil, fmt.Errorf("error fetching tables: %w", err)
	}
	defer rows.Close()

	var tables []*table.Table
	for rows.Next() {
		var t table.Table
		var createdAt, updatedAt time.Time
		var status string
		err := rows.Scan(
			&t.Number,
			&t.Capacity,
			&status,
			&t.Token,
			&createdAt,
			&updatedAt,
		)
		if err != nil {
			tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableList - Error scanning table: %v", err))
			return nil, fmt.Errorf("error scanning table: %w", err)
		}
		t.Status = table.TableStatus(table.TableStatus_value[status])
		t.CreatedAt = timestamppb.New(createdAt)
		t.UpdatedAt = timestamppb.New(updatedAt)
		tables = append(tables, &t)
	}
	if err := rows.Err(); err != nil {
		tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableList - Error iterating over tables: %v", err))
		return nil, fmt.Errorf("error iterating over tables: %w", err)
	}

	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableList - Successfully fetched %d tables", len(tables)))
	return tables, nil
}

func (tr *TableRepository) GetTableDetail(ctx context.Context, number int32) (*table.Table, error) {
	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableDetail - Fetching table detail for number: %d", number))

	query := `
		SELECT number, capacity, status, token, created_at, updated_at
		FROM tables
		WHERE number = $1
	`
	var t table.Table
	var createdAt, updatedAt time.Time
	var statusStr string
	err := tr.db.QueryRow(ctx, query, number).Scan(
		&t.Number,
		&t.Capacity,
		&statusStr,
		&t.Token,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableDetail - Error fetching table detail: %v", err))
		return nil, fmt.Errorf("error fetching table detail: %w", err)
	}

	t.Status = table.TableStatus(table.TableStatus_value[statusStr])
	t.CreatedAt = timestamppb.New(createdAt)
	t.UpdatedAt = timestamppb.New(updatedAt)

	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:GetTableDetail - Successfully fetched table detail for number: %d", number))
	return &t, nil
}

func (tr *TableRepository) UpdateTable(ctx context.Context, req *table.UpdateTableRequest) (*table.Table, error) {
	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:UpdateTable - Updating table with number: %d", req.Number))

	var newToken string
	var err error
	if req.ChangeToken {
		newToken, err = tr.generateToken(req.Number, req.Capacity, req.Status)
		if err != nil {
			tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:UpdateTable - Error generating new token: %v", err))
			return nil, fmt.Errorf("error generating new token: %w", err)
		}
	}

	query := `
		UPDATE tables
		SET capacity = $2, status = $3, token = CASE WHEN $4 THEN $5 ELSE token END, updated_at = $6
		WHERE number = $1
		RETURNING number, capacity, status, token, created_at, updated_at
	`

	var t table.Table
	var createdAt, updatedAt time.Time
	var statusStr string
	err = tr.db.QueryRow(ctx, query,
		req.Number,
		req.Capacity,
		req.Status.String(),
		req.ChangeToken,
		newToken,
		time.Now(),
	).Scan(
		&t.Number,
		&t.Capacity,
		&statusStr,
		&t.Token,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:UpdateTable - Error updating table: %v", err))
		return nil, fmt.Errorf("error updating table: %w", err)
	}

	t.Status = table.TableStatus(table.TableStatus_value[statusStr])
	t.CreatedAt = timestamppb.New(createdAt)
	t.UpdatedAt = timestamppb.New(updatedAt)

	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:UpdateTable - Successfully updated table with number: %d", t.Number))
	return &t, nil
}

func (tr *TableRepository) DeleteTable(ctx context.Context, number int32) (*table.Table, error) {
	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:DeleteTable - Deleting table with number: %d", number))

	query := `
		DELETE FROM tables
		WHERE number = $1
		RETURNING number, capacity, status, token, created_at, updated_at
	`
	var t table.Table
	var createdAt, updatedAt time.Time
	err := tr.db.QueryRow(ctx, query, number).Scan(
		&t.Number,
		&t.Capacity,
		&t.Status,
		&t.Token,
		&createdAt,
		&updatedAt,
	)
	if err != nil {
		tr.logger.Error(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:DeleteTable - Error deleting table: %v", err))
		return nil, fmt.Errorf("error deleting table: %w", err)
	}
	t.CreatedAt = timestamppb.New(createdAt)
	t.UpdatedAt = timestamppb.New(updatedAt)

	tr.logger.Info(fmt.Sprintf("golang/quanqr/tables/tables_repository.go:DeleteTable - Successfully deleted table with number: %d", number))
	return &t, nil
}

// helper
var TableStatusMap = map[table.TableStatus]string{
	table.TableStatus_AVAILABLE:      "Available",
	table.TableStatus_OCCUPIED:       "Occupied",
	table.TableStatus_RESERVED:       "Reserved",
	table.TableStatus_OUT_OF_SERVICE: "Out_Of_Service",
	table.TableStatus_TAKE_AWAY:      "Take_Away",
}

// ReverseTableStatusMap maps database status to protobuf status
var ReverseTableStatusMap = map[string]table.TableStatus{}

func init() {
	// Initialize reverse mapping
	for k, v := range TableStatusMap {
		ReverseTableStatusMap[v] = k
	}
}

func (tr *TableRepository) generateToken(tableNumber int32, capacity int32, status table.TableStatus) (string, error) {
	tr.logger.Info(fmt.Sprintf("Generating token for table number: %d", tableNumber))

	// Convert status to string using the map
	statusStr := TableStatusMap[status]
	if statusStr == "" {
		tr.logger.Warning(fmt.Sprintf("Invalid status provided for table %d, defaulting to Available", tableNumber))
		statusStr = "Available"
	}

	// Generate a short token with table-specific claims
	shortToken, err := tr.jwtMaker.CreateShortTableToken(
		tableNumber,
		capacity,
		statusStr,
		100*365*24*time.Hour,
	)
	if err != nil {
		tr.logger.Error(fmt.Sprintf("Error creating short token: %v", err))
		return "", fmt.Errorf("error creating short token: %w", err)
	}

	tr.logger.Info(fmt.Sprintf("Successfully generated token for table number: %d", tableNumber))
	return shortToken, nil
}

func (tr *TableRepository) CreateTable(ctx context.Context, req *table.CreateTableRequest) (*table.Table, error) {
	// Start logging with request details
	tr.logger.Info(fmt.Sprintf("Creating new table repository layer - Number: %d, Capacity: %d, Status: %s",
		req.Number,
		req.Capacity,
		req.Status.String()))

	// Status validation and conversion
	dbStatus, exists := TableStatusMap[req.Status]
	if !exists {
		tr.logger.Error(fmt.Sprintf("Invalid status provided: %s, defaulting to Available", req.Status.String()))
		dbStatus = "Available" // Use default status from schema
	}

	// Generate token with all table information
	token, err := tr.generateToken(req.Number, req.Capacity, req.Status)
	if err != nil {
		tr.logger.Error(fmt.Sprintf("Failed to generate token for table %d: %v", req.Number, err))
		return nil, fmt.Errorf("error generating token: %w", err)
	}

	// Token length validation
	if len(token) > 255 {
		token = token[:255]
		tr.logger.Warning(fmt.Sprintf("Token truncated to 255 characters for table %d", req.Number))
	}

	// Log the status conversion
	tr.logger.Info(fmt.Sprintf("Status conversion - Input: %s, DB Status: %s",
		req.Status.String(),
		dbStatus))

	query := `
        INSERT INTO tables (number, capacity, status, token, created_at, updated_at)
        VALUES ($1, $2, $3, $4, $5, $5)
        RETURNING number, capacity, status, token, created_at, updated_at
    `

	var t table.Table
	var createdAt, updatedAt time.Time
	var statusStr string

	// Execute query and scan results
	err = tr.db.QueryRow(ctx, query,
		req.Number,
		req.Capacity,
		dbStatus,
		token,
		time.Now(),
	).Scan(
		&t.Number,
		&t.Capacity,
		&statusStr,
		&t.Token,
		&createdAt,
		&updatedAt,
	)

	if err != nil {
		tr.logger.Error(fmt.Sprintf("Failed to create table %d: %v", req.Number, err))
		return nil, fmt.Errorf("error creating table: %w", err)
	}

	// Convert database status back to protobuf status
	if protoStatus, exists := ReverseTableStatusMap[statusStr]; exists {
		t.Status = protoStatus
	} else {
		tr.logger.Error(fmt.Sprintf("Unknown status from database: %s", statusStr))
		t.Status = table.TableStatus_OCCUPIED
	}

	t.CreatedAt = timestamppb.New(createdAt)
	t.UpdatedAt = timestamppb.New(updatedAt)

	tr.logger.Info(fmt.Sprintf("Successfully created table %d with status %s",
		t.Number,
		t.Status.String()))

	return &t, nil
}
