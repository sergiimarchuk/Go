package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "github.com/xuri/excelize/v2"
    "net/http"
    "time"
    "fmt"
)

// main page 
func HomePage(c *gin.Context) {
    c.HTML(http.StatusOK, "index.html", gin.H{
        "title": "Трекер рабочего времени",
    })
}

// page login
func LoginPage(c *gin.Context) {
    timeout := c.Query("timeout")
    
    c.HTML(http.StatusOK, "login.html", gin.H{
        "timeout": timeout == "1",
    })
}
// username 
func LoginHandler(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    
    user, err := GetUserByUsername(username)
    if err != nil {
        c.HTML(http.StatusOK, "login.html", gin.H{
            "error": "WTF your creds are wrong",
        })
        return
    }
    
    if !CheckPassword(password, user.Password) {
        c.HTML(http.StatusOK, "login.html", gin.H{
            "error": "shitty you are doing creds",
        })
        return
    }
    
    // save sessions
    session := sessions.Default(c)
    session.Set("user_id", user.ID)
    session.Set("username", user.Username)
    session.Save()
    
    c.Redirect(http.StatusFound, "/dashboard")
}

// dashboard (next page after login page)
func DashboardPage(c *gin.Context) {
    username := GetCurrentUsername(c)
    
    c.HTML(http.StatusOK, "dashboard.html", gin.H{
        "username": username,
    })
}

// new entry form 
func NewWorkLogPage(c *gin.Context) {
    c.HTML(http.StatusOK, "new_worklog.html", gin.H{})
}

// save new entry
func CreateWorkLogHandler(c *gin.Context) {
    date := c.PostForm("date")
    description := c.PostForm("description")
    hours := c.PostForm("hours")
    
    userID := GetCurrentUserID(c)
    
    _, err := db.Exec(
        "INSERT INTO worklogs (user_id, date, description, hours) VALUES (?, ?, ?, ?)",
        userID, date, description, hours,
    )
    
    if err != nil {
        c.HTML(http.StatusOK, "new_worklog.html", gin.H{
            "error": "error to save entry: " + err.Error(),
        })
        return
    }
    
    c.HTML(http.StatusOK, "new_worklog.html", gin.H{
        "success": "✅ Save new entry!",
    })
}

// List all entris 
// list after filtered
func WorkLogListPage(c *gin.Context) {
    userID := GetCurrentUserID(c)
    
    // values filters
    dateFrom := c.Query("date_from")
    dateTo := c.Query("date_to")
    search := c.Query("search")
    
    //  SQL build filters 
    query := `SELECT id, date, description, hours FROM worklogs WHERE user_id = ?`
    args := []interface{}{userID}
    
    if dateFrom != "" {
        query += ` AND date >= ?`
        args = append(args, dateFrom)
    }
    
    if dateTo != "" {
        query += ` AND date <= ?`
        args = append(args, dateTo)
    }
    
    if search != "" {
        query += ` AND description LIKE ?`
        args = append(args, "%"+search+"%")
    }
    
    query += ` ORDER BY date DESC`
    
    rows, err := db.Query(query, args...)
    
    if err != nil {
        c.HTML(http.StatusOK, "worklog_list.html", gin.H{
            "error": "errors loads ...",
        })
        return
    }
    defer rows.Close()
    
    var logs []WorkLog
    for rows.Next() {
        var log WorkLog
        var dateStr string
        rows.Scan(&log.ID, &dateStr, &log.Description, &log.Hours)
        
        log.Date, _ = time.Parse("2006-01-02", dateStr)
        logs = append(logs, log)
    }
    
    c.HTML(http.StatusOK, "worklog_list.html", gin.H{
        "logs":     logs,
        "dateFrom": dateFrom,
        "dateTo":   dateTo,
        "search":   search,
    })
}



