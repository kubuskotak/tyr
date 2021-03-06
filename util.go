package tyr

import (
	"database/sql/driver"
	"reflect"
	"strings"
)

var NameMapping = camelCaseToSnakeCase

func isUpper(b byte) bool {
	return b >= 'A' && b <= 'Z'
}

func isLower(b byte) bool {
	return b >= 'a' && b <= 'z'
}

func isDigit(b byte) bool {
	return b >= '0' && b <= '9'
}

func toLower(b byte) byte {
	if isUpper(b) {
		return b - 'A' + 'a'
	}
	return b
}

func camelCaseToSnakeCase(name string) string {
	var buf strings.Builder
	buf.Grow(len(name) * 2)

	for i := 0; i < len(name); i++ {
		if err := buf.WriteByte(toLower(name[i])); err == nil {
			if i != len(name)-1 && isUpper(name[i+1]) &&
				(isLower(name[i]) || isDigit(name[i]) ||
					(i != len(name)-2 && isLower(name[i+2]))) {
				_ = buf.WriteByte('_')
			}
		}
	}

	return buf.String()
}

var (
	typeValuer = reflect.TypeOf((*driver.Valuer)(nil)).Elem()
)

type tagStore struct {
	m map[reflect.Type][]string
}

func newTagStore() *tagStore {
	return &tagStore{
		m: make(map[reflect.Type][]string),
	}
}

func (s *tagStore) get(t reflect.Type) []string {
	if t.Kind() != reflect.Struct {
		return nil
	}
	if _, ok := s.m[t]; !ok {
		l := make([]string, t.NumField())
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)
			if field.PkgPath != "" && !field.Anonymous {
				// unexported
				continue
			}
			tag := field.Tag.Get(sqlTag)
			if tag == "-" {
				// ignore
				continue
			}
			if tag == "" {
				// no tag, but we can record the field name
				tag = NameMapping(field.Name)
			}
			l[i] = tag
		}
		s.m[t] = l
	}
	return s.m[t]
}

func (s *tagStore) findPtr(value reflect.Value, name []string, ptr []interface{}) error {
	if value.CanAddr() && value.Addr().Type().Implements(typeScanner) {
		ptr[0] = value.Addr().Interface()
		return nil
	}
	switch value.Kind() {
	case reflect.Struct:
		s.findValueByName(value, name, ptr, true)
		return nil
	case reflect.Ptr:
		if value.IsNil() {
			value.Set(reflect.New(value.Type().Elem()))
		}
		return s.findPtr(value.Elem(), name, ptr)
	default:
		ptr[0] = value.Addr().Interface()
		return nil
	}
}

func (s *tagStore) findValueByName(value reflect.Value, name []string, ret []interface{}, retPtr bool) {
	if value.Type().Implements(typeValuer) {
		return
	}
	switch value.Kind() {
	case reflect.Ptr:
		if value.IsNil() {
			return
		}
		s.findValueByName(value.Elem(), name, ret, retPtr)
	case reflect.Struct:
		l := s.get(value.Type())
		for i := 0; i < value.NumField(); i++ {
			tag := l[i]
			if tag == "" {
				continue
			}
			fieldValue := value.Field(i)
			for i, want := range name {
				if want != tag {
					continue
				}
				if ret[i] == nil {
					if retPtr {
						ret[i] = fieldValue.Addr().Interface()
					} else {
						ret[i] = fieldValue
					}
				}
			}
			s.findValueByName(fieldValue, name, ret, retPtr)
		}
	}
}

func interpolateSql(d Dialect, i Buffer, query string, value []interface{}) error {
	valueIndex := 0
	N := 0

	for {
		index := strings.Index(query, placeholder)
		if index == -1 {
			break
		}

		// escape placeholder by repeating it twice
		if strings.HasPrefix(query[index:], escapedPlaceholder) {
			_, _ = i.WriteString(query[:index+1]) // Write placeholder once, not twice
			query = query[index+len(escapedPlaceholder):]
			continue
		}

		_, _ = i.WriteString(query[:index])
		_, _ = i.WriteString(d.Placeholder(N))
		N++
		_ = i.WriteValue(value[valueIndex])
		query = query[index+len(placeholder):]
		valueIndex++
	}
	_, _ = i.WriteString(query)
	return nil
}
