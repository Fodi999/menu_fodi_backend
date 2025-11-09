# Academy Module

## Overview
The **Academy Module** provides a comprehensive culinary education platform with courses, lessons, quizzes, certificates, and progress tracking. Users can enroll in cooking courses, complete lessons, take quizzes, earn rewards, and receive PDF certificates upon completion.

## Architecture

```
academy/
├── dto/
│   └── requests.go      # DTOs for academy operations (14 types)
├── repo/
│   └── repository.go    # Data access layer (18 methods)
├── service/
│   └── service.go       # Business logic with gamification (11 methods)
├── transport/
│   └── http/
│       └── handlers.go  # HTTP handlers (10 endpoints)
├── module.go            # Route registration
└── README.md           # This file
```

## Features

### 1. Course Management
Browse and enroll in culinary courses:
- **Course Catalog**: Published courses with filtering
- **Course Details**: Comprehensive course information
- **Enrollment System**: Simple one-click enrollment
- **Multi-language Support**: Courses in different languages
- **Difficulty Levels**: Beginner to advanced courses

### 2. Lesson System
Structured learning with lessons:
- **Lesson Content**: Text and video lessons
- **Sequential Learning**: Ordered lesson progression
- **Progress Tracking**: Mark lessons as completed
- **Lesson Details**: Full content with duration

### 3. Quiz System
Knowledge assessment with rewards:
- **Random Questions**: 10 random questions per quiz
- **Multiple Choice**: 4 options per question
- **Scoring**: Percentage-based scores (0-100%)
- **Star Rewards**: 0-5 stars based on performance
- **Token Rewards**: 10 ChefTokens per star earned
- **XP Gain**: XP equal to quiz score

### 4. Progress Tracking
Monitor learning progress:
- **Completed Lessons**: Count of finished lessons
- **Total Lessons**: Total lessons in course
- **Quiz Score**: Best quiz performance
- **Stars Earned**: Total stars from quizzes
- **Completion Status**: Course completion flag

### 5. Certificate Generation
PDF certificates for completed courses:
- **Automatic Generation**: PDF created on completion
- **Personalized**: Student name, course name, level
- **Achievement Display**: Stars and quiz score shown
- **Digital Signature**: Official academy signature
- **Permanent Storage**: Certificates stored permanently

## API Endpoints

### Public Endpoints

#### GET /api/academy/courses
Get list of published courses with filtering.

**Query Parameters:**
- `language` (string): Filter by language
- `category` (string): Filter by category  
- `level` (int): Filter by difficulty level

**Response:**
```json
[
  {
    "id": "uuid",
    "title": "Italian Pasta Mastery",
    "description": "Learn authentic Italian pasta techniques",
    "category": "italian",
    "level": 2,
    "language": "en",
    "duration": 180,
    "isPublished": true,
    "createdAt": "2024-01-15T10:00:00Z"
  }
]
```

#### GET /api/academy/courses/{courseId}
Get course details.

**Response:**
```json
{
  "id": "uuid",
  "title": "Italian Pasta Mastery",
  "description": "Master the art of Italian pasta from scratch",
  "category": "italian",
  "level": 2,
  "language": "en",
  "duration": 180,
  "isPublished": true,
  "createdAt": "2024-01-15T10:00:00Z"
}
```

#### GET /api/academy/courses/{courseId}/lessons
Get lessons for a course.

**Response:**
```json
[
  {
    "id": "uuid",
    "courseId": "uuid",
    "title": "Introduction to Pasta Types",
    "content": "Learn about different pasta shapes...",
    "videoUrl": "https://...",
    "order": 1,
    "duration": 15,
    "isPublished": true,
    "createdAt": "2024-01-15T10:00:00Z"
  }
]
```

#### GET /api/academy/lessons/{lessonId}
Get lesson details.

**Response:**
```json
{
  "id": "uuid",
  "courseId": "uuid",
  "title": "Making Fresh Pasta Dough",
  "content": "Step-by-step instructions...",
  "videoUrl": "https://...",
  "order": 2,
  "duration": 20,
  "isPublished": true,
  "createdAt": "2024-01-15T10:00:00Z"
}
```

#### GET /api/academy/quizzes/{courseId}
Get random quiz questions (10 questions, without correct answers).

