package main

import (
	"database/sql"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type PersonInfo struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	City        string `json:"city"`
	State       string `json:"state"`
	Street1     string `json:"street1"`
	Street2     string `json:"street2"`
	ZipCode     string `json:"zip_code"`
}

var pInfo PersonInfo

func Task1(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	personID := c.Param("person_id")
	sqlQuery := `
        SELECT
            p.name,
            ph.number AS phone_number,
            a.city,
            a.state,
            a.street1,
            a.street2,
            a.zip_code
        FROM
            person p
        JOIN
            phone ph ON p.id = ph.person_id
        JOIN
            address_join aj ON p.id = aj.person_id
        JOIN
            address a ON aj.address_id = a.id
        WHERE
            p.id = ?
    `

	if err := db.QueryRow(sqlQuery, personID).Scan(&pInfo.Name, &pInfo.PhoneNumber, &pInfo.City, &pInfo.State, &pInfo.Street1, &pInfo.Street2, &pInfo.ZipCode); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, pInfo)
}

func Task2(c *gin.Context) {
	db := c.MustGet("db").(*sql.DB)

	if err := c.ShouldBindJSON(&pInfo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	res, err := db.Exec("INSERT INTO person (name) VALUES (?)", pInfo.Name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	personID, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = db.Exec("INSERT INTO phone (person_id, number) VALUES (?, ?)", personID, pInfo.PhoneNumber)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	res, err = db.Exec("INSERT INTO address (city, state, street1, street2, zip_code) VALUES (?, ?, ?, ?, ?)",
		pInfo.City, pInfo.State, pInfo.Street1, pInfo.Street2, pInfo.ZipCode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	addressID, err := res.LastInsertId()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	_, err = db.Exec("INSERT INTO address_join (person_id, address_id) VALUES (?, ?)", personID, addressID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Person successfully created"})
}

func main() {

	// Database connection
	db, err := sql.Open("mysql", "root:<password>@tcp(localhost:3306)/cetec")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// Routes for the Tasks
	router := gin.Default()

	// Middleware
	router.Use(func(c *gin.Context) {
		c.Set("db", db)
		c.Next()
	})

	// TASK1
	router.GET("/person/:person_id/info", Task1)

	// TASK2
	router.POST("/person/create", Task2)

	router.Run(":8080")
}
