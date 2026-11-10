package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"math"
	"reflect"
	"testing"
)

var (
	int8JSON              = []byte(`127`)
	int8StringJSON        = []byte(`"127"`)
	int8BlankJSON         = []byte(`""`)
	int8MaxValueValidJSON = []byte(`{"Int8":127,"Valid":true}`)
	int8NullInvalidGob    = `2dff810301010b496e74385772617070657201ff82000102010556616c7565010400010556616c6964010200000003ff8200`
	int8OkValidGob        = `2dff810301010b496e74385772617070657201ff82000102010556616c7565010400010556616c6964010200000008ff8201fffe010100`
)

func isInt8Valid(t *testing.T, i Int8, from string) {
	const (
		errInt8  = "ошибка %q, int8: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv, _ := i.Value()
	if dv.(int64) != int64(math.MaxInt8) {
		t.Errorf(errInt8, from, dv, int64(math.MaxInt8))
	}
	if !i.Valid {
		t.Error(from, errValid)
	}
}

func isInt8Null(t *testing.T, i Int8, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewInt8(t *testing.T) {
	v1 := NewInt8()
	isInt8Null(t, v1, "NewInt8()")
}

func TestNewInt8Value(t *testing.T) {
	v1 := NewInt8Value(int8(math.MaxInt8))
	isInt8Valid(t, v1, "NewInt8Value()")
}

func TestNewInt8PointerValue(t *testing.T) {
	var bv = int8(math.MaxInt8)

	v1 := NewInt8PointerValue(&bv)
	isInt8Valid(t, v1, "NewInt8PointerValue()")

	v2 := NewInt8PointerValue(nil)
	isInt8Null(t, v2, "NewInt8PointerValue()")
}

func TestInt8SetValid(t *testing.T) {
	v1 := NewInt8()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(int8(math.MaxInt8))
	isInt8Valid(t, v1, "SetValid()")
}

func TestInt8Reset(t *testing.T) {
	v1 := NewInt8()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestInt8NullIfDefault(t *testing.T) {
	v1 := NewInt8Value(int8(math.MaxInt8))
	isInt8Valid(t, v1, "NewInt8Value()")
	v1.NullIfDefault()
	isInt8Valid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isInt8Null(t, v1, "NullIfDefault()")
}

func TestInt8MustValue(t *testing.T) {
	var buf interface{}

	v1 := NewInt8()
	buf = v1.MustValue()
	if _, ok := buf.(int8); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(int8(math.MaxInt8))
	bf2 := v1.MustValue()
	if bf2 != math.MaxInt8 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestInt8Pointer(t *testing.T) {
	v1 := NewInt8PointerValue(nil)
	isInt8Null(t, v1, "NewInt8PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewInt8Value(int8(math.MaxInt8))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxInt8 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewInt8PointerValue(pb)
	isInt8Valid(t, v3, "NewInt8PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestInt8Scan(t *testing.T) {
	v1 := NewInt8()
	errorPanic(v1.Scan(int8(math.MaxInt8)))
	isInt8Valid(t, v1, "Scan()")

	v2 := NewInt8()
	errorPanic(v2.Scan("127"))
	isInt8Valid(t, v2, "Scan()")

	v3 := NewInt8()
	errorPanic(v3.Scan(nil))
	isInt8Null(t, v3, "Scan()")

	v4 := NewInt8()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestInt8Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item int64

	v1 := NewInt8Value(int8(math.MaxInt8))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(int64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "int64")
	}
	if item != int64(math.MaxInt8) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewInt8()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt8UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewInt8()
	err = json.Unmarshal(int8JSON, &v1)
	errorPanic(err)
	isInt8Valid(t, v1, "UnmarshalJSON()")

	v2 := NewInt8()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isInt8Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewInt8()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewInt8()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewInt8()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewInt8()
	err = json.Unmarshal(int8StringJSON, &v6)
	errorPanic(err)
	isInt8Valid(t, v6, "UnmarshalJSON()")

	v7 := NewInt8()
	err = json.Unmarshal(int8BlankJSON, &v7)
	errorPanic(err)
	isInt8Null(t, v7, "UnmarshalJSON()")

	v8 := NewInt8()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isInt8Null(t, v8, "UnmarshalJSON()")
}

func TestInt8MarshalJSON(t *testing.T) {
	v1 := NewInt8Value(int8(math.MaxInt8))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(int8JSON), "non-empty json marshal")

	v2 := NewInt8()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestInt8UnmarshalText(t *testing.T) {
	var err error

	v1 := NewInt8()
	err = v1.UnmarshalText(int8JSON)
	errorPanic(err)
	isInt8Valid(t, v1, "UnmarshalText()")

	v2 := NewInt8()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Int8 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewInt8()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isInt8Null(t, v3, "UnmarshalText()")
}

func TestInt8MarshalText(t *testing.T) {
	v1 := NewInt8Value(int8(math.MaxInt8))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(int8JSON), "Non-empty text marshal")

	v2 := NewInt8()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestInt8UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Int8{}
	btf, err = hex.DecodeString(int8NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isInt8Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Int8{}
	btf, err = hex.DecodeString(int8OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isInt8Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Int8{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestInt8MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Int8

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewInt8()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt8Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewInt8Value(int8(math.MaxInt8))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isInt8Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
