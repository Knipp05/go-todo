package main

import (
	"database/sql"
	"fmt"

	_ "modernc.org/sqlite"
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

func initTables(db *sql.DB) {
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

func addUser(db *sql.DB, name, password string) {
	query := `INSERT INTO users (name, password) VALUES (?,?)`
	_, err := db.Exec(query, name, password)
	if !checkError(err) {
		newUser := NewUser(name, password, make([]task, 0))
		users[name] = newUser
	}
}
func deleteUser(db *sql.DB, name string, users map[string]*user) {
	taskQuery := `DELETE FROM tasks WHERE user_name = ?`
	_, err := db.Exec(taskQuery, name)
	checkError(err)
	userQuery := `DELETE FROM users WHERE name = ?`
	_, err = db.Exec(userQuery, name)
	checkError(err)
	delete(users, name)
}

func (u *user) addTask(db *sql.DB, title string, desc string, category string) {
	if category == "" {
		category = "default"
	}
	query := `INSERT INTO tasks (title, desc, isDone, category, user_name) VALUES (?,?,?,?,?)`
	_, err := db.Exec(query, title, desc, false, category, u.Name)
	if !checkError(err) {
		u.loadTasks(db)
	}
}
func (u *user) deleteTask(db *sql.DB, id int) {
	u.Tasks = append(u.Tasks[:id], u.Tasks[id+1:]...)
	query := `DELETE FROM tasks WHERE id = ?`
	_, err := db.Exec(query, id)
	if !checkError(err) {
		u.loadTasks(db)
	}
}
func (u *user) changeTitle(db *sql.DB, id int, title string) {
	query := `UPDATE tasks SET title = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, title, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Title = title
	}
}
func (u *user) changeDesc(db *sql.DB, id int, desc string) {
	query := `UPDATE tasks SET desc = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, desc, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Desc = desc
	}
}
func (u *user) changeIsDone(db *sql.DB, id int, isDone bool) {
	query := `UPDATE tasks SET isDone = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, isDone, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.IsDone = isDone
	}
}
func (u *user) changeCategory(db *sql.DB, id int, category string) {
	query := `UPDATE tasks SET category = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, category, id, u.Name)
	if !checkError(err) {
		task := findTaskById(&u.Tasks, id)
		task.Category = category
	}
}

func loadUsers(db *sql.DB) {
	userQuery := `SELECT name, password FROM users;`
	usersRows, err := db.Query(userQuery)
	checkError(err)
	for usersRows.Next() {
		var name, password string
		var tasks []task
		err := usersRows.Scan(&name, &password)
		checkError(err)
		users[name] = NewUser(name, password, tasks)
		users[name].loadTasks(db)
	}
	defer usersRows.Close()
}
func (u *user) loadTasks(db *sql.DB) {
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

/*
	 func RegisterUser(c *fiber.Ctx) error {
		name := c.Params("name")
		password := c.Params("password")
		newUser = NewUser(name, password)
	}
*/

var users map[string]*user

func main() {
	users = make(map[string]*user)
	dbPath := "./go-todo.db"
	db, err := sql.Open("sqlite", dbPath)
	checkError(err)
	initTables(db)
	loadUsers(db)
	addUser(db, "Niklas", "12332")
	users["Niklas"].addTask(db, "Hallo", "i bims", "")
	users["Niklas"].changeIsDone(db, 1, true)
	fmt.Println(*users["Niklas"])
	defer db.Close()

}
