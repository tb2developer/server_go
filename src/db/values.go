package db

import (
	"database/sql"
	"fmt"
)

type mapStringScan struct {
	// cp are the column pointers
	cp []interface{}
	// row contains the final result
	row      map[string]string
	colCount int
	colNames []string
}

func ScanColumnNames(rows *sql.Rows) *mapStringScan {
	columnNames, err := rows.Columns()
	if err != nil {
		fmt.Print(err)
	}
	rc := newMapStringScan(columnNames)
	return rc
}

func GetValuesRow(rows *sql.Rows, rc *mapStringScan) map[string]string {
	err := rc.update(rows)
	if err != nil {
		fmt.Print(err)
	}
	cv := rc.get()
	return cv
}

func (s *mapStringScan) update(rows *sql.Rows) error {
	if err := rows.Scan(s.cp...); err != nil {
		return err
	}

	for i := 0; i < s.colCount; i++ {
		if rb, ok := s.cp[i].(*sql.RawBytes); ok {
			s.row[s.colNames[i]] = string(*rb)
			*rb = nil // reset pointer to discard current value to avoid a bug
		} else {
			return fmt.Errorf("Cannot convert index %d column %s to type *sql.RawBytes", i, s.colNames[i])
		}
	}
	return nil
}

func (s *mapStringScan) get() map[string]string {
	return s.row
}

func newMapStringScan(columnNames []string) *mapStringScan {
	lenCN := len(columnNames)
	s := &mapStringScan{
		cp:       make([]interface{}, lenCN),
		row:      make(map[string]string, lenCN),
		colCount: lenCN,
		colNames: columnNames,
	}
	for i := 0; i < lenCN; i++ {
		s.cp[i] = new(sql.RawBytes)
	}
	return s
}
