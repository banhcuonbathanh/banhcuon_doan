package qr_guests

import (
	"context"
	"log"

	"english-ai-full/internal/proto_qr/guest"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type GuestServiceStruct struct {
	guestRepo *GuestRepository
	guest.UnimplementedGuestServiceServer
}

func NewGuestService(guestRepo *GuestRepository) *GuestServiceStruct {
	return &GuestServiceStruct{
		guestRepo: guestRepo,
	}
}

func (gs *GuestServiceStruct) GuestLoginGRPC(ctx context.Context, req *guest.GuestLoginRequest) (*guest.GuestLoginResponse, error) {
	log.Println("Guest login attempt:",
		"Name:", req.Name,
		"Table Number:", req.TableNumber,
	)

	response, err := gs.guestRepo.GuestLogin(ctx, req)
	if err != nil {
		log.Println("Error during guest login:", err)
		return nil, err
	}

	log.Println("Guest logged in successfully. Guest ID:", response.Guest.Id)
	return response, nil
}

func (gs *GuestServiceStruct) GuestLogoutGRPC(ctx context.Context, req *guest.LogoutRequest) (*emptypb.Empty, error) {
	log.Println("Guest logout attempt")

	err := gs.guestRepo.GuestLogout(ctx, req)
	if err != nil {
		log.Println("Error during guest logout:", err)
		return nil, err
	}

	log.Println("Guest logged out successfully")
	return &emptypb.Empty{}, nil
}

func (gs *GuestServiceStruct) GuestRefreshTokenGRPC(ctx context.Context, req *guest.RefreshTokenRequest) (*guest.RefreshTokenResponse, error) {
	log.Println("Token refresh attempt")

	response, err := gs.guestRepo.GuestRefreshToken(ctx, req)
	if err != nil {
		log.Println("Error during token refresh:", err)
		return nil, err
	}

	log.Println("Token refreshed successfully")
	return response, nil
}

func (gs *GuestServiceStruct) GuestCreateOrdersGRPC(ctx context.Context, req *guest.GuestCreateOrderRequest) (*guest.OrdersResponse, error) {

	log.Print("golang/quanqr/qr_guests/qr_guests_service.go")
	log.Println("Create orders attempt:",
		"Number of items:", len(req.Items),
	)

	// Validate that all items have the same guest_Id
	if len(req.Items) > 0 {
		guestID := req.Items[0].GuestId
		for _, item := range req.Items[1:] {
			if item.GuestId != guestID {
				return nil, status.Errorf(codes.InvalidArgument, "All items must have the same guest_Id")
			}
		}
	}

	response, err := gs.guestRepo.GuestCreateOrders(ctx, req)
	if err != nil {
		log.Println("Error creating orders:", err)
		return nil, status.Errorf(codes.Internal, "Failed to create orders: %v", err)
	}

	log.Println("Orders created successfully. Number of orders:", len(response.Data))
	return response, nil
}

func (gs *GuestServiceStruct) GuestGetOrdersGRPC(ctx context.Context, req *guest.GuestGetOrdersGRPCRequest) (*guest.ListOrdersResponse, error) {
	log.Println("Get orders attempt for guest ID:", req.GuestId)

	response, err := gs.guestRepo.GuestGetOrders(ctx, req)
	if err != nil {
		log.Println("Error fetching orders:", err)
		return nil, err
	}

	log.Println("Orders fetched successfully. Number of orders:", len(response.Orders))
	return response, nil
}

// Add these methods to your GuestServiceStruct

func (gs *GuestServiceStruct) GuestCreateSession(ctx context.Context, req *guest.GuestSessionReq) (*guest.GuestSessionRes, error) {
	log.Print("golang/quanqr/qr_guests/qr_guests_service.go GuestCreateSession")
	return gs.guestRepo.GuestCreateSession(ctx, req)
}

func (gs *GuestServiceStruct) GuestGetSession(ctx context.Context, req *guest.GuestSessionReq) (*guest.GuestSessionRes, error) {
	return gs.guestRepo.GuestGetSession(ctx, req.GetId())
}

func (gs *GuestServiceStruct) GuestRevokeSession(ctx context.Context, req *guest.GuestSessionReq) (*guest.GuestSessionRes, error) {
	err := gs.guestRepo.GuestRevokeSession(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	// Return the session info after revocation
	session, err := gs.guestRepo.GuestGetSession(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return session, nil
}

func (gs *GuestServiceStruct) GuestDeleteSession(ctx context.Context, req *guest.GuestSessionReq) (*guest.GuestSessionRes, error) {
	// Get the session info before deletion for the response
	session, err := gs.guestRepo.GuestGetSession(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	err = gs.guestRepo.GuestDeleteSession(ctx, req.GetId())
	if err != nil {
		return nil, err
	}

	return session, nil
}
