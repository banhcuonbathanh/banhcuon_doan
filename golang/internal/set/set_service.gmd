package set_qr

import (
	"context"
	"fmt"

	"english-ai-full/internal/proto_qr/set"
	"english-ai-full/logger"

	"google.golang.org/protobuf/types/known/emptypb"
)

type SetServiceStruct struct {
	setRepo *SetRepository
	logger  *logger.Logger
	set.UnimplementedSetServiceServer
}

func NewSetService(setRepo *SetRepository) *SetServiceStruct {
	return &SetServiceStruct{
		setRepo: setRepo,

		logger: logger.NewLogger(),
	}
}

func (ss *SetServiceStruct) GetSetProtoList(ctx context.Context, _ *emptypb.Empty) (*set.SetProtoListResponse, error) {
	ss.logger.Info("Fetching set list")
	sets, err := ss.setRepo.GetSetProtoList(ctx)
	if err != nil {
		ss.logger.Error("Error fetching set list: " + err.Error())
		return nil, err
	}
	return &set.SetProtoListResponse{
		Data: sets,
	}, nil
}

func (ss *SetServiceStruct) GetSetProtoDetail(ctx context.Context, req *set.SetProtoIdParam) (*set.SetProtoResponse, error) {
	ss.logger.Info("Fetching set detail for ID: ")
	s, err := ss.setRepo.GetSetProtoDetail(ctx, req.Id)
	if err != nil {
		ss.logger.Error("Error fetching set detail: " + err.Error())
		return nil, err
	}
	return &set.SetProtoResponse{
		Data: s,
	}, nil
}
func (ss *SetServiceStruct) CreateSetProto(ctx context.Context, req *set.CreateSetProtoRequest) (*set.SetProtoResponse, error) {

	ss.logger.Info(fmt.Sprintf("Creating new set:CreateSetProto serviace  %+v", req))
	createdSet, err := ss.setRepo.CreateSetProto(ctx, req)
	if err != nil {
		ss.logger.Error("Error creating set: " + err.Error())
		return nil, err
	}

	ss.logger.Info("Set created successfully. ID: " + fmt.Sprint(createdSet.Id))
	return &set.SetProtoResponse{
		Data: createdSet,
	}, nil
}

func (ss *SetServiceStruct) UpdateSetProto(ctx context.Context, req *set.UpdateSetProtoRequest) (*set.SetProtoResponse, error) {
	ss.logger.Info("Updating set: ")
	updatedSet, err := ss.setRepo.UpdateSetProto(ctx, req)
	if err != nil {
		ss.logger.Error("Error updating set: " + err.Error())
		return nil, err
	}
	return &set.SetProtoResponse{
		Data: updatedSet,
	}, nil
}

func (ss *SetServiceStruct) DeleteSetProto(ctx context.Context, req *set.SetProtoIdParam) (*set.SetProtoResponse, error) {
	ss.logger.Info("Deleting set: ")
	deletedSet, err := ss.setRepo.DeleteSetProto(ctx, req.Id)
	if err != nil {
		ss.logger.Error("Error deleting set: " + err.Error())
		return nil, err
	}
	return &set.SetProtoResponse{
		Data: deletedSet,
	}, nil
}

func (ss *SetServiceStruct) GetSetProtoListDetail(ctx context.Context, _ *emptypb.Empty) (*set.SetProtoListDetailResponse, error) {
	ss.logger.Info("Fetching detailed set list")
	sets, err := ss.setRepo.GetSetProtoListDetail(ctx)
	if err != nil {
		ss.logger.Error("Error fetching detailed set list: " + err.Error())
		return nil, err
	}
	return &set.SetProtoListDetailResponse{
		Data: sets,
	}, nil
}

//-----

// type SetServiceStruct struct {
//     setRepo *SetRepository
//     set.UnimplementedSetServiceServer
// }

// func NewSetService(setRepo *SetRepository) *SetServiceStruct {
//     return &SetServiceStruct{
//         setRepo: setRepo,
//     }
// }

// func (ss *SetServiceStruct) GetSetProtoList(ctx context.Context, _ *emptypb.Empty) (*set.SetProtoListResponse, error) {
//     log.Println("Fetching set list")
//     sets, err := ss.setRepo.GetSetProtoList(ctx)
//     if err != nil {
//         log.Println("Error fetching set list:", err)
//         return nil, err
//     }
//     return &set.SetProtoListResponse{
//         Data: sets,
//     }, nil
// }

// func (ss *SetServiceStruct) GetSetProtoDetail(ctx context.Context, req *set.SetProtoIdParam) (*set.SetProtoResponse, error) {
//     log.Println("Fetching set detail for ID:", req.Id)
//     s, err := ss.setRepo.GetSetProtoDetail(ctx, req.Id)
//     if err != nil {
//         log.Println("Error fetching set detail:", err)
//         return nil, err
//     }
//     return &set.SetProtoResponse{
//         Data: s,
//     }, nil
// }

// func (ss *SetServiceStruct) CreateSetProto(ctx context.Context, req *set.CreateSetProtoRequest) (*set.SetProtoResponse, error) {
//     log.Println("Creating new set:", req.Name)
//     createdSet, err := ss.setRepo.CreateSetProto(ctx, req)
//     if err != nil {
//         log.Println("Error creating set:", err)
//         return nil, err
//     }
//     log.Println("Set created successfully. ID:", createdSet.Id)
//     return &set.SetProtoResponse{
//         Data: createdSet,
//     }, nil
// }

// func (ss *SetServiceStruct) UpdateSetProto(ctx context.Context, req *set.UpdateSetProtoRequest) (*set.SetProtoResponse, error) {
//     log.Println("Updating set:", req.Id)
//     updatedSet, err := ss.setRepo.UpdateSetProto(ctx, req)
//     if err != nil {
//         log.Println("Error updating set:", err)
//         return nil, err
//     }
//     return &set.SetProtoResponse{
//         Data: updatedSet,
//     }, nil
// }

// func (ss *SetServiceStruct) DeleteSetProto(ctx context.Context, req *set.SetProtoIdParam) (*set.SetProtoResponse, error) {
//     log.Println("Deleting set:", req.Id)
//     deletedSet, err := ss.setRepo.DeleteSetProto(ctx, req.Id)
//     if err != nil {
//         log.Println("Error deleting set:", err)
//         return nil, err
//     }
//     return &set.SetProtoResponse{
//         Data: deletedSet,
//     }, nil
// }

// func (ss *SetServiceStruct) CreateSetProto(ctx context.Context, req *CreateSetRequest) (*set.SetProtoResponse, error) {
// 	ss.logger.Info("Creating new set: " + req.Name)

// 	// Transform the request into the format expected by the repository
// 	repoReq := &set.CreateSetProtoRequest{
// 		Name:        req.Name,
// 		Description: req.Description,
// 		UserId:      int(req.UserID),
// 		Dishes:      make([]*set.SetDishProto, len(req.Dishes)),
// 	}

// 	for i, dish := range req.Dishes {
// 		repoReq.Dishes[i] = &set.SetDishProto{
// 			DishId:   dish.Dish.ID,
// 			Quantity: int32(dish.Quantity),
// 		}
// 	}

// 	createdSet, err := ss.setRepo.CreateSetProto(ctx, repoReq)
// 	if err != nil {
// 		ss.logger.Error("Error creating set: " + err.Error())
// 		return nil, err
// 	}

// 	ss.logger.Info("Set created successfully. ID: " + fmt.Sprint(createdSet.Id))
// 	return &set.SetProtoResponse{
// 		Data: createdSet,
// 	}, nil
// }
