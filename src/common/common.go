package common

import (
	"database/sql"
	"fmt"
	"reflect"
)

func GetRcsColumnPq(mtype interface{}) []string {
	t := reflect.TypeOf(mtype)
	columns := make([]string, t.NumField())
	for i := 0; i < t.NumField(); i++ {
		columns[i] = t.Field(i).Tag.Get("db")
	}
	return columns
}

func UpdateProcFlag(tx *sql.Tx, tableName string, ids []interface{}) error {
	commastr := fmt.Sprintf("update %s set proc_flag='N' where msgid in (", tableName)
	for i := range ids {
		if i == 0 {
			commastr += fmt.Sprintf("$%d", i+1)
		} else {
			commastr += fmt.Sprintf(", $%d", i+1)
		}
	}
	commastr += ")"

	_, err := tx.Exec(commastr, ids...)
	return err
}
