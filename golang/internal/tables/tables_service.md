package tables_test

import (
	"context"
	"fmt"

	"english-ai-full/internal/proto_qr/table"
	"english-ai-full/logger" // Add this import

	"google.golang.org/protobuf/types/known/emptypb"
)

type TableServiceStruct struct {
	tableRepo *TableRepository
	logger    *logger.Logger // Add this field
	table.UnimplementedTableServiceServer
}

func NewTableService(tableRepo *TableRepository) *TableServiceStruct {
	return &TableServiceStruct{
		tableRepo: tableRepo,
		logger:    logger.NewLogger(), // Initialize the logger
	}
}

func (ts *TableServiceStruct) GetTableList(ctx context.Context, _ *emptypb.Empty) (*table.TableListResponse, error) {
	ts.logger.Info("Fetching all tables golang/quanqr/tables/tables_service.go")

	tables, err := ts.tableRepo.GetTableList(ctx)
	if err != nil {
		ts.logger.Error("Error fetching tables: golang/quanqr/tables/tables_service.go" + err.Error())
		return nil, err
	}

	return &table.TableListResponse{
		Data:    tables,
		Message: "Tables fetched successfully",
	}, nil
}

func (ts *TableServiceStruct) GetTableDetail(ctx context.Context, req *table.TableNumberRequest) (*table.TableResponse, error) {

	ts.logger.Info("Fetching table detail for number: golang/quanqr/tables/tables_service.go")

	tableDetail, err := ts.tableRepo.GetTableDetail(ctx, req.Number)
	if err != nil {
		ts.logger.Error("Error fetching table detail: golang/quanqr/tables/tables_service.go" + err.Error())
		return nil, err
	}

	return &table.TableResponse{
		Data:    tableDetail,
		Message: "Table detail fetched successfully",
	}, nil
}

func (ts *TableServiceStruct) CreateTable(ctx context.Context, req *table.CreateTableRequest) (*table.TableResponse, error) {
	ts.logger.Info("golang/quanqr/tables/tables_service.go 22 create table")
	ts.logger.Info(fmt.Sprintf("Creating new table service layer - Number: %d, Capacity: %d, Status: %s",
		req.Number,
		req.Capacity,
		req.Status.String()))
	createdTable, err := ts.tableRepo.CreateTable(ctx, req)
	if err != nil {
		ts.logger.Error("golang/quanqr/tables/tables_service.go 22 err: create table" + err.Error())
		ts.logger.Error("Error creating table: " + err.Error())
		return nil, err
	}

	return &table.TableResponse{
		Data:    createdTable,
		Message: "Table created successfully",
	}, nil
}

func (ts *TableServiceStruct) UpdateTable(ctx context.Context, req *table.UpdateTableRequest) (*table.TableResponse, error) {
	ts.logger.Info("Updating table: golang/quanqr/tables/tables_service.go")

	updatedTable, err := ts.tableRepo.UpdateTable(ctx, req)
	if err != nil {
		ts.logger.Error("Error updating table: golang/quanqr/tables/tables_service.go" + err.Error())
		return nil, err
	}

	return &table.TableResponse{
		Data:    updatedTable,
		Message: "Table updated successfully",
	}, nil
}

func (ts *TableServiceStruct) DeleteTable(ctx context.Context, req *table.TableNumberRequest) (*table.TableResponse, error) {
	ts.logger.Info("Deleting table: golang/quanqr/tables/tables_service.go")

	deletedTable, err := ts.tableRepo.DeleteTable(ctx, req.Number)
	if err != nil {
		ts.logger.Error("Error deleting table: golang/quanqr/tables/tables_service.go" + err.Error())
		return nil, err
	}

	return &table.TableResponse{
		Data:    deletedTable,
		Message: "Table deleted successfully",
	}, nil
}
