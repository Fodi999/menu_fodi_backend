// ============================================================
// Backend API Response Types - TypeScript
// Auto-generated from Go backend types
// Last Updated: 2026-01-11
// ============================================================

/**
 * Standard API response format for all endpoints
 */
export interface APIResponse<T = any> {
  data: T | null;
  error: APIError | null;
  meta: ResponseMeta;
}

/**
 * API error structure
 */
export interface APIError {
  code: ErrorCode;
  message: string;
  details?: string;
}

/**
 * Response metadata
 */
export interface ResponseMeta {
  request_id: string;
  timestamp: string;
  version?: string;
}

/**
 * Paginated data structure
 */
export interface PaginatedData<T> {
  items: T[];
  pagination: PaginationInfo;
}

/**
 * Pagination metadata
 */
export interface PaginationInfo {
  page: number;
  limit: number;
  total: number;
  total_pages: number;
}

// ============================================================
// Error Codes
// ============================================================

export type ErrorCode =
  // Authentication Errors
  | 'AUTH_INVALID_TOKEN'
  | 'AUTH_MISSING_TOKEN'
  | 'AUTH_INVALID_CREDENTIALS'
  | 'AUTH_USER_EXISTS'
  | 'AUTH_INSUFFICIENT_PERMISSIONS'
  | 'AUTH_EXPIRED_TOKEN'
  
  // Ingredient Errors
  | 'INGREDIENT_NOT_FOUND'
  | 'INGREDIENT_INVALID_INPUT'
  | 'INGREDIENT_ALREADY_EXISTS'
  | 'INGREDIENT_INVALID_UNIT'
  | 'INGREDIENT_AI_FAILED'
  
  // Recipe Errors
  | 'RECIPE_NOT_FOUND'
  | 'RECIPE_INVALID_INPUT'
  | 'RECIPE_AI_GENERATION_FAILED'
  | 'RECIPE_VALIDATION_FAILED'
  | 'RECIPE_INSUFFICIENT_INGREDIENTS'
  | 'RECIPE_ALREADY_SAVED'
  
  // Fridge Errors
  | 'FRIDGE_ITEM_NOT_FOUND'
  | 'FRIDGE_INSUFFICIENT_QUANTITY'
  | 'FRIDGE_INVALID_INPUT'
  | 'FRIDGE_ITEM_EXPIRED'
  | 'FRIDGE_INVALID_PRICE'
  
  // Token Economy Errors
  | 'TOKEN_INSUFFICIENT_BALANCE'
  | 'TOKEN_INVALID_AMOUNT'
  | 'TOKEN_TRANSACTION_FAILED'
  | 'TOKEN_BANK_NOT_FOUND'
  
  // User Errors
  | 'USER_NOT_FOUND'
  | 'USER_INVALID_INPUT'
  | 'USER_ALREADY_EXISTS'
  | 'USER_INVALID_EMAIL'
  | 'USER_INVALID_ROLE'
  
  // Marketplace Errors
  | 'MARKETPLACE_RECIPE_NOT_FOUND'
  | 'MARKETPLACE_INSUFFICIENT_TOKENS'
  | 'MARKETPLACE_PURCHASE_FAILED'
  | 'MARKETPLACE_ALREADY_PURCHASED'
  
  // General Errors
  | 'GENERAL_INVALID_INPUT'
  | 'GENERAL_INVALID_JSON'
  | 'GENERAL_INTERNAL_ERROR'
  | 'GENERAL_NOT_FOUND'
  | 'GENERAL_DATABASE_ERROR'
  | 'GENERAL_MISSING_PARAMETER'
  
  // Validation Errors
  | 'VALIDATION_FAILED'
  | 'VALIDATION_INVALID_EMAIL'
  | 'VALIDATION_INVALID_URL'
  
  // Upload Errors
  | 'UPLOAD_INVALID_FILE'
  | 'UPLOAD_FILE_TOO_LARGE'
  | 'UPLOAD_INVALID_FORMAT'
  | 'UPLOAD_FAILED';

// ============================================================
// Helper Functions
// ============================================================

/**
 * Type guard to check if response is success
 */
export function isSuccess<T>(response: APIResponse<T>): response is APIResponse<T> & { data: T } {
  return response.error === null && response.data !== null;
}

/**
 * Type guard to check if response is error
 */
export function isError<T>(response: APIResponse<T>): response is APIResponse<T> & { error: APIError } {
  return response.error !== null;
}

/**
 * Extract data from response or throw error
 */
