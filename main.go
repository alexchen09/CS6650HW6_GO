package main

import (
	"database/sql"
	"log"
	"os"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
	"github.com/google/uuid"
)

// We'll store a global *sql.DB for simplicity in a demo.
var db *sql.DB

func main() {
	// Read MySQL DSN from environment variable, e.g.:
	// DB_DSN = "user:pass@tcp(mydb.xxxx.us-west-2.rds.amazonaws.com:3306)/mydemodb"
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN environment variable not set")
	}

	var err error
	db, err = sql.Open("mysql", dsn)
	if err != nil {
		log.Fatalf("Failed to open DB: %v", err)
	}

	// Test the DB connection quickly
	err = db.Ping()
	if err != nil {
		log.Fatalf("Failed to connect to DB: %v", err)
	}

	// Create table if not exists
	_, err = db.Exec(`
	CREATE TABLE IF NOT EXISTS new_table (
		albumID VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255),
		artist VARCHAR(255),
		price FLOAT,
		image BLOB
	) ENGINE=InnoDB;
	`)
	if err != nil {
		log.Fatalf("Failed to create table: %v", err)
	}

	// Setup Gin engine
	r := gin.Default()

	// Health check route
	r.GET("/count", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// GET /album/:albumID -> returns album details
	r.GET("/album/:albumID", func(c *gin.Context) {
    albumID := c.Param("albumID")
		log.Printf("Received GET request for albumID: %s", albumID)
    var album struct {
        AlbumID string  `json:"albumID"`
        Name    string  `json:"name"`
        Artist  string  `json:"artist"`
        Price   float64 `json:"price"`
        Image   []byte  `json:"image"`
    }

    query := "SELECT albumID FROM new_table WHERE albumID = ?"
    row := db.QueryRow(query, albumID)
    err := row.Scan(&album.AlbumID)
    if err != nil {
        if err == sql.ErrNoRows {
            c.JSON(404, gin.H{"message": "Album not found"})
        } else {
            c.JSON(500, gin.H{"message": "Error fetching album"})
        }
        return
    }

    c.JSON(200, gin.H{"albumID": album.AlbumID})  
	})


	// POST /add -> insert new album
	r.POST("/add", func(c *gin.Context) {
		var newAlbum struct {
			Name   string  `json:"name"`
			Artist string  `json:"artist"`
			Price  float64 `json:"price"`
		}

		if err := c.ShouldBindJSON(&newAlbum); err != nil {
			c.JSON(400, gin.H{"error": "Invalid JSON data"})
			return
		}

		// Insert album
		query := "INSERT INTO new_table (albumID, name, artist, price) VALUES (?, ?, ?, ?)"
		albumID := uuid.New().String()
		_, err := db.Exec(query, albumID, newAlbum.Name, newAlbum.Artist, newAlbum.Price)
		if err != nil {
			c.JSON(500, gin.H{"error": "Failed to insert album into database"})
			return
		}

		c.JSON(201, gin.H{
			"albumID": albumID,
		})
	})

	// Optionally, pass a port via environment variable or default to 8080
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s ...", port)
	r.Run(":" + port)
}
