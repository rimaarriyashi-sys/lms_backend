package main

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

var DB *gorm.DB

type Role struct {
	ID       uint   `gorm:"primaryKey"`
	RoleName string `gorm:"type:varchar(50);unique;not null"`
}

type User struct {
	ID           string `gorm:"primaryKey;type:varchar(36)"`
	NISN_NIP     string `gorm:"type:varchar(50)"`
	NIS          string `gorm:"type:varchar(50)"`
	Name         string `gorm:"type:varchar(100);not null"`
	Email        string `gorm:"type:varchar(100);unique;not null"`
	PasswordHash string `gorm:"type:varchar(255);not null"`
	TempatLahir  string `gorm:"type:varchar(100)"`
	TanggalLahir string `gorm:"type:date"`
	JenisKelamin string `gorm:"type:varchar(20)"`
	Specialty    string `gorm:"type:varchar(100)"`

	RoleID uint
	Role   Role `gorm:"foreignKey:RoleID"`

	ClassID *uint
	Class   *Class `gorm:"foreignKey:ClassID"`

	TaughtClasses []Class `gorm:"many2many:teacher_classes;"`

	CreatedAt time.Time
}

func (u *User) BeforeCreate(tx *gorm.DB) (err error) {
	if u.ID == "" {
		u.ID = uuid.New().String()
	}
	return
}

type Major struct {
	ID        uint   `gorm:"primaryKey"`
	MajorName string `gorm:"type:varchar(100);not null;unique"`
	CreatedAt time.Time
}

type Class struct {
	ID                uint   `gorm:"primaryKey"`
	ClassName         string `gorm:"type:varchar(50);not null;unique"`
	HomeroomTeacherID *string
	HomeroomTeacher   *User `gorm:"foreignKey:HomeroomTeacherID"`

	MajorID *uint
	Major   *Major `gorm:"foreignKey:MajorID"`

	CreatedAt time.Time
}

type Subject struct {
	ID          uint   `gorm:"primaryKey"`
	SubjectName string `gorm:"type:varchar(100);not null"`
	TeacherID   string
	Teacher     User `gorm:"foreignKey:TeacherID"`
}

type Schedule struct {
	ID        uint `gorm:"primaryKey"`
	ClassID   uint
	Class     Class `gorm:"foreignKey:ClassID"`
	SubjectID uint
	Subject   Subject `gorm:"foreignKey:SubjectID"`
	DayOfWeek string  `gorm:"type:varchar(20);not null"`
	StartTime string  `gorm:"type:varchar(10);not null"`
	EndTime   string  `gorm:"type:varchar(10);not null"`
}

type Material struct {
	ID         uint `gorm:"primaryKey"`
	SubjectID  uint
	Title      string `gorm:"type:varchar(255);not null"`
	ContentURL string `gorm:"type:text"`
	UploadedBy string
}

type Assignment struct {
	ID        uint `gorm:"primaryKey"`
	SubjectID uint
	Title     string `gorm:"type:varchar(255);not null"`
	Deadline  time.Time
	MaxScore  int `gorm:"default:100"`
}

type Submission struct {
	ID           uint `gorm:"primaryKey"`
	AssignmentID uint
	StudentID    string
	FileURL      string `gorm:"type:text;not null"`
	Score        int
	Feedback     string `gorm:"type:text"`
}

type Quiz struct {
	ID        uint `gorm:"primaryKey"`
	SubjectID uint
	Subject   Subject `gorm:"foreignKey:SubjectID"`
	TeacherID string
	Teacher   User   `gorm:"foreignKey:TeacherID"`
	Title     string `gorm:"type:varchar(255);not null"`
	QuizType  string `gorm:"type:varchar(20);not null"` // "kuis" atau "ujian"

	IsTimed     bool `gorm:"default:false"`
	DurationMin int

	StartTime time.Time
	EndTime   time.Time

	CreatedAt time.Time
}

type Question struct {
	ID           uint   `gorm:"primaryKey"`
	QuizID       uint
	QuestionText string `gorm:"type:text;not null"`
	QuestionType string `gorm:"type:varchar(20);not null"` // "pilihan_ganda" atau "esai"

	OptionA       string `gorm:"type:text"`
	OptionB       string `gorm:"type:text"`
	OptionC       string `gorm:"type:text"`
	OptionD       string `gorm:"type:text"`
	CorrectOption string `gorm:"type:varchar(1)"`

	Score int `gorm:"default:10"`
}

type QuizSubmission struct {
	ID        uint `gorm:"primaryKey"`
	QuizID    uint
	Quiz      Quiz `gorm:"foreignKey:QuizID"`
	StudentID string
	Student   User `gorm:"foreignKey:StudentID"`

	StartedAt   time.Time
	SubmittedAt *time.Time

	TotalScore int
	Status     string `gorm:"type:varchar(20);default:'mengerjakan'"`
}

