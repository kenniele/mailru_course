package main

import (
	"fmt"
	"github.com/gorilla/mux"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"net/http"
)

type ServerManager struct {
	m  *mux.Router
	db *gorm.DB
}

// сюда писать код

func GetApp() http.Handler {
	dsn := "root:love@tcp(127.0.0.1:3307)/photolist?charset=utf8&interpolateParams=true"
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		fmt.Println("Error was invoked", err)
		return nil
	}

	sm := &ServerManager{
		m:  mux.NewRouter(),
		db: db,
	}

	sm.m.HandleFunc("/api/user", sm.CurrentUser).Methods("GET")
	sm.m.HandleFunc("/api/articles", sm.GetArticles).Methods("GET")

	sm.m.HandleFunc("/api/users", sm.Register).Methods("POST")
	sm.m.HandleFunc("/api/users/login", sm.Login).Methods("POST")
	sm.m.HandleFunc("/api/articles", sm.CreateArticle).Methods("POST")
	sm.m.HandleFunc("/api/user/logout", sm.Logout).Methods("POST")

	sm.m.HandleFunc("/api/user", sm.UpdateUser).Methods("PUT")

	return sm.m
}

func indexHandle(w http.ResponseWriter, r *http.Request) {
	fmt.Println("aa")
}
