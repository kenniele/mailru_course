package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type XMLUser struct {
	Id        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type XMLUsers struct {
	Users []XMLUser `xml:"row"`
}

func SearchServer(w http.ResponseWriter, r *http.Request) {
	file, err := os.Open("dataset.xml")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	defer file.Close()

	var xmlUsers XMLUsers
	if err := xml.NewDecoder(file).Decode(&xmlUsers); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")
	orderBy, _ := strconv.Atoi(r.URL.Query().Get("order_by"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	offset, _ := strconv.Atoi(r.URL.Query().Get("offset"))

	if limit < 0 || offset < 0 {
		http.Error(w, "limit or offset must be >= 0.", http.StatusBadRequest)
		return
	}

	var filteredUsers []User
	for _, user := range xmlUsers.Users {
		name := user.FirstName + " " + user.LastName
		if strings.Contains(name, query) || strings.Contains(user.About, query) {
			filteredUsers = append(filteredUsers, User{
				Id:     user.Id,
				Name:   name,
				Age:    user.Age,
				Gender: user.Gender,
				About:  user.About,
			})
		}
	}

	switch orderField {
	case "Id":
		sort.Slice(filteredUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return filteredUsers[i].Id < filteredUsers[j].Id
			} else if orderBy == OrderByDesc {
				return filteredUsers[i].Id < filteredUsers[j].Id
			}
			return true
		})
	case "Age":
		sort.Slice(filteredUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return filteredUsers[i].Age < filteredUsers[j].Age
			} else if orderBy == OrderByDesc {
				return filteredUsers[i].Age < filteredUsers[j].Age
			}
			return true
		})
	case "Name", "":
		sort.Slice(filteredUsers, func(i, j int) bool {
			if orderBy == OrderByAsc {
				return filteredUsers[i].Name < filteredUsers[j].Name
			} else if orderBy == OrderByDesc {
				return filteredUsers[i].Name < filteredUsers[j].Name
			}
			return true
		})
	default:
		http.Error(w, "Invalid order field", http.StatusBadRequest)
		return
	}

	if offset > len(filteredUsers) {
		offset = len(filteredUsers)
	}
	if limit+offset > len(filteredUsers) {
		limit = len(filteredUsers) - offset
	}
	result := filteredUsers[offset : offset+limit]

	w.Header().Set("Content-Type", "application/xml")
	res, err := json.Marshal(result)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Write(res)
}
