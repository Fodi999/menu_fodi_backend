# 🐛 Frontend Fix: daysLeft NULL Handling

**Date:** January 15, 2026  
**Issue:** Products without expiry date show "Осталось 0 дней" instead of "Без срока годности"  
**Status:** ✅ Backend correct, ⚠️ Frontend needs fix  

---

## 🔍 Root Cause Analysis

### Backend Behavior (✅ CORRECT):
```bash
$ curl /api/fridge/items | jq '.data.items[] | select(.expiresAt == null)'

{
  "name": "Oliwa z oliwek",
  "expiresAt": null,
  "daysLeft": null    # ✅ Correctly returns NULL
}
```

### Frontend Behavior (❌ INCORRECT):
```javascript
// Console logs:
[FridgePage] Item: daysLeft = 0 (type: number)  // ❌ Should be null!
[FridgeItem] Oliwa z oliwek → daysLeft: 0 (type: number)

// UI displays:
"Осталось 0 дней"  // ❌ Wrong - implies "expires today"
```

---

## 🎯 The Problem

Somewhere in the frontend code, `null` is being converted to `0`:

### ❌ Common Anti-Patterns:

```typescript
// Pattern 1: Nullish coalescing with 0
const daysLeft = item.daysLeft ?? 0;  // ❌ Converts null → 0

// Pattern 2: Logical OR with 0
const daysLeft = item.daysLeft || 0;  // ❌ Converts null → 0

// Pattern 3: Default value in destructuring
const { daysLeft = 0 } = item;  // ❌ Converts undefined/null → 0

// Pattern 4: Math operations
const daysLeft = Math.max(item.daysLeft, 0);  // ❌ Converts null → 0
```

---

## ✅ Solution: Preserve NULL Values

### TypeScript Type Definition:

```typescript
// types/fridge.ts
export interface FridgeItem {
  id: string;
  name: string;
  quantity: number;
  unit: string;
  expiresAt: string | null;  // ✅ Can be null
  daysLeft: number | null;    // ✅ Can be null (KEY!)
  status: 'fresh' | 'expired' | 'discarded';
  ingredient?: Ingredient;
}
```

### Display Logic:

```typescript
// components/FridgeItem.tsx
interface Props {
  item: FridgeItem;
}

export function FridgeItem({ item }: Props) {
  // ✅ CORRECT: Handle null explicitly
  const expiryDisplay = item.daysLeft !== null 
    ? `Осталось ${item.daysLeft} дней`
    : 'Без срока годности';

  // ✅ CORRECT: Conditional rendering
  return (
    <div>
      <h3>{item.name}</h3>
      <p className={getUrgencyClass(item.daysLeft)}>
        {expiryDisplay}
      </p>
    </div>
  );
}

// Helper function
function getUrgencyClass(daysLeft: number | null): string {
  if (daysLeft === null) return 'no-expiry';  // ✅ Gray/neutral
  if (daysLeft < 0) return 'expired';          // Red
  if (daysLeft === 0) return 'critical';       // Red
  if (daysLeft === 1) return 'warning';        // Orange
  if (daysLeft <= 3) return 'info';            // Yellow
  return 'fresh';                               // Green
}
```

### Filtering Logic:

```typescript
// ✅ CORRECT: Filter expired items (exclude null)
const expiredItems = items.filter(item => 
  item.daysLeft !== null && item.daysLeft < 0
);

// ✅ CORRECT: Filter items expiring soon (exclude null)
const urgentItems = items.filter(item => 
  item.daysLeft !== null && item.daysLeft <= 3
);

// ✅ CORRECT: Filter items without expiry
const noExpiryItems = items.filter(item => 
  item.daysLeft === null
);
```

---

## 🔎 Where to Check

### Likely Problem Locations:

1. **API Response Parsing:**
```typescript
// utils/api.ts or services/fridgeApi.ts
// ❌ Check for: daysLeft: response.daysLeft ?? 0
// ✅ Should be: daysLeft: response.daysLeft
```

2. **State Management:**
```typescript
// contexts/FridgeContext.tsx or stores/fridgeStore.ts
// ❌ Check for default values in reducers/setters
```

3. **Component Props:**
```typescript
// components/FridgeItem.tsx
// ❌ Check for: const { daysLeft = 0 } = props
// ✅ Should be: const { daysLeft } = props
```

4. **Display Helpers:**
```typescript
// utils/formatters.ts
// ❌ Check for: Math operations without null checks
```

---

## 📊 Testing Checklist

### Manual Testing:
- [ ] Add item **without** expiry date
- [ ] Verify UI shows "Без срока годности" (not "0 дней")
- [ ] Verify no urgency color (gray/neutral)
- [ ] Check console logs: `daysLeft: null` (not `0`)

### Unit Tests:
```typescript
// FridgeItem.test.tsx
describe('FridgeItem with no expiry', () => {
  it('should display "Без срока годности" when daysLeft is null', () => {
    const item = {
      name: 'Olive Oil',
      daysLeft: null,
      expiresAt: null,
      status: 'fresh'
    };
    
    render(<FridgeItem item={item} />);
    expect(screen.getByText('Без срока годности')).toBeInTheDocument();
    expect(screen.queryByText(/0 дней/)).not.toBeInTheDocument();
  });

  it('should apply neutral styling when daysLeft is null', () => {
    const item = { daysLeft: null, /* ... */ };
    render(<FridgeItem item={item} />);
    
    const element = screen.getByTestId('expiry-badge');
    expect(element).toHaveClass('no-expiry');
    expect(element).not.toHaveClass('critical');
  });
});
```

