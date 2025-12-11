# 🔌 WebSocket Real-Time Event System

## Overview

WebSocket система для real-time обновлений токенов, Treasury и задач. Клиенты подключаются к WebSocket и получают мгновенные уведомления о любых изменениях в token economy.

## Architecture

```
┌─────────────────┐
│  Token          │
│  Operations     │──┐
└─────────────────┘  │
                     │  Publish
┌─────────────────┐  │  Events
│  Task           │──┤
│  Operations     │  │
└─────────────────┘  │
                     ▼
              ┌──────────────┐
              │  EventBus    │
              │  (Pub/Sub)   │
              └──────────────┘
                     │
                     │ Broadcast
                     ▼
              ┌──────────────┐
              │  WebSocket   │
              │     Hub      │
              └──────────────┘
                     │
        ┌────────────┼────────────┐
        ▼            ▼            ▼
    ┌────────┐  ┌────────┐  ┌────────┐
    │Client 1│  │Client 2│  │Client N│
    └────────┘  └────────┘  └────────┘
```

## WebSocket Endpoints

### 1. General WebSocket Connection
```
GET ws://localhost:8080/ws
```
Универсальное подключение для всех типов событий.

### 2. Treasury Updates (Admin)
```
GET ws://localhost:8080/ws/treasury
```
Автоматически подписывается на:
- `treasury_update`
- `treasury_allocate`
- `treasury_spend`

### 3. User Token Updates
```
GET ws://localhost:8080/ws/tokens/{userID}
```
Автоматически подписывается на:
- `token_balance_update`
- `token_earn`
- `token_spend`
- `task_completed`
- `task_reward_claimed`

### 4. WebSocket Statistics
```
GET http://localhost:8080/ws/stats
```
HTTP endpoint для получения статистики подключений.

## Event Types

### Treasury Events

#### `treasury_update`
Общее обновление баланса Treasury.
```json
{
  "type": "treasury_update",
  "timestamp": 1702259567,
  "data": {
    "balance": 999990000,
    "total_used": 10000,
    "remaining": 999990000
  }
}
```

#### `treasury_allocate`
Treasury выделил токены пользователю.
```json
{
  "type": "treasury_allocate",
  "timestamp": 1702259567,
  "data": {
    "balance": 999990000,
    "amount": 100,
    "user_id": "user_123",
    "operation": "allocate"
  }
}
```

#### `treasury_spend`
Токены вернулись в Treasury (пользователь потратил).
```json
{
  "type": "treasury_spend",
  "timestamp": 1702259567,
  "data": {
    "balance": 999990050,
    "amount": 50,
    "user_id": "user_123",
    "operation": "return"
  }
}
```

### Token Events

#### `token_balance_update`
Общее обновление баланса пользователя.
```json
{
  "type": "token_balance_update",
  "timestamp": 1702259567,
  "user_id": "user_123",
  "data": {
    "user_id": "user_123",
    "balance_before": 0,
    "balance_after": 100,
    "amount": 100,
    "reason": "allocated_from_treasury",
    "type": "earn"
  }
}
```

#### `token_earn`
Пользователь получил токены.
```json
{
  "type": "token_earn",
  "timestamp": 1702259567,
  "user_id": "user_123",
  "data": {
    "amount": 50,
    "reason": "task_reward",
    "balance_after": 150
  }
}
```

#### `token_spend`
Пользователь потратил токены.
```json
{
  "type": "token_spend",
  "timestamp": 1702259567,
  "user_id": "user_123",
  "data": {
    "user_id": "user_123",
    "balance_before": 100,
    "balance_after": 50,
    "amount": 50,
    "reason": "ai_request",
    "type": "spend"
  }
}
```

### Task Events

#### `task_completed`
Задача выполнена пользователем.
```json
{
  "type": "task_completed",
  "timestamp": 1702259567,
  "user_id": "user_123",
  "data": {
    "user_id": "user_123",
    "task_id": "task_456",
    "task_name": "Complete first recipe",
    "reward": 50,
    "status": "completed"
  }
}
```

#### `task_reward_claimed`
Награда за задачу получена.
```json
{
  "type": "task_reward_claimed",
  "timestamp": 1702259567,
  "user_id": "user_123",
  "data": {
    "user_id": "user_123",
    "task_id": "task_456",
    "reward": 50,
    "balance_before": 50,
    "balance_after": 100
  }
}
```

### User Events

#### `user_registered`
Новый пользователь зарегистрирован.

#### `user_welcome_bonus`
Приветственный бонус выдан.

## Client Usage

### JavaScript Example

```javascript
// Connect to WebSocket
const ws = new WebSocket('ws://localhost:8080/ws');

ws.onopen = function() {
    console.log('Connected to WebSocket');
    
    // Subscribe to specific events
    ws.send(JSON.stringify({
        action: 'subscribe',
        events: ['token_balance_update', 'treasury_update']
    }));
};

ws.onmessage = function(event) {
    const data = JSON.parse(event.data);
    console.log('Received event:', data);
    
    // Handle different event types
    switch(data.type) {
        case 'token_balance_update':
            updateUserBalance(data.data);
            break;
        case 'treasury_update':
            updateTreasuryDisplay(data.data);
            break;
        case 'task_completed':
            showTaskNotification(data.data);
            break;
    }
};

ws.onerror = function(error) {
    console.error('WebSocket error:', error);
};

ws.onclose = function() {
    console.log('Disconnected from WebSocket');
    // Implement reconnection logic
};
```