**Response:**
```json
[
  {
    "id": "uuid",
    "courseId": "uuid",
    "question": "What is the ideal water temperature for pasta?",
    "options": [
      "Boiling (100°C)",
      "Warm (60°C)",
      "Cold (20°C)",
      "Room temperature"
    ],
    "explanation": "Pasta should be cooked in boiling water..."
  }
]
```

### Protected Endpoints (JWT Required)

#### POST /api/academy/enroll
Enroll in a course.

**Request:**
```json
{
  "courseId": "uuid"
}
```

**Response:**
```json
{
  "message": "Successfully enrolled in course"
}
```

**Errors:**
- `400`: Already enrolled in this course
- `404`: Course not found

#### POST /api/academy/lessons/complete
Mark a lesson as completed.

**Request:**
```json
{
  "lessonId": "uuid"
}
```

**Response:**
```json
{
  "message": "Lesson completed successfully"
}
```

#### POST /api/academy/quizzes/submit
Submit quiz answers and get results.

**Request:**
```json
{
  "courseId": "uuid",
  "answers": [0, 2, 1, 3, 0, 1, 2, 3, 0, 1]
}
```

**Response:**
```json
{
  "score": 80,
  "correctAnswers": 8,
  "totalQuestions": 10,
  "starsEarned": 4,
  "reward": 40
}
```

**Star Calculation:**
- 90%+ → 5 stars
- 80-89% → 4 stars
- 70-79% → 3 stars
- 60-69% → 2 stars
- 50-59% → 1 star
- <50% → 0 stars

**Rewards:**
- **ChefTokens**: 10 tokens per star
- **XP**: Equal to quiz score percentage

#### GET /api/academy/progress/{courseId}
Get user progress for a course.

**Response:**
```json
{
  "userId": "uuid",
  "courseId": "uuid",
  "completedLessons": 8,
  "totalLessons": 12,
  "quizScore": 85,
  "starsEarned": 4,
  "isCompleted": false,
  "completedAt": null,
  "createdAt": "2024-01-20T10:00:00Z"
}
```

#### POST /api/academy/certificates/generate
Generate PDF certificate for completed course.

**Request:**
```json
{
  "courseId": "uuid"
}
```

**Response:**
```json
{
  "id": "uuid",
  "userId": "uuid",
  "courseId": "uuid",
  "courseName": "Italian Pasta Mastery",
  "userName": "Chef Mario",
  "level": 12,
  "stars": 4,
  "pdfUrl": "/certificates/certificate_Mario_123456.pdf",
  "signature": "Chef Dima Fomin - Culinary Academy AI",
  "issuedAt": "2024-02-01T15:30:00Z"
}
```

**Requirements:**
- Course must be completed (all lessons + quiz with score ≥50%)
- Certificate is generated only once per user per course

#### GET /api/academy/certificates
Get all user's certificates.

**Response:**
```json
{
  "certificates": [
    {
      "id": "uuid",
      "userId": "uuid",
      "courseId": "uuid",
      "courseName": "Italian Pasta Mastery",
      "userName": "Chef Mario",
      "level": 12,
      "stars": 4,
      "pdfUrl": "/certificates/certificate_Mario_123456.pdf",
      "signature": "Chef Dima Fomin - Culinary Academy AI",
      "issuedAt": "2024-02-01T15:30:00Z"
    }
  ],
  "total": 1
}
```

## Database Schema

### Tables Used

#### Course
- `id` (uuid): Course ID
- `title` (string): Course title
- `description` (text): Course description
- `category` (string): Course category
- `level` (int): Difficulty level
- `language` (string): Language code
- `duration` (int): Total duration in minutes
- `is_published` (bool): Publication status
- `created_at` (timestamp): Creation time

#### Lesson
- `id` (uuid): Lesson ID
- `course_id` (uuid): Parent course
- `title` (string): Lesson title
- `content` (text): Lesson content
- `video_url` (string): Video URL
- `order` (int): Lesson order
- `duration` (int): Lesson duration in minutes
- `is_published` (bool): Publication status
- `created_at` (timestamp): Creation time

#### QuizQuestion
- `id` (uuid): Question ID
- `course_id` (uuid): Parent course
- `question` (text): Question text
- `options` (json): Answer options array
- `correct_answer` (int): Index of correct answer
- `explanation` (text): Answer explanation

