package csvutil

import (
	"reflect"
	"strings"
)

type tag struct {
	name      string
	prefix    string
	empty     bool
	omitEmpty bool
	ignore    bool
	inline    bool
}

// credit to Pascal de Kloe on stackexchange for this function
// https://codereview.stackexchange.com/a/280193
func splitEscapedString(s, separator, escapeString string) []string {
	a := strings.Split(s, separator)

	for i := len(a) - 2; i >= 0; i-- {
		if strings.HasSuffix(a[i], escapeString) {
			a[i] = a[i][:len(a[i])-len(escapeString)] + separator + a[i+1]
			a = append(a[:i+1], a[i+2:]...)
		}
	}
	return a
}

func parseTag(tagname string, field reflect.StructField) (t tag) {
	tags := splitEscapedString(field.Tag.Get(tagname), ",", "\\")
	if len(tags) == 1 && tags[0] == "" {
		t.name = field.Name
		t.empty = true
		return
	}

	switch tags[0] {
	case "-":
		t.ignore = true
		return
	case "":
		t.name = field.Name
	default:
		t.name = tags[0]
	}

	for _, tagOpt := range tags[1:] {
		switch tagOpt {
		case "omitempty":
			t.omitEmpty = true
		case "inline":
			if walkType(field.Type).Kind() == reflect.Struct {
				t.inline = true
				t.prefix = tags[0]
			}
		}
	}
	return
}
