

// error_system/validation_helper.go
package error_system

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
)

// ValidationErrorDetails contains detailed validation error information
type ValidationErrorDetails struct {
	Field        string      `json:"field"`
	Value        interface{} `json:"value,omitempty"`
	Tag          string      `json:"tag"`
	Param        string      `json:"param,omitempty"`
	Message      string      `json:"message"`
	MessageVN    string      `json:"message_vn"`
	AllowedValues []string   `json:"allowed_values,omitempty"`
}

// Enhanced ValidationError with detailed information
func EnhancedValidationError(field string, value interface{}, tag string, param string) *AppError {
	details := getValidationDetails(field, value, tag, param)
	
	return &AppError{
		Code:       ErrValidationFailed,
		Message:    details.Message,
		MessageVN:  details.MessageVN,
		HTTPStatus: 400,
		Details: map[string]interface{}{
			"field":          details.Field,
			"value":          details.Value,
			"tag":            details.Tag,
			"message":        details.Message,
			"message_vn":     details.MessageVN,
			"allowed_values": details.AllowedValues,
		},
	}
}

// ProcessValidatorErrors converts go-playground validator errors to meaningful AppErrors
func ProcessValidatorErrors(err error) *AppError {
	if validationErrors, ok := err.(validator.ValidationErrors); ok {
		// Get the first validation error (you can modify to handle multiple)
		firstError := validationErrors[0]
		
		field := firstError.Field()
		value := firstError.Value()
		tag := firstError.Tag()
		param := firstError.Param()
		
		return EnhancedValidationError(field, value, tag, param)
	}
	
	// Fallback for other validation errors
	return ValidationError("request", "Invalid request format")
}

// getValidationDetails returns detailed validation information based on tag
func getValidationDetails(field string, value interface{}, tag string, param string) ValidationErrorDetails {
	switch tag {
	case "required":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' is required", field),
			MessageVN: fmt.Sprintf("Trường '%s' là bắt buộc", field),
		}
		
	case "email":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' must be a valid email address", field),
			MessageVN: fmt.Sprintf("Trường '%s' phải là địa chỉ email hợp lệ", field),
		}
		
	case "min":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Param:     param,
			Message:   fmt.Sprintf("Field '%s' must be at least %s characters long", field, param),
			MessageVN: fmt.Sprintf("Trường '%s' phải có ít nhất %s ký tự", field, param),
		}
		
	case "max":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Param:     param,
			Message:   fmt.Sprintf("Field '%s' must be at most %s characters long", field, param),
			MessageVN: fmt.Sprintf("Trường '%s' không được vượt quá %s ký tự", field, param),
		}
		
	case "oneof":
		allowedValues := strings.Split(param, " ")
		return ValidationErrorDetails{
			Field:         field,
			Value:         value,
			Tag:           tag,
			Param:         param,
			AllowedValues: allowedValues,
			Message:       fmt.Sprintf("Field '%s' must be one of: %s. Received: '%v'", field, strings.Join(allowedValues, ", "), value),
			MessageVN:     fmt.Sprintf("Trường '%s' phải là một trong: %s. Nhận được: '%v'", field, strings.Join(allowedValues, ", "), value),
		}
		
	case "url":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' must be a valid URL", field),
			MessageVN: fmt.Sprintf("Trường '%s' phải là URL hợp lệ", field),
		}
		
	case "numeric":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' must contain only numeric characters", field),
			MessageVN: fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự số", field),
		}
		
	case "alpha":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' must contain only alphabetic characters", field),
			MessageVN: fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự chữ cái", field),
		}
		
	case "alphanum":
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Message:   fmt.Sprintf("Field '%s' must contain only alphanumeric characters", field),
			MessageVN: fmt.Sprintf("Trường '%s' chỉ được chứa các ký tự chữ và số", field),
		}
		
	default:
		return ValidationErrorDetails{
			Field:     field,
			Value:     value,
			Tag:       tag,
			Param:     param,
			Message:   fmt.Sprintf("Field '%s' failed validation constraint '%s'", field, tag),
			MessageVN: fmt.Sprintf("Trường '%s' không thỏa mãn ràng buộc xác thực '%s'", field, tag),
		}
	}
}