// reports
// reports analytics 
func ReportsPage(c *gin.Context) {
    userID := GetCurrentUserID(c)
    
    rows, err := db.Query(`
        SELECT date, hours 
        FROM worklogs 
        WHERE user_id = ? 
        ORDER BY date ASC
    `, userID)
    
    if err != nil {
        c.HTML(http.StatusOK, "reports.html", gin.H{
            "error": "errors loads data",
        })
        return
    }
    defer rows.Close()
    
    var dates []string
    var hours []float64
    var totalHours float64
    
    // for goupe monthe 
    monthsMap := make(map[string]float64)
    // for grpoup weeks  
    weeksMap := make(map[string]fl at64)
    
    for rows.Next() {
        var date string
        var hour float64
        rows.Scan(&date, &hour)
        
        t, _ := time.Parse("2006-01-02", date)
        
        // group dates
	dates = append(dates, t.Format("02.01"))
        hours = append(hours, hour)
        totalHours += hour
        
        // group month
        monthKey := t.Format("2006-01")
        monthsMap[monthKey] += hour
        
        // group (ISO week)
        year, week := t.ISOWeek()
        weekKey := fmt.Sprintf("%d-W%02d", year, week)
        weeksMap[weekKey] += hour
    }
    
    // 
    var months []string
    var monthHours []float64
    for month, hours := range monthsMap {
        t, _ := time.Parse("2006-01", month)
        months = append(months, t.Format("01/2006"))
        monthHours = append(monthHours, hours)
    }
    
    // 
    var weeks []string
    var weekHours []float64
    for week, hours := range weeksMap {
        weeks = append(weeks, week)
        weekHours = append(weekHours, hours)
    }
    
    // 
    avgHours := 0.0
    if len(hours) > 0 {
        avgHours = totalHours / float64(len(hours))
    }
    
    c.HTML(http.StatusOK, "reports.html", gin.H{
        "dates":      dates,
        "hours":      hours,
        "months":     months,
        "monthHours": monthHours,
        "weeks":      weeks,
        "weekHours":  weekHours,
        "totalHours": totalHours,
        "avgHours":   avgHours,
        "daysCount":  len(hours),
    })
}

// 
func EditWorkLogPage(c *gin.Context) {
    id := c.Param("id")
    
    var log WorkLog
    var dateStr string
    err := db.QueryRow("SELECT id, date, description, hours FROM worklogs WHERE id = ?", id).
        Scan(&log.ID, &dateStr, &log.Description, &log.Hours)
    
    if err != nil {
        c.Redirect(http.StatusFound, "/worklog/list")
        return
    }
    
    log.Date, _ = time.Parse("2006-01-02", dateStr)
    
    c.HTML(http.StatusOK, "edit_worklog.html", gin.H{
        "log": log,
    })
}

//
func UpdateWorkLogHandler(c *gin.Context) {
    id := c.Param("id")
    date := c.PostForm("date")
    description := c.PostForm("description")
    hours := c.PostForm("hours")
    
    _, err := db.Exec(
        "UPDATE worklogs SET date = ?, description = ?, hours = ? WHERE id = ?",
        date, description, hours, id,
    )
    
    if err != nil {
        c.HTML(http.StatusOK, "edit_worklog.html", gin.H{
            "error": "Ошибка обновления",
        })
        return
    }
    
    c.Redirect(http.StatusFound, "/worklog/list")
}

// 
func DeleteWorkLogHandler(c *gin.Context) {
    id := c.Param("id")
    
    _, err := db.Exec("DELETE FROM worklogs WHERE id = ?", id)
    if err != nil {
        c.String(http.StatusInternalServerError, "Ошибка удаления")
        return
    }
    
    c.Redirect(http.StatusFound, "/worklog/list")
}

// 
func LogoutHandler(c *gin.Context) {
    session := sessions.Default(c)
    session.Clear()
    session.Save()
    
    c.Redirect(http.StatusFound, "/login")
}

