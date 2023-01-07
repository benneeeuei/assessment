package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/lib/pq"
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
	id := c.Param("id")
	stmt, err := db.Prepare("SELECT id,title,amount,note,tags FROM expenses WHERE id = $1")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare query expenses statment:" + err.Error()})
	}

	rows := stmt.QueryRow(id)
	t := Expense{}

	err = rows.Scan(&t.Id, &t.Title, &t.Amount, &t.Note, pq.Array(&t.Tags))

	switch err {
	case sql.ErrNoRows:
		return c.JSON(http.StatusNotFound, Err{Message: "expenses not found"})
	case nil:
		return c.JSON(http.StatusOK, t)
	default:
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't scan expenses:" + err.Error()})
	}
}

func updateExpenseHandler(c echo.Context) error {
	id := c.Param("id")
	stmt, err := db.Prepare("UPDATE expenses SET title=$2,amount=$3,note= $4,tags=$5 WHERE id = $1  ; ")
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't prepare update expenses statment:" + err.Error()})
	}

	var t Expense
	err = c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	if _, err := stmt.Exec(id, t.Title, t.Amount, t.Note, pq.Array(t.Tags)); err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "error execute update:" + err.Error()})
	}

	fmt.Println("update success")

	t.Id, err = strconv.Atoi(id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: "can't convert id to int :" + err.Error()})
	}

	return c.JSON(http.StatusCreated, t)

}

func getExpensesHandler(c echo.Context) error {

	return c.JSON(http.StatusOK, "OK")

}

func createExpensesHandler(c echo.Context) error {
	var t Expense
	err := c.Bind(&t)
	if err != nil {
		return c.JSON(http.StatusBadRequest, Err{Message: err.Error()})
	}

	row := db.QueryRow("INSERT INTO expenses (title, amount, note , tags ) values ($1, $2, $3 ,$4)  RETURNING id", t.Title, t.Amount, t.Note, pq.Array(t.Tags))
	err = row.Scan(&t.Id)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, Err{Message: err.Error()})
	}

	return c.JSON(http.StatusCreated, t)
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