#### UserProgress
- `id` (uuid): Progress ID
- `user_id` (uuid): User ID
- `course_id` (uuid): Course ID
- `completed_lessons` (int): Number of completed lessons
- `total_lessons` (int): Total lessons in course
- `quiz_score` (int): Best quiz score
- `stars_earned` (int): Stars from quiz
- `is_completed` (bool): Completion status
- `last_accessed_at` (timestamp): Last access time
- `completed_at` (timestamp): Completion time

#### UserQuiz
- `id` (uuid): Quiz result ID
- `user_id` (uuid): User ID
- `course_id` (uuid): Course ID
- `score` (int): Score percentage
- `total_questions` (int): Number of questions
- `correct_answers` (int): Correct answers count
- `answers` (json): User answers array
- `stars_earned` (int): Stars awarded
- `completed_at` (timestamp): Completion time

#### Certificate
- `id` (uuid): Certificate ID
- `user_id` (uuid): User ID
- `course_id` (uuid): Course ID
- `course_name` (string): Course name
- `user_name` (string): Student name
- `level` (int): Student level at completion
- `stars` (int): Stars earned
- `pdf_url` (string): PDF file path
- `signature` (string): Official signature
- `issued_at` (timestamp): Issue time

## Business Logic

### Enrollment
1. Check if user already enrolled → Error if yes
2. Check if course exists → Error if not found
3. Create UserProgress record with zero progress
4. Return success message

### Lesson Completion
1. Get lesson details
2. Get or create user progress for the course
3. Increment completed lessons counter
4. Update total lessons count
5. Save progress

### Quiz Submission
1. Fetch all questions for the course
2. Compare user answers with correct answers
3. Calculate score percentage
4. Determine stars based on score
5. Save quiz result
6. Update user profile:
   - Add stars to total
   - Add XP equal to score
   - Add ChefToken reward (stars × 10)
7. Update course progress:
   - Save quiz score
   - Save stars earned
   - Check if course completed (all lessons + quiz ≥50%)
8. Return quiz results

### Certificate Generation
1. Check if certificate already exists → Return existing
2. Verify course completion → Error if not complete
3. Get course and user profile data
4. Generate PDF using CertificateService
5. Save certificate record to database
6. Return certificate details with PDF URL

## Repository Layer

### AcademyRepository Interface

```go
type AcademyRepository interface {
    // Courses
    GetCourses(filters *dto.CourseFilters) ([]*models.Course, error)
    GetCourseByID(courseID uuid.UUID) (*models.Course, error)
    
    // Lessons
    GetCourseLessons(courseID uuid.UUID) ([]*models.Lesson, error)
    GetLessonByID(lessonID uuid.UUID) (*models.Lesson, error)
    
    // Quiz
    GetQuizQuestions(courseID uuid.UUID) ([]*models.QuizQuestion, error)
    CreateUserQuiz(quiz *models.UserQuiz) error
    
    // Progress & Enrollment
    CheckEnrollment(userID, courseID uuid.UUID) (bool, error)
    CreateEnrollment(userID, courseID uuid.UUID) error
    GetUserProgress(userID, courseID uuid.UUID) (*models.UserProgress, error)
    CreateOrUpdateProgress(progress *models.UserProgress) error
    CompleteLesson(userID, lessonID uuid.UUID) error
    
    // Certificate
    GetCertificate(userID, courseID uuid.UUID) (*models.Certificate, error)
    CreateCertificate(cert *models.Certificate) error
    GetUserCertificates(userID uuid.UUID) ([]*models.Certificate, error)
    
    // User Profile
    GetUserProfile(userID uuid.UUID) (*models.UserProfile, error)
    UpdateUserProfile(profile *models.UserProfile) error
    AddWalletReward(userID uuid.UUID, amount float64, description string, relatedID uuid.UUID) error
}
```

## Service Layer

### Gamification System

**Star Rewards:**
- 5 stars = 90%+ quiz score → 50 ChefTokens
- 4 stars = 80-89% → 40 ChefTokens
- 3 stars = 70-79% → 30 ChefTokens
- 2 stars = 60-69% → 20 ChefTokens
- 1 star = 50-59% → 10 ChefTokens
- 0 stars = <50% → 0 ChefTokens

**XP System:**
- Quiz XP = quiz score percentage
- Example: 85% score = 85 XP

**Course Completion:**
- All lessons must be completed
- Quiz score must be ≥50%
- Automatically marks course as completed
- Enables certificate generation

## Error Handling

