package delivery_grpc

//
//import (
//	"context"
//	"english-ai-full/logger"
//	"english-ai-full/quanqr/proto_qr/delivery"
//	"fmt"
//)
//
//type DeliveryServiceStruct struct {
//	deliveryRepo *DeliveryRepository
//	logger       *logger.Logger
//	delivery.UnimplementedDeliveryServiceServer
//}
//
//func NewDeliveryService(deliveryRepo *DeliveryRepository) *DeliveryServiceStruct {
//	return &DeliveryServiceStruct{
//		deliveryRepo: deliveryRepo,
//		logger:       logger.NewLogger(),
//	}
//}
//
//func (ds *DeliveryServiceStruct) CreateOrder(ctx context.Context, req *delivery.CreateDeliverRequest) (*delivery.DeliverResponse, error) {
//	ds.logger.Info(fmt.Sprintf("Creating new delivery: %+v", req))
//
//	createdDelivery, err := ds.deliveryRepo.CreateDelivery(ctx, req)
//	if err != nil {
//		ds.logger.Error("Error creating delivery: " + err.Error())
//		return nil, err
//	}
//
//	ds.logger.Info("Delivery created successfully. ID: " + fmt.Sprint(createdDelivery.Id))
//	return createdDelivery, nil
//}
//
//func (ds *DeliveryServiceStruct) GetDeliveryDetailById(ctx context.Context, req *delivery.DeliveryIdParam) (*delivery.DeliverResponse, error) {
//	ds.logger.Info("Fetching delivery detail for ID: " + fmt.Sprint(req.Id))
//
//	deliveryDetail, err := ds.deliveryRepo.GetDeliveryById(ctx, req.Id)
//	if err != nil {
//		ds.logger.Error("Error fetching delivery detail: " + err.Error())
//		return nil, err
//	}
//
//	return deliveryDetail, nil
//}
//
//func (ds *DeliveryServiceStruct) GetDeliveryDetailByClientName(ctx context.Context, req *delivery.DeliveryClientNameParam) (*delivery.DeliveryDetailedListResponse, error) {
//	ds.logger.Info("Fetching delivery detail for client name: service layer" + req.Name)
//
//	deliveryDetail, err := ds.deliveryRepo.GetDeliveryByClientName(ctx, req.Name)
//	if err != nil {
//		ds.logger.Error("Error fetching delivery detail by client name: " + err.Error())
//		return nil, err
//	}
//
//	return deliveryDetail, nil
//}
//
//func (ds *DeliveryServiceStruct) UpdateDelivery(ctx context.Context, req *delivery.UpdateDeliverRequest) (*delivery.DeliverResponse, error) {
//	ds.logger.Info("Updating delivery: " + fmt.Sprint(req.Id))
//
//	updatedDelivery, err := ds.deliveryRepo.UpdateDelivery(ctx, req)
//	if err != nil {
//		ds.logger.Error("Error updating delivery: " + err.Error())
//		return nil, err
//	}
//
//	return updatedDelivery, nil
//}
//
//func (ds *DeliveryServiceStruct) GetDeliveriesListDetail(ctx context.Context, req *delivery.GetDeliveriesRequest) (*delivery.DeliveryDetailedListResponse, error) {
//	ds.logger.Info("Fetching detailed delivery list with pagination service ")
//
//	// Validate pagination parameters
//	if req.Page < 1 {
//		req.Page = 1
//	}
//	if req.PageSize < 1 {
//		req.PageSize = 10 // Default page size
//	}
//
//	detailedList, err := ds.deliveryRepo.GetDeliveriesList(ctx, req)
//	if err != nil {
//		ds.logger.Error("Error fetching detailed delivery list: " + err.Error())
//		return nil, fmt.Errorf("failed to fetch detailed delivery list: %w", err)
//	}
//
//	return detailedList, nil
//}
//
//func (ds *DeliveryServiceStruct) DeleteDeliveryDetailById(ctx context.Context, req *delivery.DeliveryIdParam) (*delivery.DeliveryIdParam, error) {
//	ds.logger.Info("Deleting delivery with ID: " + fmt.Sprint(req.Id))
//
//	err := ds.deliveryRepo.DeleteDeliveryById(ctx, req.Id)
//	if err != nil {
//		ds.logger.Error("Error deleting delivery: " + err.Error())
//		return nil, err
//	}
//
//	return req, nil
//}
