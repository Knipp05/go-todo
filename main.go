package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	_ "modernc.org/sqlite"

	"github.com/gofiber/contrib/websocket"
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
	Owner    string   `json:"owner"`
	Shared   []string `json:"shared"`
}

type category struct {
	ID           int    `json:"id"`
	Cat_name     string `json:"cat_name"`
	Color_header string `json:"color_header"`
	Color_body   string `json:"color_body"`
}

type Claims struct {
	Name string `json:"name"`
	jwt.RegisteredClaims
}

var jwtSecret = []byte("3F6C8DC3EEBB3987C95E87E15D629") // Key generieren und der Einfachheit halber hier fest codieren

func generateJWT(username string) (string, error) {
	claims := &Claims{
		Name: username,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 1)),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		return "", err
	}
	return tokenString, nil
}

func jwtMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		tokenString := c.Get("Authorization")
		if tokenString == "" {
			return c.Status(401).JSON(fiber.Map{"error": "Kein Token bereitgestellt"})
		}

		if !strings.HasPrefix(tokenString, "Bearer ") {
			return c.Status(401).JSON(fiber.Map{"error": "Ungültiges Token-Format"})
		}
		tokenString = strings.Replace(tokenString, "Bearer ", "", 1)

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			return c.Status(401).JSON(fiber.Map{"error": "Ungültiges Token", "details": err.Error()})
		}

		if claims, ok := token.Claims.(*Claims); ok && token.Valid {
			c.Locals("name", claims.Name)
			fmt.Printf("Benutzername aus Token: %s\n", claims.Name)
		} else {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Next()
	}
}

func NewTask(id int, title string, desc string, isDone bool, category category, owner string, shared []string) *task {
	newTask := task{ID: id, Title: title, Desc: desc, IsDone: isDone, Category: category, Owner: owner, Shared: shared}
	return &newTask
}
func NewCategory(id int, cat_name, color_header, color_body string) *category {
	newTask := category{ID: id, Cat_name: cat_name, Color_header: color_header, Color_body: color_body}
	return &newTask
}

func initTables() {
	usersTable := `CREATE TABLE IF NOT EXISTS users (
		name TEXT PRIMARY KEY,
		password TEXT NOT NULL
	);`

	categoriesTable := `CREATE TABLE IF NOT EXISTS categories (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		cat_name TEXT NOT NULL,
		color_header TEXT,
		color_body TEXT,
		user_name TEXT,
		FOREIGN KEY (user_name) REFERENCES users(name)
	);`

	tasksTable := `CREATE TABLE IF NOT EXISTS tasks (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        title TEXT NOT NULL,
        desc TEXT,
		isDone BOOL,
		category_id INTEGER,
        user_name TEXT,
		FOREIGN KEY (category_id) REFERENCES categories(id),
        FOREIGN KEY (user_name) REFERENCES users(name)
    );`

	sharedTable := `CREATE TABLE IF NOT EXISTS sharing (
	task_id INTEGER,
	target_name TEXT,
	PRIMARY KEY(target_name, task_id),
	FOREIGN KEY (target_name) REFERENCES users(name),
	FOREIGN KEY (task_id) REFERENCES tasks(id)
	)`

	_, err := db.Exec(usersTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der User Tabelle")
	}
	_, err = db.Exec(categoriesTable)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(tasksTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der Task Tabelle")
	}
	_, err = db.Exec(sharedTable)
	if err != nil {
		log.Fatal("Fehler beim Erstellen der Task Tabelle")
	}
}

