package psy

import (
	"encoding/csv"
	"fmt"
	"io"
	"math/rand"
	"os"
	"strings"
)

// FileCache is a cache of file paths to their corresponding plain text contents.
var FileCache = map[string]string{}

// Record is a map of field names to values read from a CSV file.
type Record map[string]string

// Table is a 2-dimensional collection of data read from a CSV file.
type Table struct {
	FieldNames []string
	Records    []Record
}

// FieldCount returns the number of fields/columns in the Table.
func (t *Table) FieldCount() int {
	return len(t.FieldNames)
}

// RecordCount returns the number of records/rows in the Table.
func (t *Table) RecordCount() int {
	return len(t.Records)
}

// HasField returns true if the field/column name is valid.
func (t *Table) HasField(name string) bool {
	for _, n := range t.FieldNames {
		if n == name {
			return true
		}
	}
	return false
}

// AddField appends a new field/column name to the Table if it doesn't already exist.
func (t *Table) AddField(name string) {
	if !t.HasField(name) {
		t.FieldNames = append(t.FieldNames, name)
	}
}

// Record returns the first Record with a matching value for the given field/column name.
func (t *Table) Record(name, value string) Record {
	for _, record := range t.Records {
		if record[name] == value {
			return record
		}
	}
	return nil
}

// Random returns a random Record from the Table.
func (t *Table) Random() Record {
	if len(t.Records) == 0 {
		return nil
	}
	return t.Records[rand.Intn(len(t.Records))]
}

// WriteCSV writes a Table of Records to a CSV file.
func (t *Table) WriteCSV(path string) error {
	// Open a CSV file writer:
	f, err := os.Create(path)
	if err != nil {
		return fmt.Errorf("write csv file %s: %w", path, err)
	}
	defer f.Close()
	w := csv.NewWriter(f)

	// Write the field/column names to the first row:
	if err := w.Write(t.FieldNames); err != nil {
		return fmt.Errorf("write csv file %s header: %w", path, err)
	}

	// Write the remaining rows:
	for _, record := range t.Records {
		row := []string{}
		for _, name := range t.FieldNames {
			row = append(row, record[name])
		}
		if err := w.Write(row); err != nil {
			return fmt.Errorf("write csv file %s: %w", path, err)
		}
	}
	w.Flush()
	return nil
}

// ReadCSVTable reads a CSV file and returns a Table of Records.
func ReadCSVTable(path string) (*Table, error) {
	table := &Table{
		FieldNames: []string{},
		Records:    []Record{},
	}

	// Open a CSV file reader:
	f, err := os.Open(path)
	if err != nil {
		return table, fmt.Errorf("read csv file %s: %w", path, err)
	}
	defer f.Close()
	r := csv.NewReader(f)

	// Read the field/column names from the first row:
	names, err := r.Read()
	if err != nil {
		return table, fmt.Errorf("read csv file %s header: %w", path, err)
	}

	// Collect and validate the field/column names. They must be unique, and not blank.
	index := map[string]int{}
	for i, name := range names {
		name = strings.TrimSpace(name)
		if name == "" {
			name = fmt.Sprintf("column%d", i+1)
		}
		if _, ok := index[name]; ok {
			return table, fmt.Errorf("read csv file %s header: duplicate column name %s", path, name)
		}
		index[name] = i
		table.FieldNames = append(table.FieldNames, name)
	}

	// Read the remaining rows:
	for {
		// Read the next row
		row, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return table, fmt.Errorf("read csv file %s: %w", path, err)
		}

		// Add new columns, not found in the header:
		if len(row) > len(table.FieldNames) {
			for i := len(table.FieldNames); i < len(row); i++ {
				name := fmt.Sprintf("column%d", i+1)
				index[name] = i
				table.FieldNames = append(table.FieldNames, name)
			}
		}

		// Create a new record from the row:
		record := Record{}
		for i, name := range table.FieldNames {
			if len(row) > i {
				record[name] = row[i]
			}
		}
		table.Records = append(table.Records, record)
	}

	// Verify that the Table has at least one record:
	if len(table.Records) == 0 {
		return table, fmt.Errorf("read csv file %s: no records found", path)
	}

	return table, nil
}

