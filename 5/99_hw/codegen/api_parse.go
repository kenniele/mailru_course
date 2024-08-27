package main

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strconv"
	"strings"
)

type Params struct {
	Name string
	Typ string
	Tags map[string]string
}


func ParamsParse(r *http.Request, paramNames []string) []Params {
	result := []Params{}
	for _, v := range paramNames {
		sp := strings.Split(v, "_")
		result = append(result, Params{
			Name: sp[0],
			Typ: sp[1],
			Tags: ParseTag(r.URL.Query().Get(v)),
		})
	}
	return result
}

func ParseTag(tag string) map[string]string {
	result := map[string]string{}
	split := strings.Split(strings.Replace(strings.TrimLeft(tag, "apivalidator:"), "\"", "", -1), ",")
	for _, v := range split {
		if v == "required" {
			result[v] = "true"
		} else {
			sp := strings.Split(v, "=")
			result[sp[0]] = sp[1]
		}
	}
	return result
}

func (h *MyApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/profile":
		h.wrapperDoSomeJob(w, r, []string{"Login_string"}, ""Profile"")

	case "/user/create":
		h.wrapperDoSomeJob(w, r, []string{"Login_string", "Name_string", "Status_string", "Age_int"}, ""Create"")
default:
		w.WriteHeader(404)
	}
}

func (h *MyApi) wrapperDoSomeJob(w http.ResponseWriter, r *http.Request, paramNames []string) {
	params := ParamsParse(r, paramNames)
	for _, v := range params {
		switch v.Typ {
		case "int":
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required"))
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if k < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if k > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}

		case "string":
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required")
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if len(k) < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if len(k) > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}

		default:
			panic("Undefined fieldType")
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	MyApi := nil
}
func (h *OtherApi) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.URL.Path {
	case "/user/create":
		h.wrapperDoSomeJob(w, r, []string{"Username_string", "Name_string", "Class_string", "Level_int"}, ""Create"")
default:
		w.WriteHeader(404)
	}
}

func (h *OtherApi) wrapperDoSomeJob(w http.ResponseWriter, r *http.Request, paramNames []string) {
	params := ParamsParse(r, paramNames)
	for _, v := range params {
		switch v.Typ {
		case "int":
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required"))
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if k < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if k > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}

		case "string":
			for fieldName, fieldValue := range v.Tags {
				switch fieldName {
				case "required":
					if k == 0 {
						return ApiError(500, errors.New("Field is required")
					}
				case "enum":
					cases := strings.Split(fieldValue[5:len(v)-1], "|")
					if !slices.Contains(cases, k) {
						return ApiError(500, errors.New("Field doesn't contain required value")
					}
				case "default" && k == 0:
					k = fieldValue
				case "min":
					if len(k) < strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too small")
					}
				case "max":
					if len(k) > strconv.Atoi(fieldValue) {
						return ApiError(500, errors.New("Field is too big"))
					}
				default:
					panic("Undefined field: " + fieldName)
				}
			}
		}

		default:
			panic("Undefined fieldType")
		}
	}
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)
	cancel()
	OtherApi := nil
}