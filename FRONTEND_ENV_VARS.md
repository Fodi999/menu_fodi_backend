# 🔐 Frontend Environment Variables

**Для фронтенд-команды:** Создайте файл `.env.local` в корне Next.js проекта с этими переменными:

---

## 📋 Required Environment Variables

```bash
# ============================================
# 🌐 Backend API
# ============================================
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app

# ============================================
# ☁️ Cloudinary (Image Upload)
# ============================================
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=dwrn0ohbp
NEXT_PUBLIC_CLOUDINARY_API_KEY=954374638575439
NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=ml_default

# ============================================
# 🔑 API Keys (Optional - Backend Only)
# ============================================
# DO NOT add backend API keys here!
# Groq API key is configured on backend server only
```

---

## 📸 Cloudinary Configuration

### Cloud Name
```
dwrn0ohbp
```

### API Key
```
954374638575439
```

### Upload Preset
Для публичной загрузки создайте **unsigned upload preset** в Cloudinary:

1. Зайдите в [Cloudinary Dashboard](https://cloudinary.com/console)
2. Settings → Upload → Upload presets
3. Add upload preset:
   - **Name:** `ml_default` (или любое другое)
   - **Signing Mode:** Unsigned
   - **Folder:** `culinary-academy`
   - **Allowed formats:** jpg, png, jpeg, gif, webp
   - **Transformation:** Optional (можно добавить auto resize)

---

## 🧪 Test Cloudinary Upload

### Using Fetch API (Frontend)
```typescript
const uploadImage = async (file: File) => {
  const formData = new FormData();
  formData.append('file', file);
  formData.append('upload_preset', 'ml_default');
  formData.append('folder', 'culinary-academy');

  const response = await fetch(
    `https://api.cloudinary.com/v1_1/dwrn0ohbp/image/upload`,
    {
      method: 'POST',
      body: formData
    }
  );

  const data = await response.json();
  return data.secure_url; // https://res.cloudinary.com/dwrn0ohbp/...
};
```

### Using Backend Endpoint (Recommended)
```typescript
const uploadImage = async (imageUrl: string) => {
  const response = await fetch(
    'https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/upload/image',
    {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ imageUrl })
    }
  );

  const { data } = await response.json();
  return data.url;
};
```

---

## 🔗 API Endpoints

### Production Server
```
https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
```

### Health Check
```bash
curl https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/health
```

### Upload Image
```bash
curl -X POST https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app/api/upload/image \
  -H 'Content-Type: application/json' \
  -d '{"imageUrl":"https://images.unsplash.com/photo-1579584425555-c3ce17fd4351?w=500"}'
```

---

## 👤 Test User Credentials

### User 1 - Dima Fomin
```json
{
  "email": "dima@example.com",
  "password": "password123",
  "userId": "ef03cd81-71fd-429f-bb5f-8be5c9172ca8"
}
```

### User 2 - Anna Kowalska
```json
{
  "email": "anna@example.com",
  "password": "password123",
  "userId": "fba50be3-e3c5-4d73-8ed8-cfb6422f7034"
}
```

### User 3 - Maksym Fomin
```json
{
  "email": "fodi85@gmail.ru",
  "password": "password123",
  "userId": "407582be-59d5-4d21-873b-1a72d31b0d42"
}
```

---

## 📝 Example .env.local File

Create this file in your Next.js project root:

```bash
# .env.local
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=dwrn0ohbp
NEXT_PUBLIC_CLOUDINARY_API_KEY=954374638575439
NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=ml_default
```

---

## 🚀 Quick Start Commands

```bash
# 1. Create .env.local file
cat > .env.local << 'EOF'
NEXT_PUBLIC_API_URL=https://yeasty-madelaine-fodi999-671ccdf5.koyeb.app
NEXT_PUBLIC_CLOUDINARY_CLOUD_NAME=dwrn0ohbp
NEXT_PUBLIC_CLOUDINARY_API_KEY=954374638575439
NEXT_PUBLIC_CLOUDINARY_UPLOAD_PRESET=ml_default
EOF

# 2. Restart Next.js dev server
npm run dev

# 3. Test API connection
curl $NEXT_PUBLIC_API_URL/api/health
```

---

## 🔒 Security Notes

⚠️ **Important:**
- ✅ `NEXT_PUBLIC_*` variables are **exposed to browser** - это нормально для Cloudinary
- ✅ Unsigned upload preset безопасен для публичной загрузки
- ✅ Backend API endpoint `/api/upload/image` проверяет и валидирует изображения
- ❌ **НЕ добавляйте** `CLOUDINARY_API_SECRET` в `.env.local` - только на backend!

---

## 📚 Documentation Links

- **Backend API:** [ALL_ENDPOINTS.md](ALL_ENDPOINTS.md)
- **Quick Start:** [QUICK_START.md](QUICK_START.md)
- **Cloudinary Docs:** https://cloudinary.com/documentation/upload_images
- **Next.js Env Vars:** https://nextjs.org/docs/app/building-your-application/configuring/environment-variables

---

**Created:** 4 November 2025  
**Backend:** Production Ready ✅  
**Frontend:** Requires .env.local configuration