// ReadCSVFields reads a map of key/value pairs from a specified CSV file.
// path is the path to the CSV file (optional). If empty, an empty map is returned.
// keyField is the field name of an key column (required).
// valuefield is the field name of a value column (required).
func ReadCSVFields(path, keyField, valueField string) (map[string]string, error) {
	var fields = map[string]string{}
	// Validate the parameters:
	if len(path) == 0 {
		return fields, nil
	}
	if len(keyField) == 0 {
		return fields, fmt.Errorf("read csv fields: key field name is required")
	}
	if len(valueField) == 0 {
		return fields, fmt.Errorf("read csv fields: value field name is required")
	}
	// Load the CSV file:
	table, err := ReadCSVTable(path)
	if err != nil {
		return fields, fmt.Errorf("read csv fields: %w", err)
	}
	// Validate the field names:
	if !table.HasField(keyField) {
		return fields, fmt.Errorf("read csv fields: key field %s not found in file %s", keyField, path)
	}
	if !table.HasField(valueField) {
		return fields, fmt.Errorf("read csv fields: value field %s not found in file %s", valueField, path)
	}
	// Read the key/value pairs:
	for _, record := range table.Records {
		key := CleanText(record[keyField])
		value := CleanText(record[valueField])
		if len(key) > 0 && len(value) > 0 {
			fields[key] = value
		}
	}
	return fields, nil
}

// ReadCSVField reads a specified field from a specified row in a CSV file.
// path is the path to the CSV file (optional). If empty, an empty string is returned.
// id is a unique row ID field, specified as a name=value pair (required).
// field is the field name to be read (required).
func ReadCSVField(path, id, field string) (string, error) {
	// Validate the parameters:
	if len(path) == 0 {
		return "", nil
	}
	if len(id) == 0 {
		return "", fmt.Errorf("read csv field: row ID (name=value) is required")
	}
	if len(field) == 0 {
		return "", fmt.Errorf("read csv field: field name is required")
	}
	// Load the CSV file:
	table, err := ReadCSVTable(path)
	if err != nil {
		return "", fmt.Errorf("read csv field: %w", err)
	}
	// Parse the row ID field name and value:
	nameValue := strings.Split(id, "=")
	if len(nameValue) != 2 {
		return "", fmt.Errorf("read csv field: ID %s is not a name=value pair", id)
	}
	if !table.HasField(nameValue[0]) {
		return "", fmt.Errorf("read csv field: ID field %s not found in file %s", nameValue[0], path)
	}
	// Read the specified record:
	record := table.Record(nameValue[0], nameValue[1])
	if record == nil {
		return "", fmt.Errorf("read csv field: ID %s not found in file %s", id, path)
	}
	// Read the specified field:
	if !table.HasField(field) {
		return "", fmt.Errorf("read csv field: field %s not found in file %s", field, path)
	}
	text := CleanText(record[field])
	if len(text) == 0 {
		return "", fmt.Errorf("read csv field: field %s for %s is empty in file %s", field, id, path)
	}
	return text, nil
}

// RandomCSVField returns a random value from a specified field in a CSV file.
func RandomCSVField(path, field string) (string, error) {
	// Validate the parameters:
	if len(path) == 0 || len(field) == 0 {
		return "", fmt.Errorf("random csv field: file path and field name are required")
	}
	// Load the CSV file:
	table, err := ReadCSVTable(path)
	if err != nil {
		return "", err
	}
	// Read the specified field from a random record:
	if !table.HasField(field) {
		return "", fmt.Errorf("random csv field: field %s not found in file %s", field, path)
	}
	text := CleanText(table.Random()[field])
	if len(text) == 0 {
		return "", fmt.Errorf("random csv field: field %s is empty in file %s", field, path)
	}
	return text, nil
}

// Read a text file into a string, using the file cache for performance.
func ReadTextFile(path string) (string, error) {
	// If the path is empty, return an empty string:
	if len(path) == 0 {
		return "", nil
	}
	// Check the cache:
	if s, ok := FileCache[path]; ok {
		return s, nil
	}
	// Open the file:
	f, err := os.Open(path)
	if err != nil {
		return "", fmt.Errorf("read text file: open file %s: %w", path, err)
	}
	defer f.Close()
	// Read the file:
	b, err := io.ReadAll(f)
	if err != nil {
		return "", fmt.Errorf("read text file: read file %s: %w", path, err)
	}
	// Cache the file contents:
	FileCache[path] = string(b)
	return FileCache[path], nil
}

// CleanText removes any extra whitespace from the provided text.
func CleanText(t string) string {
	return strings.Join(strings.Fields(t), " ")
}
