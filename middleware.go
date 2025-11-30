package main

import (
    "github.com/gin-gonic/gin"
    "github.com/gin-contrib/sessions"
    "time"
)

const inactivityTimeout = 30 * time.Minute // 30 

// Middleware check is it authenticity 
func CheckInactivity() gin.HandlerFunc {
    return func(c *gin.Context) {
        session := sessions.Default(c)
        lastActivity := session.Get("last_activity")
        
        if lastActivity != nil {
            lastTime, ok := lastActivity.(int64)
            if ok {
                elapsed := time.Since(time.Unix(lastTime, 0))
                
                // if more than 30 minutes - go out 
                if elapsed > inactivityTimeout {
                    session.Clear()
                    session.Save()
                    c.Redirect(302, "/login?timeout=1")
                    c.Abort()
                    return
                }
            }
        }
        
        // update time from last activity here 
        session.Set("last_activity", time.Now().Unix())
        session.Save()
        
        c.Next()
    }
}
