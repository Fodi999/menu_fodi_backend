# 🚀 AI Chef Mentor - Professional Recipe Assistant

## ✨ New Features

### 1. **Persistent Sessions with Neon PostgreSQL**
- ✅ Sessions survive server restarts
- ✅ Automatic conversation history storage
- ✅ Recipe drafts persist in database
- ✅ 24-hour session TTL with auto-cleanup

**Tables created:**
- `chef_mentor_sessions` - Session metadata + recipe JSONB
- `chef_mentor_messages` - Full conversation history

### 2. **Smart Ingredients Parsing**
- ✅ Natural language: "рис 100г, вугор 200г, норі 1 шт"
- ✅ Multi-language support (UA, EN, RU, PL)
- ✅ Automatic gross/net weight calculation
- ✅ Cooking loss factors per ingredient type

**Examples:**
```bash
"рис 100г"     → gross: 100г, net: 120г (absorbs water +20%)
"вугор 200г"   → gross: 200г, net: 170г (cooking loss -15%)
"морква 50г"   → gross: 50г,  net: 42.5г (water loss -15%)
```

### 3. **Automatic Nutrition & Cost Analysis**
- ✅ Calories per 100g database (20+ ingredients)
- ✅ Protein, fats, carbs calculation
- ✅ Cost estimation (Ukrainian market prices)
- ✅ Total yield calculation

**Example output:**
```
📊 Цей рецепт має приблизно 544 ккал, вихід 337 г, собівартість 0.09 грн.
```

### 4. **Human-like AI Commentary**
- ✅ Natural conversation flow
- ✅ Automatic nutrition commentary after ingredient updates
- ✅ Multi-language support
- ✅ Contextual quick replies

### 5. **SSE Streaming Endpoint** (ChatGPT-like UX)
- ✅ Real-time token streaming
- ✅ `POST /api/ai/chef-mentor/stream`
- ✅ EventSource compatible
- ✅ Automatic session management

**Events:**
- `session` - Session ID
- `token` - AI response tokens (streaming)
- `metadata` - Recipe updates, quick replies
- `done` - Completion signal

### 6. **Enhanced Smart Extraction**
- ✅ Difficulty: "легка складність" → "easy"
- ✅ Time: "45 хвилин" → time: 45
- ✅ Portions: "6 порцій" → portions: 6
- ✅ Multi-language keyword matching

### 7. **Context-aware Quick Replies**
- ✅ Removes already-filled suggestions
- ✅ Language-specific options
- ✅ Dynamic based on recipe state

---

## 📊 Database Schema

```sql
CREATE TABLE chef_mentor_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID,
    language VARCHAR(5) NOT NULL DEFAULT 'ua',
    context JSONB DEFAULT '{}',
    recipe JSONB DEFAULT '{}',
    is_complete BOOLEAN DEFAULT false,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW(),
    last_activity TIMESTAMP NOT NULL
);

CREATE TABLE chef_mentor_messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    session_id UUID NOT NULL,
    role VARCHAR(20) NOT NULL, -- 'user' | 'assistant'
    content TEXT NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_chef_mentor_messages_session_id 
ON chef_mentor_messages(session_id);
```

---

## 🔌 API Endpoints

### 1. **Session-based Chat** (Recommended)
```bash
POST /api/ai/chef-mentor/session
Content-Type: application/json

{
  "message": "рис 100г, вугор 200г",
  "sessionId": "optional-uuid", # Auto-created if missing
  "language": "ua"
}
```

**Response:**
```json
{
  "status": "success",
  "data": {
    "sessionId": "7ed40d08-...",
    "message": "Чудово! Які назва страви?\n\n📊 Цей рецепт має 468 ккал...",
    "recipe": {
      "ingredients": [
        {
          "name": "рис",
          "amount": 100,
          "unit": "г",
          "gross": 100,
          "net": 120
        }
      ],
      "calories": 468,
      "protein": 8.5,
      "fats": 3.2,
      "carbs": 45.0,
      "cost": 0.081,
      "yield": 290
    },
    "quickReplies": ["4 порції", "30 хвилин", "Додати інгредієнти"],
    "isComplete": false
  }
}
```

### 2. **Streaming Chat** (Real-time)
```bash
POST /api/ai/chef-mentor/stream
Content-Type: application/json

{
  "message": "Хочу зробити роли Дракон",
  "sessionId": "optional-uuid",
  "language": "ua"
}
```

**Response (SSE):**
```
event: session
data: {"sessionId":"7ed40d08-..."}

event: token
data: {"content":"Чу"}

event: token
data: {"content":"дово"}

event: token
data: {"content":"!"}

event: metadata
data: {"recipe":{...},"quickReplies":[...]}

event: done
data: {"status":"success"}
```

### 3. **Get Session**
```bash
GET /api/ai/chef-mentor/session?id={sessionId}
```

### 4. **Delete Session**
```bash
DELETE /api/ai/chef-mentor/session?id={sessionId}
```

---

## 🧪 Testing Examples

### Test 1: Ingredients Parsing
```bash
curl -X POST http://localhost:8080/api/ai/chef-mentor/session \
  -H "Content-Type: application/json" \
  -d '{"message":"рис 100г, вугор 200г, авокадо 1 шт","language":"ua"}' | jq
```

**Expected:**
- ✅ 3 ingredients parsed
- ✅ Gross/net calculated
- ✅ Nutrition analysis: ~544 ккал
- ✅ Cost: ~0.09 грн
- ✅ Human commentary added

### Test 2: Session Persistence
```bash
# Create session
SESSION_ID=$(curl -s -X POST http://localhost:8080/api/ai/chef-mentor/session \
  -d '{"message":"Привіт!","language":"ua"}' | jq -r '.data.sessionId')

# Continue conversation
curl -X POST http://localhost:8080/api/ai/chef-mentor/session \
  -d "{\"message\":\"рис 100г\",\"sessionId\":\"$SESSION_ID\"}" | jq

# Restart server...

# Retrieve session (should still exist!)
curl "http://localhost:8080/api/ai/chef-mentor/session?id=$SESSION_ID" | jq
```

### Test 3: Streaming
```bash
curl -N -X POST http://localhost:8080/api/ai/chef-mentor/stream \
  -H "Content-Type: application/json" \
  -d '{"message":"Хочу зробити суші","language":"ua"}'
```

---

## 🎯 Production Ready

✅ **Tested Features:**
- Session creation & continuation
- Ingredients parsing (5+ ingredients)
- Gross/net calculation (rice, eel, avocado, cucumber)
- Nutrition analysis (calories, protein, fats, carbs)
- Cost estimation
- Human commentary generation
- Session persistence after server restart

✅ **Performance:**
- Auto-cleanup of 24h+ old sessions
- GORM optimized queries
- Indexed session lookups

✅ **Security:**
- Input validation
- SQL injection protected (GORM)
- Rate limiting ready

---

## 🚀 Deployment

1. **Environment Variables:**
```bash
DATABASE_URL=postgres://...@neon.tech/...
GROQ_API_KEY=gsk_...
```

2. **Auto-migration:**
```go
database.AutoMigrate() // Creates tables automatically
```

3. **Production URL:**
```
https://menu-fodifood-backend.koyeb.app/api/ai/chef-mentor/session
```

---

## 📈 Future Improvements

- [ ] AI Memory Compression (summarize long conversations)
- [ ] Recipe Steps parsing
- [ ] Image upload for ingredients recognition
- [ ] Voice input support
- [ ] Recipe templates library
- [ ] Social sharing

---

**Created by:** Dima Fomin
**Version:** 2.0 (Professional)
**Date:** 2025-11-06
