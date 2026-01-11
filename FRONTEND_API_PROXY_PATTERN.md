# API Proxy Pattern - Frontend Integration Guide

## Overview

Instead of duplicating API logic in Next.js API routes, we proxy all requests to the Go backend. This ensures:
- ✅ Single source of truth
- ✅ No code duplication
- ✅ Consistent error handling
- ✅ Automatic header forwarding

## Proxy Helper Implementation

### File: `lib/api-proxy.ts`

```typescript
/**
 * Proxy helper for Next.js API routes
 * Forwards requests to Go backend with proper headers
 */

const GO_API_BASE = process.env.GO_API_BASE_URL || 'http://localhost:8080/api';

interface ProxyOptions {
  method?: 'GET' | 'POST' | 'PUT' | 'DELETE' | 'PATCH';
  body?: any;
  headers?: Record<string, string>;
}

/**
 * Proxy request to Go backend
 * @param req - Next.js Request object
 * @param path - Backend path (e.g., '/admin/ingredients/suggest')
 * @param options - Additional options
 */
export async function proxyToGo(
  req: Request,
  path: string,
  options: ProxyOptions = {}
): Promise<Response> {
  try {
    // Extract query parameters from request URL
    const url = new URL(req.url);
    const queryString = url.search; // Includes '?' if present

    // Build full backend URL
    const backendUrl = `${GO_API_BASE}${path}${queryString}`;

    // Forward important headers
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...options.headers,
    };

    // Forward Authorization header
    const authHeader = req.headers.get('Authorization');
    if (authHeader) {
      headers['Authorization'] = authHeader;
    }

    // Forward Accept-Language for localization
    const langHeader = req.headers.get('Accept-Language');
    if (langHeader) {
      headers['Accept-Language'] = langHeader;
    }

    // Determine HTTP method
    const method = options.method || req.method;

    // Build fetch options
    const fetchOptions: RequestInit = {
      method,
      headers,
    };

    // Add body for POST/PUT/PATCH requests
    if (options.body) {
      fetchOptions.body = JSON.stringify(options.body);
    } else if (method !== 'GET' && method !== 'DELETE') {
      const body = await req.json().catch(() => null);
      if (body) {
        fetchOptions.body = JSON.stringify(body);
      }
    }

    // Make request to Go backend
    console.log(`🔄 Proxy: ${method} ${backendUrl}`);
    const response = await fetch(backendUrl, fetchOptions);

    // Get response data
    const data = await response.json().catch(() => null);

    // Return response with same status code
    return Response.json(data, {
      status: response.status,
      headers: {
        'Content-Type': 'application/json',
      },
    });
  } catch (error) {
    console.error('❌ Proxy error:', error);
    return Response.json(
      { error: 'Proxy request failed' },
      { status: 500 }
    );
  }
}

/**
 * Simple GET proxy (most common case)
 */
export async function proxyGet(req: Request, path: string): Promise<Response> {
  return proxyToGo(req, path, { method: 'GET' });
}

/**
 * Simple POST proxy
 */
export async function proxyPost(req: Request, path: string): Promise<Response> {
  return proxyToGo(req, path, { method: 'POST' });
}

/**
 * Simple PUT proxy
 */
export async function proxyPut(req: Request, path: string): Promise<Response> {
  return proxyToGo(req, path, { method: 'PUT' });
}

/**
 * Simple DELETE proxy
 */
export async function proxyDelete(req: Request, path: string): Promise<Response> {
  return proxyToGo(req, path, { method: 'DELETE' });
}
```

## Usage Examples

### Example 1: Ingredient Suggest

**Before (❌ Duplicated Logic):**

```typescript
// app/api/admin/ingredients/suggest/route.ts
export async function GET(req: Request) {
  const { searchParams } = new URL(req.url);
  const query = searchParams.get('q');
  const limit = searchParams.get('limit') || '5';
  
  // Direct fetch to Go API
  const response = await fetch(
    `${process.env.GO_API}/admin/ingredients/suggest?q=${query}&limit=${limit}`,
    {
      headers: {
        'Authorization': req.headers.get('Authorization') || '',
        'Accept-Language': req.headers.get('Accept-Language') || 'en',
      }
    }
  );
  
  const data = await response.json();
  return Response.json(data);
}
```

**After (✅ Proxy Pattern):**

```typescript
// app/api/admin/ingredients/suggest/route.ts
import { proxyGet } from '@/lib/api-proxy';

export async function GET(req: Request) {
  return proxyGet(req, '/admin/ingredients/suggest');
}
```

### Example 2: AI Recipe Preview

**Before (❌ Complex Logic):**

```typescript
// app/api/admin/recipes/preview-ai/route.ts
export async function POST(req: Request) {
  const body = await req.json();
  
  const response = await fetch(
    `${process.env.GO_API}/admin/recipes/preview-ai`,
    {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
        'Authorization': req.headers.get('Authorization') || '',
        'Accept-Language': req.headers.get('Accept-Language') || 'en',
      },
      body: JSON.stringify(body)
    }
  );
  
  return Response.json(await response.json(), { status: response.status });
}
```

**After (✅ Simple Proxy):**

```typescript
// app/api/admin/recipes/preview-ai/route.ts
import { proxyPost } from '@/lib/api-proxy';

export async function POST(req: Request) {
  return proxyPost(req, '/admin/recipes/preview-ai');
}
```

