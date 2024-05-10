package common

import "reflect"

func GetRcsColumnPq(mtype interface{}) []string {
	t := reflect.TypeOf(mtype)
	columns := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		columns[i] = t.Field(i).Tag.Get("db")
	}
	return columns
}
