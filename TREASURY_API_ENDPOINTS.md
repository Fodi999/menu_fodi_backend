# 💰 Treasury API Endpoints

## Backend Endpoints (Go)

### 1. Get Treasury Balance (Simplified)
```
GET /api/admin/token-bank/treasury
```

**Response:**
```json
{
  "balance": 999992000
}
```

**Authorization:** Requires Admin role

**Used by:** Admin dashboard to display current treasury balance

---

### 2. Get Full Treasury Info
```
GET /api/admin/treasury
```

**Response:**
```json
{
  "id": "uuid",
  "user_id": "TREASURY",
  "balance": 999992000,
  "total_allocated": 8000,
  "total_used": 8000,
  "total_supply": 8000,
  "distributed": 8000,
  "remaining": 999992000,
  "created_at": "2024-12-11T00:00:00Z",
  "updated_at": "2024-12-11T01:00:00Z"
}
```

**Authorization:** Requires Admin role

---

### 3. Real-Time Treasury Stream (SSE)
```
GET /api/treasury/stream
```

**Response Type:** `text/event-stream`

**Authorization:** Requires authentication (not admin-only)

**Events:**
```
data: {"balance": 999992000, "type": "initial"}

data: {"balance": 999991900, "type": "treasury_update", "timestamp": 1702259567}

data: {"balance": 999991800, "type": "treasury_allocate", "amount": 100, "user_id": "user_123"}
```

**Usage:**
```javascript
const eventSource = new EventSource('/api/treasury/stream', {
  headers: {
    'Authorization': `Bearer ${token}`
  }
});

eventSource.onmessage = (event) => {
  const data = JSON.parse(event.data);
  console.log('Treasury balance:', data.balance);
  updateUI(data.balance);
};

eventSource.onerror = () => {
  console.log('SSE connection lost, reconnecting...');
};
```

---

## Frontend Integration

### React Component Example

```typescript
import { useEffect, useState } from 'react';

export function RealTimeTreasuryBalance() {
  const [balance, setBalance] = useState<number | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // 1. Fetch initial balance
    const fetchInitial = async () => {
      try {
        const response = await fetch('/api/admin/token-bank/treasury', {
          headers: {
            'Authorization': `Bearer ${getToken()}`
          }
        });
        
        if (!response.ok) {
          throw new Error(`HTTP ${response.status}`);
        }
        
        const data = await response.json();
        setBalance(data.balance);
      } catch (err) {
        console.error('Error fetching treasury:', err);
        setError('Failed to load treasury balance');
      }
    };

    fetchInitial();

    // 2. Connect to SSE stream
    const eventSource = new EventSource('/api/treasury/stream');

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        setBalance(data.balance);
        setError(null);
      } catch (err) {
        console.error('Error parsing SSE data:', err);
      }
    };

    eventSource.onerror = (err) => {
      console.error('SSE connection error:', err);
      eventSource.close();
      
      // Fallback to polling
      const pollInterval = setInterval(fetchInitial, 5000);
      return () => clearInterval(pollInterval);
    };

    return () => {
      eventSource.close();
    };
  }, []);

  if (error) return <div className="error">{error}</div>;
  if (balance === null) return <div>Loading...</div>;

  return (
    <div className="treasury-balance">
      <h3>Treasury Balance</h3>
      <div className="balance">
        {balance.toLocaleString()} tokens
      </div>
    </div>
  );
}
```

---

## Testing

### Using curl

**1. Test balance endpoint:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/api/admin/token-bank/treasury
```

**2. Test SSE stream:**
```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  -N http://localhost:8080/api/treasury/stream
```

**Expected output:**
```
data: {"balance":999992000,"type":"initial"}

data: {"balance":999991900,"type":"treasury_update","timestamp":1702259567}
```

---

## Troubleshooting

### 404 Not Found Errors

**Problem:**
```
GET http://localhost:3000/api/admin/token-bank/treasury 404 (Not Found)
GET http://localhost:3000/api/treasury/stream 404 (Not Found)
```

**Solutions:**

1. **Check if backend is running:**
```bash
curl http://localhost:8080/health
# Should return: ok
```

2. **Check if you're using correct port:**
   - Backend runs on port **8080** (Go server)
   - Frontend runs on port **3000** (Next.js)
   - Make sure Next.js is proxying to backend

3. **Verify Next.js proxy configuration** (`next.config.js`):
```javascript
module.exports = {
  async rewrites() {
    return [
      {
        source: '/api/:path*',
        destination: 'http://localhost:8080/api/:path*',
      },
    ];
  },
};
```

4. **Check authentication:**
```bash
# Get token first
TOKEN=$(curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"admin@example.com","password":"password"}' \
  | jq -r .token)

# Then use it
curl -H "Authorization: Bearer $TOKEN" \
  http://localhost:8080/api/admin/token-bank/treasury
```

5. **Verify role permissions:**
   - User must have `admin` role for `/api/admin/*` endpoints
   - Check user role in database

---

## CORS Configuration

If you're getting CORS errors, ensure backend has correct CORS settings:

```go
// In routes_modular.go
r.Use(cors.Handler(cors.Options{
    AllowedOrigins:   []string{"http://localhost:3000", "https://yourdomain.com"},
    AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
    AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
    ExposedHeaders:   []string{"Link"},
    AllowCredentials: true,
    MaxAge:           300,
}))
```

---

## Current Status

✅ **Endpoints Implemented:**
- GET `/api/admin/token-bank/treasury` - Simplified balance
- GET `/api/admin/treasury` - Full treasury info
- GET `/api/treasury/stream` - SSE real-time stream

✅ **Features:**
- Real-time balance updates via SSE
- Initial balance fetch
- Fallback to polling on SSE failure
- Authentication required
- Admin-only for balance endpoint

⏳ **TODO:**
- Full EventBus integration for SSE (currently sends initial data only)
- WebSocket alternative to SSE
- Rate limiting for SSE connections

---

**Last Updated:** 11 декабря 2025 г.  
**Commit:** c7aef35  
**Status:** ✅ Ready for frontend integration
