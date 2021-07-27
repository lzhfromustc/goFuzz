package gooracle

import (
	"reflect"
	"runtime"
)

func IsInterestingType(kind reflect.Kind) bool {
	switch kind {
	case reflect.Chan, reflect.Slice, reflect.Array, reflect.Map, reflect.Struct, reflect.Ptr:
		return true
	default:
		return false
	}
}

// TODO: shared struct updated in one goroutine;
func CurrentGoAddValue(v interface{}, vecScanned []interface{}) {

	if vecScanned == nil {
		vecScanned = []interface{}{}
	}
	for _, existing := range vecScanned {
		if existing == v {
			return
		}
	}
	vecScanned = append(vecScanned, v)

	reflectValue := reflect.ValueOf(v)
	reflectType := reflect.TypeOf(v)

	switch typeKind := reflectType.Kind(); typeKind {
	case reflect.Struct:
		for i := 0; i < reflectType.NumField(); i++ {
			field := reflectType.Field(i)
			fieldType := field.Type
			fieldValue := reflectValue.Field(i).Interface()
			fieldKind := fieldType.Kind()
			if fieldKind == reflect.Chan {
				runtime.AddRefGoroutine(runtime.FindChanInfo(fieldValue), runtime.CurrentGoInfo())
			} else if IsInterestingType(fieldKind) {
				CurrentGoAddValue(fieldValue, vecScanned)
			}
		}
	case reflect.Map:
		if reflectValue.IsNil() {
			return
		}
		kindKey := reflectType.Key().Kind()
		if kindKey == reflect.Chan {
			for _, key := range reflectValue.MapKeys() {
				keyValue := key.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(keyValue), runtime.CurrentGoInfo())
			}
		} else if IsInterestingType(kindKey) {
			for _, key := range reflectValue.MapKeys() {
				keyValue := key.Interface()
				CurrentGoAddValue(keyValue, vecScanned)
			}
		}
		kindElem := reflectType.Elem().Kind()
		if kindElem == reflect.Chan {
			for _, key := range reflectValue.MapKeys() {
				elem := reflectValue.MapIndex(key)
				elemValue := elem.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
			}
		} else if IsInterestingType(kindElem) {
			for _, key := range reflectValue.MapKeys() {
				elem := reflectValue.MapIndex(key)
				elemValue := elem.Interface()
				CurrentGoAddValue(elemValue, vecScanned)
			}
		}

	case reflect.Slice, reflect.Array:
		if typeKind == reflect.Slice {
			if reflectValue.IsNil() {
				return
			}
		}
		kind := reflectType.Elem().Kind()
		if kind == reflect.Chan {
			for i := 0; i < reflectValue.Len(); i++ {
				elem := reflectValue.Index(i)
				elemValue := elem.Interface()
				runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
			}
		} else if IsInterestingType(kind) {
			for i := 0; i < reflectValue.Len(); i++ {
				elem := reflectValue.Index(i)
				elemValue := elem.Interface()
				CurrentGoAddValue(elemValue, vecScanned)
			}
		}
	case reflect.Ptr:
		if reflectValue.IsNil() {
			return
		}
		kind := reflectType.Elem().Kind()
		if kind == reflect.Chan {
			elem := reflectValue.Elem()
			elemValue := elem.Interface()
			runtime.AddRefGoroutine(runtime.FindChanInfo(elemValue), runtime.CurrentGoInfo())
		} else if IsInterestingType(kind) {
			elem := reflectValue.Elem()
			elemValue := elem.Interface()
			CurrentGoAddValue(elemValue, vecScanned)
		}
	case reflect.Chan:
		runtime.AddRefGoroutine(runtime.FindChanInfo(v), runtime.CurrentGoInfo())
	}
}