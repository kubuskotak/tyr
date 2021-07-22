package tyr

import (
	"database/sql"
	"reflect"
)

type interfaceLoader struct {
	v   interface{}
	typ reflect.Type
}

func InterfaceLoader(value, concreteType interface{}) interface{} {
	return interfaceLoader{value, reflect.TypeOf(concreteType)}
}

// Load loads any value from sql.Rows.
//
// value can be:
//
// 1. simple type like int64, string, etc.
//
// 2. sql.Scanner, which allows loading with custom types.
//
// 3. map; the first column from SQL result loaded to the key,
// and the rest of columns will be loaded into the value.
// This is useful to dedup SQL result with first column.
//
// 4. map of slice; like map, values with the same key are
// collected with a slice.
func Load(rows *sql.Rows, value interface{}) (int, error) {
	defer rows.Close()

	column, errCol := rows.Columns()
	if errCol != nil {
		return 0, errCol
	}
	ptr := make([]interface{}, len(column))

	var v reflect.Value
	var elemType reflect.Type

	if il, ok := value.(interfaceLoader); ok {
		v = reflect.ValueOf(il.v)
		elemType = il.typ
	} else {
		v = reflect.ValueOf(value)
	}

	if v.Kind() != reflect.Ptr || v.IsNil() {
		return 0, ErrInvalidPointer
	}
	v = v.Elem()
	isScanner := v.Addr().Type().Implements(typeScanner)
	isSlice := v.Kind() == reflect.Slice && v.Type().Elem().Kind() != reflect.Uint8 && !isScanner
	isMap := v.Kind() == reflect.Map && !isScanner
	isMapOfSlices := isMap && v.Type().Elem().Kind() == reflect.Slice && v.Type().Elem().Elem().Kind() != reflect.Uint8
	if isMap {
		v.Set(reflect.MakeMap(v.Type()))
	}

	s := newTagStore()
	count := 0
	for rows.Next() {
		var elem, keyElem reflect.Value

		if elemType != nil {
			elem = reflectAlloc(elemType)
		} else if isMapOfSlices {
			elem = reflectAlloc(v.Type().Elem().Elem())
		} else if isSlice || isMap {
			elem = reflectAlloc(v.Type().Elem())
		} else {
			elem = v
		}

		if isMap {
			if err := s.findPtr(elem, column[1:], ptr[1:]); err != nil {
				return 0, err
			}
			keyElem = reflectAlloc(v.Type().Key())
			if err := s.findPtr(keyElem, column[:1], ptr[:1]); err != nil {
				return 0, err
			}
		} else if err := s.findPtr(elem, column, ptr); err != nil {
			return 0, err
		}

		// Before scanning, set nil pointer to dummy dest.
		// After that, reset pointers to nil for the next batch.
		for i := range ptr {
			if ptr[i] == nil {
				ptr[i] = dummyDest
			}
		}
		if err := rows.Scan(ptr...); err != nil {
			return 0, err
		}
		for i := range ptr {
			ptr[i] = nil
		}

		count++

		switch {
		case isSlice:
			v.Set(reflect.Append(v, elem))
		case isMapOfSlices:
			s := v.MapIndex(keyElem)
			if !s.IsValid() {
				s = reflect.Zero(v.Type().Elem())
			}
			v.SetMapIndex(keyElem, reflect.Append(s, elem))
		case isMap:
			v.SetMapIndex(keyElem, elem)
		default:
			break
		}
	}
	return count, rows.Err()
}

func reflectAlloc(typ reflect.Type) reflect.Value {
	if typ.Kind() == reflect.Ptr {
		return reflect.New(typ.Elem())
	}
	return reflect.New(typ).Elem()
}

type dummyScanner struct{}

func (dummyScanner) Scan(interface{}) error {
	return nil
}

var (
	dummyDest   sql.Scanner = dummyScanner{}
	typeScanner             = reflect.TypeOf((*sql.Scanner)(nil)).Elem()
)

func GetTagColumns(structValue interface{}) []string {
	t := reflect.Indirect(reflect.ValueOf(structValue))
	m := newTagStore()
	return m.get(t.Type())
}