export function unwrapResponse<T>(response: APIResponse<T>): T {
  if (isError(response)) {
    throw new APIResponseError(response.error, response.meta);
  }
  if (response.data === null) {
    throw new Error('Response data is null');
  }
  return response.data;
}

/**
 * Custom error class for API errors
 */
export class APIResponseError extends Error {
  constructor(
    public readonly apiError: APIError,
    public readonly meta: ResponseMeta
  ) {
    super(`[${apiError.code}] ${apiError.message}`);
    this.name = 'APIResponseError';
  }

  get code(): ErrorCode {
    return this.apiError.code;
  }

  get details(): string | undefined {
    return this.apiError.details;
  }

  get requestId(): string {
    return this.meta.request_id;
  }

  /**
   * Format error for logging
   */
  toLogFormat(): string {
    return `[${this.meta.request_id}] ${this.apiError.code}: ${this.apiError.message}${
      this.details ? ` (${this.details})` : ''
    }`;
  }
}

// ============================================================
// API Client Helper
// ============================================================

export interface FetchOptions extends RequestInit {
  token?: string;
  requestId?: string;
  language?: 'en' | 'pl' | 'ru';
}

/**
 * Enhanced fetch wrapper with automatic error handling
 */
export async function apiFetch<T = any>(
  url: string,
  options: FetchOptions = {}
): Promise<APIResponse<T>> {
  const { token, requestId, language, headers: customHeaders, ...fetchOptions } = options;

  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...(customHeaders as Record<string, string>),
  };

  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }

  if (requestId) {
    headers['X-Request-ID'] = requestId;
  }

  if (language) {
    headers['Accept-Language'] = language;
  }

  try {
    const response = await fetch(url, {
      ...fetchOptions,
      headers,
    });

    const data: APIResponse<T> = await response.json();

    // Log request ID for debugging
    console.debug(`[${data.meta.request_id}] ${fetchOptions.method || 'GET'} ${url}`);

    return data;
  } catch (error) {
    // Network error or invalid JSON
    throw new Error(`Network error: ${error instanceof Error ? error.message : 'Unknown error'}`);
  }
}

/**
 * Enhanced fetch wrapper that throws on error
 */
export async function apiFetchUnwrap<T = any>(
  url: string,
  options: FetchOptions = {}
): Promise<T> {
  const response = await apiFetch<T>(url, options);
  return unwrapResponse(response);
}

// ============================================================
// Usage Examples
// ============================================================

/*

// Example 1: Basic usage with type safety
const response = await apiFetch<Ingredient>('/api/ingredients/123');

if (isSuccess(response)) {
  console.log(response.data.name); // TypeScript knows data exists
} else {
  console.error(response.error.code); // TypeScript knows error exists
}

// Example 2: Using unwrap (throws on error)
try {
  const ingredient = await apiFetchUnwrap<Ingredient>('/api/ingredients/123');
  console.log(ingredient.name); // Direct access, no null check needed
} catch (error) {
  if (error instanceof APIResponseError) {
    console.error(error.toLogFormat());
    
    // Handle specific error codes
    switch (error.code) {
      case 'INGREDIENT_NOT_FOUND':
        showNotFoundUI();
        break;
      case 'AUTH_INVALID_TOKEN':
        redirectToLogin();
        break;
      default:
        showGenericError(error.message);
    }
  }
}

// Example 3: With authentication
const response = await apiFetch<Recipe[]>('/api/recipes', {
  token: userToken,
  language: 'pl',
  requestId: 'my-custom-id-123',
});

// Example 4: POST request
const response = await apiFetch<Recipe>('/api/recipes', {
  method: 'POST',
  body: JSON.stringify({
    title: 'New Recipe',
    ingredients: [...]
  }),
  token: userToken,
});

// Example 5: Paginated data
const response = await apiFetch<PaginatedData<Ingredient>>('/api/ingredients?page=1&limit=20');

if (isSuccess(response)) {
  const { items, pagination } = response.data;
  console.log(`Page ${pagination.page} of ${pagination.total_pages}`);
  console.log(`Total items: ${pagination.total}`);
  items.forEach(item => console.log(item.name));
}

// Example 6: Error tracking with Sentry
try {
  const data = await apiFetchUnwrap('/api/recipes/123');
} catch (error) {
  if (error instanceof APIResponseError) {
    Sentry.captureException(error, {
      extra: {
        requestId: error.requestId,
        errorCode: error.code,
        details: error.details,
      },
    });
  }
}

*/
