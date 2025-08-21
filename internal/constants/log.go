package constants

// Log Generation Constants
const (
	// Sample Services
	ServiceAPIGateway          = "api-gateway"
	ServiceUserService         = "user-service"
	ServicePaymentService      = "payment-service"
	ServiceOrderService        = "order-service"
	ServiceNotificationService = "notification-service"

	// HTTP Methods
	MethodGET    = "GET"
	MethodPOST   = "POST"
	MethodPUT    = "PUT"
	MethodDELETE = "DELETE"

	// HTTP Paths
	PathAPIUsers    = "/api/users"
	PathAPIOrders   = "/api/orders"
	PathAPIPayments = "/api/payments"
	PathAPIProducts = "/api/products"
	PathAPIAuth     = "/api/auth"

	// Response Status Codes
	StatusOK    = 200
	StatusError = 500

	// Response Time Ranges (in milliseconds)
	MinResponseTime = 10
	MaxResponseTime = 2010

	// User ID Format
	UserIDFormat = "user_%d"
	MaxUserID    = 1000

	// Log Generation Timing
	LogGenerationInterval = 1 // seconds
	MaxLogsPerSecond      = 5

	// Log Configuration
	DefaultLogLevel  = "info"
	DefaultLogFormat = "json"

	// Environment Variable Keys
	EnvKeyLogLevel  = "LOG_LEVEL"
	EnvKeyLogFormat = "LOG_FORMAT"
)

// Log Message Templates
const (
	// Debug Messages
	DebugMessageTemplate = "Debug: Processing %s request to %s"

	// Info Messages
	InfoMessageTemplate = "Info: %s request to %s completed successfully"

	// Warning Messages
	WarningMessageTemplate = "Warning: Slow response time detected for %s %s"

	// Error Messages
	ErrorMessageTemplate = "Error: Failed to process %s request to %s"
	FatalMessageTemplate = "Fatal: Critical error in %s service"

	// Error Message Varieties
	ErrorDatabaseConnection = "Database connection failed"
	ErrorExternalTimeout    = "External service timeout"
	ErrorInvalidPayload     = "Invalid request payload"
	ErrorAuthentication     = "Authentication failed"
	ErrorResourceNotFound   = "Resource not found"
	ErrorInternalServer     = "Internal server error"
	ErrorRateLimit          = "Rate limit exceeded"
)

// Response Time Ranges for Different Log Levels
const (
	// Warning level - slow responses
	WarningMinResponseTime = 1000
	WarningMaxResponseTime = 2000

	// Error level - very slow responses
	ErrorMinResponseTime = 2000
	ErrorMaxResponseTime = 2500

	// Fatal level - extremely slow responses
	FatalMinResponseTime = 3000
	FatalMaxResponseTime = 4000
)
