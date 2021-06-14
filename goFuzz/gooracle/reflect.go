package gooracle

import (
	"reflect"
	"runtime"
)

// TODO: shared struct updated in one goroutine;
func CurrentGoAddValue(v interface{}) {
	reflectValue := reflect.ValueOf(v)
	reflectType := reflect.TypeOf(v)
	switch reflectType.Kind() {
	case reflect.Struct:
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			fieldType := field.Type
			if fieldType.Kind() == reflect.Chan {
				fieldValue := reflectValue.Field(i).Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(fieldValue), runtime.CurrentGoInfo())
			}
		}
	case reflect.Map:
		if reflectType.Key().Kind() == reflect.Chan {
			for _, key := range reflectValue.MapKeys() {
				keyValue := key.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(keyValue), runtime.CurrentGoInfo())
			}
		} else if reflectType.Elem().Kind() == reflect.Chan {
			for _, key := range reflectValue.MapKeys() {
				elem := reflectValue.MapIndex(key)
				elemValue := elem.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
			}
		}
	case reflect.Slice, reflect.Array:
		if reflectType.Elem().Kind() == reflect.Chan {
			for i := 0; i < reflectValue.Len(); i++ {
				elem := reflectValue.Index(i)
				elemValue := elem.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
			}
		}
	case reflect.Ptr:
		if reflectType.Elem().Kind() == reflect.Chan {
			elem := reflectValue.Elem()
			elemValue := elem.Interface()
			runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
		}
	}
}
