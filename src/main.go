package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

// ContactController struct
type ContactController struct {
	db *sql.DB
}

// Contact struct
type Contact struct {
	ID      int    `json:"id"`
	Name    string `json:"name"`
	Address string `json:"address"`
	Email   string `json:"email"`
}

// APIError struct
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func panicOnError(err error) {
	if err != nil {
		panic(err)
	}
}

func getEnv(name string, defaultValue string) string {
	value := os.Getenv(name)

	if value == "" {
		return defaultValue
	}

	return value
}

// -- DB Methods
func dbInit() (db *sql.DB, e error) {
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbUser := os.Getenv("DB_USER")
	dbPass := os.Getenv("DB_PASS")
	dbName := os.Getenv("DB_NAME")

	dbConStr := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", dbUser, dbPass, dbHost, dbPort, dbName)

	db, err := sql.Open("mysql", dbConStr)

	if err != nil {
		return nil, err
	}

	return db, nil
}

func createContactFromDb(db *sql.DB, c *Contact) (*Contact, error) {
	stm, err := db.Prepare("INSERT INTO address_contact (name, address, email) VALUES (?, ?, ?)")

	if err != nil {
		return nil, err
	}

	defer stm.Close()

	res, err := stm.Exec(c.Name, c.Address, c.Email)

	if err != nil {
		return nil, err
	}

	id, err := res.LastInsertId()

	if err != nil {
		return nil, err
	}

	c.ID = int(id)

	return c, nil
}

func getContactsFromDb(db *sql.DB) ([]Contact, error) {
	var contacts []Contact

	rows, err := db.Query("SELECT id, name, address, email FROM address_contact")

	if err != nil {
		return nil, err
	}

	defer rows.Close()

	var c Contact

	for rows.Next() {
		err = rows.Scan(&c.ID, &c.Name, &c.Address, &c.Email)

		if err != nil {
			return nil, err
		}

		contacts = append(contacts, c)
	}

	return contacts, nil
}

func getContactFromDb(db *sql.DB, id int) (*Contact, error) {
	var c Contact
	row := db.QueryRow("SELECT id, name, address, email FROM address_contact WHERE id = ? LIMIT 1", id)
	err := row.Scan(&c.ID, &c.Name, &c.Address, &c.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}

		return nil, err
	}

	return &c, nil
}

func updateContactFromDb(db *sql.DB, c Contact) error {
	stm, err := db.Prepare("UPDATE address_contact SET name = ?, address = ?, email = ? WHERE id = ?")

	if err != nil {
		return err
	}

	defer stm.Close()

	_, err = stm.Exec(c.Name, c.Address, c.Email, c.ID)

	return err
}

func deleteContactFromDb(db *sql.DB, c Contact) error {
	stm, err := db.Prepare("DELETE FROM address_contact WHERE id = ?")
	if err != nil {
		return err
	}

	defer stm.Close()

	_, err = stm.Exec(c.ID)

	if err != nil {
		return err
	}

	return nil
}

// -- REST Methods
func errorHandler(w http.ResponseWriter, apiError APIError) {
	w.WriteHeader(http.StatusBadRequest)
	json.NewEncoder(w).Encode(apiError)
}

func (controller *ContactController) getContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	contact, err := getContactFromDb(controller.db, id)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	if contact == nil {
		errorHandler(w, APIError{Code: http.StatusNotFound, Message: "Not Found"})
		return
	}

	encoder.Encode(contact)
}

func (controller *ContactController) createContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	c := new(Contact)
	json.NewDecoder(r.Body).Decode(c)

	c, err := createContactFromDb(controller.db, c)

	if err != nil {
		errorHandler(w, APIError{Code: http.StatusBadRequest, Message: "Bad Request"})
		return
	}

	w.WriteHeader(http.StatusCreated)
	encoder.Encode(*c)
}

func (controller *ContactController) updateContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	contact, err := getContactFromDb(controller.db, id)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	if contact == nil {
		errorHandler(w, APIError{Code: http.StatusNotFound, Message: "Not Found"})
		return
	}

	json.NewDecoder(r.Body).Decode(contact)

	err = updateContactFromDb(controller.db, *contact)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	encoder.Encode(contact)
}

func (controller *ContactController) deleteContactHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	params := mux.Vars(r)
	id, err := strconv.Atoi(params["id"])
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	contact, err := getContactFromDb(controller.db, id)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	if contact == nil {
		errorHandler(w, APIError{Code: http.StatusNotFound, Message: "Not Found"})
		return
	}

	err = deleteContactFromDb(controller.db, *contact)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusBadRequest, Message: err.Error()})
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (controller *ContactController) getContactsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	encoder := json.NewEncoder(w)

	contacts, err := getContactsFromDb(controller.db)
	if err != nil {
		errorHandler(w, APIError{Code: http.StatusInternalServerError, Message: err.Error()})
		return
	}

	encoder.Encode(contacts)
}

func main() {
	_, err := os.Stat("./.env")

	if err == nil {
		err := godotenv.Load()
		if err != nil {
			panicOnError(err)
		}
	}

	port := getEnv("PORT", "5000")

	db, err := dbInit()
	panicOnError(err)
	defer db.Close()

	contactController := &ContactController{db: db}

	router := mux.NewRouter()

	apiV1Router := router.PathPrefix("/api/v1").Subrouter()
	contactsRouter := apiV1Router.PathPrefix("/contacts").Subrouter()

	contactsRouter.HandleFunc("", contactController.getContactsHandler).Methods(http.MethodGet)
	contactsRouter.HandleFunc("", contactController.createContactHandler).Methods(http.MethodPost)
	contactsRouter.HandleFunc("/{id}", contactController.getContactHandler).Methods(http.MethodGet)
	contactsRouter.HandleFunc("/{id}", contactController.updateContactHandler).Methods(http.MethodPut)
	contactsRouter.HandleFunc("/{id}", contactController.deleteContactHandler).Methods(http.MethodDelete)

	fmt.Printf("Listen on: http://0.0.0.0:%s\n", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), router))
}
