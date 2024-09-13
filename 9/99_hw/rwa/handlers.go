package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

func (sm *ServerManager) CurrentUser(w http.ResponseWriter, r *http.Request) {
	headerToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(headerToken) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, "Invalid token")
		return
	}

	token := headerToken[1]

	profile, err := sm.GetUserByToken(token)

	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, err)
		return
	}

	newBody, err := json.Marshal(map[string]interface{}{
		"user": profile,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(newBody))
}

func (sm *ServerManager) GetArticles(w http.ResponseWriter, r *http.Request) {
	articles, err := sm.GetAllArticles(r.URL.Query())

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	newBody, err := json.Marshal(map[string]interface{}{
		"articles":      articles,
		"articlesCount": len(articles),
	})

	fmt.Println(string(newBody))

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(newBody))
}

func (sm *ServerManager) Register(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	var resp map[string]json.RawMessage
	var profile Profile

	err = json.Unmarshal(body, &resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err = json.Unmarshal(resp["user"], &profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	profile.CreatedAt = time.Now().Format(time.RFC3339)
	profile.UpdatedAt = time.Now().Format(time.RFC3339)

	err = sm.RegisterProfile(&profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	sm.CreateSession(&w, profile.ID)

	newBody, err := json.Marshal(map[string]interface{}{
		"user": profile,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(newBody))
}

func (sm *ServerManager) Login(w http.ResponseWriter, r *http.Request) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	var resp map[string]json.RawMessage
	var profile Profile

	err = json.Unmarshal(body, &resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err = json.Unmarshal(resp["user"], &profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	exProfile, err := sm.CheckPassword(&profile)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	newBody, err := json.Marshal(map[string]interface{}{
		"user": exProfile,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(newBody))
}

func (sm *ServerManager) CreateArticle(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]

	author, err := sm.GetUserByToken(token)
	fmt.Printf("User - %#v\n", author)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, err)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	var resp map[string]json.RawMessage
	var article Article
	err = json.Unmarshal(body, &resp)

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	err = json.Unmarshal(resp["article"], &article)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	sm.CreateArticleDB(&article, author.ID)

	article.Author = Profile{
		Bio:      author.Bio,
		Username: author.Username,
	}

	newBody, err := json.Marshal(map[string]interface{}{
		"article": article,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusCreated)
	fmt.Fprint(w, string(newBody))
}

func (sm *ServerManager) Logout(w http.ResponseWriter, r *http.Request) {
	headerToken := strings.Split(r.Header.Get("Authorization"), " ")
	if len(headerToken) != 2 {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	token := headerToken[1]
	sm.LogoutDB(token)
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Success")
}

func (sm *ServerManager) UpdateUser(w http.ResponseWriter, r *http.Request) {
	token := strings.Split(r.Header.Get("Authorization"), " ")[1]

	body, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	var resp map[string]json.RawMessage
	var profile Profile

	err = json.Unmarshal(body, &resp)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
	}

	err = json.Unmarshal(resp["user"], &profile)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	updProfile, err := sm.UpdateUserDB(&profile, token)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	newBody, err := json.Marshal(map[string]interface{}{
		"user": updProfile,
	})

	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, err)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, string(newBody))
}
