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
	Order    int      `json:"order"`
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

var jwtSecret = []byte("3F6C8DC3EEBB3987C95E87E15D629") // Key der Einfachheit halber hier statisch eingebettet

// generateJWT erstellt ein Token für den anfragenden Benutzer
//
// Parameter:
//   - user: Der Name des anfragenden Benutzers
//
// Rückgabewert:
//   - tokenString: Der erstellte und signierte String des token; "", falls ein Fehler auftritt
//   - error: Ein Fehler, falls die Erstellung des Token nicht funktioniert hat; "nil", falls kein Fehler auftritt
func generateJWT(name string) (string, error) {
	claims := &Claims{
		Name: name,
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

// jwtMiddleware prüft vor dem Aufruf jeder der geschützten http-Routen, ob der anfragende Benutzer ein gültiges Token besitzt
// falls nicht, wird der Zugriff diese Route verweigert
//
// Rückgabewert:
//   - c.Next: Eine Funktion, welche die nächste Methode auf dem Stack der aktuellen Route ausführt
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
		} else {
			return c.Status(401).JSON(fiber.Map{"error": "Unauthorized"})
		}
		return c.Next()
	}
}

// NewTask erstellt ein neues Objekt vom Typ task
//
// Parameter:
//   - taskID: Die ID der neuen Aufgabe
//   - title: Der Titel der Aufgabe
//   - desc: Die Beschreibung der Aufgabe
//   - isDone: Der Status der Aufgabe
//   - category: Die Kategorie der Aufgabe
//   - owner: Der Besitzer der Aufgabe
//   - shared: Die Benutzer, mit denen die Aufgabe geteilt wurde
//   - order: Die Nummer in der Reihenfolge dieser Aufgabe
//   - colorBody: Der Hex-Wert der Farbe des Body, die alle Aufgaben dieser Kategorie besitzen
//
// Rückgabewert:
//   - newTask: Ein Pointer auf die neu erstellte Aufgabe
func NewTask(taskID int, title string, desc string, isDone bool, category category, owner string, shared []string, order int) *task {
	newTask := task{ID: taskID, Title: title, Desc: desc, IsDone: isDone, Category: category, Owner: owner, Shared: shared, Order: order}
	return &newTask
}

// NewCategory erstellt ein neues Objekt vom Typ category
//
// Parameter:
//   - catID: Die ID der neuen Kategorie
//   - catName: Der Name der Kategorie
//   - colorHeader: Der Hex-Wert der Farbe des Header, die alle Aufgaben dieser Kategorie besitzen
//   - colorBody: Der Hex-Wert der Farbe des Body, die alle Aufgaben dieser Kategorie besitzen
//
// Rückgabewert:
//   - newCategory: Ein Pointer auf die neu erstellte Kategorie
func NewCategory(catID int, catName, colorHeader, colorBody string) *category {
	newCategory := category{ID: catID, Cat_name: catName, Color_header: colorHeader, Color_body: colorBody}
	return &newCategory
}

// initTables initialisiert die Tabellen der Datenbank, falls diese noch nicht existieren
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

	orderTable := `CREATE TABLE IF NOT EXISTS task_order (
	user_name TEXT,
	task_id INTEGER,
	order_id INTEGER,
	PRIMARY KEY(user_name, task_id),
	FOREIGN KEY (user_name) REFERENCES users(name),
	FOREIGN KEY (task_id) REFERENCES tasks(id)
	)`

	_, err := db.Exec(usersTable)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(categoriesTable)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(tasksTable)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(sharedTable)
	if err != nil {
		log.Fatal(err)
	}
	_, err = db.Exec(orderTable)
	if err != nil {
		log.Fatal(err)
	}
}

