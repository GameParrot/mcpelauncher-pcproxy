package utils

import (
	"reflect"
	"unsafe"
)

// UpdatePrivateField sets a private field to the value passed.
func UpdatePrivateField[T any](v any, name string, value T) {
	reflectedValue := reflect.ValueOf(v).Elem()
	privateFieldValue := reflectedValue.FieldByName(name)

	privateFieldValue = reflect.NewAt(privateFieldValue.Type(), unsafe.Pointer(privateFieldValue.UnsafeAddr())).Elem()

	privateFieldValue.Set(reflect.ValueOf(value))
}

// FetchPrivateField fetches a private field.
func FetchPrivateField[T any](s any, name string) T {
	reflectedValue := reflect.ValueOf(s).Elem()
	privateFieldValue := reflectedValue.FieldByName(name)
	privateFieldValue = reflect.NewAt(privateFieldValue.Type(), unsafe.Pointer(privateFieldValue.UnsafeAddr())).Elem()

	return privateFieldValue.Interface().(T)
}

func UnsafeCast[T any](s any) *T {
	v := reflect.ValueOf(s)
	if v.Kind() != reflect.Ptr {
		return (*T)(unsafe.Pointer(&s))
	}
	return (*T)(unsafe.Pointer(v.Pointer()))
}
