package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"slices"
	"strconv"
	"strings"
)

// тут вы пишете код
// обращаю ваше внимание - в этом задании запрещены глобальные переменные

type Handlers struct {
	db *sql.DB
}

func intValid(values ...string) error {
	for _, v := range values {
		if _, err := strconv.Atoi(v); err != nil {
			return err
		}
	}
	return nil
}

func listTables(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SHOW TABLES")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var tables []string
	for rows.Next() {
		var table string
		if err := rows.Scan(&table); err != nil {
			return nil, err
		}
		tables = append(tables, table)
	}
	return tables, nil
}

func getColumnTypes(db *sql.DB, tableName string) (map[string]interface{}, error) {
	query := `SELECT column_name, data_type, is_nullable FROM INFORMATION_SCHEMA.COLUMNS WHERE table_name=?;`
	rows, err := db.Query(query, tableName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make(map[string]interface{})

	var columnName, dataType, is_nullable string
	for rows.Next() {
		err = rows.Scan(&columnName, &dataType, &is_nullable)
		if err != nil {
			return nil, err
		}
		result[columnName] = dataType + "_" + is_nullable
	}
	return result, nil
}

func getPrimaryKeyName(db *sql.DB, tableName string) (string, error) {
	query := `SELECT COLUMN_NAME 
FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
WHERE TABLE_NAME = ? AND CONSTRAINT_NAME = 'PRIMARY' AND TABLE_SCHEMA = DATABASE();`
	rows, err := db.Query(query, tableName)
	if err != nil {
		return "", err
	}
	defer rows.Close()
	var primaryKey string
	for rows.Next() {
		err = rows.Scan(&primaryKey)
		if err != nil {
			return "", errors.New("no primary key found")
		}
	}
	return primaryKey, nil
}

func scanningDb(rows *sql.Rows) ([]map[string]interface{}, int, error) {
	var items []map[string]interface{}
	columns, err := rows.Columns()
	if err != nil {
		return nil, http.StatusInternalServerError, err
	}

	for rows.Next() {
		item := make(map[string]interface{})
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for i := range columns {
			valuePtrs[i] = &values[i]
		}

		if err = rows.Scan(valuePtrs...); err != nil {
			return nil, http.StatusInternalServerError, err
		}

		for i, col := range columns {
			switch values[i].(type) {
			case string:
				item[col] = values[i].(string)
			case json.Number:
				item[col] = strconv.Itoa(values[i].(int))
			case []byte:
				item[col] = string(values[i].([]byte))
			default:
				item[col] = values[i]
			}
		}

		items = append(items, item)
	}

	if len(items) == 0 {
		return nil, http.StatusNotFound, errors.New("no rows found")
	}

	return items, 0, nil
}

func getValues(mapp []map[string]interface{}) []interface{} {
	var result []interface{}
	for _, v := range mapp {
		for _, val := range v {
			result = append(result, val)
		}
	}
	return result
}

func NewDbExplorer(db *sql.DB) (http.Handler, error) {
	r := http.NewServeMux()
	h := &Handlers{db: db}

	r.HandleFunc("/", h.handlerFunc)

	return r, nil
}

func (h *Handlers) handlerFunc(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		getHandler(w, r, h)
	case http.MethodPost:
		postHandler(w, r, h)
	case http.MethodPut:
		putHandler(w, r, h)
	case http.MethodDelete:
		deleteHandler(w, r, h)
	default:
		panic("Undefined method LUL")
	}
}

