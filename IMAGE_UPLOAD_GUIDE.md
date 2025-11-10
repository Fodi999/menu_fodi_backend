# Image Upload API Guide

## Overview

The backend now supports image uploads via the `/api/upload/image` endpoint. Images are uploaded to Cloudinary for cloud storage and retrieval.

## Endpoint Details

### POST /api/upload/image

Upload an image file to Cloudinary and get the image URL.

**Authentication**: Required (JWT Bearer Token)

**Content-Type**: `multipart/form-data`

### Request

```bash
curl -X POST http://localhost:8080/api/upload/image \
  -H "Authorization: Bearer {JWT_TOKEN}" \
  -F "image=@/path/to/image.jpg"
```

### Response (Success - 200 OK)

```json
{
  "success": true,
  "url": "http://res.cloudinary.com/...",
  "secureUrl": "https://res.cloudinary.com/...",
  "publicId": "culinary-academy/uuid-string",
  "message": "Image uploaded successfully"
}
```

### Error Responses

**401 Unauthorized** - Missing or invalid JWT token
```json
{
  "error": "unauthorized"
}
```

**400 Bad Request** - Invalid file or form data
```json
{
  "error": "image file is required"
}
```

**400 Bad Request** - Invalid file type
```json
{
  "error": "invalid file type - only JPEG, PNG, WebP, GIF, and SVG are allowed"
}
```

**400 Bad Request** - File too large
```json
{
  "error": "file size exceeds 10MB limit"
}
```

**500 Internal Server Error** - Cloudinary upload failed
```json
{
  "error": "failed to upload image"
}
```

## Frontend Integration

### Using Fetch API

```javascript
async function uploadImage(file, authToken) {
  const formData = new FormData();
  formData.append('image', file);

  try {
    const response = await fetch('/api/upload/image', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${authToken}`
      },
      body: formData
    });

    if (!response.ok) {
      const error = await response.json();
      console.error('Upload failed:', error.error);
      return null;
    }

    const data = await response.json();
    console.log('Image URL:', data.secureUrl);
    return data.secureUrl;
  } catch (error) {
    console.error('Upload error:', error);
    return null;
  }
}
```

### Using localStorage for Token

```javascript
// Store token after login
const loginResponse = await fetch('/api/auth/login', {
  method: 'POST',
  headers: { 'Content-Type': 'application/json' },
  body: JSON.stringify({ email, password })
});

const { token } = await loginResponse.json();
localStorage.setItem('authToken', token);

// Use token for image upload
const authToken = localStorage.getItem('authToken');
const imageUrl = await uploadImage(fileInput.files[0], authToken);
```

### In React Component

```jsx
import { useState } from 'react';