//  Excel
func ExportWorkLogHandler(c *gin.Context) {
    userID := GetCurrentUserID(c)
    username := GetCurrentUsername(c)

    // 
    dateFrom := c.Query("date_from")
    dateTo := c.Query("date_to")
    search := c.Query("search")

    // SQL
    query := `SELECT date, description, hours FROM worklogs WHERE user_id = ?`
    args := []interface{}{userID}

    if dateFrom != "" {
        query += ` AND date >= ?`
        args = append(args, dateFrom)
    }

    if dateTo != "" {
        query += ` AND date <= ?`
        args = append(args, dateTo)
    }

    if search != "" {
        query += ` AND description LIKE ?`
        args = append(args, "%"+search+"%")
    }

    query += ` ORDER BY date DESC`

    rows, err := db.Query(query, args...)
    if err != nil {
        c.String(http.StatusInternalServerError, "Ошибка получения данных")
        return
    }
    defer rows.Close()

    // Excel 
    f := excelize.NewFile()
    sheetName := "Рабочие часы"
    index, _ := f.NewSheet(sheetName)

    // 
    f.SetCellValue(sheetName, "A1", "Дата")
    f.SetCellValue(sheetName, "B1", "Описание")
    f.SetCellValue(sheetName, "C1", "Часы")

    //
    headerStyle, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true, Size: 12},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"#667eea"}, Pattern: 1},
        Alignment: &excelize.Alignment{Horizontal: "center"},
    })
    f.SetCellStyle(sheetName, "A1", "C1", headerStyle)

    // 
    row := 2
    totalHours := 0.0

    for rows.Next() {
        var date, description string
        var hours float64
        rows.Scan(&date, &description, &hours)

        t, _ := time.Parse("2006-01-02", date)

        f.SetCellValue(sheetName, "A"+fmt.Sprintf("%d", row), t.Format("02.01.2006"))
        f.SetCellValue(sheetName, "B"+fmt.Sprintf("%d", row), description)
        f.SetCellValue(sheetName, "C"+fmt.Sprintf("%d", row), hours)

        totalHours += hours
        row++
    }

    // 
    row++
    f.SetCellValue(sheetName, "B"+fmt.Sprintf("%d", row), "ИТОГО:")
    f.SetCellValue(sheetName, "C"+fmt.Sprintf("%d", row), totalHours)

    totalStyle, _ := f.NewStyle(&excelize.Style{
        Font: &excelize.Font{Bold: true, Size: 12},
        Fill: excelize.Fill{Type: "pattern", Color: []string{"#4CAF50"}, Pattern: 1},
    })
    f.SetCellStyle(sheetName, "B"+fmt.Sprintf("%d", row), "C"+fmt.Sprintf("%d", row), totalStyle)

    //
    f.SetColWidth(sheetName, "A", "A", 15)
    f.SetColWidth(sheetName, "B", "B", 50)
    f.SetColWidth(sheetName, "C", "C", 10)

    f.SetActiveSheet(index)
    f.DeleteSheet("Sheet1")

    // 
    fileName := fmt.Sprintf("worklog_%s_%s.xlsx", username, time.Now().Format("2006-01-02"))

    c.Header("Content-Type", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet")
    c.Header("Content-Disposition", "attachment; filename="+fileName)

    if err := f.Write(c.Writer); err != nil {
        c.String(http.StatusInternalServerError, "Ошибка создания файла")
    }
}

// 
func RegisterPage(c *gin.Context) {
    c.HTML(http.StatusOK, "register.html", gin.H{})
}

// 
func RegisterHandler(c *gin.Context) {
    username := c.PostForm("username")
    password := c.PostForm("password")
    passwordConfirm := c.PostForm("password_confirm")

    // 
    if username == "" || password == "" {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "Заполните все поля",
        })
        return
    }

    if len(username) < 3 {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "more then 3 symbols",
        })
        return
    }

    if len(password) < 6 {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "more than 6 symbols",
        })
        return
    }

    if password != passwordConfirm {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "passwords are not equals",
        })
        return
    }

    // 
    existingUser, _ := GetUserByUsername(username)
    if existingUser != nil {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "Пользователь с таким логином уже существует",
        })
        return
    }

    // 
    err := CreateUser(username, password)
    if err != nil {
        c.HTML(http.StatusOK, "register.html", gin.H{
            "error": "error: " + err.Error(),
        })
        return
    }

    // 
    user, _ := GetUserByUsername(username)
    session := sessions.Default(c)
    session.Set("user_id", user.ID)
    session.Set("username", user.Username)
    session.Save()

    c.Redirect(http.StatusFound, "/dashboard")
}
