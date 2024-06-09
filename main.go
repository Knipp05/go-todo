package main

import (
	"database/sql"
	"errors"
	"log"
	"strconv"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/golang-jwt/jwt/v5"
)

type task struct {
	ID       int    `json:"id"`
	Title    string `json:"title"`
	Desc     string `json:"desc"`
	IsDone   bool   `json:"isDone"`
	Category string `json:"category"`
}

type user struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Tasks    []task `json:"tasks"`
}

var jwtSecret = []byte("3F6C8DC3EEBB3987C95E87E15D629") // Key generieren und der Einfachheit halber hier fest codieren

func generateJWT(username string) (string, error) {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"name": username,
		"exp":  time.Now().Add(time.Hour * 1).Unix(),
	})

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func parseJWT(tokenString string) (*jwt.Token, error) {
	token, err := jwt.Parse(tokenString, func(t *jwt.Token) (interface{}, error) {
		return jwtSecret, nil
	})
	if err != nil {
		return nil, err
	}
	return token, nil
}

func jwtMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Kein Token bereitgestellt"})
		}

		token, err := parseJWT(tokenString)
		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Ungültiges Token"})
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("name", claims["name"])
		} else {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Next()
	}
}

func NewUser(name, password string, tasks []task) *user {
	newUser := user{Name: name, Password: password, Tasks: tasks}
	return &newUser
}
func NewTask(id int, title string, desc string, isDone bool, category string) *task {
	newTask := task{ID: id, Title: title, Desc: desc, IsDone: isDone, Category: category}
	return &newTask
}

func initTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		name TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);`

	tasksTable := `CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        desc TEXT,
		isDone BOOL,
		category TEXT,
        user_name TEXT,
        FOREIGN KEY (user_name) REFERENCES users(name)
    );`

	_, err := db.Exec(usersTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der User Tabelle")
	}
	_, err = db.Exec(tasksTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der Task Tabelle")
	}
}

func addUser(name, password string) error {
	query := `INSERT INTO users (name, password) VALUES (?,?)`
	_, err := db.Exec(query, name, password)
	if err != nil {
		return errors.New("Benutzer existiert bereits")
	}
	return nil
}

/* func deleteUser(name string) {
	taskQuery := `DELETE FROM tasks WHERE user_name = ?`
	_, err := db.Exec(taskQuery, name)
	checkError(err)
	userQuery := `DELETE FROM users WHERE name = ?`
	_, err = db.Exec(userQuery, name)
	checkError(err)
	delete(users, name)
} */

func loginUser(inputName, inputPassword string) (token string, name string, tasks []task, err error) {
	query := `SELECT name, password FROM users WHERE name=?`
	user, err := db.Query(query, inputName)
	if err != nil {
		return "", "", nil, err
	}
	defer user.Close()

	for user.Next() {
		var name, password string
		err = user.Scan(&name, &password)
		if err == nil && password == inputPassword {
			token, err := generateJWT(name)
			if err != nil {
				return "", "", nil, err
			}
			tasks, err := loadTasks(name)
			if err != nil {
				return "", "", nil, err
			}
			return token, name, tasks, nil
		}
	}

	return "", "", nil, errors.New("Die Anmeldedaten sind nicht korrekt")
}

func addTask(name string, title string, desc string, category string) *task {
	if category == "" {
		category = "default"
	}
	query := `INSERT INTO tasks (title, desc, isDone, category, user_name) VALUES (?,?,?,?,?)`
	newTask, err := db.Exec(query, title, desc, false, category, name)
	if err != nil {
		return nil
	}
	addedTaskId, _ := newTask.LastInsertId()
	addedTask := NewTask(int(addedTaskId), title, desc, false, category)

	return addedTask
}
func deleteTask(name string, id int) error {
	query := `DELETE FROM tasks WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, id, name)
	if err != nil {
		return err
	}
	return nil
}
func changeContent(name string, id int, title string, desc string) error {
	query := `UPDATE tasks SET title = ?, desc = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, title, desc, id, name)
	if err != nil {
		return err
	}
	return nil
}
func changeIsDone(name string, id int, isDone bool) error {
	query := `UPDATE tasks SET isDone = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, isDone, id, name)
	if err != nil {
		return err
	}
	return nil
}
func changeCategory(name string, id int, category string) error {
	query := `UPDATE tasks SET category = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, category, id, name)
	if err != nil {
		return err
	}
	return nil
}

func loadTasks(name string) ([]task, error) {
	query := `SELECT id, title, desc, isDone, category FROM tasks WHERE user_name = ?`
	rows, err := db.Query(query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	loadedTasks := make([]task, 0)
	for rows.Next() {
		var id int
		var title, desc, category string
		var isDone bool
		err := rows.Scan(&id, &title, &desc, &isDone, &category)
		if err != nil {
			return nil, err
		}
		loadedTasks = append(loadedTasks, *NewTask(id, title, desc, isDone, category))
	}
	return loadedTasks, nil
}

func RegisterUser(c *fiber.Ctx) error {
	type Credentials struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	var creds Credentials
	if err := c.BodyParser(&creds); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if creds.Name != "" && creds.Password != "" {
		err := addUser(creds.Name, creds.Password)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Dieser Benutzer existiert bereits"})
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein"})
	}
	return c.Status(201).JSON("Benutzer erfolgreich hinzugefügt")
}

func LogInUser(c *fiber.Ctx) error {
	var err error
	type Credentials struct {
		Name     string `json:"name"`
		Password string `json:"password"`
	}

	var creds Credentials
	if err = c.BodyParser(&creds); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if creds.Name != "" && creds.Password != "" {
		token, name, tasks, err := loginUser(creds.Name, creds.Password)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(200).JSON(fiber.Map{"token": token, "name": name, "tasks": tasks})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein"})
	}
}

func LogOutUser(c *fiber.Ctx) error {
	return c.Status(200).JSON(fiber.Map{"msg": "Logout erfolgreich"})
}

func AddTask(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	type TaskInput struct {
		Title    string `json:"title"`
		Desc     string `json:"desc"`
		Category string `json:"category"`
	}

	var input TaskInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if input.Title != "" {
		addedTask := addTask(name, input.Title, input.Desc, input.Category)
		if addedTask == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Aufgabe konnte nicht erstellt werden"})
		}
		return c.Status(201).JSON(fiber.Map{"id": addedTask.ID, "title": addedTask.Title, "desc": addedTask.Desc, "isDone": addedTask.IsDone, "category": addedTask.Category})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel darf nicht leer sein"})
	}
}
func DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	name := c.Locals("name").(string)
	if id != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
		}
		deleteTask(name, i)
		return c.Status(200).JSON(fiber.Map{"msg": "Aufgabe erfolgreich gelöscht"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
	}
}
func ChangeContent(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	type TaskInput struct {
		Title string `json:"title"`
		Desc  string `json:"desc"`
	}
	var input TaskInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if id != "" && input.Title != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		err = changeContent(name, i, input.Title, input.Desc)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Titel konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Titel erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel darf nicht leer sein"})
	}
}

func ChangeIsDone(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	isDone := c.Params("isDone")
	if id != "" && isDone != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		d, err := strconv.ParseBool(isDone)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		err = changeIsDone(name, i, !d)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Status konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Status erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Status darf nicht leer sein"})
	}
}
func ChangeCategory(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	type TaskInput struct {
		ID       string `json:"id"`
		Category string `json:"category"`
	}
	var input TaskInput
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if input.ID != "" {
		i, err := strconv.Atoi(input.ID)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		err = changeCategory(name, i, input.Category)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Kategorie erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
	}
}

/* func logQuery(query string, args ...interface{}) {
	log.Printf("Executing query: %s with args: %v\n", query, args)
}

func logDBStats() {
	stats := db.Stats()
	log.Printf("Open connections: %d, In use: %d, Idle: %d\n", stats.OpenConnections, stats.InUse, stats.Idle)
} */

var db *sql.DB

func main() {
	var err error

	db, err = sql.Open("sqlite", "go-todo.db")
	if err != nil {
		log.Fatal("Fehler beim Erstellen/Öffnen der Datenbank")
	}

	initTables()

	/* go func() {
		for {
			logDBStats()
			time.Sleep(10 * time.Second)
		}
	}() */

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))
	app.Post("/api/users/new", RegisterUser)
	app.Post("/api/users", LogInUser)

	app.Use(jwtMiddleware())

	app.Delete("/api/users/logout", LogOutUser)
	app.Delete("/api/tasks/:id", DeleteTask)
	app.Patch("/api/tasks/:id/isdone/:isDone", ChangeIsDone)
	app.Patch("/api/tasks/:id/content", ChangeContent)
	app.Post("/api/tasks", AddTask)
	app.Listen(":5000")
	defer db.Close()
}
