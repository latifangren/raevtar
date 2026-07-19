package main
import (
    "database/sql"
    "fmt"
    _ "modernc.org/sqlite"
    "log"
)
func main() {
    db, err := sql.Open("sqlite", "data.db")
    if err != nil { log.Fatal(err) }
    rows, err := db.Query("SELECT title FROM posts WHERE content_md LIKE '%BUILD FASTER%'")
    if err != nil { log.Fatal(err) }
    for rows.Next() {
        var t sql.NullString
        rows.Scan(&t)
        fmt.Println("Found in posts:", t.String)
    }
}
