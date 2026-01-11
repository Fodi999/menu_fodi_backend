package models

// Error codes for the API
// Format: DOMAIN_ERROR_TYPE

// ============================================================
// Authentication Errors (AUTH_*)
// ============================================================
const (
	ErrorAuthInvalidToken             = "AUTH_INVALID_TOKEN"
	ErrorAuthMissingToken             = "AUTH_MISSING_TOKEN"
	ErrorAuthInvalidCredentials       = "AUTH_INVALID_CREDENTIALS"
	ErrorAuthUserExists               = "AUTH_USER_EXISTS"
	ErrorAuthInsufficientPermissions  = "AUTH_INSUFFICIENT_PERMISSIONS"
	ErrorAuthExpiredToken             = "AUTH_EXPIRED_TOKEN"
	ErrorAuthInvalidRefreshToken      = "AUTH_INVALID_REFRESH_TOKEN"
)

// ============================================================
// Ingredient Errors (INGREDIENT_*)
// ============================================================
const (
	ErrorIngredientNotFound      = "INGREDIENT_NOT_FOUND"
	ErrorIngredientInvalidInput  = "INGREDIENT_INVALID_INPUT"
	ErrorIngredientAlreadyExists = "INGREDIENT_ALREADY_EXISTS"
	ErrorIngredientInvalidUnit   = "INGREDIENT_INVALID_UNIT"
	ErrorIngredientAIFailed      = "INGREDIENT_AI_CLASSIFICATION_FAILED"
)

// ============================================================
// Recipe Errors (RECIPE_*)
// ============================================================
const (
	ErrorRecipeNotFound             = "RECIPE_NOT_FOUND"
	ErrorRecipeInvalidInput         = "RECIPE_INVALID_INPUT"
	ErrorRecipeAIGenerationFailed   = "RECIPE_AI_GENERATION_FAILED"
	ErrorRecipeValidationFailed     = "RECIPE_VALIDATION_FAILED"
	ErrorRecipeInsufficientIngredients = "RECIPE_INSUFFICIENT_INGREDIENTS"
	ErrorRecipeAlreadySaved         = "RECIPE_ALREADY_SAVED"
	ErrorRecipeNotSaved             = "RECIPE_NOT_SAVED"
)

// ============================================================
// Fridge Errors (FRIDGE_*)
// ============================================================
const (
	ErrorFridgeItemNotFound         = "FRIDGE_ITEM_NOT_FOUND"
	ErrorFridgeInsufficientQuantity = "FRIDGE_INSUFFICIENT_QUANTITY"
	ErrorFridgeInvalidInput         = "FRIDGE_INVALID_INPUT"
	ErrorFridgeItemExpired          = "FRIDGE_ITEM_EXPIRED"
	ErrorFridgeInvalidPrice         = "FRIDGE_INVALID_PRICE"
)

// ============================================================
// Token Economy Errors (TOKEN_*)
// ============================================================
const (
	ErrorTokenInsufficientBalance = "TOKEN_INSUFFICIENT_BALANCE"
	ErrorTokenInvalidAmount       = "TOKEN_INVALID_AMOUNT"
	ErrorTokenTransactionFailed   = "TOKEN_TRANSACTION_FAILED"
	ErrorTokenBankNotFound        = "TOKEN_BANK_NOT_FOUND"
	ErrorTokenAllocationFailed    = "TOKEN_ALLOCATION_FAILED"
)

// ============================================================
// User Errors (USER_*)
// ============================================================
const (
	ErrorUserNotFound      = "USER_NOT_FOUND"
	ErrorUserInvalidInput  = "USER_INVALID_INPUT"
	ErrorUserAlreadyExists = "USER_ALREADY_EXISTS"
	ErrorUserInvalidEmail  = "USER_INVALID_EMAIL"
	ErrorUserInvalidRole   = "USER_INVALID_ROLE"
)

// ============================================================
// Marketplace Errors (MARKETPLACE_*)
// ============================================================
const (
	ErrorMarketplaceRecipeNotFound  = "MARKETPLACE_RECIPE_NOT_FOUND"
	ErrorMarketplaceInsufficientTokens = "MARKETPLACE_INSUFFICIENT_TOKENS"
	ErrorMarketplacePurchaseFailed  = "MARKETPLACE_PURCHASE_FAILED"
	ErrorMarketplaceAlreadyPurchased = "MARKETPLACE_ALREADY_PURCHASED"
	ErrorMarketplaceInvalidPrice    = "MARKETPLACE_INVALID_PRICE"
)

// ============================================================
// Academy Errors (ACADEMY_*)
// ============================================================
const (
	ErrorAcademyCourseNotFound     = "ACADEMY_COURSE_NOT_FOUND"
	ErrorAcademyAlreadyEnrolled    = "ACADEMY_ALREADY_ENROLLED"
	ErrorAcademyEnrollmentFailed   = "ACADEMY_ENROLLMENT_FAILED"
	ErrorAcademyLessonNotFound     = "ACADEMY_LESSON_NOT_FOUND"
	ErrorAcademyQuizNotFound       = "ACADEMY_QUIZ_NOT_FOUND"
)

// ============================================================
// General Errors (GENERAL_*)
// ============================================================
const (
	ErrorGeneralInvalidInput     = "GENERAL_INVALID_INPUT"
	ErrorGeneralInternalError    = "GENERAL_INTERNAL_ERROR"
	ErrorGeneralNotFound         = "GENERAL_NOT_FOUND"
	ErrorGeneralDatabaseError    = "GENERAL_DATABASE_ERROR"
	ErrorGeneralInvalidJSON      = "GENERAL_INVALID_JSON"
	ErrorGeneralMissingParameter = "GENERAL_MISSING_PARAMETER"
	ErrorGeneralInvalidParameter = "GENERAL_INVALID_PARAMETER"
)

// ============================================================
// Validation Errors (VALIDATION_*)
// ============================================================
const (
	ErrorValidationFailed        = "VALIDATION_FAILED"
	ErrorValidationInvalidEmail  = "VALIDATION_INVALID_EMAIL"
	ErrorValidationInvalidURL    = "VALIDATION_INVALID_URL"
	ErrorValidationMinLength     = "VALIDATION_MIN_LENGTH"
	ErrorValidationMaxLength     = "VALIDATION_MAX_LENGTH"
	ErrorValidationInvalidFormat = "VALIDATION_INVALID_FORMAT"
)

// ============================================================
// File Upload Errors (UPLOAD_*)
// ============================================================
const (
	ErrorUploadInvalidFile     = "UPLOAD_INVALID_FILE"
	ErrorUploadFileTooLarge    = "UPLOAD_FILE_TOO_LARGE"
	ErrorUploadInvalidFormat   = "UPLOAD_INVALID_FORMAT"
	ErrorUploadFailed          = "UPLOAD_FAILED"
)

// ============================================================
// Localization Errors (LOCALIZATION_*)
// ============================================================
const (
	ErrorLocalizationInvalidLanguage = "LOCALIZATION_INVALID_LANGUAGE"
	ErrorLocalizationNotSupported    = "LOCALIZATION_NOT_SUPPORTED"
)
