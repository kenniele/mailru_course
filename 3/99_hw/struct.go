package main

type User struct {
	Browsers []string `json:"browsers"`
	Company  string   `json:"-"`
	Country  string   `json:"-"`
	Email    string   `json:"email"`
	Job      string   `json:"-"`
	Name     string   `json:"name"`
	Phone    string   `json:"-"`
}
