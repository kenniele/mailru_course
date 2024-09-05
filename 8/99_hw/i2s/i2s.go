package main

import (
	"fmt"
	"reflect"
)

func i2s(data interface{}, out interface{}) error {
	if reflect.ValueOf(out).Kind() != reflect.Ptr && reflect.ValueOf(out).Kind() != reflect.Slice {
		return fmt.Errorf("OUT is not a pointer")
	}

	switch v := reflect.ValueOf(data); v.Kind() {
	case reflect.Map:

		reflOut := reflect.ValueOf(out).Elem()
		for _, mapKey := range v.MapKeys() {

			mapValue := v.MapIndex(mapKey)
			field := reflOut.FieldByName(mapKey.String())
			switch field.Kind() {
			case reflect.Int:
				switch mapValue.Interface().(type) {
				case int:
					field.SetInt(int64(mapValue.Interface().(int)))
				case float64:
					field.SetInt(int64(mapValue.Interface().(float64)))
				default:
					return fmt.Errorf("mapValue is not an integer")
				}
			case reflect.String:
				switch mapValue.Interface().(type) {
				case string:
					field.SetString(mapValue.Interface().(string))
				default:
					return fmt.Errorf("mapValue is not a string")
				}
			case reflect.Bool:
				switch mapValue.Interface().(type) {
				case bool:
					field.SetBool(mapValue.Interface().(bool))
				default:
					return fmt.Errorf("mapValue is not a bool")
				}
			case reflect.Slice:
				switch mapValue.Interface().(type) {
				case []interface{}:
					if field.Kind() == reflect.Struct {
						return fmt.Errorf("mapValue is not a struct")
					}
					sliceMapValue := mapValue.Interface().([]interface{})
					newSlice := reflect.MakeSlice(field.Type(), 0, len(sliceMapValue))

					for i := 0; i < len(sliceMapValue); i++ {
						cur := reflect.ValueOf(sliceMapValue[i]).Interface()
						newOut := reflect.New(field.Type().Elem()).Interface()

						err := i2s(cur, newOut)
						if err != nil {
							return err
						}
						newSlice = reflect.Append(newSlice, reflect.ValueOf(newOut).Elem())
					}
					field.Set(newSlice)
				default:
					return fmt.Errorf("mapValue is not an array")
				}

			case reflect.Struct:
				newOut := reflect.New(field.Type()).Interface()
				err := i2s(mapValue.Interface(), newOut)
				if err != nil {
					return err
				}

				field.Set(reflect.ValueOf(newOut).Elem())
			default:
				return fmt.Errorf("expected a slice, got %v", field.Kind())
			}
		}
	case reflect.Slice:
		reflOut := reflect.ValueOf(out).Elem()
		if reflOut.Kind() != reflect.Slice {
			return fmt.Errorf("OUT is not a slice")
		}
		newSlice := reflect.MakeSlice(reflOut.Type(), 0, v.Len())
		for i := 0; i < v.Len(); i++ {
			val := v.Index(i).Interface()
			newOut := reflect.New(reflOut.Type().Elem()).Interface()
			err := i2s(val, newOut)

			if err != nil {
				return err
			}

			newSlice = reflect.Append(newSlice, reflect.ValueOf(newOut).Elem())
		}

		reflOut.Set(newSlice)
	default:
		return fmt.Errorf("data is not a struct")
	}
	return nil
}
