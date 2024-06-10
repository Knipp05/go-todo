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
	ID       int      `json:"id"`
	Title    string   `json:"title"`
	Desc     string   `json:"desc"`
	IsDone   bool     `json:"isDone"`
	Category category `json:"category"`
}

type category struct {
	Cat_name     string `json:"cat_name"`
	Color_header string `json:"color_header"`
	Color_body   string `json:"color_body"`
}

/* type user struct {
	Name     string `json:"name"`
	Password string `json:"password"`
	Tasks    []task `json:"tasks"`
} */

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

/*
	 func NewUser(name, password string, tasks []task) *user {
		newUser := user{Name: name, Password: password, Tasks: tasks}
		return &newUser
	}
*/
func NewTask(id int, title string, desc string, isDone bool, category category) *task {
	newTask := task{ID: id, Title: title, Desc: desc, IsDone: isDone, Category: category}
	return &newTask
}
func NewCategory(cat_name, color_header, color_body string) *category {
	newTask := category{Cat_name: cat_name, Color_header: color_header, Color_body: color_body}
	return &newTask
}

func initTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		name TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);`

	categoriesTable := `CREATE TABLE IF NOT EXISTS categories (
		cat_name TEXT NOT NULL,
		color_header TEXT,
		color_body TEXT,
		user_name TEXT,
		PRIMARY KEY (cat_name, user_name),
		FOREIGN KEY (user_name) REFERENCES users(name)
	);`

	tasksTable := `CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        desc TEXT,
		isDone BOOL,
		category TEXT,
        user_name TEXT,
		FOREIGN KEY (category) REFERENCES categories(cat_name),
        FOREIGN KEY (user_name) REFERENCES users(name)
    );`

	_, err := db.Exec(usersTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der User Tabelle")
	}
	_, err = db.Exec(categoriesTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der Kategorie Tabelle")
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
	query = `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	_, err = db.Exec(query, "default", "#00a4ba", "#00ceea", name)
	if err != nil {
		return errors.New("Kategorie default konnte nicht angelegt werden")
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

func loginUser(inputName, inputPassword string) (token string, name string, tasks []task, categories []category, err error) {
	query := `SELECT name, password FROM users WHERE name=?`
	user, err := db.Query(query, inputName)
	if err != nil {
		return "", "", nil, nil, err
	}
	defer user.Close()

	for user.Next() {
		var name, password string
		err = user.Scan(&name, &password)
		if err == nil && password == inputPassword {
			token, err := generateJWT(name)
			if err != nil {
				return "", "", nil, nil, err
			}
			tasks := loadTasks(name)
			if tasks == nil {
				return "", "", nil, nil, errors.New("Fehler beim Laden der Tasks")
			}
			categories := loadCategories(name)
			if categories == nil {
				return "", "", nil, nil, errors.New("Fehler beim Laden der Kategorien")
			}
			return token, name, tasks, categories, nil
		}
	}

	return "", "", nil, nil, errors.New("Die Anmeldedaten sind nicht korrekt")
}

func addTask(name string, title string, desc string, category category) *task {
	query := `INSERT INTO tasks (title, desc, isDone, category, user_name) VALUES (?,?,?,?,?)`
	newTask, err := db.Exec(query, title, desc, false, category.Cat_name, name)
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
func changeCategory(name string, id int, cat_name string) error {
	query := `UPDATE tasks SET category = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, cat_name, id, name)
	if err != nil {
		return err
	}
	return nil
}

func loadTasks(name string) []task {
	query := `
		SELECT t.id, t.title, t.desc, t.isDone, t.category, c.color_header, c.color_body 
		FROM tasks t
		LEFT JOIN categories c ON t.category = c.cat_name AND t.user_name = c.user_name
		WHERE t.user_name = ?`

	rows, err := db.Query(query, name)
	if err != nil {
		return nil
	}
	defer rows.Close()

	loadedTasks := make([]task, 0)
	for rows.Next() {
		var id int
		var title, desc, category, color_header, color_body string
		var isDone bool

		err := rows.Scan(&id, &title, &desc, &isDone, &category, &color_header, &color_body)
		if err != nil {
			return nil
		}

		loadedTasks = append(loadedTasks, *NewTask(id, title, desc, isDone, *NewCategory(category, color_header, color_body)))
	}

	if err := rows.Err(); err != nil {
		return nil
	}

	return loadedTasks
}

func addCategory(cat_name, color_header, color_body, user_name string) *category {
	query := `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	_, err := db.Exec(query, cat_name, color_header, color_body, user_name)
	if err != nil {
		return nil
	}
	addedCategory := NewCategory(cat_name, color_header, color_body)
	return addedCategory
}
func deleteCategory(cat_name, user_name string) error {
	query := `DELETE FROM categories WHERE cat_name = ? AND user_name = ?`
	_, err := db.Exec(query, cat_name, user_name)
	if err != nil {
		return err
	}
	return nil
}
func loadCategories(name string) []category {
	query := `SELECT cat_name, color_header, color_body FROM categories WHERE user_name = ?`
	rows, err := db.Query(query, name)
	if err != nil {
		return nil
	}
	defer rows.Close()
	loadedCategories := make([]category, 0)
	for rows.Next() {
		var cat_name, color_header, color_body string
		err := rows.Scan(&cat_name, &color_header, &color_body)
		if err != nil {
			return nil
		}
		loadedCategories = append(loadedCategories, *NewCategory(cat_name, color_header, color_body))
	}
	return loadedCategories
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
		token, name, tasks, categories, err := loginUser(creds.Name, creds.Password)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(200).JSON(fiber.Map{"token": token, "name": name, "tasks": tasks, "categories": categories})
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
		Title    string   `json:"title"`
		Desc     string   `json:"desc"`
		Category category `json:"category"`
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
		Cat_name string `json:"category"`
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
		err = changeCategory(name, i, input.Cat_name)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Kategorie erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
	}
}

func AddCategory(c *fiber.Ctx) error {
	name := c.Locals("name").(string)

	var input category
	if err := c.BodyParser(&input); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if input.Cat_name != "" {
		addedCategory := addCategory(input.Cat_name, input.Color_header, input.Color_body, name)
		if addedCategory == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht erstellt werden"})
		}
		return c.Status(201).JSON(fiber.Map{"msg": "Kategorie erfolgreich angelegt"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategoriename darf nicht leer sein"})
	}
}
func DeleteCategory(c *fiber.Ctx) error {
	cat_name := c.Params("cat_name")
	name := c.Locals("name").(string)
	if cat_name != "" {
		deleteCategory(cat_name, name)
		return c.Status(200).JSON(fiber.Map{"msg": "Kategorie erfolgreich gelöscht"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
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
	app.Delete("/api/:name/:cat_name", DeleteCategory)
	app.Patch("/api/tasks/:id/isdone/:isDone", ChangeIsDone)
	app.Patch("/api/tasks/:id/content", ChangeContent)
	app.Post("/api/tasks", AddTask)
	app.Post("/api/:name/categories", AddCategory)
	app.Listen(":5000")
	defer db.Close()
}
