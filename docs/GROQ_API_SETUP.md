# Groq API Configuration for Smart Kitchen

## 🔑 API Key Status
- **Status**: ✅ Active and working
- **Provider**: Groq (https://groq.com)
- **Tier**: Free Tier

## 🤖 Model Configuration

### Current Model
```env
GROQ_MODEL="llama-3.3-70b-versatile"
```

### Why Llama 3.3 70B Versatile?
- ✅ **Latest model** (December 2024)
- ✅ **70B parameters** = high quality responses
- ✅ **Versatile** = optimized for all tasks (not just chat)
- ✅ **Multilingual** = excellent Polish/English/Russian support
- ✅ **Fast** = ~250 tokens/second
- ✅ **Smart** = follows complex culinary instructions

### Alternative Models (if needed)
```bash
# Faster but less powerful
llama-3.1-8b-instant         # Good for simple tasks

# Previous generation (still good)
llama3-70b-8192              # Older 70B model
```

## 📊 Rate Limits (Free Tier)

### Llama 3.3 70B Versatile
| Metric | Limit | Smart Kitchen Usage |
|--------|-------|---------------------|
| **RPM** (Requests Per Minute) | 30 | ~50 concurrent users |
| **TPM** (Tokens Per Minute) | 6,000 | ~15-20 fridge analyses/min |
| **RPD** (Requests Per Day) | 14,400 | ~400 users/day |

### Token Usage Estimates
- **Today Meals**: ~300-400 tokens
- **3-Day Plan**: ~500-700 tokens
- **Reduce Waste**: ~300-500 tokens
- **Budget Review**: ~400-600 tokens

### Capacity Planning
```
MVP/Demo:   Free Tier is perfect
            14,400 requests/day = plenty

Production: If you hit limits:
            Pay-As-You-Go: $0.59 / 1M tokens
            Example: 1,000 analyses ≈ $0.30

Scale:      >1,000 users/day:
            Contact Groq for Enterprise plan
```

## 🚀 Deployment on Koyeb

### Environment Variables
Set these in Koyeb dashboard:

```bash
GROQ_API_KEY=gsk_YOUR_GROQ_API_KEY_HERE
GROQ_MODEL=llama-3.3-70b-versatile
AI_DEFAULT_LANG=pl
```

### Steps to Update Model
1. Go to Koyeb Dashboard → Your Service → Settings
2. Click "Environment Variables"
3. Update `GROQ_MODEL` to `llama-3.3-70b-versatile`
4. Click "Save"
5. Service will auto-redeploy (~90 seconds)

## 🧪 Testing the API

### Test Model Availability
```bash
curl -X GET "https://api.groq.com/openai/v1/models" \
  -H "Authorization: Bearer $GROQ_API_KEY" | \
  jq -r '.data[] | select(.id | contains("llama")) | .id'
```

### Test Polish Culinary Response
```bash
curl -X POST "https://api.groq.com/openai/v1/chat/completions" \
  -H "Authorization: Bearer $GROQ_API_KEY" \
  -H "Content-Type: application/json" \
  -d '{
    "model": "llama-3.3-70b-versatile",
    "messages": [
      {
        "role": "system",
        "content": "Jesteś kuchennym asystentem. Odpowiadaj TYLKO po polsku."
      },
      {
        "role": "user",
        "content": "Co mogę zrobić z wołowiną i cebulą?"
      }
    ],
    "max_tokens": 200,
    "temperature": 0.7
  }' | jq -r '.choices[0].message.content'
```

Expected output: Polish culinary recommendations

### Test Smart Kitchen Endpoint
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/ai/fridge/analyze \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "goal": "today_meals",
    "language": "pl"
  }' | jq '.data.result'
```

## 🔒 Security Best Practices

### ✅ DO:
- Keep API key in environment variables
- Use `.gitignore` for `.env` file
- Rotate key if exposed
- Monitor usage in Groq dashboard

### ❌ DON'T:
- Commit `.env` to Git
- Share API key publicly
- Hardcode key in source code
- Use same key for dev/prod

## 📈 Monitoring

### Check Usage
Visit: https://console.groq.com/settings/usage

### Key Metrics to Watch:
- **Daily requests**: Stay under 14,400/day
- **Token usage**: Monitor if approaching 6,000/min
- **Error rate**: Should be <1%
- **Latency**: Usually 1-3 seconds

### If You Hit Rate Limits:
1. Implement request queuing
2. Add caching for common requests
3. Upgrade to Pay-As-You-Go ($0.59/1M tokens)

## 🆘 Troubleshooting

### API Key Invalid
```bash
# Test key validity
curl -X GET "https://api.groq.com/openai/v1/models" \
  -H "Authorization: Bearer $GROQ_API_KEY"
```

### Model Not Found
- Check available models: `curl https://api.groq.com/openai/v1/models`
- Verify exact model name (case-sensitive)
- Update to latest model if deprecated

### Rate Limit Errors
```json
{"error": {"message": "Rate limit exceeded", "type": "rate_limit_error"}}
```
**Solution**: Wait 60 seconds or upgrade plan

### Slow Responses
- Normal: 1-3 seconds
- If >5 seconds: Check Groq status page
- Consider using `llama-3.1-8b-instant` for faster responses

## 📚 Resources

- **Groq Console**: https://console.groq.com
- **API Docs**: https://console.groq.com/docs
- **Pricing**: https://groq.com/pricing
- **Status Page**: https://status.groq.com

## 🎯 Summary

**Current Setup**: ✅ Optimal for Smart Kitchen MVP
- Model: Llama 3.3 70B Versatile (best quality)
- Tier: Free (14,400 requests/day)
- Capacity: 400+ users/day
- Cost: $0 (upgrade if needed: ~$0.30/1000 requests)

**Production Ready**: ✅ Yes
- API key: Valid
- Model: Tested and working
- Polish support: Excellent
- Error handling: Complete
- Fallbacks: Implemented

---

*Last Updated: December 17, 2024*
*Contact: AI Team Lead for any questions*
