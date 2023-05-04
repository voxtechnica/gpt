package psy

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCSVTable(t *testing.T) {
	expect := assert.New(t)
	tbl, err := ReadCSVTable("testdata/table.csv")
	if expect.NoError(err) {
		expect.Equal(3, tbl.FieldCount(), "Correct FieldCount")
		expect.Equal(2, tbl.RecordCount(), "Correct RecordCount")

		// Check the field names
		expect.Equal([]string{"pid", "qid", "essay"}, tbl.FieldNames)
		expect.True(tbl.HasField("pid"))
		expect.True(tbl.HasField("qid"))
		expect.True(tbl.HasField("essay"))
		expect.False(tbl.HasField("missing"), "Missing field name")
		expect.False(tbl.HasField(""), "Empty field name")

		// Add fields
		tbl.AddField("pid")
		expect.Equal(3, tbl.FieldCount(), "Existing field count")
		tbl.AddField("score")
		expect.Equal(4, tbl.FieldCount(), "New field count")
		tbl.AddField("")
		expect.Equal(4, tbl.FieldCount(), "Skip empty field name")

		// Read records
		rec := tbl.Record("pid", "1")
		if expect.NotNil(rec) {
			expect.Equal("1", rec["pid"])
		}
		rec = tbl.Record("pid", "1000")
		expect.Nil(rec, "Missing record")
		rec = tbl.Record("missing", "1")
		expect.Nil(rec, "Missing field name")
		rec = tbl.Record("", "1")
		expect.Nil(rec, "Empty field name")

		// Random record
		rec = tbl.Random()
		if expect.NotNil(rec) {
			expect.Equal(3, len(rec))
		}
		rec = (&Table{}).Random()
		expect.Nil(rec, "Empty table")
	}
}

func TestCSVFields(t *testing.T) {
	expect := assert.New(t)
	fields, err := ReadCSVFields("testdata/table.csv", "pid", "qid")
	if expect.NoError(err) {
		expect.Equal(2, len(fields))
		expect.Equal("A", fields["1"])
		expect.Equal("B", fields["2"])
	}
}

func TestCSVFieldsEmptyPath(t *testing.T) {
	expect := assert.New(t)
	fields, err := ReadCSVFields("", "pid", "qid")
	if expect.NoError(err, "Optional Fields") {
		expect.Equal(0, len(fields))
	}
}

func TestCSVFieldsMissingKey(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVFields("testdata/table.csv", "", "qid")
	expect.Error(err, "Missing key field")
}

func TestCSVFieldsInvalidKeyField(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVFields("testdata/table.csv", "missing", "qid")
	expect.Error(err, "Invalid key field name")
}

func TestCSVFieldsMissingValue(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVFields("testdata/table.csv", "pid", "")
	expect.Error(err, "Missing value field")
}

func TestCSVFieldsInvalidValueField(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVFields("testdata/table.csv", "pid", "missing")
	expect.Error(err, "Invalid value field name")
}

func TestCSVFieldsMissingFile(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVFields("testdata/missing.csv", "pid", "qid")
	expect.Error(err, "Missing file")
}

func TestCSVField(t *testing.T) {
	expect := assert.New(t)
	value, err := ReadCSVField("testdata/table.csv", "pid=1", "qid")
	if expect.NoError(err) {
		expect.Equal("A", value)
	}
}

func TestCSVFieldEmptyPath(t *testing.T) {
	expect := assert.New(t)
	value, err := ReadCSVField("", "pid=1", "qid")
	if expect.NoError(err, "Optional Field") {
		expect.Equal("", value)
	}
}

func TestCSVFieldMissingID(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVField("testdata/table.csv", "", "qid")
	expect.Error(err, "Missing ID field")
}

func TestCSVFieldInvalidID(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVField("testdata/table.csv", "pid", "qid")
	expect.Error(err, "Invalid ID (expect key=value)")
}

func TestCSVFieldInvalidIDField(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVField("testdata/table.csv", "missing=1000", "qid")
	expect.Error(err, "Invalid ID field name")
}

func TestCSVFieldMissingRecord(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVField("testdata/table.csv", "pid=1000", "qid")
	expect.Error(err, "Missing record")
}

func TestCSVFieldMissingField(t *testing.T) {
	expect := assert.New(t)
	_, err := ReadCSVField("testdata/table.csv", "pid=1", "missing")
	expect.Error(err, "Missing field")
}

func TestRandomCSVField(t *testing.T) {
	expect := assert.New(t)
	value, err := RandomCSVField("testdata/table.csv", "qid")
	if expect.NoError(err) {
		expect.NotEmpty(value)
		expect.True(value == "A" || value == "B")
	}
}

func TestRandomCSVFieldEmptyPath(t *testing.T) {
	expect := assert.New(t)
	_, err := RandomCSVField("", "qid")
	expect.Error(err, "Required path")
}

func TestRandomCSVFieldMissingField(t *testing.T) {
	expect := assert.New(t)
	_, err := RandomCSVField("testdata/table.csv", "missing")
	expect.Error(err, "Missing field")
}
