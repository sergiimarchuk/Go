package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/gin-contrib/sessions/cookie"
    "log"
    "html/template"
    "net/http"
)

func main() {
    // Инициализируем БД
    if err := InitDB(); err != nil {
        log.Fatal("Ошибка БД:", err)
    }

    r := gin.Default()
    
    // Настраиваем сессии
    store := cookie.NewStore([]byte("super-secret-key-change-me-in-production"))
    store.Options(sessions.Options{
        Path:     "/",
        MaxAge:   0, // Session cookie - умрёт при закрытии браузера
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
    })
    r.Use(sessions.Sessions("mysession", store))
    
    // Добавляем функцию сложения для шаблонов
    r.SetFuncMap(template.FuncMap{
        "add": func(a, b float64) float64 {
            return a + b
        },
    })
    
    // Загружаем HTML шаблоны
    r.LoadHTMLGlob("templates/*")
    r.Static("/static", "./static")

    // ========== WEB ROUTES ==========
    
    // Публичные роуты
    r.GET("/", HomePage)
    r.GET("/login", LoginPage)
    r.POST("/login", LoginHandler)
    r.GET("/register", RegisterPage)
    r.POST("/register", RegisterHandler)
    
    // Защищённые роуты (требуют авторизации)
    authorized := r.Group("/")
    authorized.Use(AuthRequired())
    authorized.Use(CheckInactivity())
    {
        authorized.GET("/dashboard", DashboardPage)
        authorized.GET("/worklog/new", NewWorkLogPage)
        authorized.POST("/worklog/create", CreateWorkLogHandler)
        authorized.GET("/worklog/list", WorkLogListPage)
        authorized.GET("/reports", ReportsPage)
        authorized.GET("/worklog/edit/:id", EditWorkLogPage)
        authorized.POST("/worklog/update/:id", UpdateWorkLogHandler)
        authorized.POST("/worklog/delete/:id", DeleteWorkLogHandler)
        authorized.GET("/worklog/export", ExportWorkLogHandler)
        authorized.GET("/logout", LogoutHandler)
    }
    
    // ========== API ROUTES ==========
    
    api := r.Group("/api/v1")
    {
        // Публичные API endpoints
        api.POST("/auth/login", APILogin)
        api.POST("/auth/register", APIRegister)
        
        // Защищённые API endpoints (требуют JWT токен)
        apiAuth := api.Group("/")
        apiAuth.Use(JWTAuthMiddleware())
        {
            // Worklogs
            apiAuth.GET("/worklogs", APIGetWorkLogs)
            apiAuth.POST("/worklogs", APICreateWorkLog)
            apiAuth.PUT("/worklogs/:id", APIUpdateWorkLog)
            apiAuth.DELETE("/worklogs/:id", APIDeleteWorkLog)
            
            // Statistics
            apiAuth.GET("/stats", APIGetStats)
        }
    }
 
    log.Println("🚀 Сервер запущен на http://localhost:8080")
    log.Println("📡 API доступно на http://localhost:8080/api/v1")
    r.Run(":8080")
}
