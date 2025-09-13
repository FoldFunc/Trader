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
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}
type loginReqStruct struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}
type createPortfolioReqStruct struct {
	PortfolioName string `json:"portfolioname"`
}

func fetchNameDB(email string, c *fiber.Ctx) (string, error) {
	var name string
	err := DB.QueryRow(context.Background(),
		"SELECT name FROM users WHERE email=$1", email).Scan(&name)
	if err != nil {
		log.Println("Error checking db: ", err)
		return "", res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return name, nil
}
func loginCheckDB(req loginReqStruct, c *fiber.Ctx) error {
	var exsists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exsists)
	if err != nil {
		log.Println("Error checking db: ", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	if !exsists {
		return res(resStruct{Code: 404, ErrorMessage: "User Not found"}, c)
	}
	var row loginReqStruct
	err = DB.QueryRow(context.Background(),
		"SELECT email, password FROM users WHERE email=$1", req.Email).
		Scan(&row.Email, &row.Password)
	if err != nil {
		log.Println("Error checking db: ", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	if row.Password != req.Password {
		return res(resStruct{Code: 400, ErrorMessage: "Invalid credentials"}, c)
	}
	_, err = DB.Exec(context.Background(),
		"UPDATE users SET logged=$2 WHERE email=$1",
		req.Email, true)
	if err != nil {
		log.Println("Error updating db:", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return res(resStruct{Code: 200, ErrorMessage: "Logged in"}, c)
}
func logCheckDB(email string, c *fiber.Ctx) (bool, error) {
	var is bool
	err := DB.QueryRow(context.Background(),
		"SELECT logged FROM users WHERE email=$1", email).Scan(&is)
	if err != nil {
		log.Println("Error checking db:", err)
		return false, res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return is, nil
}
func addPortfolioToDB(email string, req createPortfolioReqStruct, c *fiber.Ctx) error {
	log.Println("email: ", email)
	log.Println("req.PortfolioName", req.PortfolioName)
	var exists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM portfolios WHERE owner=$1 AND name=$2)", email, req.PortfolioName).Scan(&exists)
	if err != nil {
		log.Println("Error checking db:", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	if exists {
		return res(resStruct{Code: 400, ErrorMessage: "Portfolio with that name alredy exists"}, c)
	}
	_, err = DB.Exec(context.Background(),
		"INSERT INTO portfolios (owner, name, money, stocks) VALUES ($1, $2, $3, $4)",
		email, req.PortfolioName, 100, nil)
	if err != nil {
		log.Println("Error inserting to db:", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return nil
}
func registerAddToDB(req registerReqStruct, c *fiber.Ctx) error {
	var exists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists)
	if err != nil {
		log.Println("Error checking db:", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	if exists {
		return res(resStruct{Code: 400, ErrorMessage: "User already exists"}, c)
	}

	_, err = DB.Exec(context.Background(),
		"INSERT INTO users (email, password, name, logged) VALUES ($1, $2, $3, $4)",
		req.Email, req.Password, req.Name, false)
	if err != nil {
		log.Println("Error inserting to db:", err)
		return res(resStruct{Code: 500, ErrorMessage: "DB error"}, c)
	}
	return nil
}

// TODO: Returns status 201 no matter what.
func createPortfolioFunc(c *fiber.Ctx) error {
	log.Println("createPortfolioFunc called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 400, Message: "No cookie found"}, c)
	}
	req := new(createPortfolioReqStruct)
	if err := c.BodyParser(req); err != nil {
		return res(resStruct{Code: 400, ErrorMessage: "Invalid body of the request."}, c)
	}
	if err := addPortfolioToDB(cookie, *req, c); err != nil {
		return err
	}
	return res(resStruct{Code: 201, Message: "Portfolio craeted."}, c)
}
func fetchNameFunc(c *fiber.Ctx) error {
	log.Println("fetchNameFunc called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 400, Message: "No cookie found"}, c)
	}
	name, err := fetchNameDB(cookie, c)
	if err != nil {
		return err
	}
	return res(resStruct{Code: 200, Message: name}, c)
}
func logCheckFunc(c *fiber.Ctx) error {
	log.Println("logCheckFunc called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 200, Message: "false"}, c)
	}
	is, err := logCheckDB(cookie, c)
	if err != nil {
		return err
	}
	if is {
		return res(resStruct{Code: 200, Message: "true"}, c)
	}
	return res(resStruct{Code: 200, Message: "false"}, c)
}
func loginFunc(c *fiber.Ctx) error {
	log.Println("Login handler called")
	req := new(loginReqStruct)
	if err := c.BodyParser(req); err != nil {
		return res(resStruct{Code: 400, ErrorMessage: "Invalid body of the request."}, c)
	}
	if err := loginCheckDB(*req, c); err != nil {
		return err
	}
	c.Cookie(&fiber.Cookie{
		Name:     "session",
		Value:    req.Email,
		HTTPOnly: true,
		Secure:   true,
		SameSite: "Strict",
		Path:     "/",
		MaxAge:   60 * 60 * 24,
	})
	return res(resStruct{Code: 200, Message: "Logged in."}, c)
}
func registerFunc(c *fiber.Ctx) error {
	log.Println("Register handler called")
	req := new(registerReqStruct)
	if err := c.BodyParser(req); err != nil {
		return res(resStruct{Code: 400, ErrorMessage: "Invalid body of the request."}, c)
	}
	if err := registerAddToDB(*req, c); err != nil {
		return err
	}
	return res(resStruct{Code: 201, Message: "User created."}, c)
}

var DB *pgx.Conn

func setupDB() {
	connstring := "postgres://fold:1234@localhost:5432/trader"
	var err error
	DB, err = pgx.Connect(context.Background(), connstring)
	if err != nil {
		log.Fatal("Unable to connect to the db:", err)
	}

	_, err = DB.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS users (
			id SERIAL PRIMARY KEY,
			email TEXT NOT NULL,
			password TEXT NOT NULL,
			name TEXT NOT NULL,
			logged BOOLEAN NOT NULL
		)`)
	if err != nil {
		log.Fatal("Failed to migrate the db:", err)
	}
	_, err = DB.Exec(context.Background(),
		`CREATE TABLE IF NOT EXISTS portfolios (
			id SERIAL PRIMARY KEY,
			owner TEXT NOT NULL,
			name TEXT NOT NULL,
			money INTEGER NOT NULL,
			stocks TEXT
		)`)
	if err != nil {
		log.Fatal("Failed to migrate the db:", err)
	}
}

func main() {
	setupDB()
	defer DB.Close(context.Background())

	app := fiber.New()
	app.Get("/sidebar.css", func(c *fiber.Ctx) error {
		return c.SendFile("./static/sidebar.css")
	})
	app.Static("/", "./static")
	app.Get("/register", func(c *fiber.Ctx) error {
		return c.SendFile("./static/register/register.html")
	})
	app.Get("/login", func(c *fiber.Ctx) error {
		return c.SendFile("./static/login/login.html")
	})
	app.Get("/profile", func(c *fiber.Ctx) error {
		return c.SendFile("./static/profile/profile.html")
	})
	app.Get("/createPortfolio", func(c *fiber.Ctx) error {
		return c.SendFile("./static/createPortfolio/createPortfolio.html")
	})
	app.Post("/api/register", registerFunc)
	app.Post("/api/login", loginFunc)
	app.Post("/api/createPortfolio", createPortfolioFunc)
	app.Get("/api/logcheck", logCheckFunc)
	app.Get("/api/fetch/name", fetchNameFunc)
	log.Fatal(app.Listen(":42069"))
}