func addUser(name, password string) error {
	query := `INSERT INTO users (name, password) VALUES (?,?)`
	_, err := db.Exec(query, name, password)
	if err != nil {
		return errors.New("benutzer existiert bereits")
	}
	query = `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	_, err = db.Exec(query, "default", "#00a4ba", "#00ceea", name)
	if err != nil {
		return errors.New("kategorie default konnte nicht angelegt werden")
	}
	return nil
}

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
				return "", "", nil, nil, errors.New("fehler beim Laden der Tasks")
			}
			categories := loadCategories(name)
			if categories == nil {
				return "", "", nil, nil, errors.New("fehler beim Laden der Kategorien")
			}
			return token, name, tasks, categories, nil
		}
	}

	return "", "", nil, nil, errors.New("die Anmeldedaten sind nicht korrekt")
}

func addTask(name string, title string, desc string, category category) *task {
	query := `INSERT INTO tasks (title, desc, isDone, category_id, user_name) VALUES (?,?,?,?,?)`
	newTask, err := db.Exec(query, title, desc, false, category.ID, name)
	if err != nil {
		return nil
	}
	addedTaskId, _ := newTask.LastInsertId()
	addedTask := NewTask(int(addedTaskId), title, desc, false, category, name, []string{})

	return addedTask
}
func deleteTask(name string, id int) error {
	existQuery := `SELECT EXISTS(SELECT 1 FROM sharing WHERE task_id = ?)`
	taskQuery := `DELETE FROM tasks WHERE id = ? AND user_name = ?`
	var exists bool

	tx, err := db.Begin()
	if err != nil {
		return err
	}

	err = tx.QueryRow(existQuery, id).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return err
	}

	if exists {
		targetQuery := `SELECT target_name FROM sharing WHERE task_id = ?`
		targetRows, err := tx.Query(targetQuery, id)
		if err != nil {
			fmt.Println(err)
			return err
		}
		defer targetRows.Close()

		var targetName string
		for targetRows.Next() {
			err = targetRows.Scan(&targetName)
			if err != nil {
				fmt.Println(err)
				continue
			}
			message, err := json.Marshal(id)
			if err != nil {
				fmt.Println(err)
				return err
			}
			mu.Lock()
			if conn, ok := clients[targetName]; ok {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					fmt.Println(err)
					return err
				}
			}
			mu.Unlock()
		}
	}
	_, err = tx.Exec(taskQuery, id, name)
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		return err
	}
	return nil
}
func changeTask(changedTask task, user string) error {
	var changeQuery string
	existQuery := `SELECT EXISTS(SELECT 1 FROM sharing WHERE task_id = ?)`
	var exists bool

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()
	if changedTask.Owner == user {
		changeQuery = `UPDATE tasks SET title = ?, desc = ?, isDone = ?, category_id = ? WHERE id = ? AND user_name = ?`
		_, err = tx.Exec(changeQuery, changedTask.Title, changedTask.Desc, changedTask.IsDone, changedTask.Category.ID, changedTask.ID, changedTask.Owner)
	} else {
		fmt.Println("Changing foreign task")
		changeQuery = `UPDATE tasks SET isDone = ? WHERE id = ?`
		_, err = tx.Exec(changeQuery, changedTask.IsDone, changedTask.ID)
	}
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tx.QueryRow(existQuery, changedTask.ID).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}

	if exists {
		err = updateSharedTask(changedTask, user)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

func changeCategory(name string, id int, cat_name, color_header, color_body string) error {
	query := `UPDATE categories SET cat_name = ?, color_header = ?, color_body = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, cat_name, color_header, color_body, id, name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

func loadTasks(name string) []task {
	query := `SELECT t.id, t.title, t.desc, t.isDone, t.user_name, c.id AS category_id, c.cat_name, c.color_header, c.color_body 
	FROM tasks t
	LEFT JOIN categories c ON t.category_id = c.id
	WHERE t.user_name = ?

	UNION

	SELECT t.id, t.title, t.desc, t.isDone, t.user_name, c.id AS category_id, c.cat_name, c.color_header, c.color_body 
	FROM tasks t
	LEFT JOIN categories c ON t.category_id = c.id
	INNER JOIN sharing s ON t.id = s.task_id
	WHERE s.target_name = ?`

	rows, err := db.Query(query, name, name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	loadedTasks := make([]task, 0)
	for rows.Next() {
		var shared *[]string
		var task_id, cat_id int
		var title, desc, cat_name, color_header, color_body, owner string
		var isDone bool

		err := rows.Scan(&task_id, &title, &desc, &isDone, &owner, &cat_id, &cat_name, &color_header, &color_body)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if name == owner {
			shared, err = loadSharedUsers(task_id)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			loadedTasks = append(loadedTasks, *NewTask(task_id, title, desc, isDone, *NewCategory(cat_id, cat_name, color_header, color_body), owner, *shared))
		} else {
			loadedTasks = append(loadedTasks, *NewTask(task_id, title, desc, isDone, *NewCategory(cat_id, cat_name, color_header, color_body), owner, []string{}))
		}

	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return nil
	}

	return loadedTasks
}

func loadSharedUsers(taskID int) (*[]string, error) {
	sharedQuery := `SELECT target_name FROM sharing WHERE task_id = ?`
	shared := make([]string, 0)

	sharedRows, err := db.Query(sharedQuery, taskID)
	if err != nil {
		fmt.Println(err)
		return nil, err
	}
	defer sharedRows.Close()

	var targetName string

	for sharedRows.Next() {
		err = sharedRows.Scan(&targetName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		shared = append(shared, targetName)
	}
	return &shared, nil
}

func shareTask(sharedTask task, target string) error {
	existQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE name = ?)`
	shareQuery := `INSERT INTO sharing (task_id, target_name) VALUES (?,?)`

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return err
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	var exists bool
	err = tx.QueryRow(existQuery, target).Scan(&exists)
	if err != nil {
		fmt.Println(err)
		return err
	}

	if !exists {
		return errors.New("benutzer konnte nicht gefunden werden")
	}

	_, err = tx.Exec(shareQuery, sharedTask.ID, target)
	if err != nil {
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		fmt.Println(err)
		return err
	}
	message, err := json.Marshal(sharedTask)
	if err != nil {
		fmt.Println(err)
		return err
	}
	mu.Lock()
	if conn, ok := clients[target]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println(err)
			return err
		}
	}
	mu.Unlock()
	return nil
}

func removeSharing(id int, target string) error {
	removeQuery := `DELETE FROM sharing WHERE task_id = ? AND target_name = ?`
	_, err := db.Exec(removeQuery, id, target)
	if err != nil {
		fmt.Println(err)
		return err
	}
	message, err := json.Marshal(id)
	if err != nil {
		fmt.Println(err)
		return err
	}
	mu.Lock()
	if conn, ok := clients[target]; ok {
		if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
			fmt.Println(err)
			return err
		}
	}
	mu.Unlock()
	return nil
}

func updateSharedTask(task task, user string) error {
	targetQuery := `SELECT target_name FROM sharing WHERE task_id = ?`

	targetRows, err := db.Query(targetQuery, task.ID)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer targetRows.Close()

	var targetName string

	message, err := json.Marshal(task)
	if err != nil {
		fmt.Println(err)
		return err
	}
	for targetRows.Next() {
		err = targetRows.Scan(&targetName)
		if err != nil {
			fmt.Println(err)
			continue
		}
		mu.Lock()
		if conn, ok := clients[targetName]; ok {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println(err)
				return err
			}
		}
		mu.Unlock()
	}
	if task.Owner != user {
		mu.Lock()
		if conn, ok := clients[task.Owner]; ok {
			if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
				fmt.Println(err)
				return err
			}
		}
		mu.Unlock()
	}

	return nil
}

func addCategory(cat_name, color_header, color_body, user_name string) *category {
	query := `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	newCategory, err := db.Exec(query, cat_name, color_header, color_body, user_name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	addedCategoryId, _ := newCategory.LastInsertId()
	addedCategory := NewCategory(int(addedCategoryId), cat_name, color_header, color_body)
	return addedCategory
}
func deleteCategory(user_name string, id int) ([]task, error) {
	taskQuery := `UPDATE tasks SET category_id = 1 WHERE category_id = ? AND user_name = ?`
	categoryQuery := `DELETE FROM categories WHERE id = ? AND user_name = ?`

	tx, err := db.Begin()
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec(taskQuery, id, user_name)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	_, err = tx.Exec(categoryQuery, id, user_name)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return nil, err
	}

	return loadTasks(user_name), nil
}
func loadCategories(name string) []category {
	query := `SELECT id, cat_name, color_header, color_body FROM categories WHERE user_name = ?`
	rows, err := db.Query(query, name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()
	loadedCategories := make([]category, 0)
	for rows.Next() {
		var id int
		var cat_name, color_header, color_body string
		err := rows.Scan(&id, &cat_name, &color_header, &color_body)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		loadedCategories = append(loadedCategories, *NewCategory(id, cat_name, color_header, color_body))
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
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if strings.TrimSpace(creds.Name) != "" && strings.TrimSpace(creds.Password) != "" {
		err := addUser(creds.Name, creds.Password)
		if err != nil {
			fmt.Println(err)
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
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if strings.TrimSpace(creds.Name) != "" && strings.TrimSpace(creds.Password) != "" {
		token, name, tasks, categories, err := loginUser(creds.Name, creds.Password)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(200).JSON(fiber.Map{"token": token, "name": name, "tasks": tasks, "categories": categories})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein"})
	}
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
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if strings.TrimSpace(input.Title) != "" {
		addedTask := addTask(name, input.Title, input.Desc, input.Category)
		if addedTask == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Aufgabe konnte nicht erstellt werden"})
		}
		return c.Status(201).JSON(fiber.Map{"id": addedTask.ID, "category": addedTask.Category})
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
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
		}
		deleteTask(name, i)
		return c.Status(200).JSON(fiber.Map{"msg": "Aufgabe erfolgreich gelöscht"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
	}
}
func ChangeTask(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	type TaskInput struct {
		Title    string   `json:"title"`
		Desc     string   `json:"desc"`
		IsDone   bool     `json:"isDone"`
		Category category `json:"category"`
		Owner    string   `json:"owner"`
	}
	var input TaskInput
	if err := c.BodyParser(&input); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if id != "" && strings.TrimSpace(input.Title) != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		err = changeTask(*NewTask(i, input.Title, input.Desc, input.IsDone, input.Category, input.Owner, []string{}), name)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Aufgabe konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Aufgabe erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel darf nicht leer sein"})
	}
}

func ShareTask(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	target := c.Params("target")

	type TaskInput struct {
		Title    string   `json:"title"`
		Desc     string   `json:"desc"`
		IsDone   bool     `json:"isDone"`
		Category category `json:"category"`
		Owner    string   `json:"owner"`
	}
	var input TaskInput

	if name != target && id != "" {
		if err := c.BodyParser(&input); err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Parsen"})
		}
		i, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Konvertieren von ID"})
		}
		err = shareTask(*NewTask(i, input.Title, input.Desc, input.IsDone, input.Category, name, []string{}), target)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Benutzer konnte nicht gefunden werden"})
		}
		return c.Status(201).JSON(fiber.Map{"msg": "Task erfolgreich freigegeben"})
	}
	return c.Status(400).JSON(fiber.Map{"error": "Besitzer und Zielperson dürfen nicht identisch sein"})
}

func RemoveSharing(c *fiber.Ctx) error {
	id := c.Params("id")
	target := c.Params("target")

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	err = removeSharing(i, target)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Freigabe konnte nicht beendet werden"})
	}
	return c.Status(200).JSON(fiber.Map{"msg": "Freigabe erfolgreich beendet"})
}

func ChangeCategory(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	type CategoryInput struct {
		Cat_name     string `json:"cat_name"`
		Color_header string `json:"color_header"`
		Color_body   string `json:"color_body"`
	}
	var input CategoryInput
	if err := c.BodyParser(&input); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if strings.TrimSpace(input.Cat_name) != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
		}
		err = changeCategory(name, i, input.Cat_name, input.Color_header, input.Color_body)
		if err != nil {
			fmt.Println(err)
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
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if strings.TrimSpace(input.Cat_name) != "" {
		addedCategory := addCategory(input.Cat_name, input.Color_header, input.Color_body, name)
		if addedCategory == nil {
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie existiert bereits"})
		}
		return c.Status(201).JSON(fiber.Map{"id": addedCategory.ID})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategoriename darf nicht leer sein"})
	}
}
func DeleteCategory(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id := c.Params("id")
	if id != "" {
		i, err := strconv.Atoi(id)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
		}
		updatedTasks, err := deleteCategory(name, i)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
		}
		return c.Status(200).JSON(fiber.Map{"tasks": updatedTasks})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Fehler beim Löschen aufgetreten"})
	}
}

