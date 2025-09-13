package account_handler

import (
	"fmt"
	"strings"
	"english-ai-full/error_system"
	"github.com/go-playground/validator/v10"
)

func (h *AccountHandler) handleValidationError(err error) *error_system.AppError {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// Get the first validation error
		firstError := validationErrors[0]
		
		field := firstError.Field()
		value := firstError.Value()
		tag := firstError.Tag()
		param := firstError.Param()
		
		// Create more descriptive error messages based on validation tag
		switch tag {
		case "required":
			message := fmt.Sprintf("The '%s' field is required", field)
			messageVN := fmt.Sprintf("Trường '%s' là bắt buộc", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"constraint": tag,
			}
			return appErr

		case "email":
			message := fmt.Sprintf("The '%s' field must be a valid email address", field)
			messageVN := fmt.Sprintf("Trường '%s' phải là địa chỉ email hợp lệ", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"value":      maskSensitiveValue(field, value),
				"constraint": tag,
			}
			return appErr

		case "min":
			message := fmt.Sprintf("The '%s' field must be at least %s characters long", field, param)
			messageVN := fmt.Sprintf("Trường '%s' phải có ít nhất %s ký tự", field, param)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"constraint": tag,
				"min_length": param,
			}
			return appErr

		case "max":
			message := fmt.Sprintf("The '%s' field must be at most %s characters long", field, param)
			messageVN := fmt.Sprintf("Trường '%s' không được vượt quá %s ký tự", field, param)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"constraint": tag,
				"max_length": param,
			}
			return appErr

		case "oneof":
			allowedValues := strings.Split(param, " ")
			allowedValuesStr := strings.Join(allowedValues, ", ")
			message := fmt.Sprintf("The '%s' field must be one of: %s (received: %v)", field, allowedValuesStr, value)
			messageVN := fmt.Sprintf("Trường '%s' phải là một trong: %s (nhận được: %v)", field, allowedValuesStr, value)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":          field,
				"value":          value,
				"constraint":     tag,
				"allowed_values": allowedValues,
			}
			return appErr

		case "url":
			message := fmt.Sprintf("The '%s' field must be a valid URL", field)
			messageVN := fmt.Sprintf("Trường '%s' phải là URL hợp lệ", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"value":      maskSensitiveValue(field, value),
				"constraint": tag,
			}
			return appErr

		case "numeric":
			message := fmt.Sprintf("The '%s' field must contain only numeric characters", field)
			messageVN := fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự số", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"value":      maskSensitiveValue(field, value),
				"constraint": tag,
			}
			return appErr

		case "alpha":
			message := fmt.Sprintf("The '%s' field must contain only alphabetic characters", field)
			messageVN := fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự chữ cái", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"value":      maskSensitiveValue(field, value),
				"constraint": tag,
			}
			return appErr

		case "alphanum":
			message := fmt.Sprintf("The '%s' field must contain only alphanumeric characters", field)
			messageVN := fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự chữ và số", field)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"value":      maskSensitiveValue(field, value),
				"constraint": tag,
			}
			return appErr

		case "len":
			message := fmt.Sprintf("The '%s' field must be exactly %s characters long", field, param)
			messageVN := fmt.Sprintf("Trường '%s' phải có đúng %s ký tự", field, param)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":          field,
				"constraint":     tag,
				"required_length": param,
			}
			return appErr

		default:
			// For unknown validation tags, provide a generic but informative message
			message := fmt.Sprintf("The '%s' field failed validation (constraint: %s)", field, tag)
			messageVN := fmt.Sprintf("Trường '%s' không thỏa mãn ràng buộc xác thực (%s)", field, tag)
			appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
			appErr.Details = map[string]interface{}{
				"field":      field,
				"constraint": tag,
				"param":      param,
			}
			if param != "" {
				appErr.Details["param"] = param
			}
			return appErr
		}
	}

	// Handle other types of validation errors
	errorMsg := err.Error()
	
	// Check for common JSON parsing errors
	if strings.Contains(strings.ToLower(errorMsg), "json") {
		message := "Invalid JSON format in request body"
		messageVN := "Định dạng JSON trong request body không hợp lệ"
		appErr := error_system.NewErrorWithMessages(error_system.ErrInvalidInput, message, messageVN)
		appErr.Details = map[string]interface{}{
			"error_type": "json_parse_error",
			"hint":       "Please check your JSON syntax",
		}
		return appErr
	}

	// Check for field type mismatch errors
	if strings.Contains(strings.ToLower(errorMsg), "cannot unmarshal") {
		message := "Invalid data type in request body"
		messageVN := "Kiểu dữ liệu trong request body không hợp lệ"
		appErr := error_system.NewErrorWithMessages(error_system.ErrInvalidInput, message, messageVN)
		appErr.Details = map[string]interface{}{
			"error_type": "type_mismatch",
			"hint":       "Please check that all field types match the expected format",
		}
		return appErr
	}

	// Fallback for truly unknown validation errors
	message := "Request validation failed"
	messageVN := "Xác thực request thất bại"
	appErr := error_system.NewErrorWithMessages(error_system.ErrValidationFailed, message, messageVN)
	appErr.Details = map[string]interface{}{
		"error_type":    "unknown_validation_error",
		"original_error": errorMsg,
	}
	return appErr
}

// maskSensitiveValue masks sensitive field values for logging/error responses
func maskSensitiveValue(field string, value interface{}) interface{} {
	fieldLower := strings.ToLower(field)
	
	// Mask password fields
	if strings.Contains(fieldLower, "password") {
		return "***"
	}
	
	// Mask email partially
	if strings.Contains(fieldLower, "email") {
		if strValue, ok := value.(string); ok {
			return maskEmail(strValue)
		}
	}
	
	// Return original value for non-sensitive fields
	return value
}

