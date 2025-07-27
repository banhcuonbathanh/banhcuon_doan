package account_handler

import (
	"context"

	"net/http"
	"strconv"

	errorcustom "english-ai-full/internal/error_custom"
	pb "english-ai-full/internal/proto_qr/account"


	"github.com/go-playground/validator/v10"

)

type BaseAccountHandler struct {
	userClient pb.AccountServiceClient
	validator  *validator.Validate
}

func NewBaseHandler(userClient pb.AccountServiceClient) *BaseAccountHandler {
	v := validator.New()
	v.RegisterValidation("password", ValidatePassword)
	v.RegisterValidation("role", ValidateRole)
	v.RegisterValidation("uniqueemail", ValidateEmailUnique(userClient))

	return &BaseAccountHandler{
		userClient: userClient,
		validator:  v,
	}
}

func (h *BaseAccountHandler) getUserIDFromContext(ctx context.Context) (int64, error) {
	userID, ok := ctx.Value("user_id").(int64)
	if !ok {
		return 0, errorcustom.NewAPIError(
			errorcustom.ErrCodeInvalidInput,
			"User ID not found in context",
			http.StatusUnauthorized,
		)
	}
	return userID, nil
}

func (h *BaseAccountHandler) getPaginationParams(r *http.Request) (page, pageSize int32, apiErr *errorcustom.APIError) {
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page, pageSize = 1, 10

	if pageStr != "" {
		if p, err := strconv.ParseInt(pageStr, 10, 32); err != nil || p < 1 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page parameter: must be a positive integer",
				http.StatusBadRequest,
			)
		} else {
			page = int32(p)
		}
	}

	if pageSizeStr != "" {
		if ps, err := strconv.ParseInt(pageSizeStr, 10, 32); err != nil || ps < 1 || ps > 100 {
			return 0, 0, errorcustom.NewAPIError(
				errorcustom.ErrCodeInvalidInput,
				"Invalid page_size parameter: must be between 1 and 100",
				http.StatusBadRequest,
			)
		} else {
			pageSize = int32(ps)
		}
	}

	return page, pageSize, nil
}