type Answer struct {
	ID               uint `gorm:"primaryKey"`
	QuizSubmissionID uint
	QuestionID       uint
	Question         Question `gorm:"foreignKey:QuestionID"`

	SelectedOption string `gorm:"type:varchar(1)"`
	EssayAnswer    string `gorm:"type:text"`

	Score          int
	TeacherComment string `gorm:"type:text"`
}

type Grade struct {
	ID        uint `gorm:"primaryKey"`
	StudentID string
	Student   User `gorm:"foreignKey:StudentID"`
	SubjectID uint
	Subject   Subject `gorm:"foreignKey:SubjectID"`

	AssignmentAvg float64 // rata-rata nilai tugas/projek
	QuizAvg       float64 // rata-rata nilai kuis/ujian
	FinalScore    float64 // nilai akhir gabungan

	GeneratedAt time.Time
}

func seedRoles() {
	roles := []Role{
		{ID: 1, RoleName: "Admin"},
		{ID: 2, RoleName: "Guru"},
		{ID: 3, RoleName: "Siswa"},
		{ID: 4, RoleName: "Kurikulum"},
		{ID: 5, RoleName: "Kepsek"},
	}
	for _, role := range roles {
		DB.FirstOrCreate(&role, Role{ID: role.ID})
	}
	log.Println("✅ Data Role berhasil disuntikkan!")
}

func main() {
	var err error
	DB, err = gorm.Open(sqlite.Open("lms_data.db"), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	DB.AutoMigrate(
	&Role{}, &User{}, &Class{}, &Major{}, &Subject{}, &Schedule{},
	&Material{}, &Assignment{}, &Submission{},
	&Quiz{}, &Question{}, &QuizSubmission{}, &Answer{},
	&Grade{},
)
	seedRoles()

	r := gin.Default()
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{"http://localhost:3000"}, AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization"}, AllowCredentials: true, MaxAge: 12 * time.Hour,
	}))

	r.GET("/api/status", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "sukses"}) })
	r.POST("/api/register", Register)
	r.POST("/api/login", Login)

	protected := r.Group("/api")
	protected.Use(AuthMiddleware())
	{
		protected.GET("/admin/classes", GetClasses)
		protected.GET("/admin/classes/:id/details", GetClassDetails)
		protected.POST("/admin/classes", CreateClass)
		protected.PUT("/admin/classes/:id", UpdateClass)
		protected.DELETE("/admin/classes/:id", DeleteClass)

		protected.GET("/admin/users", GetUsers)
		protected.POST("/admin/users", CreateUser)
		protected.PUT("/admin/users/:id", UpdateUser)
		protected.DELETE("/admin/users/:id", DeleteUser)

		protected.PUT("/admin/students/:student_id/class", AssignStudentToClass)
		protected.POST("/admin/import", ImportDataExcel)

		protected.GET("/admin/majors", GetMajors)
		protected.POST("/admin/majors", CreateMajor)
		protected.PUT("/admin/majors/:id", UpdateMajor)
		protected.DELETE("/admin/majors/:id", DeleteMajor)

		// ↓↓↓ TAMBAHIN INI ↓↓↓
		protected.POST("/guru/quizzes", CreateQuiz)
		protected.GET("/guru/quizzes", GetMyQuizzes)
		protected.POST("/guru/quizzes/:quiz_id/questions", AddQuestion)
		protected.GET("/guru/quizzes/:quiz_id/questions", GetQuizQuestions)
		protected.GET("/guru/quizzes/:quiz_id/submissions", GetQuizSubmissions)
		protected.GET("/guru/submissions/:submission_id/answers", GetSubmissionAnswers)
		protected.PUT("/guru/answers/:answer_id/grade", GradeEssayAnswer)
		// ↑↑↑ SAMPAI SINI ↑↑↑

		protected.GET("/siswa/quizzes", GetAvailableQuizzes)
		protected.GET("/siswa/quizzes/:quiz_id/questions", GetQuizForStudent)
		protected.POST("/siswa/quizzes/:quiz_id/start", StartQuiz)
		protected.POST("/siswa/submissions/:submission_id/submit", SubmitQuiz)
		protected.GET("/siswa/submissions/:submission_id/result", GetMyQuizResult)

		// ↓↓↓ TAMBAHIN INI (route Grade) ↓↓↓
		protected.POST("/guru/subjects/:subject_id/generate-grades", GenerateGrades)
		protected.GET("/guru/subjects/:subject_id/grades", GetGradesBySubject)
		protected.GET("/siswa/grades", GetMyGrades)
		// ↑↑↑ SAMPAI SINI ↑↑↑

		protected.GET("/viewer/grades", ViewAllGrades)
		protected.GET("/viewer/assignments", ViewAllAssignments)
		protected.GET("/viewer/quizzes", ViewQuizzesByTeacher)
		protected.GET("/viewer/teachers", ViewAllTeachers)
		protected.GET("/viewer/students", ViewAllStudents)
	
	}

	log.Println("server on di: http://localhost:8080")
	r.Run(":8080")
}