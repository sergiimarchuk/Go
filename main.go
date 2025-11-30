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
    // initio. db 
    if err := InitDB(); err != nil {
        log.Fatal("fehler db:", err)
    }

    r := gin.Default()
    
    // conf or put settings for sessions 
    store := cookie.NewStore([]byte("super-secret-key-change-me-in-production"))
    store.Options(sessions.Options{
        Path:     "/",
        MaxAge:   0, // Session cookie - out session after browser closed
        HttpOnly: true,
        Secure:   false,
        SameSite: http.SameSiteLaxMode,
    })
    r.Use(sessions.Sessions("mysession", store))
    
    // added func for tamplates 
    r.SetFuncMap(template.FuncMap{
        "add": func(a, b float64) float64 {
            return a + b
        },
    })
    
    // load HTML temlates
    r.LoadHTMLGlob("templates/*")
    r.Static("/static", "./static")

    // ========== WEB ROUTES ==========
    
    // public routes
    r.GET("/", HomePage)
    r.GET("/login", LoginPage)
    r.POST("/login", LoginHandler)
    r.GET("/register", RegisterPage)
    r.POST("/register", RegisterHandler)
    
    // secured routes only after creds done successful 
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
        // public API endpoints
        api.POST("/auth/login", APILogin)
        api.POST("/auth/register", APIRegister)
        
        // isecured API endpoints (needs JWT token)
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
 
    log.Println("🚀 up see there http://localhost:8080")
    log.Println("📡 API has to be available there http://localhost:8080/api/v1")
    r.Run(":8080")
}