var db *sql.DB
var clients = make(map[string]*websocket.Conn)
var mu sync.Mutex

func main() {
	var err error

	db, err = sql.Open("sqlite", "go-todo.db")
	if err != nil {
		log.Fatal("Fehler beim Erstellen/Öffnen der Datenbank")
	}
	defer db.Close()

	initTables()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, http://192.168.178.69:5173",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
	}))

	app.Use("/ws", func(c *fiber.Ctx) error {
		if websocket.IsWebSocketUpgrade(c) {
			c.Locals("allowed", true)
			return c.Next()
		}
		return fiber.ErrUpgradeRequired
	})

	app.Get("/ws", websocket.New(func(c *websocket.Conn) {
		tokenString := c.Query("token")
		if tokenString == "" {
			c.Close()
			return
		}

		claims := &Claims{}

		token, err := jwt.ParseWithClaims(tokenString, claims, func(t *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			log.Println("Invalid token:", err)
			c.Close()
			return
		}

		name := claims.Name
		mu.Lock()
		clients[name] = c
		mu.Unlock()

		defer func() {
			mu.Lock()
			delete(clients, name)
			mu.Unlock()
			c.Close()
		}()

		log.Println("User:", claims.Name)

		var (
			mt  int
			msg []byte
		)
		for {
			if mt, msg, err = c.ReadMessage(); err != nil {
				log.Println("read:", err)
				break
			}
			log.Printf("recv: %s", msg)

			if err = c.WriteMessage(mt, msg); err != nil {
				log.Println("write", err)
				break
			}
		}
	}))

	app.Post("/api/users/new", RegisterUser)
	app.Post("/api/users", LogInUser)

	app.Use(jwtMiddleware())

	// Task Routen
	app.Post("/api/:name/tasks", AddTask)
	app.Delete("/api/:name/tasks/:id", DeleteTask)
	app.Patch("/api/:name/tasks/:id", ChangeTask)
	app.Post("/api/:name/tasks/:id/:target", ShareTask)
	app.Delete("/api/:name/tasks/:id/:target", RemoveSharing)

	// Category Routen
	app.Post("/api/:name/categories", AddCategory)
	app.Patch("/api/:name/categories/:id/delete", DeleteCategory)
	app.Patch("/api/:name/categories/:id", ChangeCategory)

	app.Listen(":5000")
}
