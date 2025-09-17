// error_system/messages_vn.go
package error_system

// vietnameseMessages contains all Vietnamese translations for error codes
// This can be easily maintained and updated by translators
var vietnameseMessages = map[ErrorCode]string{
	// User/Account errors - Lỗi tài khoản người dùng
	ErrAccountDuplicate:   "Email đã được sử dụng",
	ErrAccountNotFound:    "Không tìm thấy tài khoản",
	ErrInvalidCredentials: "Thông tin đăng nhập không chính xác",
	ErrAccountSuspended:   "Tài khoản đã bị tạm khóa",
	// In getVietnameseMessage function, add:
ErrInvalidToken:       "Token không hợp lệ",
	// Validation errors - Lỗi xác thực dữ liệu
	ErrInvalidInput:     "Dữ liệu đầu vào không hợp lệ",
	ErrValidationFailed: "Xác thực dữ liệu thất bại",
	ErrMissingField:     "Thiếu thông tin bắt buộc",
	
	// System errors - Lỗi hệ thống
	ErrDatabaseError:      "Lỗi hệ thống cơ sở dữ liệu",
	ErrServiceUnavailable: "Dịch vụ hiện không khả dụng",
	ErrTimeout:            "Yêu cầu đã hết thời gian chờ",
	ErrContextCancelled:   "Yêu cầu đã bị hủy",
	
	// Permission errors - Lỗi quyền truy cập
	ErrUnauthorized: "Chưa được xác thực",
	ErrForbidden:    "Không có quyền truy cập",
	
	// Generic fallback - Lỗi chung
	ErrSystemError: "Lỗi hệ thống",
}

// GetVietnameseMessage returns Vietnamese message for error code
func GetVietnameseMessage(code ErrorCode) string {
	if msg, exists := vietnameseMessages[code]; exists {
		return msg
	}
	return "Lỗi hệ thống" // Default fallback
}

// AddVietnameseMessage allows adding new Vietnamese messages dynamically
func AddVietnameseMessage(code ErrorCode, message string) {
	vietnameseMessages[code] = message
}

// UpdateVietnameseMessage allows updating existing Vietnamese messages
func UpdateVietnameseMessage(code ErrorCode, message string) {
	vietnameseMessages[code] = message
}