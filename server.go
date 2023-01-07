package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	_ "github.com/lib/pq"
)

type Expense struct {
	Id     int      `json:"id"`
	Title  string   `json:"title"`
	Amount float64  `json:"amount"`
	Note   string   `json:"note"`
	Tags   []string `json:"tags"`
}

type Err struct {
	Message string `json:"message"`
}

func getExpenseHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "OK")

}

func updateExpenseHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "OK")

}

func getExpensesHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "OK")

}

func createExpensesHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "OK")

}

var db *sql.DB

func main() {
	fmt.Println("Please use server.go for main file")
	fmt.Println("start at port:", os.Getenv("PORT"))

	url := os.Getenv("DATABASE_URL")
	port := os.Getenv("PORT")

	var err error
	db, err = sql.Open("postgres", url)
	if err != nil {
		log.Fatal("Connect to database error", err)
	}
	defer db.Close()

	createTb := `
	CREATE TABLE IF NOT EXISTS expenses (
		id SERIAL PRIMARY KEY,
		title TEXT,
		amount FLOAT,
		note TEXT,
		tags TEXT[]
	);
	`
	_, err = db.Exec(createTb)
	if err != nil {
		log.Fatal("can't create table", err)
	}
	fmt.Println("create table success")

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.GET("/expenses", getExpensesHandler)
	e.GET("/expenses/:id", getExpenseHandler)
	e.PUT("/expenses/:id", updateExpenseHandler)
	e.POST("/expenses", createExpensesHandler)

	go func() {
		if err := e.Start(port); err != nil && err != http.ErrServerClosed { // Start server
			e.Logger.Fatal("shutting down the server")
		}
	}()

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	<-shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Println("bye bye!")
}
