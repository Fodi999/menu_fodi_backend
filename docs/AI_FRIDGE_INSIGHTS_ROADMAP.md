# AI Fridge Insights - Architecture & Roadmap

## 📌 Current Status: FOUNDATION COMPLETE ✅

### What's Already Built (and it's solid):

#### 1️⃣ **Data Layer** ✅
- `history_events` table with full metadata
- `user_fridge_items` with pricing and expiry tracking
- `ExpiredItemMetadata` capturing:
  - Cost (quantity × price)
  - Days in fridge
  - Reason (expiry_date_passed)
  - Context (fridge_cleanup)

#### 2️⃣ **Event Layer** ✅
- `EventTypeExpired` - automatic expiry tracking
- `SourceTypeAuto` - system-generated events
- Full audit trail of what happened

#### 3️⃣ **Cleanup Logic** ✅
- `cmd/cleanup_expired/main.go`
- Automatic removal of expired items
- History logging before deletion
- Cost calculation and statistics

#### 4️⃣ **Analytics API** ✅
- `GET /api/history?type=expired` - filtered events
- `GET /api/history/losses?days=30` - aggregated analytics
- Summary: totalItems, totalCost, avgCostPerItem
- Period-based filtering

---

## 🔮 Next Layer: AI as Mentor (NOT URGENT)

### Philosophy:
> "You did this → this happened → here's what you could change"

**NOT:**
- ❌ "Buy this"
- ❌ "Cook that"
- ❌ Judge user decisions

**BUT:**
- ✅ Observe patterns
- ✅ Reflect on outcomes
- ✅ Suggest experiments

---

## 🎯 Future Endpoint: `/api/ai/insights/fridge-losses`

### What It Does:

**Input:** User's history (last 30-90 days)

**Analysis:**
1. **Pattern Recognition**
   - Which categories expire most? (dairy, protein, vegetable)
   - Time patterns: weekday vs weekend
   - Quantity patterns: buying too much?
   - Storage duration: too short vs spoiled

2. **Context Understanding**
   - User's cooking frequency (from cook events)
   - Recipe preferences (from recommendations)
   - Budget constraints (from budget module)
   - Shopping patterns (from fridge_add events)

3. **Reflection Generation**
   - "Here's what I noticed..."
   - "This pattern suggests..."
   - "You might experiment with..."

### Example Response:

```json
{
  "success": true,
  "data": {
    "period": {
      "days": 30,
      "analyzed_events": 15
    },
    "patterns": [
      {
        "type": "category",
        "insight": "Dairy products (mleko, jogurt) expired 5 times",
        "context": "These items had avg 3 days in fridge before expiry",
        "reflection": "You're buying dairy but not using it quickly enough"
      },
      {
        "type": "timing",
        "insight": "Most expirations happen on weekends",
        "context": "You cook more during weekdays",
        "reflection": "Consider smaller shopping trips before weekends"
      },
      {
        "type": "cost",
        "insight": "Łosoś (salmon) accounts for 35% of total loss (45.50 PLN)",
        "context": "High-value protein stored 2-3 days before expiry",
        "reflection": "Plan salmon recipes within 48h of purchase"
      }
    ],
    "experiments": [
      {
        "title": "Smaller dairy portions",
        "description": "Try buying 1L milk instead of 2L for next 2 weeks",
        "expected_outcome": "Reduce dairy waste by 50%"
      },
      {
        "title": "Cook-first rule",
        "description": "Plan recipes for expensive proteins (salmon, beef) on purchase day",
        "expected_outcome": "Reduce high-value losses"
      }
    ],
    "positive_patterns": [
      {
        "observation": "You used 8 recipes this month",
        "impact": "Active cooking reduces waste"
      },
      {
        "observation": "Vegetables rarely expire",
        "impact": "Good planning for produce"
      }
    ]
  }
}
```

---

## 🏗️ Implementation Plan (When Ready)

### Phase 1: Data Analysis Service
**File:** `internal/modules/ai_recommendations/service/fridge_insights_service.go`

**Methods:**
- `AnalyzeExpiredPatterns(userID, days)` - pattern detection
- `GenerateReflections(patterns)` - insight generation
- `SuggestExperiments(patterns)` - actionable recommendations

### Phase 2: AI Integration
**Use existing:** `internal/modules/ai_core/groq_client.go`

**Prompt Engineering:**
```
You are a kitchen mentor analyzing user's fridge history.

Data:
- 5 expired items in 30 days
- Total loss: 45.50 PLN
- Categories: dairy (3), protein (2)
- Avg storage: 3.2 days before expiry

Task:
Generate 3 observations and 2 experiments.

Style:
- Observational, not prescriptive
- Data-driven, not judgemental
- Encouraging, not critical

Language: {user.settings.language}
```

### Phase 3: Frontend Integration
**Dashboard Widget:**
- "Fridge Insights" card
- 1-2 key observations
- Link to full insights page

**Insights Page:**
- Timeline of patterns
- Cost visualization
- Experiment tracker

---

## 🎨 Design Principles

### 1. Observer, Not Judge
```
❌ "You're wasting food"
✅ "I noticed dairy expires often - here's the pattern"
```

