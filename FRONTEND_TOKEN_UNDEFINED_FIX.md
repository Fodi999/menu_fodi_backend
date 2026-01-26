# 🐛 Исправление проблемы "Bearer undefined"

**Дата:** 2026-01-26  
**Проблема:** Фронтенд отправляет `Bearer undefined` вместо реального JWT токена

---

## 🔍 Проблема

В логах видно:
```
🔍 Auth header preview: "Bearer undefined"
🎫 Token extracted, length: 9
⚠️ Token preview: "undefined"
```

**Причина:** Переменная `token` равна `undefined` на фронтенде при отправке запроса.

---

## ✅ Решение для фронтенда

### Проблема 1: Токен не сохраняется после логина

**Симптомы:**
- Логин успешен (200 OK)
- Токен приходит в ответе
- Но при следующем запросе токен = `undefined`

**Решение:**

```javascript
// ✅ ПРАВИЛЬНО: Сохранение токена после логина
async function login(email, password) {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  
  const data = await response.json();
  
  if (data.success && data.token) {
    // ✅ Сохраняем токен
    localStorage.setItem('token', data.token);
    console.log('✅ Token saved:', data.token.substring(0, 20) + '...');
    return data;
  }
  
  throw new Error('Login failed');
}
```

### Проблема 2: Токен не извлекается перед запросом

**Симптомы:**
- Токен сохранен в localStorage
- Но при запросе используется `undefined`

**Решение:**

```javascript
// ✅ ПРАВИЛЬНО: Извлечение токена перед запросом
function getAuthHeaders() {
  const token = localStorage.getItem('token');
  
  if (!token) {
    console.error('❌ No token found in localStorage');
    // Перенаправить на логин
    window.location.href = '/login';
    return {};
  }
  
  // Проверка формата токена
  if (token.length < 50) {
    console.error('❌ Token seems invalid (too short):', token);
    localStorage.removeItem('token');
    window.location.href = '/login';
    return {};
  }
  
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };
}

// Использование
async function getUserProfile() {
  const headers = getAuthHeaders();
  
  if (!headers['Authorization']) {
    return; // Уже перенаправлено на логин
  }
  
  const response = await fetch('/api/user/profile', {
    method: 'GET',
    headers: headers
  });
  
  return response.json();
}
```

### Проблема 3: Токен не обновляется после логина

**Симптомы:**
- Старый токен в localStorage
- Новый токен не сохраняется

**Решение:**

```javascript
// ✅ ПРАВИЛЬНО: Обновление токена после логина
async function login(email, password) {
  const response = await fetch('/api/auth/login', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ email, password })
  });
  
  const data = await response.json();
  
  if (data.success && data.token) {
    // ✅ Удаляем старый токен (если есть)
    localStorage.removeItem('token');
    
    // ✅ Сохраняем новый токен
    localStorage.setItem('token', data.token);
    
    // ✅ Проверяем, что токен сохранен
    const savedToken = localStorage.getItem('token');
    if (savedToken === data.token) {
      console.log('✅ Token saved successfully, length:', savedToken.length);
    } else {
      console.error('❌ Token save failed!');
    }
    
    return data;
  }
  
  throw new Error('Login failed');
}
```

---

## 🔧 Полный пример исправления

### React/Next.js пример:

```javascript
// utils/auth.js
export function getToken() {
  if (typeof window === 'undefined') {
    return null; // SSR
  }
  
  const token = localStorage.getItem('token');
  
  if (!token) {
    return null;
  }
  
  // Валидация токена
  if (token.length < 50) {
    console.error('❌ Invalid token format');
    localStorage.removeItem('token');
    return null;
  }
  
  return token;
}

export function getAuthHeaders() {
  const token = getToken();
  
  if (!token) {
    return {
      'Content-Type': 'application/json'
    };
  }
  
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  };
}

export function saveToken(token) {
  if (typeof window === 'undefined') {
    return;
  }
  
  if (!token || token.length < 50) {
    console.error('❌ Invalid token to save');
    return;
  }
  
  localStorage.setItem('token', token);
  console.log('✅ Token saved, length:', token.length);
}

export function removeToken() {
  if (typeof window === 'undefined') {
    return;
  }
  
  localStorage.removeItem('token');
}
```