function ImageUploader() {
  const [uploading, setUploading] = useState(false);
  const [imageUrl, setImageUrl] = useState(null);

  const handleImageUpload = async (event) => {
    const file = event.target.files[0];
    if (!file) return;

    setUploading(true);
    try {
      const formData = new FormData();
      formData.append('image', file);

      const response = await fetch('/api/upload/image', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`
        },
        body: formData
      });

      if (!response.ok) {
        throw new Error('Upload failed');
      }

      const data = await response.json();
      setImageUrl(data.secureUrl);
    } catch (error) {
      console.error('Upload error:', error);
    } finally {
      setUploading(false);
    }
  };

  return (
    <div>
      <input 
        type="file" 
        accept="image/*" 
        onChange={handleImageUpload}
        disabled={uploading}
      />
      {uploading && <p>Uploading...</p>}
      {imageUrl && <img src={imageUrl} alt="Uploaded" />}
    </div>
  );
}

export default ImageUploader;
```

## Supported File Types

- **JPEG** (`image/jpeg`, `image/jpg`)
- **PNG** (`image/png`)
- **WebP** (`image/webp`)
- **GIF** (`image/gif`)
- **SVG** (`image/svg+xml`)

## Constraints

| Constraint | Value |
|-----------|-------|
| **Max File Size** | 10 MB |
| **Authentication** | Required (JWT) |
| **Method** | POST |
| **Content-Type** | multipart/form-data |
| **Form Field Name** | `image` |

## Implementation Details

### Backend Architecture

1. **Handler**: `MarketplaceHandlers.UploadImage()`
   - Validates JWT token via `middleware.GetUserID()`
   - Parses multipart form data (max 10MB)
   - Extracts image file from form
   - Validates file type by MIME type
   - Uploads to Cloudinary using `CloudinaryService`
   - Returns Cloudinary URLs and metadata

2. **Service**: `CloudinaryService.UploadImage()`
   - Generates signed upload request to Cloudinary
   - Uses folder: `culinary-academy`
   - Handles authentication with API credentials from environment variables
   - Returns upload response with URLs and public ID

3. **Route**: `/api/upload/image`
   - Path: `/upload/image` (registered in marketplace module)
   - Full path: `/api/upload/image`
   - Protected by JWT middleware
   - Method: `POST`

### Environment Variables Required

```bash
CLOUDINARY_CLOUD_NAME=your_cloud_name
CLOUDINARY_API_KEY=your_api_key
CLOUDINARY_API_SECRET=your_api_secret
```

## Security Considerations

1. **Authentication**: All requests require valid JWT token
2. **File Type Validation**: Only image MIME types are accepted
3. **File Size Limit**: Max 10MB to prevent abuse
4. **Cloudinary**: Uses signed uploads with API credentials
5. **Secure URLs**: Returns HTTPS URLs from Cloudinary (secureUrl)

## Cloudinary Integration

- **Folder**: `culinary-academy` - All images stored under this folder
- **Public ID**: Unique UUID generated for each upload
- **Base URL**: `https://res.cloudinary.com/{cloud_name}/`

## Error Handling

| Error | Cause | Solution |
|-------|-------|----------|
| 401 Unauthorized | Missing or invalid token | Ensure authToken is in Authorization header |
| 400 Bad Request (image file is required) | No file in form | Include `image` field in multipart form |
| 400 Bad Request (invalid file type) | Unsupported format | Only use supported image formats |
| 400 Bad Request (exceeds 10MB) | File too large | Compress or resize image before upload |
| 500 Internal Server Error | Cloudinary upload failed | Check Cloudinary credentials and service status |

## Testing

### Test with cURL

```bash
# Generate a test image
convert -size 100x100 xc:red test.jpg

# Upload with auth token
curl -X POST http://localhost:8080/api/upload/image \
  -H "Authorization: Bearer YOUR_JWT_TOKEN" \
  -F "image=@test.jpg"
```

### Test Response

```bash
curl -v -X POST http://localhost:8080/api/upload/image \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -F "image=@image.jpg" 2>&1 | grep -E "HTTP/|success|secureUrl"
```

## Common Issues

### Issue: "unauthorized" Error
**Solution**: Ensure JWT token is included in Authorization header and is valid

### Issue: "image file is required" Error
**Solution**: Verify form field name is exactly `image` and file is actually attached

### Issue: "invalid file type" Error
**Solution**: Check file MIME type - use supported formats (JPEG, PNG, WebP, GIF, SVG)

### Issue: "failed to upload image" Error
**Solution**: Check Cloudinary credentials in environment variables and service availability

## Example Workflow

1. **User Authentication**
   ```
   POST /api/auth/login
   Response: { "token": "eyJ..." }
   Store in localStorage.authToken
   ```

2. **Image Selection**
   - User selects image file via file input
   - Validate file locally (optional)

3. **Upload Image**
   ```
   POST /api/upload/image (with JWT token)
   Form: image=<file>
   Response: { "secureUrl": "https://..." }
   ```

4. **Use Image URL**
   - Display image using returned URL
   - Save URL to database if needed
   - Use URL in recipe or profile updates

## Next Steps

- Integrate image upload into recipe creation form
- Integrate image upload into user profile editor
- Add drag-and-drop upload UI
- Add image preview before upload
- Implement image compression/resizing on frontend

## Support

For issues or questions about the image upload endpoint:
1. Check Cloudinary credentials in `.env` file
2. Review error logs in server output
3. Verify file format and size requirements
4. Ensure JWT token is valid and not expired