### React Example

```typescript
import { useEffect, useState } from 'react';

function useWebSocket(url: string) {
    const [ws, setWs] = useState<WebSocket | null>(null);
    const [events, setEvents] = useState<any[]>([]);
    
    useEffect(() => {
        const websocket = new WebSocket(url);
        
        websocket.onopen = () => {
            console.log('WebSocket connected');
            // Subscribe to all events
            websocket.send(JSON.stringify({
                action: 'subscribe',
                events: ['token_balance_update', 'treasury_update']
            }));
        };
        
        websocket.onmessage = (event) => {
            const data = JSON.parse(event.data);
            setEvents(prev => [data, ...prev]);
        };
        
        setWs(websocket);
        
        return () => {
            websocket.close();
        };
    }, [url]);
    
    return { ws, events };
}

// Usage in component
function TokenDisplay() {
    const { events } = useWebSocket('ws://localhost:8080/ws/tokens/user_123');
    
    const latestBalance = events
        .find(e => e.type === 'token_balance_update')
        ?.data?.balance_after || 0;
    
    return <div>Balance: {latestBalance} tokens</div>;
}
```

## Client Commands

### Subscribe to Events
```json
{
    "action": "subscribe",
    "events": ["treasury_update", "token_balance_update"]
}
```

### Unsubscribe from Events
```json
{
    "action": "unsubscribe",
    "events": ["task_completed"]
}
```

### Ping (Keep-Alive)
```json
{
    "action": "ping"
}
```

Response:
```json
{
    "type": "pong",
    "timestamp": 1702259567
}
```

## Integration in Code

### Publishing Events from Token Operations

```go
// In AllocateFromTreasury method
eventBus := wsservice.GetEventBus()

// Publish Treasury event
eventBus.Publish(wsservice.TreasuryAllocateEvent, map[string]interface{}{
    "balance": treasuryBalance,
    "amount": amount,
    "user_id": userID,
})

// Publish User event (only to specific user)
eventBus.PublishUserEvent(
    wsservice.TokenBalanceUpdateEvent,
    userID,
    map[string]interface{}{
        "balance_after": newBalance,
        "amount": amount,
    },
)
```

### Custom Event Publishing

```go
import wsservice "github.com/dmitrijfomin/menu-fodifood/backend/internal/modules/websocket/service"

// Get EventBus instance
eventBus := wsservice.GetEventBus()

// Publish to all clients
eventBus.Publish("custom_event", map[string]interface{}{
    "message": "Something happened",
    "value": 123,
})

// Publish to specific user
eventBus.PublishUserEvent(
    "user_notification",
    "user_123",
    map[string]interface{}{
        "title": "New Achievement",
        "description": "You earned a badge!",
    },
)
```

## Testing

### Using Test Server

1. Start test server:
```bash
go run cmd/test_ws/main.go
```

2. Open browser at `http://localhost:8080`

3. Click "Connect" button

4. Test events with curl:
```bash
# Test allocation
curl -X POST http://localhost:8080/api/test/allocate

# Test spending
curl -X POST http://localhost:8080/api/test/spend

# Test task completion
curl -X POST http://localhost:8080/api/test/task-complete
```

### Using HTML Test Page

Open `websocket_test.html` in your browser for an interactive WebSocket monitor with:
- Connection status
- Event subscription controls
- Real-time event log
- Uptime tracking

## Production Considerations

### Security

1. **Authentication**: Add JWT validation to WebSocket upgrade:
```go
// Extract token from query or header
token := r.URL.Query().Get("token")
userID := validateJWT(token)
```

2. **CORS**: Configure proper origins:
```go
var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        origin := r.Header.Get("Origin")
        return origin == "https://yourdomain.com"
    },
}
```

3. **Rate Limiting**: Limit connections per user/IP

### Performance

1. **Connection Limits**: Monitor and limit max connections
2. **Message Buffering**: Adjust buffer sizes based on load
3. **Graceful Shutdown**: Close all connections properly on server shutdown

### Monitoring

Track metrics:
- Active connections count
- Events per second
- Failed broadcasts
- Connection duration
- Memory usage

## Benefits

✅ **Real-time Updates**: Instant UI updates without polling  
✅ **Reduced Server Load**: No repeated API calls  
✅ **Better UX**: Users see changes immediately  
✅ **Scalable**: Pub/Sub pattern allows easy scaling  
✅ **Type-Safe**: Defined event types and structures  
✅ **Decoupled**: Business logic separated from WebSocket  

## Next Steps

1. Add authentication to WebSocket connections
2. Implement reconnection logic in frontend
3. Add message acknowledgments
4. Create admin dashboard with live Treasury monitoring
5. Add notification system for important events
6. Implement event history/replay functionality

---

**Status**: ✅ Fully Implemented and Tested  
**Commit**: b4ce836  
**Date**: 11 декабря 2025 г.
