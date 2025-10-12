// package main

// import (
//     "context"
//     "database/sql"
//     "fmt"
//     "time"
// )

// type User struct {
//     ID   int
//     Name string
// }

// func getUserByID(ctx context.Context, db *sql.DB, userID int) (*User, error) {
//     // Create query context with timeout
//     queryCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
//     defer cancel()
    
//     var user User
    
//     // Execute query with context
//     err := db.QueryRowContext(queryCtx, 
//         "SELECT id, name FROM users WHERE id = ?", userID).
//         Scan(&user.ID, &user.Name)
    
//     if err != nil {
//         return nil, err
//     }
    
//     return &user, nil
// }

// func main() {
//     // db, _ := sql.Open("mysql", "connection_string")
//     // defer db.Close()
    
//     // Request-level context with 5-second timeout
//     ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
//     defer cancel()
    
//     // Query inherits timeout (but adds its own 2s limit)
//     // user, err := getUserByID(ctx, db, 123)
//     // if err != nil {
//     //     fmt.Println("Error:", err)
//     //     return
//     // }
    
//     // fmt.Printf("User: %+v\n", user)
// }