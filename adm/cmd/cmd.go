package cmd

import (
	"errors"
	"reflect"
)

func bind(args []string, input interface{}) error {
	typeOf := reflect.TypeOf(input)
	valueOf := reflect.ValueOf(input)

	if typeOf.Kind() != reflect.Ptr || typeOf.Elem().Kind() != reflect.Struct {
		return errors.New("input must be a pointer to a structure")
	}

	for i := 0; i < len(args); i++ {
		valueOf.Elem().
			FieldByName(typeOf.Elem().FieldByIndex([]int{i}).Name).
			SetString(args[i])
	}

	return nil
}