---

## 🎨 UI/UX Recommendations

### Visual Design:

```tsx
// No expiry date - Neutral/Gray
<Badge variant="neutral" icon={<CheckCircle />}>
  Без срока годности
</Badge>

// Expires today/tomorrow - Critical/Red
<Badge variant="critical" icon={<AlertTriangle />}>
  Истекает сегодня
</Badge>

// Expired - Critical/Red
<Badge variant="critical" icon={<XCircle />}>
  Просрочено {Math.abs(daysLeft)} дней назад
</Badge>

// Expires soon (2-3 days) - Warning/Yellow
<Badge variant="warning" icon={<Clock />}>
  Осталось {daysLeft} дня
</Badge>
```

### Sorting/Filtering:

```typescript
// ✅ Priority sorting (null goes to end)
items.sort((a, b) => {
  // Expired first (negative daysLeft)
  if (a.daysLeft !== null && a.daysLeft < 0) return -1;
  if (b.daysLeft !== null && b.daysLeft < 0) return 1;
  
  // Expiring soon (0-3 days)
  if (a.daysLeft !== null && a.daysLeft <= 3) {
    if (b.daysLeft !== null && b.daysLeft <= 3) {
      return a.daysLeft - b.daysLeft;
    }
    return -1;
  }
  
  // Items with expiry
  if (a.daysLeft !== null && b.daysLeft !== null) {
    return a.daysLeft - b.daysLeft;
  }
  
  // Items without expiry go last
  if (a.daysLeft === null) return 1;
  if (b.daysLeft === null) return -1;
  
  return 0;
});
```

---

## 🧠 Why This Matters

### User Trust:
- ❌ **"0 дней"** → User thinks: "It expires TODAY, urgent!"
- ✅ **"Без срока"** → User thinks: "No rush, long-term storage"

### Correct Behavior:
```
Expired (-2 days)     → ❌ Red badge     "Просрочено 2 дня назад"
Expires today (0)     → ⚠️ Red badge     "Истекает сегодня"
Expires tomorrow (1)  → ⚠️ Orange badge  "Завтра истечет"
Expires soon (2-3)    → ⚡ Yellow badge  "Осталось 2 дня"
Fresh (4+)            → ✅ Green badge   "Осталось 5 дней"
No expiry (null)      → ⚪ Gray badge    "Без срока годности"
```

---

## 📚 Backend Contract (Reference)

### API Response Format:

```json
{
  "data": {
    "items": [
      {
        "id": "uuid",
        "name": "Oliwa z oliwek",
        "quantity": 500,
        "unit": "ml",
        "expiresAt": null,        // ← Can be null
        "daysLeft": null,          // ← MUST preserve null!
        "status": "fresh",
        "ingredient": { /* ... */ }
      },
      {
        "id": "uuid",
        "name": "Milk",
        "quantity": 1,
        "unit": "l",
        "expiresAt": "2026-01-16T00:00:00Z",
        "daysLeft": 1,             // ← Number when expiry exists
        "status": "fresh",
        "ingredient": { /* ... */ }
      }
    ]
  },
  "success": true
}
```

### Backend Logic (Reference):
```go
// internal/modules/fridge/service/fridge_service.go:394
func (s *FridgeService) calculateDaysLeft(expiresAt *time.Time) *int {
    if expiresAt == nil {
        return nil  // ✅ Returns nil, not 0
    }
    duration := time.Until(*expiresAt)
    days := int(duration.Hours() / 24)
    return &days
}
```

---

## ✅ Verification Steps

After fixing the frontend:

1. **API Response Check:**
```bash
$ curl /api/fridge/items | jq '.data.items[0]'
{
  "daysLeft": null  # ✅ Backend sends null
}
```

2. **Console Logs Check:**
```javascript
console.log('daysLeft:', item.daysLeft);
// Should show: daysLeft: null (not 0)
```

3. **UI Check:**
```
Oliwa z oliwek
Свежее
Без срока годности  ✅ (not "Осталось 0 дней")
```

4. **Component Props Check:**
```typescript
// React DevTools:
<FridgeItem item={{daysLeft: null}} />  ✅
// Not: <FridgeItem item={{daysLeft: 0}} />  ❌
```

---

## 🎯 Summary

### Backend Status: ✅ **PERFECT**
- Returns `null` when `expiresAt` is `null`
- No conversion to `0` anywhere
- JSON serialization preserves `null`

### Frontend Fix Required: ⚠️
- **Find:** Where `null` becomes `0`
- **Fix:** Preserve `null` values
- **Display:** Show "Без срока годности"
- **Style:** Use neutral color (not red/urgent)

---

**Last Updated:** January 15, 2026  
**Backend Version:** ✅ Production-ready  
**Frontend Action Required:** Check TypeScript type coercion
