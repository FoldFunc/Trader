package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5"
)

type resStruct struct {
	Code         int    `json:"code"`
	Message      string `json:"message,omitempty"`
	ErrorMessage string `json:"errormessage,omitempty"`
}

func res(spec resStruct, c *fiber.Ctx) error {
	if spec.Message == "" {
		return c.Status(spec.Code).JSON(fiber.Map{
			"error": spec.ErrorMessage,
		})
	}
	return c.Status(spec.Code).JSON(fiber.Map{
		"message": spec.Message,
	})
}

type registerReqStruct struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func registerAddToDB(req registerReqStruct, c *fiber.Ctx) error {
	rows, err := DB.Query(context.Background(),
		"SELECT FROM users where email = $1",
		req.Email)
	if err != nil {
		log.Fatal("Error while checking db: ", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	defer rows.Close()
	if rows != nil {
		log.Fatal("User already exsists", err)
		return res(resStruct{Code: 400, ErrorMessage: "User already exists"}, c)
	}
	_, err = DB.Exec(context.Background(),
		"INSERT INTO user (email, password) VALUES ($1, $2)",
		req.Email, req.Password)
	if err != nil {
		log.Fatal("Error while inserting to db: ", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return nil
}
func registerFunc(c *fiber.Ctx) error {
	req := new(registerReqStruct)
	err := c.BodyParser(req)
	if err != nil {
		return res(resStruct{Code: 400, ErrorMessage: "Invalid body of the request."}, c)
	}
	err = registerAddToDB(*req, c)
	if err != nil {
		return err
	}
	return res(resStruct{Code: 201, Message: "User created."}, c)
}

var DB *pgx.Conn

func init() {
	connstring := "postgres://fold:1234@localhost:5432/trader"
	var err error
	DB, err = pgx.Connect(context.Background(), connstring)
	if err != nil {
		log.Fatal("Unable to connect to the db: ", err)
	}
	defer DB.Close(context.Background())

	_, err = DB.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		email TEXT NOT NULL,
		password TEXT NOT NULL
	)
	`)
	if err != nil {
		log.Fatal("Failed to migrate the db: ", err)
	}
}
func main() {
	app := fiber.New()
	app.Static("/", "./static")
	app.Post("/api/register", registerFunc)
	log.Fatal(app.Listen(":42069"))
}
