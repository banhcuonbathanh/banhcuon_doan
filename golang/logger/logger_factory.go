// logger/factory.go - Factory constructors for different component loggers
package logger

// Enhanced component loggers with layer support
func NewComponentLogger(component string) *Logger {
	logger := NewLogger()
	logger.SetComponent(component)
	return logger
}

func NewLayerLogger(layer string) *Logger {
	logger := NewLogger()
	logger.SetLayer(layer)
	return logger
}

func NewHandlerLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("handler")
	logger.SetLayer(LayerHandler)
	return logger
}

func NewServiceLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("service")
	logger.SetLayer(LayerService)
	return logger
}

func NewRepositoryLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("repository")
	logger.SetLayer(LayerRepository)
	return logger
}

func NewMiddlewareLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("middleware")
	logger.SetLayer(LayerMiddleware)
	return logger
}

func NewAuthLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("auth")
	logger.SetLayer(LayerAuth)
	return logger
}

func NewValidationLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("validation")
	logger.SetLayer(LayerValidation)
	return logger
}

func NewCacheLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("cache")
	logger.SetLayer(LayerCache)
	return logger
}

func NewDatabaseLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("database")
	logger.SetLayer(LayerDatabase)
	return logger
}

func NewExternalLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("external")
	logger.SetLayer(LayerExternal)
	return logger
}

func NewSecurityLogger() *Logger {
	logger := NewLogger()
	logger.SetComponent("security")
	logger.SetLayer(LayerSecurity)
	return logger
}