package dish_grpc

//
//
//import (
//	"context"
//	"log"
//
//	"english-ai-full/quanqr/proto_qr/dish"
//
//
//	"google.golang.org/protobuf/types/known/emptypb"
//)
//
//type DishServiceStruct struct {
//	dishRepo *DishRepository
//	dish.UnimplementedDishServiceServer
//}
//
//func NewDishService(dishRepo *DishRepository) *DishServiceStruct {
//	return &DishServiceStruct{
//		dishRepo: dishRepo,
//	}
//}
//
//func (ds *DishServiceStruct) GetDishList(ctx context.Context, _ *emptypb.Empty) (*dish.DishListResponse, error) {
//	dishes, err := ds.dishRepo.GetDishList(ctx)
//	if err != nil {
//		log.Println("Error fetching dish list:", err)
//		return nil, err
//	}
//	return &dish.DishListResponse{
//		Data:    dishes,
//		Message: "Dishes fetched successfully",
//	}, nil
//}
//
//func (ds *DishServiceStruct) GetDishDetail(ctx context.Context, req *dish.DishIdParam) (*dish.DishResponse, error) {
//	d, err := ds.dishRepo.GetDishDetail(ctx, req.Id)
//	if err != nil {
//		log.Println("Error fetching dish detail:", err)
//		return nil, err
//	}
//	return &dish.DishResponse{
//		Data:    d,
//		Message: "Dish detail fetched successfully",
//	}, nil
//}
//
//func (ds *DishServiceStruct) CreateDish(ctx context.Context, req *dish.CreateDishRequest) (*dish.DishResponse, error) {
//	log.Println("Creating new dish:",
//		"Name:", req.Name,
//		"Price:", req.Price,
//		"Description:", req.Description,
//		"Image:", req.Image,
//		"Status:", req.Status,
//	)
//
//	createdDish, err := ds.dishRepo.CreateDish(ctx, req)
//	if err != nil {
//		log.Println("Error creating dish:", err)
//		return nil, err
//	}
//
//	log.Println("Dish created successfully. ID:", createdDish.Id)
//	return &dish.DishResponse{
//		Data:    createdDish,
//		Message: "Dish created successfully",
//	}, nil
//}
//
//func (ds *DishServiceStruct) UpdateDish(ctx context.Context, req *dish.UpdateDishRequest) (*dish.DishResponse, error) {
//	updatedDish, err := ds.dishRepo.UpdateDish(ctx, req)
//	if err != nil {
//		log.Println("Error updating dish:", err)
//		return nil, err
//	}
//	return &dish.DishResponse{
//		Data:    updatedDish,
//		Message: "Dish updated successfully",
//	}, nil
//}
//
//func (ds *DishServiceStruct) DeleteDish(ctx context.Context, req *dish.DishIdParam) (*dish.DishResponse, error) {
//	deletedDish, err := ds.dishRepo.DeleteDish(ctx, req.Id)
//	if err != nil {
//		log.Println("Error deleting dish:", err)
//		return nil, err
//	}
//	return &dish.DishResponse{
//		Data:    deletedDish,
//		Message: "Dish deleted successfully",
//	}, nil
//}
