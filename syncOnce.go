package main

import (
	"fmt"
	"sync"
)

var initialized bool

func main() {
	var once sync.Once
	var wg sync.WaitGroup

	initialize := func() {
		initialized = true
		// Complex initialization logic can go here
	}

	for i := 0; i < 10; i++ {
		// spawn a goroutine
		go func() {
			wg.Add(1)
			defer wg.Done()
			once.Do(initialize)
			fmt.Println(initialized)
		}()
	}
	
	wg.Wait()
	fmt.Println("All goroutines completed")
}

// Inside sync.Once simplified psedudo-code

// type Once struct {
//     done uint32
//     m    Mutex
// }

// func (o *Once) Do(f func()) {
//     if atomic.LoadUint32(&o.done) == 1 {
//         return
//     }
//     o.doSlow(f)
// }

// func (o *Once) doSlow(f func()) {
//     o.m.Lock()
//     defer o.m.Unlock()
//     if o.done == 0 {
//         defer atomic.StoreUint32(&o.done, 1)
//         f()
//     }
// }

// // Example usage in singleton pattern
// var db *sql.DB
// var once sync.Once

// func GetDB() *sql.DB {
//     once.Do(func() {
//         db, _ = sql.Open("postgres", connStr)
//     })
//     return db
// }

// Example usage in caching
// var cache map[string]string
// var once sync.Once

// func LoadCache() {
//     once.Do(func() {
//         cache = loadFromDisk()
//     })
// }

// Example usage in server initialization
// func initServer() {
//     once.Do(startServer)
// }



package main

import (
    "database/sql"
    "fmt"
    "log"
    "sync"
    "time"
)

// // Each Once manages independent initialization
// var (
//     dbOnce     sync.Once
//     cacheOnce  sync.Once
//     loggerOnce sync.Once
//     configOnce sync.Once
// )

// var (
//     db     *sql.DB
//     cache  map[string]string
//     logger *log.Logger
//     config map[string]interface{}
// )

// // Database initialization - might be called from many goroutines
// func GetDB() *sql.DB {
//     dbOnce.Do(func() {
//         fmt.Println("Initializing database connection...")
//         time.Sleep(100 * time.Millisecond) // Simulate DB connection time
//         // db, _ = sql.Open("postgres", connStr)
//         db = &sql.DB{} // Simplified for demo
//         fmt.Println("Database initialized!")
//     })
//     return db
// }

// // Cache initialization - independent of DB
// func GetCache() map[string]string {
//     cacheOnce.Do(func() {
//         fmt.Println("Loading cache from disk...")
//         time.Sleep(50 * time.Millisecond)
//         cache = make(map[string]string)
//         cache["key1"] = "value1"
//         fmt.Println("Cache loaded!")
//     })
//     return cache
// }

// // Logger initialization
// func GetLogger() *log.Logger {
//     loggerOnce.Do(func() {
//         fmt.Println("Initializing logger...")
//         time.Sleep(20 * time.Millisecond)
//         logger = log.Default()
//         fmt.Println("Logger initialized!")
//     })
//     return logger
// }

// // Config initialization
// func GetConfig() map[string]interface{} {
//     configOnce.Do(func() {
//         fmt.Println("Loading configuration...")
//         time.Sleep(30 * time.Millisecond)
//         config = make(map[string]interface{})
//         config["app_name"] = "MyApp"
//         fmt.Println("Config loaded!")
//     })
//     return config
// }

// func main() {
//     var wg sync.WaitGroup

//     // Simulate multiple services starting concurrently
//     // Each might need different resources
    
//     // Service 1: Needs DB and Logger
//     for i := 0; i < 3; i++ {
//         wg.Add(1)
//         go func(id int) {
//             defer wg.Done()
//             fmt.Printf("Service1-%d: Getting DB...\n", id)
//             GetDB()
//             fmt.Printf("Service1-%d: Getting Logger...\n", id)
//             GetLogger()
//             fmt.Printf("Service1-%d: Ready!\n", id)
//         }(i)
//     }

//     // Service 2: Needs Cache and Config
//     for i := 0; i < 3; i++ {
//         wg.Add(1)
//         go func(id int) {
//             defer wg.Done()
//             fmt.Printf("Service2-%d: Getting Cache...\n", id)
//             GetCache()
//             fmt.Printf("Service2-%d: Getting Config...\n", id)
//             GetConfig()
//             fmt.Printf("Service2-%d: Ready!\n", id)
//         }(i)
//     }

//     // Service 3: Needs everything
//     for i := 0; i < 2; i++ {
//         wg.Add(1)
//         go func(id int) {
//             defer wg.Done()
//             fmt.Printf("Service3-%d: Getting all resources...\n", id)
//             GetDB()
//             GetCache()
//             GetLogger()
//             GetConfig()
//             fmt.Printf("Service3-%d: Ready!\n", id)
//         }(i)
//     }

//     wg.Wait()
//     fmt.Println("\n✅ All services initialized!")
// }