### Custom Errors
```go
var (
    ErrCourseNotFound      = errors.New("course not found")
    ErrLessonNotFound      = errors.New("lesson not found")
    ErrAlreadyEnrolled     = errors.New("already enrolled in this course")
    ErrNotEnrolled         = errors.New("not enrolled in this course")
    ErrCourseNotCompleted  = errors.New("course not completed yet")
    ErrCertificateExists   = errors.New("certificate already exists")
    ErrNoQuizQuestions     = errors.New("no quiz questions found")
)
```

### HTTP Status Codes
- `200 OK`: Successful operation
- `400 Bad Request`: Already enrolled, course not completed
- `401 Unauthorized`: Missing/invalid JWT
- `404 Not Found`: Course/lesson not found, not enrolled
- `500 Internal Server Error`: Database/system errors

## Security

### Authentication
- JWT middleware protects enrollment, progress, quiz, and certificate endpoints
- Public endpoints for browsing courses and lessons
- User ID extracted from JWT (cannot be spoofed)

### Quiz Security
- Correct answers not sent to client
- Only question text and options provided
- Answers validated server-side
- Random question selection prevents memorization

## Integration

### Dependencies
- **Wallet Module**: ChefToken rewards for quiz completion
- **User Module**: User profiles, levels, XP
- **Certificate Service**: PDF generation (internal/services)

### Used By
- Mobile app academy section
- Web platform learning portal
- User dashboard progress display
- Gamification system

## Performance Considerations

### Database Indexes
Recommended indexes for optimal performance:
```sql
CREATE INDEX idx_course_language ON Course(language);
CREATE INDEX idx_course_category ON Course(category);
CREATE INDEX idx_lesson_course ON Lesson(course_id);
CREATE INDEX idx_quiz_course ON QuizQuestion(course_id);
CREATE INDEX idx_progress_user_course ON UserProgress(user_id, course_id);
CREATE INDEX idx_certificate_user ON Certificate(user_id);
```

### Caching Opportunities
- Course catalog (rarely changes)
- Lesson content (static)
- Quiz questions (static per course)
- Certificate PDFs (generated once)

## Testing

### Unit Tests
```go
func TestSubmitQuiz(t *testing.T) {
    // Test cases:
    // - Correct star calculation
    // - Token reward calculation
    // - XP award
    // - Progress update
    // - Course completion check
}
```

### Integration Tests
```bash
# Enroll in course
curl -X POST http://localhost:8080/api/academy/enroll \
  -H "Authorization: Bearer $JWT" \
  -d '{"courseId": "uuid"}'

# Complete lesson
curl -X POST http://localhost:8080/api/academy/lessons/complete \
  -H "Authorization: Bearer $JWT" \
  -d '{"lessonId": "uuid"}'

# Submit quiz
curl -X POST http://localhost:8080/api/academy/quizzes/submit \
  -H "Authorization: Bearer $JWT" \
  -d '{"courseId": "uuid", "answers": [0,2,1,3,0,1,2,3,0,1]}'
```

## Monitoring

### Key Metrics
- Total enrollments
- Course completion rate
- Average quiz scores
- Certificates issued
- Most popular courses
- Student engagement (lessons completed)

### Logging
All operations logged with structured fields:
- User ID
- Course/Lesson ID
- Quiz scores
- Stars earned
- Certificate generation

## Future Enhancements

### Planned Features
1. **Discussion Forums**: Course-specific discussions
2. **Live Classes**: Real-time cooking sessions
3. **Peer Review**: Student recipe submissions
4. **Badges System**: Achievement badges
5. **Course Recommendations**: AI-powered suggestions
6. **Mobile Notifications**: Reminder for incomplete courses
7. **Leaderboards**: Top students per course
8. **Course Bundles**: Package multiple courses

### Technical Improvements
1. Video streaming optimization
2. Offline lesson access
3. Interactive quizzes with timers
4. Certificate blockchain verification
5. Multi-language quiz support
6. Adaptive learning paths

## Conclusion

The Academy module provides a complete e-learning platform for culinary education with:
- ✅ Comprehensive course management
- ✅ Structured lesson progression
- ✅ Gamified quiz system
- ✅ Reward mechanism (tokens + XP)
- ✅ Progress tracking
- ✅ PDF certificate generation
- ✅ Clean DDD architecture
- ✅ RESTful API design

Built with DDD principles and clean architecture for maintainability and scalability.