### 2. Data-Driven
```
❌ "Buy less milk"
✅ "3 of 5 losses were dairy, avg 3 days in fridge"
```

### 3. Experimental Mindset
```
❌ "You must do this"
✅ "Try this for 2 weeks and see what happens"
```

### 4. Celebrate Wins
```
✅ "Vegetables rarely expire - great planning!"
✅ "You cooked 8 recipes - active kitchen!"
```

---

## 📊 Technical Architecture

```
┌──────────────────────────────────────────────────────┐
│                   Frontend                            │
│  ┌────────────────────────────────────────────────┐ │
│  │  Dashboard: "Fridge Insights" Widget           │ │
│  │  Insights Page: Full Analysis + Experiments    │ │
│  └────────────────────────────────────────────────┘ │
└──────────────────────┬───────────────────────────────┘
                       │
                       │ GET /api/ai/insights/fridge-losses
                       ▼
┌──────────────────────────────────────────────────────┐
│           AI Recommendations Module                   │
│  ┌────────────────────────────────────────────────┐ │
│  │  FridgeInsightsService                         │ │
│  │    ├─ AnalyzeExpiredPatterns()                 │ │
│  │    ├─ GenerateReflections()                    │ │
│  │    └─ SuggestExperiments()                     │ │
│  └────────────────────────────────────────────────┘ │
└──────────────────────┬───────────────────────────────┘
                       │
                       │ Reads from
                       ▼
┌──────────────────────────────────────────────────────┐
│               History Repository                      │
│  ┌────────────────────────────────────────────────┐ │
│  │  GetByFilters(userID, {type: "expired"})      │ │
│  │  GetStatsByUser(userID, dateRange)            │ │
│  └────────────────────────────────────────────────┘ │
└──────────────────────┬───────────────────────────────┘
                       │
                       │ Queries
                       ▼
┌──────────────────────────────────────────────────────┐
│             Database (PostgreSQL)                     │
│  ┌────────────────────────────────────────────────┐ │
│  │  history_events (event_type = 'expired')      │ │
│  │    ├─ metadata.cost                            │ │
│  │    ├─ metadata.ingredient_name                 │ │
│  │    ├─ metadata.days_in_fridge                  │ │
│  │    └─ metadata.reason                          │ │
│  └────────────────────────────────────────────────┘ │
└──────────────────────────────────────────────────────┘
```

---

## 🚀 Why This Works

### 1. **Foundation is Solid**
- Data exists (history_events)
- Events are logged (expired items)
- Analytics are available (losses API)

### 2. **AI is Optional Enhancement**
- System works without AI
- AI adds interpretation layer
- Users can ignore insights

### 3. **Scalable Architecture**
- AI reads from history (read-only)
- No tight coupling
- Can add more insight types easily

### 4. **User-Centric Design**
- Language-aware (uses user.settings.language)
- Respects user autonomy
- Focuses on learning, not lecturing

---

## 📝 Implementation Checklist (Future)

### Backend
- [ ] Create `FridgeInsightsService` in ai_recommendations
- [ ] Add `/api/ai/insights/fridge-losses` endpoint
- [ ] Implement pattern detection algorithms
- [ ] Integrate with Groq AI for reflection generation
- [ ] Add caching for expensive AI calls

### Frontend
- [ ] Create "Fridge Insights" dashboard widget
- [ ] Build full insights page with visualizations
- [ ] Add experiment tracking UI
- [ ] Implement timeline view for patterns

### Testing
- [ ] Unit tests for pattern detection
- [ ] Integration tests for insights API
- [ ] User testing for reflection clarity
- [ ] A/B testing for insight effectiveness

---

## 🎯 Success Metrics

### Quantitative
- % reduction in expired items after insights
- Cost savings (PLN) from reduced waste
- User engagement with insights page

### Qualitative
- User feedback: "This helped me understand my patterns"
- Behavioral change: smaller purchases, faster cooking
- Trust: "The AI noticed what I didn't see"

---

## 🏆 The Big Picture

This is not just a "food waste app".

This is a **decision-making system** that:
- Observes reality (expired items)
- Records consequences (cost, patterns)
- Reflects on behavior (AI insights)
- Suggests experiments (actionable changes)

**You built:**
- ✅ The sensors (fridge tracking)
- ✅ The memory (history events)
- ✅ The analysis (losses API)

**What's left:**
- 🔮 The mentor (AI interpretation)

And it's **not urgent** because the foundation is so solid that users already get value from:
- Knowing what expired (transparency)
- Seeing costs (accountability)
- Tracking patterns (awareness)

The AI will add:
- Understanding WHY (insight)
- Suggesting HOW to improve (mentorship)

---

## 💡 Final Thought

> "Ты строишь не приложение. Ты строишь систему принятия решений."

And you did it right:
- **Honest** - shows reality, doesn't hide losses
- **Scalable** - can add more insight types
- **Professional** - clean architecture, clear separation
- **Explainable** - every decision is logged

The AI layer will be the cherry on top. But the cake is already perfect.

---

**Status:** Ready for AI when needed, but already valuable without it.  
**Next:** Enjoy the solid foundation you built. AI can wait.
