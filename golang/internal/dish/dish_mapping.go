package dish_grpc

import (
	"english-ai-full/internal/proto_qr/dish"

	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToPBCreateDishRequest(req CreateDishRequest) *dish.CreateDishRequest {
	return &dish.CreateDishRequest{
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Image:       req.Image,
		Status:      &req.Status,
	}
}

func ToPBUpdateDishRequest(req UpdateDishRequest) *dish.UpdateDishRequest {
	return &dish.UpdateDishRequest{
		Id:          req.ID,
		Name:        req.Name,
		Price:       req.Price,
		Description: req.Description,
		Image:       req.Image,
		Status:      &req.Status,
	}
}

func ToDishResFromPbDishResponse(pbResp *dish.DishResponse) DishResponse {
	return DishResponse{
		Data:    toDishFromPbDish(pbResp.Data),
		Message: pbResp.Message,
	}
}

func ToDishResListFromPbDishListResponse(pbResp *dish.DishListResponse) DishListResponse {
	dishes := make([]Dish, len(pbResp.Data))
	for i, pbDish := range pbResp.Data {
		dishes[i] = toDishFromPbDish(pbDish)
	}
	return DishListResponse{
		Data:    dishes,
		Message: pbResp.Message,
	}
}

func toDishFromPbDish(pbDish *dish.Dish) Dish {
	return Dish{
		ID:          pbDish.Id,
		Name:        pbDish.Name,
		Price:       pbDish.Price,
		Description: pbDish.Description,
		Image:       pbDish.Image,
		Status:      pbDish.Status,
		CreatedAt:   pbDish.CreatedAt.AsTime(),
		UpdatedAt:   pbDish.UpdatedAt.AsTime(),
	}
}

func ToPbDishFromDish(d Dish) *dish.Dish {
	return &dish.Dish{
		Id:          d.ID,
		Name:        d.Name,
		Price:       d.Price,
		Description: d.Description,
		Image:       d.Image,
		Status:      d.Status,
		CreatedAt:   timestamppb.New(d.CreatedAt),
		UpdatedAt:   timestamppb.New(d.UpdatedAt),
	}
}
