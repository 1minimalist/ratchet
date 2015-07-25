package ratchet

import (
	"database/sql"
	"fmt"
	"sort"
	"strings"
)

// GetDataFromSQLQuery is a util function that, given a properly intialized sql.DB
// and a valid SQL query, will handle executing the query and getting back Data
// objects. This function is asynch, and Data should be received on teh return
// data channel. If there was a problem setting up the query, then an error will also be
// returned immediately. It is also possible for errors to occur during execution as data
// is retrieved from the query. If this happens, the Data object returned will be a JSON
// object in the form of {"Error": "description"}.
func GetDataFromSQLQuery(db *sql.DB, query string, batchSize int) (chan Data, error) {
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}

	columns, err := rows.Columns()
	if err != nil {
		return nil, err
	}

	dataChan := make(chan Data)

	go func(rows *sql.Rows, columns []string) {
		defer rows.Close()

		tableData := []map[string]interface{}{}
		values := make([]interface{}, len(columns))
		valuePtrs := make([]interface{}, len(columns))

		for rows.Next() {
			for i := 0; i < len(columns); i++ {
				valuePtrs[i] = &values[i]
			}
			rows.Scan(valuePtrs...)
			entry := make(map[string]interface{})
			for i, col := range columns {
				var v interface{}
				val := values[i]
				b, ok := val.([]byte)
				if ok {
					v = string(b)
				} else {
					v = val
				}
				entry[col] = v
			}
			tableData = append(tableData, entry)

			if batchSize > 0 && len(tableData) >= batchSize {
				sendTableData(tableData, dataChan)
				tableData = []map[string]interface{}{}
			}
		}
		if rows.Err() != nil {
			sendErr(rows.Err(), dataChan)
		}

		// Flush remaining tableData
		if len(tableData) > 0 {
			sendTableData(tableData, dataChan)
		}

		close(dataChan) // signal completion to caller
	}(rows, columns)

	return dataChan, nil
}

func sendTableData(tableData []map[string]interface{}, dataChan chan Data) {
	data, err := NewData(tableData)
	if err != nil {
		sendErr(err, dataChan)
	} else {
		dataChan <- data
	}
}

func sendErr(err error, dataChan chan Data) {
	dataChan <- []byte("{\"Error\":\"" + err.Error() + "\"}")
}

// SQLInsertData abstracts building and executing a SQL INSERT
// statement for the given Data object.
//
// Note that the Data valid JSON object
// (or a slice of valid objects all with the same keys),
// where the keys are column names and the
// the values are SQL values to be inserted into those columns.
func SQLInsertData(db *sql.DB, data Data, tableName string, onDupKeyUpdate bool) error {
	var v interface{}
	err := ParseData(data, &v)
	if err != nil {
		return err
	}

	var objects []map[string]interface{}
	// check if we have a single object or a slice of objects
	switch vv := v.(type) {
	case []interface{}:
		for _, o := range vv {
			objects = append(objects, o.(map[string]interface{}))
		}
	case map[string]interface{}:
		objects = []map[string]interface{}{vv}
	case []map[string]interface{}:
		objects = vv
	default:
		return fmt.Errorf("SQLInsertData: unsupported data type: %T", vv)
	}

	return insertObjects(db, objects, tableName, onDupKeyUpdate)
}

func insertObjects(db *sql.DB, objects []map[string]interface{}, tableName string, onDupKeyUpdate bool) error {
	insertSQL := buildInsertSQL(objects, tableName, onDupKeyUpdate)

	stmt, err := db.Prepare(insertSQL)
	if err != nil {
		return err
	}

	cols := sortedColumns(objects[0])
	vals := []interface{}{}
	for _, object := range objects {
		// Since maps aren't ordered, must iterate over sorted columns
		// for each object.
		for _, k := range cols {
			vals = append(vals, object[k])
		}
	}

	res, err := stmt.Exec(vals...)
	if err != nil {
		return err
	}
	lastID, err := res.LastInsertId()
	if err != nil {
		return err
	}
	rowCnt, err := res.RowsAffected()
	if err != nil {
		return err
	}

	LogInfo(fmt.Sprintf("SQLInsertData: rows affected = %d, last insert ID = %d", rowCnt, lastID))
	return nil
}

func buildInsertSQL(objects []map[string]interface{}, tableName string, onDupKeyUpdate bool) (insertSQL string) {
	cols := sortedColumns(objects[0])

	// Format: INSERT INTO tablename(col1,col2) VALUES(?,?),(?,?)
	insertSQL = fmt.Sprintf("INSERT INTO %v(%v) VALUES", tableName, strings.Join(cols, ","))
	// builds the (?,?) part
	vals := "("
	for i := 0; i < len(cols); i++ {
		if i > 0 {
			vals += ","
		}
		vals += "?"
	}
	vals += ")"
	// append as many (?,?) parts as there are objects to insert
	for i := 0; i < len(objects); i++ {
		if i > 0 {
			insertSQL += ","
		}
		insertSQL += vals
	}

	if onDupKeyUpdate {
		// format: ON DUPLICATE KEY UPDATE a=VALUES(a), b=VALUES(b), c=VALUES(c)
		insertSQL += " ON DUPLICATE KEY UPDATE "
		for i, c := range cols {
			if i > 0 {
				insertSQL += ","
			}
			insertSQL += "`" + c + "`=VALUES(`" + c + "`)"
		}
	}

	return insertSQL
}

func sortedColumns(object map[string]interface{}) []string {
	cols := []string{}
	for k := range object {
		cols = append(cols, k)
	}
	sort.Strings(cols)
	return cols
}
