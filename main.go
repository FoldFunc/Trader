package main

import (
	"context"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/jackc/pgx/v5/pgxpool"
)

type resStruct struct {
	Code         int    `json:"code"`
	Message      any    `json:"message,omitempty"`
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
type fetchPortfoliosReqStruct struct {
	PortfolioNames []string `json:"portfolionames"`
}
type portfolios struct {
	Portfolios []portfolio `json:"portfolios"`
}
type portfolio struct {
	Id     int    `json:"id"`
	Owner  string `json:"owner"`
	Name   string `json:"name"`
	Money  int    `json:"money"`
	Stocks string `json:"stocks"` // this will be fun
}

func fetchPortfoliosDB(email string, c *fiber.Ctx) (portfolios, error) {
	var portfolioss portfolios
	rows, err := DB.Query(context.Background(),
		"SELECT * FROM portfolios WHERE owner=$1", email)
	if err != nil {
		log.Println("Error checking db: ", err)
		return *new(portfolios), err
	}
	defer rows.Close()
	for rows.Next() {
		var portfolioo portfolio
		if err := rows.Scan(&portfolioo.Id, &portfolioo.Owner, &portfolioo.Name, &portfolioo.Money, &portfolioo.Stocks); err != nil {
			log.Println("Error in scaning rows: ", err)
			return *new(portfolios), err
		}
		portfolioss.Portfolios = append(portfolioss.Portfolios, portfolioo)
	}
	if rows.Err() != nil {
		log.Println("Error in rows")
		return *new(portfolios), rows.Err()
	}
	return portfolioss, nil
}
func fetchPortfolioNamesDB(email string, c *fiber.Ctx) ([]string, error) {
	var names []string
	rows, err := DB.Query(context.Background(),
		"SELECT name FROM portfolios WHERE owner=$1", email)
	if err != nil {
		log.Println("Error checking db: ", err)
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		names = append(names, name)
	}
	if rows.Err() != nil {
		return nil, rows.Err()
	}
	return names, nil
}
func fetchNameDB(email string, c *fiber.Ctx) (string, error) {
	var name string
	err := DB.QueryRow(context.Background(),
		"SELECT name FROM users WHERE email=$1", email).Scan(&name)
	if err != nil {
		log.Println("Error checking db: ", err)
		return "", err
	}
	return name, nil
}
func loginCheckDB(req loginReqStruct, c *fiber.Ctx) error {
	var exsists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exsists)
	if err != nil {
		log.Println("Error checking db: ", err)
		return err
	}
	if !exsists {
		return err
	}
	var row loginReqStruct
	err = DB.QueryRow(context.Background(),
		"SELECT email, password FROM users WHERE email=$1", req.Email).
		Scan(&row.Email, &row.Password)
	if err != nil {
		log.Println("Error checking db: ", err)
		return err
	}
	if row.Password != req.Password {
		return err
	}
	_, err = DB.Exec(context.Background(),
		"UPDATE users SET logged=$2 WHERE email=$1",
		req.Email, true)
	if err != nil {
		log.Println("Error updating db:", err)
		return err
	}
	return nil
}
func logCheckDB(email string, c *fiber.Ctx) (bool, error) {
	var is bool
	err := DB.QueryRow(context.Background(),
		"SELECT logged FROM users WHERE email=$1", email).Scan(&is)
	if err != nil {
		log.Println("Error checking db:", err)
		return false, err
	}
	return is, nil
}
func addPortfolioToDB(email string, req createPortfolioReqStruct, c *fiber.Ctx) (string, error) {
	log.Println("email: ", email)
	log.Println("req.PortfolioName", req.PortfolioName)
	var exists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM portfolios WHERE owner=$1 AND name=$2)", email, req.PortfolioName).Scan(&exists)
	if err != nil {
		log.Println("Error checking db:", err)
		return "error", err
	}
	log.Println("exsists: ", exists)
	if exists {
		return "error", err
	}
	_, err = DB.Exec(context.Background(),
		"INSERT INTO portfolios (owner, name, money, stocks) VALUES ($1, $2, $3, $4)",
		email, req.PortfolioName, 100, "")
	if err != nil {
		log.Println("Error inserting to db:", err)
		return "error", err
	}
	return "", nil
}
func registerAddToDB(req registerReqStruct, c *fiber.Ctx) error {
	var exists bool
	err := DB.QueryRow(context.Background(),
		"SELECT EXISTS(SELECT 1 FROM users WHERE email=$1)", req.Email).Scan(&exists)
	if err != nil {
		log.Println("Error checking db:", err)
		return err
	}
	if exists {
		return err
	}
	_, err = DB.Exec(context.Background(),
		"INSERT INTO users (email, password, name, logged) VALUES ($1, $2, $3, $4)",
		req.Email, req.Password, req.Name, false)
	if err != nil {
		log.Println("Error inserting to db:", err)
		return err
	}
	return nil
}

func createPortfolioFunc(c *fiber.Ctx) error {
	log.Println("createPortfolioFunc called")
	cookie := c.Cookies("session")
	if cookie == "" {
		log.Println("No cookie")
		return res(resStruct{Code: 400, Message: "No cookie found"}, c)
	}
	req := new(createPortfolioReqStruct)
	if err := c.BodyParser(req); err != nil {
		log.Println("Body parser error: ", err)
		return res(resStruct{Code: 400, ErrorMessage: "Invalid body of the request."}, c)
	}
	if string, err := addPortfolioToDB(cookie, *req, c); string != "" {
		log.Println("addPortfolioToDB error: ", err)
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
	}
	return res(resStruct{Code: 201, Message: "Portfolio craeted."}, c)
}
func fetchPortfolios(c *fiber.Ctx) error {
	log.Println("fetchPortfolios called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 400, ErrorMessage: "No cookie found"}, c)
	}
	infos, err := fetchPortfoliosDB(cookie, c)
	if err != nil {
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
	}
	return res(resStruct{Code: 200, Message: infos}, c)
}
func fetchPortfolioNames(c *fiber.Ctx) error {
	log.Println("fetchPortfolioNames called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 400, ErrorMessage: "No cookie found"}, c)
	}
	names, err := fetchPortfolioNamesDB(cookie, c)
	if err != nil {
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
	}
	return res(resStruct{Code: 200, Message: names}, c)
}
func fetchNameFunc(c *fiber.Ctx) error {
	log.Println("fetchNameFunc called")
	cookie := c.Cookies("session")
	if cookie == "" {
		return res(resStruct{Code: 400, ErrorMessage: "No cookie found"}, c)
	}
	name, err := fetchNameDB(cookie, c)
	if err != nil {
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
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
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
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
		return res(resStruct{Code: 500, ErrorMessage: "Interanal server error"}, c)
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
		return res(resStruct{Code: 500, ErrorMessage: "Internal server error"}, c)
	}
	return res(resStruct{Code: 201, Message: "User created."}, c)
}

var DB *pgxpool.Pool

func setupDB() {
	connstring := "postgres://fold:1234@localhost:5432/trader"
	var err error
	DB, err = pgxpool.New(context.Background(), connstring)
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
			stocks TEXT NOT NULL
		)`)
	if err != nil {
		log.Fatal("Failed to migrate the db:", err)
	}
}

func main() {
	setupDB()
	defer DB.Close()

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
	app.Get("/api/fetch/portfolioNames", fetchPortfolioNames)
	app.Get("/api/fetch/portfolios", fetchPortfolios)
	log.Fatal(app.Listen(":42069"))
}