### Использование в компоненте:

```javascript
// components/LoginForm.jsx
import { saveToken } from '@/utils/auth';

async function handleLogin(email, password) {
  try {
    const response = await fetch('/api/auth/login', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ email, password })
    });
    
    const data = await response.json();
    
    if (data.success && data.token) {
      // ✅ Сохраняем токен
      saveToken(data.token);
      
      // Перенаправляем на главную
      router.push('/');
    } else {
      alert('Login failed');
    }
  } catch (error) {
    console.error('Login error:', error);
    alert('Login failed');
  }
}
```

### Использование в API запросах:

```javascript
// utils/api.js
import { getAuthHeaders } from '@/utils/auth';

export async function apiRequest(url, options = {}) {
  const headers = {
    ...getAuthHeaders(),
    ...options.headers
  };
  
  // ✅ Проверяем, что токен есть
  if (!headers['Authorization']) {
    console.error('❌ No token available, redirecting to login');
    window.location.href = '/login';
    return;
  }
  
  const response = await fetch(url, {
    ...options,
    headers
  });
  
  if (response.status === 401) {
    // Токен невалидный - удаляем и перенаправляем
    localStorage.removeItem('token');
    window.location.href = '/login';
    return;
  }
  
  return response.json();
}

// Использование
const profile = await apiRequest('/api/user/profile');
```

---

## 🧪 Чеклист для проверки

- [ ] Токен сохраняется в localStorage после логина
- [ ] Токен извлекается перед каждым запросом
- [ ] Проверка на `undefined` перед использованием
- [ ] Токен имеет длину 200+ символов
- [ ] Заголовок отправляется как `Authorization: Bearer <token>`
- [ ] Обработка ошибки 401 (перенаправление на логин)
- [ ] Удаление токена при логауте

---

## 🔍 Отладка

### Проверка в консоли браузера:

```javascript
// 1. Проверить, есть ли токен
console.log('Token:', localStorage.getItem('token'));

// 2. Проверить длину токена
const token = localStorage.getItem('token');
console.log('Token length:', token?.length);

// 3. Проверить формат токена
const parts = token?.split('.');
console.log('Token parts:', parts?.length); // Должно быть 3

// 4. Проверить заголовок перед отправкой
const headers = {
  'Authorization': `Bearer ${localStorage.getItem('token')}`
};
console.log('Headers:', headers);
```

### Проверка в Network tab:

1. Открыть DevTools → Network
2. Выполнить запрос
3. Проверить заголовок `Authorization` в запросе
4. Должно быть: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`
5. НЕ должно быть: `Bearer undefined`

---

## 📝 Типичные ошибки

### ❌ Ошибка 1: Не сохраняется токен
```javascript
// ❌ НЕПРАВИЛЬНО:
const data = await response.json();
// Токен не сохраняется!

// ✅ ПРАВИЛЬНО:
const data = await response.json();
if (data.token) {
  localStorage.setItem('token', data.token);
}
```

### ❌ Ошибка 2: Использование undefined
```javascript
// ❌ НЕПРАВИЛЬНО:
const token = localStorage.getItem('token'); // может быть null
headers: {
  'Authorization': `Bearer ${token}` // Bearer null
}

// ✅ ПРАВИЛЬНО:
const token = localStorage.getItem('token');
if (!token) {
  // Обработка отсутствия токена
  return;
}
headers: {
  'Authorization': `Bearer ${token}`
}
```

### ❌ Ошибка 3: Неправильное имя ключа
```javascript
// ❌ НЕПРАВИЛЬНО:
localStorage.setItem('authToken', token); // Другое имя
const token = localStorage.getItem('token'); // null

// ✅ ПРАВИЛЬНО:
localStorage.setItem('token', token);
const token = localStorage.getItem('token');
```

---

## 🚀 После исправления

После исправления в логах должно быть:
```
🔍 Auth header preview: "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
🎫 Token extracted, length: 250+
✅ Auth OK for user ...
```

Вместо:
```
🔍 Auth header preview: "Bearer undefined"
🎫 Token extracted, length: 9
❌ JWT validation failed
```

---

**Статус:** ✅ Документация готова  
**Следующий шаг:** Исправить фронтенд согласно этой документации