// WORK
func getHandler(w http.ResponseWriter, r *http.Request, h *Handlers) {
	url := r.URL.Path
	switch {
	case url == "/":
		getAllTables(w, r, h)
	case strings.Count(url, "/") == 1:
		getSingleTable(w, r, h)
	case strings.Count(url, "/") == 2:
		getSingleRecord(w, r, h)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// WORK
func getAllTables(w http.ResponseWriter, r *http.Request, h *Handlers) {
	rows, err := h.db.Query("SHOW TABLES")
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rows == nil {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	var tables []map[string]interface{}

	tables, code, err := scanningDb(rows)
	if err != nil {
		w.WriteHeader(code)
		return
	}

	err = rows.Close()
	if err != nil {
		return
	}
	result := map[string]interface{}{"response": map[string]interface{}{"tables": getValues(tables)}}
	jsoniz, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	fmt.Fprint(w, string(jsoniz))
}

// WORK
func getSingleTable(w http.ResponseWriter, r *http.Request, h *Handlers) {
	table := strings.Split(r.URL.Path, "/")[1]
	limit := r.URL.Query().Get("limit")
	offset := r.URL.Query().Get("offset")
	if _, err := strconv.Atoi(limit); err != nil {
		limit = "5"
	}
	if _, err := strconv.Atoi(offset); err != nil {
		offset = "0"
	}

	tables, err := listTables(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(tables, table) {
		w.WriteHeader(http.StatusNotFound)
		res, err := json.Marshal(map[string]interface{}{"error": "unknown table"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(res))
		return
	}

	err = intValid(limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	query := "SELECT * FROM" + " " + table + " LIMIT ? OFFSET ?"
	rows, err := h.db.Query(query, limit, offset)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	if rows == nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var items []map[string]interface{}

	items, code, err := scanningDb(rows)
	if err != nil {
		w.WriteHeader(code)
		return
	}

	result := map[string]interface{}{"response": map[string]interface{}{"records": items}}

	jsoniz, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(jsoniz))
}

// WORK
func getSingleRecord(w http.ResponseWriter, r *http.Request, h *Handlers) {
	splittedUrl := strings.Split(r.URL.Path, "/")
	table := splittedUrl[1]
	id := splittedUrl[2]

	tables, err := listTables(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(tables, table) {
		res, err := json.Marshal(map[string]interface{}{"error": "unknown table"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, string(res))
		return
	}
	primaryKey, err := getPrimaryKeyName(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("SELECT * FROM %v WHERE %v=?", table, primaryKey)
	rows, err := h.db.Query(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var items []map[string]interface{}

	items, code, err := scanningDb(rows)
	if err != nil {
		w.WriteHeader(code)
		res, _ := json.Marshal(map[string]interface{}{"error": "record not found"})
		fmt.Fprint(w, string(res))
		return
	}

	result := map[string]interface{}{"response": map[string]interface{}{"record": items[0]}}

	jsoniz, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(jsoniz))
}

// WORK
func postHandler(w http.ResponseWriter, r *http.Request, h *Handlers) {
	url := r.URL.Path
	switch {
	case strings.Count(url, "/") == 2:
		postSingleRecord(w, r, h)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// WORK
func postSingleRecord(w http.ResponseWriter, r *http.Request, h *Handlers) {
	splittedUrl := strings.Split(r.URL.Path, "/")
	table := splittedUrl[1]
	id := splittedUrl[2]

	tables, err := listTables(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(tables, table) {
		res, _ := json.Marshal(map[string]interface{}{"error": "unknown table"})
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, string(res))
		return
	}

	var newItem []byte
	newItem, err = io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var item map[string]interface{}
	err = json.Unmarshal(newItem, &item)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var keys []string
	var placeholders []string
	var values []interface{}

	cols, err := getColumnTypes(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	types := map[string]string{
		"varchar": "string",
		"text":    "string",
		"int":     "int",
	}

	primaryKey, err := getPrimaryKeyName(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	for k, v := range item {
		var dataType, isNull string
		if k != primaryKey {
			keys = append(keys, k)
			placeholders = append(placeholders, "?")
			values = append(values, v) // добавляем как interface{}, чтобы избежать принудительного приведения типа
			splittes := strings.Split(cols[k].(string), "_")
			dataType, isNull = types[splittes[0]], splittes[1]
		}
		switch {
		case (v == nil && isNull == "NO") || (v != nil && reflect.TypeOf(v).Name() != dataType):
			errorSt := fmt.Sprintf("field %v have invalid type", k)
			w.WriteHeader(http.StatusBadRequest)
			result := map[string]interface{}{"error": errorSt}
			jsoniz, _ := json.Marshal(result)
			fmt.Fprint(w, string(jsoniz))
			return
		case v == nil:
			values[len(values)-1] = interface{}(nil)
		}
	}

	values = append(values, id)

	if len(keys) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		result := map[string]interface{}{"error": "field id have invalid type"}
		jsoniz, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(jsoniz))
		return
	}

	query := fmt.Sprintf("UPDATE %s SET %s WHERE %s=?", table, strings.Join(keys, "=?,")+"=?", primaryKey)

	rows, err := h.db.Exec(query, values...)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	affRows, err := rows.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	if affRows == 0 {
		result := map[string]interface{}{"error": "No rows affected, check field types or id"}
		jsoniz, _ := json.Marshal(result)
		fmt.Fprint(w, string(jsoniz))
		return
	}

	result := map[string]interface{}{"response": map[string]interface{}{"updated": affRows}}
	jsoniz, _ := json.Marshal(result)
	fmt.Fprint(w, string(jsoniz))
}

// WORK
func putHandler(w http.ResponseWriter, r *http.Request, h *Handlers) {
	url := strings.TrimRight(r.URL.Path, "/")
	switch {
	case strings.Count(url, "/") == 1:
		putSingleRecord(w, r, h)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

// WORK
func putSingleRecord(w http.ResponseWriter, r *http.Request, h *Handlers) {
	var newItem interface{}

	splittedUrl := strings.Split(r.URL.Path, "/")
	table := splittedUrl[1]

	tables, err := listTables(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(tables, table) {
		w.WriteHeader(http.StatusNotFound)
		res, err := json.Marshal(map[string]interface{}{"error": "unknown table"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(res))
		return
	}

	newItem, err = io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	var item map[string]interface{}
	newItem = json.Unmarshal(newItem.([]byte), &item)

	var keys []string
	var values []interface{}

	primaryKey, err := getPrimaryKeyName(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	cols, err := getColumnTypes(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	names := []string{}
	isNull := []string{}
	for col, v := range cols {
		names = append(names, col)
		if strings.Split(v.(string), "_")[1] == "YES" {
			isNull = append(isNull, col)
		}

	}

	for _, k := range names {
		v := item[k]
		if !slices.Contains(names, k) {
			continue
		}
		if k != primaryKey {
			keys = append(keys, k)
			if v == nil && !slices.Contains(isNull, k) {
				v = ""
			}
			values = append(values, v)
		}
	}
	if len(keys) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		result := map[string]interface{}{"error": fmt.Sprintf("field %s have invalid type", primaryKey)}
		jsoniz, err := json.Marshal(result)
		if err != nil {
			return
		}
		fmt.Fprint(w, string(jsoniz))
	}

	query := fmt.Sprintf("INSERT INTO %v ("+strings.Join(keys, ",")+") VALUES (", table) + strings.Repeat("?, ", len(values))[:len(values)*3-2] + ")"
	rows, err := h.db.Exec(query, values...)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	id, err := rows.LastInsertId()

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{"response": map[string]interface{}{primaryKey: id}}

	jsoniz, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(jsoniz))

}

func deleteHandler(w http.ResponseWriter, r *http.Request, h *Handlers) {
	url := r.URL.Path
	switch {
	case strings.Count(url, "/") == 2:
		deleteSingleRecord(w, r, h)
	default:
		w.WriteHeader(http.StatusNotFound)
	}
}

func deleteSingleRecord(w http.ResponseWriter, r *http.Request, h *Handlers) {
	splittedUrl := strings.Split(r.URL.Path, "/")
	table := splittedUrl[1]
	id := splittedUrl[2]

	tables, err := listTables(h.db)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if !slices.Contains(tables, table) {
		w.WriteHeader(http.StatusNotFound)
		res, err := json.Marshal(map[string]interface{}{"error": "unknown table"})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		fmt.Fprint(w, string(res))
		return
	}
	primaryKey, err := getPrimaryKeyName(h.db, table)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	query := fmt.Sprintf("DELETE FROM %v WHERE %v=?", table, primaryKey)
	rows, err := h.db.Exec(query, id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	rowsAffected, err := rows.RowsAffected()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	result := map[string]interface{}{"response": map[string]interface{}{"deleted": rowsAffected}}

	jsoniz, err := json.Marshal(result)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	fmt.Fprintln(w, string(jsoniz))

}