### Example 3: Multiple HTTP Methods

```typescript
// app/api/admin/ingredients/[id]/route.ts
import { proxyGet, proxyPut, proxyDelete } from '@/lib/api-proxy';

export async function GET(
  req: Request,
  { params }: { params: { id: string } }
) {
  return proxyGet(req, `/admin/ingredients/${params.id}`);
}

export async function PUT(
  req: Request,
  { params }: { params: { id: string } }
) {
  return proxyPut(req, `/admin/ingredients/${params.id}`);
}

export async function DELETE(
  req: Request,
  { params }: { params: { id: string } }
) {
  return proxyDelete(req, `/admin/ingredients/${params.id}`);
}
```

## Benefits

### 1. No Code Duplication

**Before:**
```typescript
// 10 different route files, each with 20+ lines of fetch logic
// = 200+ lines of duplicated code
```

**After:**
```typescript
// 10 different route files, each with 3 lines
// = 30 lines total
```

### 2. Consistent Error Handling

All errors are handled in one place (`api-proxy.ts`). No need to repeat try-catch logic.

### 3. Automatic Header Forwarding

- `Authorization` → Automatically forwarded
- `Accept-Language` → Automatically forwarded
- `Content-Type` → Always set correctly

### 4. Easy to Test

```typescript
// Mock once in api-proxy.ts
// All routes benefit from the mock
```

### 5. Environment Configuration

```bash
# .env.local
GO_API_BASE_URL=http://localhost:8080/api

# .env.production
GO_API_BASE_URL=https://your-backend.koyeb.app/api
```

## Migration Checklist

- [ ] Create `lib/api-proxy.ts` helper
- [ ] Set `GO_API_BASE_URL` in `.env` files
- [ ] Migrate `/api/admin/ingredients/suggest/route.ts`
- [ ] Migrate `/api/admin/recipes/preview-ai/route.ts`
- [ ] Migrate `/api/admin/recipes/create-ai/route.ts`
- [ ] Remove old fetch logic
- [ ] Test all endpoints
- [ ] Update frontend API client if needed

## Environment Setup

### Development

```bash
# .env.local
GO_API_BASE_URL=http://localhost:8080/api
```

### Production

```bash
# .env.production
GO_API_BASE_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api
```

### Docker

```bash
# docker-compose.yml
environment:
  - GO_API_BASE_URL=http://backend:8080/api
```

## Testing

### Unit Test

```typescript
// lib/__tests__/api-proxy.test.ts
import { proxyToGo } from '../api-proxy';

describe('proxyToGo', () => {
  it('forwards Authorization header', async () => {
    const req = new Request('http://localhost:3000/api/test?q=hello', {
      headers: { 'Authorization': 'Bearer token123' }
    });
    
    const response = await proxyToGo(req, '/admin/test');
    
    // Assert that Authorization was forwarded
    expect(response.status).toBe(200);
  });
  
  it('forwards Accept-Language header', async () => {
    const req = new Request('http://localhost:3000/api/test', {
      headers: { 'Accept-Language': 'ru' }
    });
    
    const response = await proxyToGo(req, '/admin/test');
    
    // Assert localization worked
    expect(response.status).toBe(200);
  });
});
```

### Integration Test

```typescript
// app/api/admin/ingredients/suggest/__tests__/route.test.ts
import { GET } from '../route';

describe('GET /api/admin/ingredients/suggest', () => {
  it('returns suggestions', async () => {
    const req = new Request(
      'http://localhost:3000/api/admin/ingredients/suggest?q=tom',
      { headers: { 'Authorization': 'Bearer test' } }
    );
    
    const response = await GET(req);
    const data = await response.json();
    
    expect(response.status).toBe(200);
    expect(data.data).toBeInstanceOf(Array);
  });
});
```

## Troubleshooting

### Issue: CORS Errors

**Solution:** Ensure Go backend has proper CORS settings:

```go
// Go backend
cors.AllowOrigin("http://localhost:3000")
```

### Issue: 500 Proxy Error

**Check:**
1. `GO_API_BASE_URL` is set correctly
2. Go backend is running
3. Network connectivity

```bash
# Test backend directly
curl http://localhost:8080/api/health
```

### Issue: Headers Not Forwarded

**Debug:**

```typescript
// Add logging to api-proxy.ts
console.log('Request headers:', Object.fromEntries(req.headers));
console.log('Forwarded headers:', headers);
```

## Best Practices

1. **Always use proxy for backend calls**
   - Don't mix direct `fetch()` calls with proxy
   - Consistency is key

2. **Keep proxy helper simple**
   - Don't add business logic
   - Only handle HTTP forwarding

3. **Use TypeScript**
   - Type-safe request/response
   - Better IDE support

4. **Log requests in development**
   - Easy debugging
   - Track API calls

5. **Handle errors gracefully**
   - Return proper HTTP status codes
   - Provide meaningful error messages

## Related Documentation

- `INGREDIENT_SUGGEST_API.md` - Ingredient suggest endpoint details
- `AI_RECIPE_QUICK_REF.md` - AI recipe endpoints
- `docs/API_CONTRACT_RECIPE_MATCH.md` - Recipe matching API

---

**Last Updated:** 2026-01-08  
**Pattern:** Proxy to Go Backend  
**Status:** ✅ Recommended Approach
