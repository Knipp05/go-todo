package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strconv"

	_ "modernc.org/sqlite"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
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
	checkError(err)
	_, err = db.Exec(tasksTable)
	checkError(err)
}

func addUser(name, password string) {
	query := `INSERT INTO users (name, password) VALUES (?,?)`
	_, err := db.Exec(query, name, password)
	if !checkError(err) {
		newUser := NewUser(name, password, make([]task, 0))
		activeUsers[name] = newUser
		fmt.Printf("User erstellt: ", name)
		loadUsers()
	}
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

func loginUser(inputName, inputPassword string) error {
	foundUser := NewUser(inputName, inputPassword, make([]task, 0))
	query := `SELECT name, password FROM users WHERE name=?`
	user, err := db.Query(query, inputName)
	if err == nil {
		_, alreadyLoggedIn := activeUsers[inputName]
		if !alreadyLoggedIn {
			for user.Next() {
				var name, password string
				err = user.Scan(&name, &password)
				if err == nil && password == inputPassword {
					foundUser.loadTasks()
					activeUsers[inputName] = foundUser
					fmt.Printf("User angemeldet: ", inputName)
					return nil
				}
			}

		}
		return errors.New("Dieser Benutzer ist bereits eingeloggt")

	}
	defer user.Close()
	return errors.New("Anmeldedaten sind nicht korrekt")
}

func (u *user) addTask(title string, desc string, category string) {
	if category == "" {
		category = "default"
	}
	query := `INSERT INTO tasks (title, desc, isDone, category, user_name) VALUES (?,?,?,?,?)`
	_, err := db.Exec(query, title, desc, false, category, u.Name)
	if !checkError(err) {
		u.loadTasks()
	}
}
func (u *user) deleteTask(id int) {
	u.Tasks = append(u.Tasks[:id], u.Tasks[id+1:]...)
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Exec(query, id)
	if !checkError(err) {
		u.loadTasks()
	}
}
func (u *user) changeTitle(id int, title string) {
	query := `UPDATE tasks SET title = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, title, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Title = title
	}
}
func (u *user) changeDesc(id int, desc string) {
	query := `UPDATE tasks SET desc = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, desc, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Desc = desc
	}
}
func (u *user) changeIsDone(id int, isDone bool) {
	query := `UPDATE tasks SET isDone = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, isDone, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.IsDone = isDone
	}
}
func (u *user) changeCategory(id int, category string) {
	query := `UPDATE tasks SET category = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, category, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Category = category
	}
}

func loadUsers() {
	test := make([]user, 0)
	userQuery := `SELECT name, password FROM users;`
	usersRows, err := db.Query(userQuery)
	checkError(err)
	for usersRows.Next() {
		var name, password string
		var tasks []task
		err := usersRows.Scan(&name, &password)
		checkError(err)
		test = append(test, *NewUser(name, password, tasks))
		fmt.Println(test)
	}
	defer usersRows.Close()
}
func (u *user) loadTasks() {
	query := `SELECT id, title, desc, isDone, category FROM tasks WHERE user_name = ?`
	rows, err := db.Query(query, u.Name)
	checkError(err)
	loadedTasks := make([]task, 0)
	for rows.Next() {
		var id int
		var title, desc, category string
		var isDone bool
		err := rows.Scan(&id, &title, &desc, &isDone, &category)
		checkError(err)
		loadedTasks = append(loadedTasks, *NewTask(id, title, desc, isDone, category))
	}
	u.Tasks = loadedTasks
	defer rows.Close()
}

func checkError(err error) bool {
	if err != nil {
		fmt.Println(err)
		return true
	}
	return false
}

func findTaskById(tasks *[]task, id int) *task {
	for index := range *tasks {
		if (*tasks)[index].ID == id {
			return &(*tasks)[index]
		}
	}
	return &task{}
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
		addUser(creds.Name, creds.Password)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein!"})
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
		err = loginUser(creds.Name, creds.Password)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein!"})
	}
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": err})
	} else {
		return c.Status(201).JSON(activeUsers[creds.Name])
	}

}
func LogOutUser(c *fiber.Ctx) error {
	var err error
	var name string
	if err = c.BodyParser(&name); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "Logout fehlgeschlagen"})
	}
	delete(activeUsers, name)
	fmt.Println("User abgemeldet: ", name)
	return c.Status(200).JSON(fiber.Map{"msg": "Logout erfolgreich"})
}

func AddTask(c *fiber.Ctx) error {
	title := c.Params("title")
	desc := c.Params("desc")
	category := c.Params("category")
	user := c.Params("user_name")
	if title != "" && user != "" {
		activeUsers[user].addTask(title, desc, category)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(201).JSON("Aufgabe erfolgreich erstellt")
}
func DeleteTask(c *fiber.Ctx) error {
	id := c.Params("id")
	user := c.Params("user_name")
	if id != "" && user != "" {
		i, err := strconv.Atoi(id)
		if !checkError(err) {
			activeUsers[user].deleteTask(i)
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "ID und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(200).JSON("Aufgabe erfolgreich gelöscht")
}
func ChangeTitle(c *fiber.Ctx) error {
	id := c.Params("id")
	title := c.Params("title")
	user := c.Params("user_name")
	if id != "" && title != "" && user != "" {
		i, err := strconv.Atoi(id)
		if !checkError(err) {
			activeUsers[user].changeTitle(i, title)
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "ID, Titel und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(200).JSON("Titel erfolgreich geändert")
}
func ChangeDesc(c *fiber.Ctx) error {
	id := c.Params("id")
	desc := c.Params("desc")
	user := c.Params("user_name")
	if id != "" && user != "" {
		i, err := strconv.Atoi(id)
		if !checkError(err) {
			activeUsers[user].changeDesc(i, desc)
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "ID und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(200).JSON("Beschreibung erfolgreich geändert")
}
func ChangeIsDone(c *fiber.Ctx) error {
	id := c.Params("id")
	isDone := c.Params("isDone")
	user := c.Params("user_name")
	if id != "" && user != "" {
		i, err := strconv.Atoi(id)
		if !checkError(err) {
			done, err := strconv.ParseBool(isDone)
			if !checkError(err) {
				activeUsers[user].changeIsDone(i, done)
			}
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "ID und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(200).JSON("Aufgabenstatus erfolgreich geändert")
}
func ChangeCategory(c *fiber.Ctx) error {
	id := c.Params("id")
	category := c.Params("category")
	user := c.Params("user_name")
	if id != "" && user != "" {
		i, err := strconv.Atoi(id)
		if !checkError(err) {
			activeUsers[user].changeCategory(i, category)
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "ID und Benutzer dürfen nicht leer sein!"})
	}
	return c.Status(200).JSON("Aufgabenstatus erfolgreich geändert")
}

func GetTasks(c *fiber.Ctx) error {
	name := c.Params("name")
	if name != "" {
		return c.JSON(activeUsers[name].Tasks)
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Name darf nicht leer sein!"})
	}
}

var activeUsers map[string]*user
var db *sql.DB

func main() {
	var err error
	activeUsers = make(map[string]*user)

	db, err = sql.Open("sqlite", "go-todo.db")
	checkError(err)
	initTables()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173",
		AllowHeaders: "Origin,Content-Type,Accept",
	}))
	app.Post("/api/users/new", RegisterUser)
	app.Post("/api/users", LogInUser)
	app.Post("/api/users/logout", LogOutUser)
	app.Listen(":5000")
	defer db.Close()
}
