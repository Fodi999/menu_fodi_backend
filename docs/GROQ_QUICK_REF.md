# ⚡ GROQ API - Quick Reference

## 🔑 Key Info
```
Status: ✅ Active
Model:  llama-3.3-70b-versatile (70B, Dec 2024)
Tier:   Free (14,400 req/day, 30 req/min)
```

## 📊 Limits
| Limit | Value | Enough for |
|-------|-------|-----------|
| RPM   | 30    | ~50 concurrent users |
| RPD   | 14,400 | ~400 users/day |
| Cost  | FREE  | MVP perfect! |

## 🚀 Koyeb Env Vars
```bash
GROQ_API_KEY=gsk_YOUR_GROQ_API_KEY_HERE
GROQ_MODEL=llama-3.3-70b-versatile
```

## ✅ Health Check
```bash
# Test API
curl https://api.groq.com/openai/v1/models \
  -H "Authorization: Bearer $GROQ_API_KEY" | \
  jq '.data[0].id'
```

## 💰 Pricing (if you outgrow free tier)
```
Free:           14,400 req/day ($0)
Pay-As-You-Go:  $0.59 / 1M tokens
Enterprise:     Custom limits
```

## 📚 Full Docs
→ See `docs/GROQ_API_SETUP.md` for complete guide

---
*Smart Kitchen AI Ready for Production* 🎉