// addNewUser fügt eine neuen Benutzer mit angegebenem Benutznamen und Passwort in die Datenbank ein
// für jeden neuen Benutzer wird außerdem die Standardkategorie "default" angelegt
//
// Parameter:
//   - name: Der Name des neuen Benutzers (jeder Benutzername kann nur einmal vergeben werden)
//   - password: Das festgelegte Passwort für diesen Benutzer
//
// Rückgabewert:
//   - error: Ein Fehler, falls der Benutzer bereits existiert oder ein Fehler beim Anlegen der Standardkategorie auftritt
//     Git "nil" zurück, wenn die Operation erfolgreich ausgeführt wurde
func addNewUser(name, password string) error {
	query := `INSERT INTO users (name, password) VALUES (?,?)`
	_, err := db.Exec(query, name, password)
	if err != nil {
		fmt.Println(err)
		return errors.New("Benutzer existiert bereits")
	}
	query = `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	_, err = db.Exec(query, "default", "#00a4ba", "#00ceea", name)
	if err != nil {
		fmt.Println(err)
		return errors.New("Kategorie default konnte nicht angelegt werden")
	}
	return nil
}

// loginUser führt den Login für einen bestimmten Benutzer durch, indem Benutzername und Passwort geprüft werden
// ist der Login erfolgreich, wird ein neues token generiert, sowie die Aufgaben und Kategorien des Benutzers geladen und an den Client übermittelt
//
// Parameter:
//   - inputName: Der Name des Benutzers, welcher eingeloggt werden soll
//   - inputPassword: Das für den Login eingegebene Passwort
//
// Rückgabewert:
//   - token: Das für diesen Benutzer generierte token zur Authentifizierung und Authorisierung, falls der Login erfolgreich war; "", falls Login nicht erfolgreich
//   - tasks: Die für diesen Benutzer bereits vorhandenen Aufgaben, falls der Login erfolgreich war; "nil", falls Login nicht erfolgreich oder Fehler beim Laden
//   - categories: Die für diesen Benutzer bereits angelegten Kategorien, falls der Login erfolgreich war; "nil", falls nicht erfolgreich oder Fehler beim Laden
//   - error: Ein Fehler, falls der Login nicht erfolgreich war oder beim Laden der Aufgaben bzw. Kategorien ein Fehler aufgetreten ist; "nil", falls kein Fehler auftritt
func loginUser(inputName, inputPassword string) (token string, tasks []task, categories []category, err error) {
	query := `SELECT name, password FROM users WHERE name=?`
	user, err := db.Query(query, inputName)
	if err != nil {
		return "", nil, nil, err
	}
	defer user.Close()

	for user.Next() {
		var name, password string
		err = user.Scan(&name, &password)
		if err == nil && password == inputPassword {
			token, err := generateJWT(name)
			if err != nil {
				return "", nil, nil, err
			}
			tasks := getTasksForUser(name)
			if tasks == nil {
				return "", nil, nil, errors.New("Fehler beim Laden der Tasks")
			}
			categories := getCategoriesForUser(name)
			if categories == nil {
				return "", nil, nil, errors.New("Fehler beim Laden der Kategorien")
			}
			return token, tasks, categories, nil
		}
	}

	return "", nil, nil, errors.New("Die Anmeldedaten sind nicht korrekt")
}

// addTask führt eine Transaktion in der Datenbank aus, um eine neue Aufgabe mit Titel, Beschreibung und Kategorie hinzuzufügen
// Außerdem wird für die Aufgabe eine neue Nummer in der Tabelle task_order zur Speicherung der Reihenfolge der Aufgaben angelegt. Initial wird eine neue Aufgabe ganz zuletzt angezeigt
//
// Parameter:
//   - name: Der Name des Benutzers, welcher eine neue Aufgabe erstellen möchte
//   - title: Der Titel der neuen Aufgabe
//	 - desc: Die Beschreibung der neuen Aufgabe
//	 - category: Die Kategorie, welcher die neue Aufgabe zugeteilt wird
//
// Rückgabewert:
//   - addedTaskID: Gibt die von der Datenbank erstellte ID der neuen Aufgabe zurück
//	 Gibt 0 zurück, wenn bei der Erstellung ein Fehler aufgetreten ist

func addTask(name string, title string, desc string, category category, order int) int {
	taskQuery := `INSERT INTO tasks (title, desc, isDone, category_id, user_name) VALUES (?,?,?,?,?)`
	orderQuery := `INSERT INTO task_order (user_name, task_id, order_id) VALUES (?,?,?)`

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return 0
	}

	newTask, err := tx.Exec(taskQuery, title, desc, false, category.ID, name)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return 0
	}
	addedTaskID, _ := newTask.LastInsertId()

	_, err = tx.Exec(orderQuery, name, addedTaskID, order)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return 0
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return 0
	}

	return int(addedTaskID)
}

// deleteTask führt eine Transaktion in der Datenbank aus, um eine gewünschte Aufgabe zu löschen
// dazu wird außerdem geprüft, ob die Aufgabe mit anderen Benutzern geteilt wird und diese benachrichtigt werden müssen
// weiterhin muss ggf die Reihenfolge der Tasks angepasst werden, damit keine Lücken entstehen
//
// Parameter:
//   - name: Der Name des Benutzers, der eine Aufgabe löschen möchte
//   - taskID: Die Kategorie, welcher die neue Aufgabe zugeteilt wird
//
// Rückgabewert:
//   - error: Gibt einen Fehler zurück, wenn im Löschvorgang ein Fehler auftritt
//     Gibt "nil" zurück, wenn bei der Erstellung kein Fehler aufgetreten ist
func deleteTask(name string, taskID int) error {
	existQuery := `SELECT EXISTS(SELECT 1 FROM sharing WHERE task_id = ?)`
	sharingQuery := `DELETE FROM sharing WHERE task_id = ?`
	taskQuery := `DELETE FROM tasks WHERE id = ? AND user_name = ?`
	removeOrderQuery := `DELETE FROM task_order WHERE task_id = ? AND user_name = ?`
	getOrderQuery := `SELECT order_id FROM task_order WHERE task_id = ?`
	updateOrderQuery := `UPDATE task_order SET order_id = order_id - 1 WHERE user_name = ? AND order_id > ?`
	var exists bool
	var taskOrder int

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		return err
	}

	err = tx.QueryRow(getOrderQuery, taskID).Scan(&taskOrder)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.QueryRow(existQuery, taskID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		return err
	}

	if exists {
		targetQuery := `SELECT target_name FROM sharing WHERE task_id = ?`
		targetRows, err := tx.Query(targetQuery, taskID)
		if err != nil {
			tx.Rollback()
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
			message, err := json.Marshal(taskID)
			if err != nil {
				tx.Rollback()
				fmt.Println(err)
				return err
			}
			mu.Lock()
			if conn, ok := clients[targetName]; ok {
				if err := conn.WriteMessage(websocket.TextMessage, message); err != nil {
					tx.Rollback()
					fmt.Println(err)
					return err
				}
			}
			mu.Unlock()
		}
		_, err = tx.Exec(sharingQuery, taskID)
		if err != nil {
			tx.Rollback()
			return err
		}

	}
	_, err = tx.Exec(taskQuery, taskID, name)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(removeOrderQuery, taskID, name)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(updateOrderQuery, name, taskOrder)
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

// updateTask führt eine Transaktion in der Datenbank aus, um eine gewünschte Aufgabe zu aktualisieren
// dazu wird außerdem geprüft, ob die Aufgabe mit anderen Benutzern geteilt wird und diese benachrichtigt werden müssen
//
// Parameter:
//   - name: Der Benutzer, welcher eine Aufgabe ändert
//   - changedTask: Die Aufgabe, mit den aktualisierten Attributen
//
// Rückgabewert:
//   - error: Gibt einen Fehler zurück, wenn bei der Aktualisierung ein Fehler auftritt
//     Gibt "nil" zurück, wenn bei der Erstellung kein Fehler aufgetreten ist
func updateTask(name string, changedTask task) error {
	var changeQuery string
	existQuery := `SELECT EXISTS(SELECT 1 FROM sharing WHERE task_id = ?)`
	var exists bool

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	if changedTask.Owner == name {
		changeQuery = `UPDATE tasks SET title = ?, desc = ?, isDone = ?, category_id = ? WHERE id = ? AND user_name = ?`
		_, err = tx.Exec(changeQuery, changedTask.Title, changedTask.Desc, changedTask.IsDone, changedTask.Category.ID, changedTask.ID, changedTask.Owner)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return err
		}
	} else {
		changeQuery = `UPDATE tasks SET isDone = ? WHERE id = ?`
		_, err = tx.Exec(changeQuery, changedTask.IsDone, changedTask.ID)
		if err != nil {
			tx.Rollback()
			fmt.Println(err)
			return err
		}
	}

	err = tx.QueryRow(existQuery, changedTask.ID).Scan(&exists)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	if exists {
		err = updateSharedTask(changedTask, name)
		if err != nil {
			fmt.Println(err)
			return err
		}
	}
	return nil
}

// updateCategory aktualisiert eine gewünschte Kategorie
//
// Parameter:
//   - name: Der Benutzer, welcher eine Kategorie ändert
//   - catID: Die ID der zu aktualisierenden Kategorie
//   - catName: Der ggf aktualisierte Name der Kategorie
//   - colorHeader: Der Hex-Wert der ggf aktualisierten Farbe des Headers, welche alle Aufgaben dieser Kategorie besitzen
//   - colorBody: Der Hex-Wert der ggf aktualisierten Farbe des Body, welche alle Aufgaben dieser Kategorie besitzen
//
// Rückgabewert:
//   - error: Gibt einen Fehler zurück, wenn bei der Aktualisierung ein Fehler auftritt
//     Gibt "nil" zurück, wenn bei der Erstellung kein Fehler aufgetreten ist
func updateCategory(name string, catID int, catName, colorHeader, colorBody string) error {
	query := `UPDATE categories SET cat_name = ?, color_header = ?, color_body = ? WHERE id = ? AND user_name = ?`
	_, err := db.Exec(query, catName, colorHeader, colorBody, catID, name)
	if err != nil {
		fmt.Println(err)
		return err
	}
	return nil
}

// getTasksForUser gibt alle Aufgaben zurück, die einem Benutzer gehören bzw. die für ihn freigegeben sind
// dabei wird gleichzeitig die Kategorie jeder Aufgabe abgerufen und die dazugehörigen Attribute mitgegeben
// die Aufgabe werden nach ihrer gespeicherten Reihenfolge geordnet
// ist der Benutzer gleichzeitig der Besitzer einer Aufgabe, werden weiterhin alle Benutzer mitgegeben, mit denen er die Aufgabe geteilt hat
//
// Parameter:
//   - name: Der Benutzer, für welchen die Aufgaben abgerufen werden sollen
//
// Rückgabewert:
//   - loadedTasks: Alle Aufgaben, die dem Benutzer zugeordnet werden; "nil", falls ein Fehler auftritt
func getTasksForUser(name string) []task {
	query := `SELECT t.id, t.title, t.desc, t.isDone, t.user_name, c.id AS category_id, c.cat_name, c.color_header, c.color_body, o.order_id 
	FROM tasks t
	LEFT JOIN categories c ON t.category_id = c.id
	LEFT JOIN task_order o ON t.id = o.task_id AND o.user_name = ?
	WHERE t.user_name = ?

	UNION

	SELECT t.id, t.title, t.desc, t.isDone, t.user_name, c.id AS category_id, c.cat_name, c.color_header, c.color_body, o.order_id 
	FROM tasks t
	LEFT JOIN categories c ON t.category_id = c.id
	LEFT JOIN task_order o ON t.id = o.task_id AND o.user_name = ?
	INNER JOIN sharing s ON t.id = s.task_id
	WHERE s.target_name = ?
	
	ORDER BY o.order_id;`

	rows, err := db.Query(query, name, name, name, name)
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer rows.Close()

	loadedTasks := []task{}
	for rows.Next() {
		var shared *[]string
		var task_id, cat_id, order int
		var title, desc, cat_name, color_header, color_body, owner string
		var isDone bool

		err := rows.Scan(&task_id, &title, &desc, &isDone, &owner, &cat_id, &cat_name, &color_header, &color_body, &order)
		if err != nil {
			fmt.Println(err)
			return nil
		}
		if name == owner {
			shared, err = getSharedUsersForTask(task_id)
			if err != nil {
				fmt.Println(err)
				return nil
			}
			loadedTasks = append(loadedTasks, *NewTask(task_id, title, desc, isDone, *NewCategory(cat_id, cat_name, color_header, color_body), owner, *shared, order))
		} else {
			loadedTasks = append(loadedTasks, *NewTask(task_id, title, desc, isDone, *NewCategory(cat_id, cat_name, color_header, color_body), owner, []string{}, order))
		}

	}

	if err := rows.Err(); err != nil {
		fmt.Println(err)
		return nil
	}

	return loadedTasks
}

// getSharedUsersForTask bestimmt für eine geteilte Aufgabe alle Benutzer, für welche die Aufgabe freigegeben wurde
//
// Parameter:
//   - taskID: Die ID der Aufgabe, für die nach den Benutzern gesucht werden soll
//
// Rückgabewert:
//   - &shared: Ein Pointer auf die gefundenen Benutzer, für die die Aufgabe freigegeben ist; "nil", falls ein Fehler aufgetreten ist
//   - error: Ein Fehler, falls bei der Ermittlung der Benutzer ein Fehler aufgetreten ist; "nil", falls nicht
func getSharedUsersForTask(taskID int) (*[]string, error) {
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

// shareTask führt eine Transaktion in der Datenbank aus, wobei eine Aufgabe für einen bestimmten Benutzer freigegeben und dieser darüber benachrichtigt wird, indem die Aufgabe übermittelt wird
// dabei wird zunächst geprüft, ob der Zielbenutzer existiert
// weiterhin wird die freigegebene Aufgabe in die Reihenfolgetabelle des Zielbenutzers eingetragen
//
// Parameter:
//   - sharedTask: Die Aufgabe, welche freigegeben werden soll
//   - target: Der Benutzername der Zielperson
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Freigabe ein Fehler aufgetreten ist; "nil", falls nicht
func shareTask(sharedTask task, target string) error {
	existQuery := `SELECT EXISTS(SELECT 1 FROM users WHERE name = ?)`
	shareQuery := `INSERT INTO sharing (task_id, target_name) VALUES (?,?)`
	orderQuery := `INSERT INTO task_order (user_name, task_id, order_id) VALUES (?,?,?)`
	totalTaskQuery := `SELECT COUNT(*) AS total_tasks
	FROM (
		SELECT id
		FROM tasks
		WHERE user_name = ?

		UNION

		SELECT t.id
		FROM tasks t
		INNER JOIN sharing s ON t.id = s.task_id
		WHERE s.target_name = ?
	) AS combined_tasks;`

	var exists bool
	var totalTasks int

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.QueryRow(totalTaskQuery, target, target).Scan(&totalTasks)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.QueryRow(existQuery, target).Scan(&exists)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	if !exists {
		tx.Rollback()
		return errors.New("Benutzer konnte nicht gefunden werden")
	}

	_, err = tx.Exec(shareQuery, sharedTask.ID, target)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return errors.New("Task bereits für diesen Benutzer freigegeben")
	}
	sharedTask.Order = totalTasks + 1
	_, err = tx.Exec(orderQuery, target, sharedTask.ID, sharedTask.Order)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
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

// removeSharingForUser führt eine Transaktion in der Datenbank aus, welche die Freigabe einer Aufgabe für einen bestimmten Benutzer aufhebt und diesen darüber benachrichtigt
// dabei wird die Reihenfolge der Aufgaben für den betroffenen Benutzer angepasst
//
// Parameter:
//   - taskID: Die ID der Aufgabe, für die die Freigabe aufgehoben werden soll
//   - target: Der Benutzername der Zielperson
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Aufhebung der Freigabe ein Fehler aufgetreten ist; "nil", falls nicht
func removeSharingForUser(taskID int, target string) error {
	removeShareQuery := `DELETE FROM sharing WHERE task_id = ? AND target_name = ?`
	getOrderQuery := `SELECT order_id FROM task_order WHERE task_id = ?`
	removeOrderQuery := `DELETE FROM task_order WHERE task_id = ? AND user_name = ?`
	updateOrderQuery := `UPDATE task_order SET order_id = order_id - 1 WHERE user_name = ? AND order_id > ?`
	var taskOrder int

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.QueryRow(getOrderQuery, taskID).Scan(&taskOrder)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec(removeShareQuery, taskID, target)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec(removeOrderQuery, taskID, target)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec(updateOrderQuery, target, taskOrder)
	if err != nil {
		tx.Rollback()
		return err
	}
	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	message, err := json.Marshal(taskID)
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

// updateSharedTask benachrichtigt alle Benutzer über die Änderung einer für sie freigegebenen Aufgabe
// sorgt dafür, dass die Kommunikation auch von einem Benutzer zum Besitzer der Aufgabe funktioniert
//
// Parameter:
//   - task: Die Aufgabe, welche geändert wurde
//   - user: Der Benutzername, welcher die Änderung vorgenommen hat
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Benachrichtigung ein Fehler aufgetreten ist; "nil", falls nicht
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

// addCategory führt eine Transaktion in der Datenbank aus, um eine neue Kategorie für einen bestimmten Benutzer hinzuzufügen
//
// Parameter:
//   - catName: Der Name der Kategorie
//   - colorHeader: Der Hex-Wert der Farbe des Headers, welche alle Aufgaben dieser Kategorie besitzen
//   - colorBody: Der Hex-Wert der Farbe des Body, welche alle Aufgaben dieser Kategorie besitzen
//   - name: Der Name des Benutzers, welcher die Kategorie anlegt
//
// Rückgabewert:
//   - addedCategoryID: Die von der Datenbank zurückgegebene ID der angelegten Kategorie; 0, falls ein Fehler aufgetreten ist
func addCategory(catName, colorHeader, colorBody, name string) int {
	query := `INSERT INTO categories (cat_name, color_header, color_body, user_name) VALUES (?,?,?,?)`
	newCategory, err := db.Exec(query, catName, colorHeader, colorBody, name)
	if err != nil {
		fmt.Println(err)
		return 0
	}
	addedCategoryID, _ := newCategory.LastInsertId()
	return int(addedCategoryID)
}
func deleteCategory(user_name string, id int) ([]task, error) {
	taskQuery := `UPDATE tasks SET category_id = 1 WHERE category_id = ? AND user_name = ?`
	categoryQuery := `DELETE FROM categories WHERE id = ? AND user_name = ?`

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
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

	return getTasksForUser(user_name), nil
}

// getCategoriesForUser lädt alle Kategorien aus der Datenbank, die für den Benutzer bereits existieren
//
// Parameter:
//   - name: Der Name des Benutzers, für welchen die Kategorien geladen werden sollen
//
// Rückgabewert:
//   - loadedCategories: Die von der Datenbank gefundenen Kategorien für diesen Benutzer; "nil", falls ein Fehler auftritt
func getCategoriesForUser(name string) []category {
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

// updateOrder führt eine Transaktion in der Datenbank aus, um die Reihenfolge der Aufgaben für einen Benutzer zu aktualisieren
// dabei werden zwei benachbarte Aufgaben getauscht
// Parameter:
//   - name: Der Name des Benutzers, für welchen die Reihenfolge geändert werden soll
//   - taskIDUp: Die ID der Aufgabe, die einen Platz nach unten rutschen soll
//   - taskIDDown: Die ID der Aufgabe, die einen Platz nach oben rutschen soll
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Transaktion ein Fehler auftritt; "nil", falls nicht
func updateOrder(name string, taskIDUp, taskIDDown int) error {
	queryOrderUp := `UPDATE task_order SET order_id = order_id + 1 WHERE user_name = ? AND task_id = ?`
	queryOrderDown := `UPDATE task_order SET order_id = order_id - 1 WHERE user_name = ? AND task_id = ?`

	tx, err := db.Begin()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	_, err = tx.Exec(queryOrderUp, name, taskIDUp)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	_, err = tx.Exec(queryOrderDown, name, taskIDDown)
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}

	err = tx.Commit()
	if err != nil {
		tx.Rollback()
		fmt.Println(err)
		return err
	}
	return nil
}

// HandleAddNewUser nimmt die mitgeschickten Parameter des Clients entgegen und ruft addNewUser damit auf, um einen neuen Benutzer anzulegen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleAddNewUser(c *fiber.Ctx) error {
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
		err := addNewUser(creds.Name, creds.Password)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Dieser Benutzer existiert bereits"})
		}
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein"})
	}
	return c.Status(201).JSON("Benutzer erfolgreich hinzugefügt")
}

// HandleLogInUser nimmt die mitgeschickten Parameter des Clients entgegen und ruft loginUser damit auf, um einen neuen Benutzer einzuloggen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls beim Login Benutzers ein Fehler auftritt - wird an Client gesendet
//     Bei Erfolg werden token, Aufgaben und Kategorien an den Client gesendet
func HandleLogInUser(c *fiber.Ctx) error {
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
		token, tasks, categories, err := loginUser(creds.Name, creds.Password)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(200).JSON(fiber.Map{"token": token, "tasks": tasks, "categories": categories})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Benutzername und Passwort dürfen nicht leer sein"})
	}
}

// HandleAddTask nimmt die mitgeschickten Parameter des Clients entgegen und ruft addTask damit auf, um eine neue Aufgabe anzulegen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
//     Bei Erfolg wird die ID der neu erstellen Aufgabe an den Client gesendet
func HandleAddTask(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	type TaskInput struct {
		Title    string   `json:"title"`
		Desc     string   `json:"desc"`
		Category category `json:"category"`
		Order    int      `json:"order"`
	}

	var input TaskInput
	if err := c.BodyParser(&input); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	if strings.TrimSpace(input.Title) != "" {
		addedTaskID := addTask(name, input.Title, input.Desc, input.Category, input.Order)
		if addedTaskID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Aufgabe konnte nicht erstellt werden"})
		}
		return c.Status(201).JSON(fiber.Map{"id": addedTaskID})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel darf nicht leer sein"})
	}
}

// HandleDeleteTask nimmt die mitgeschickten Parameter des Clients entgegen und ruft deleteTask damit auf, um eine Aufgabe zu löschen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleDeleteTask(c *fiber.Ctx) error {
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

// HandleUpdateTask nimmt die mitgeschickten Parameter des Clients entgegen und ruft updateTask damit auf, um eine Aufgabe zu aktualisieren
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleUpdateTask(c *fiber.Ctx) error {
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
		err = updateTask(name, *NewTask(i, input.Title, input.Desc, input.IsDone, input.Category, input.Owner, []string{}, 0))
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Aufgabe konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Aufgabe erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Titel darf nicht leer sein"})
	}
}

// HandleShareTask nimmt die mitgeschickten Parameter des Clients entgegen und ruft shareTask damit auf, um eine Aufgabe mit einem anderen Benutzer zu teilen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleShareTask(c *fiber.Ctx) error {
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
		err = shareTask(*NewTask(i, input.Title, input.Desc, input.IsDone, input.Category, name, []string{}, 0), target)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		return c.Status(201).JSON(fiber.Map{"msg": "Task erfolgreich freigegeben"})
	}
	return c.Status(400).JSON(fiber.Map{"error": "Besitzer und Zielperson dürfen nicht identisch sein"})
}

// HandleRemoveSharingForUser nimmt die mitgeschickten Parameter des Clients entgegen und ruft removeSharingForUser damit auf, um eine Freigabe mit einem Benutzer zu beenden
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleRemoveSharingForUser(c *fiber.Ctx) error {
	id := c.Params("id")
	target := c.Params("target")

	i, err := strconv.Atoi(id)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	err = removeSharingForUser(i, target)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Freigabe konnte nicht beendet werden"})
	}
	return c.Status(200).JSON(fiber.Map{"msg": "Freigabe erfolgreich beendet"})
}

// HandleUpdateCategory nimmt die mitgeschickten Parameter des Clients entgegen und ruft updateCategory damit auf, um eine Kategorie zu aktualisieren
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleUpdateCategory(c *fiber.Ctx) error {
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
		err = updateCategory(name, i, input.Cat_name, input.Color_header, input.Color_body)
		if err != nil {
			fmt.Println(err)
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
		}
		return c.Status(200).JSON(fiber.Map{"msg": "Kategorie erfolgreich geändert"})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategorie konnte nicht geändert werden"})
	}
}

// HandleAddCategory nimmt die mitgeschickten Parameter des Clients entgegen und ruft addCategory damit auf, um eine neue Kategorie anzulegen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
//     Bei Erfolg wird die ID der neu erstellen Kategorie an den Client gesendet
func HandleAddCategory(c *fiber.Ctx) error {
	name := c.Locals("name").(string)

	var input category
	if err := c.BodyParser(&input); err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}

	if strings.TrimSpace(input.Cat_name) != "" {
		addedCategoryID := addCategory(input.Cat_name, input.Color_header, input.Color_body, name)
		if addedCategoryID == 0 {
			return c.Status(400).JSON(fiber.Map{"error": "Kategorie existiert bereits"})
		}
		return c.Status(201).JSON(fiber.Map{"id": addedCategoryID})
	} else {
		return c.Status(400).JSON(fiber.Map{"error": "Kategoriename darf nicht leer sein"})
	}
}

// HandleDeleteCategory nimmt die mitgeschickten Parameter des Clients entgegen und ruft deleteCategory damit auf, um eine Kategorie zu löschen
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
//     Bei Erfolg werden die aktualisierten Aufgabe zurück an den Client geschickt
func HandleDeleteCategory(c *fiber.Ctx) error {
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

// HandleUpdateOrder nimmt die mitgeschickten Parameter des Clients entgegen und ruft updateOrder damit auf, um die Reihenfolge der Aufgaben für einen Benutzer zu ändern
//
// Parameter:
//   - c: Ein Pointer auf ein Context-Objekt von fiber
//
// Rückgabewert:
//   - error: Ein Fehler, falls bei der Erstellung des Benutzers ein Fehler auftritt - wird an Client gesendet
func HandleUpdateOrder(c *fiber.Ctx) error {
	name := c.Locals("name").(string)
	id_up := c.Params("idUp")
	id_down := c.Params("idDown")

	up, err := strconv.Atoi(id_up)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	down, err := strconv.Atoi(id_down)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Ungültige Eingabedaten"})
	}
	err = updateOrder(name, up, down)
	if err != nil {
		fmt.Println(err)
		return c.Status(400).JSON(fiber.Map{"error": "Reihenfolge konnte nicht geändert werden"})
	}
	return c.Status(200).JSON(fiber.Map{"msg": "Reihenfolge erfolgreich geändert"})

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

	app.Post("/api/users/new", HandleAddNewUser)
	app.Post("/api/users", HandleLogInUser)

	app.Use(jwtMiddleware())

	// Task Routen
	app.Post("/api/tasks", HandleAddTask)
	app.Delete("/api/tasks/:id", HandleDeleteTask)
	app.Patch("/api/tasks/:id", HandleUpdateTask)
	app.Post("/api/tasks/:id/:target", HandleShareTask)
	app.Delete("/api/tasks/:id/:target", HandleRemoveSharingForUser)
	app.Patch("/api/tasks/:idUp/:idDown", HandleUpdateOrder)

	// Category Routen
	app.Post("/api/categories", HandleAddCategory)
	app.Patch("/api/categories/:id/delete", HandleDeleteCategory)
	app.Patch("/api/categories/:id", HandleUpdateCategory)

	app.Listen(":5000")
}
