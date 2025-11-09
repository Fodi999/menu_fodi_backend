# 🔐 Setting Up Tests with Groq API Key

## Important Security Note

The `.env.test` file in the repository uses a **placeholder** for `GROQ_API_KEY` for security reasons:

```bash
GROQ_API_KEY="your_groq_api_key_here"
```

## To Run E2E Tests Locally

You need to add your actual Groq API key to `.env.test`:

### Option 1: Direct Edit (Not Recommended)
```bash
# Edit .env.test
nano .env.test

# Replace:
GROQ_API_KEY="your_groq_api_key_here"
# With your actual key from https://console.groq.com/keys
```

### Option 2: Environment Variable (Recommended)
```bash
# Export before running tests
export GROQ_API_KEY="your_actual_key_here"

# Then run tests
make test-ai
make e2e
make all
```

### Option 3: Load from .env Before Running
```bash
# If you have your real .env file with the key:
source .env
make test
```

## Getting Your Groq API Key

1. Visit [Groq Console](https://console.groq.com/keys)
2. Create or copy your API key
3. Add it to `.env.test` or export as environment variable

## GitHub Actions / CI/CD

In GitHub Actions, the API key is configured via **Repository Secrets**:

```yaml
# In .github/workflows/test.yml
env:
  GROQ_API_KEY: ${{ secrets.GROQ_API_KEY }}
```

This is stored securely in GitHub and **never exposed** in logs.

## Test Results Without Real Key

Tests will still run but some will fail:

- ✅ **Input Validation Tests** - Pass (test edge cases)
- ✅ **API Structure Tests** - Pass (verify endpoints exist)
- ❌ **API Call Tests** - Fail (require real Groq key)

This is expected and shows that validation layer is working correctly.

## Verify Setup

```bash
# Check if key is loaded
echo $GROQ_API_KEY

# Run all tests (some may fail without real key)
make all

# Run only validation tests (will pass)
make short

# Run with verbose output to see what's happening
go test ./tests/unit/... -v
```

## Files Changed

- `.env.test` - Placeholder key (safe to commit)
- `.env` - Real key (never commit this!)
- `.github/workflows/test.yml` - Uses GitHub secrets

---

**Security Rule:** Never commit real API keys. Use environment variables or GitHub secrets instead.
