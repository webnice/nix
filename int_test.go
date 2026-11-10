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
	intJSON              = []byte(`9223372036854775807`)
	intStringJSON        = []byte(`"9223372036854775807"`)
	intBlankJSON         = []byte(`""`)
	intMaxValueValidJSON = []byte(`{"Int":9223372036854775807,"Valid":true}`)
	intNullInvalidGob    = `2b7f0301010a496e745772617070657201ff80000102010556616c7565010400010556616c6964010200000003ff8000`
	intOkValidGob        = `2b7f0301010a496e745772617070657201ff80000102010556616c7565010400010556616c696401020000000fff8001f8fffffffffffffffe010100`
)

func isIntValid(t *testing.T, i Int, from string) {
	const (
		errInt   = "ошибка %q, int: \"%d\" ≠ \"%d\""
		errValid = "является недействительным, ожидалось действительное"
	)

	dv, _ := i.Value()
	if dv.(int64) != int64(math.MaxInt64) {
		t.Errorf(errInt, from, dv, int64(math.MaxInt64))
	}
	if !i.Valid {
		t.Error(from, errValid)
	}
}

func isIntNull(t *testing.T, i Int, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewInt(t *testing.T) {
	v1 := NewInt()
	isIntNull(t, v1, "NewInt()")
}

func TestNewIntValue(t *testing.T) {
	v1 := NewIntValue(int(math.MaxInt64))
	isIntValid(t, v1, "NewIntValue()")
}

func TestNewIntPointerValue(t *testing.T) {
	var bv = int(math.MaxInt64)

	v1 := NewIntPointerValue(&bv)
	isIntValid(t, v1, "NewIntPointerValue()")

	v2 := NewIntPointerValue(nil)
	isIntNull(t, v2, "NewIntPointerValue()")
}

func TestIntSetValid(t *testing.T) {
	v1 := NewInt()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(int(math.MaxInt64))
	isIntValid(t, v1, "SetValid()")
}

func TestIntReset(t *testing.T) {
	v1 := NewInt()
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

func TestIntNullIfDefault(t *testing.T) {
	v1 := NewIntValue(int(math.MaxInt64))
	isIntValid(t, v1, "NewIntValue()")
	v1.NullIfDefault()
	isIntValid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isIntNull(t, v1, "NullIfDefault()")
}

func TestIntMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewInt()
	buf = v1.MustValue()
	if _, ok := buf.(int); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(int(math.MaxInt64))
	bf2 := v1.MustValue()
	if bf2 != math.MaxInt64 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestIntPointer(t *testing.T) {
	v1 := NewIntPointerValue(nil)
	isIntNull(t, v1, "NewIntPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewIntValue(int(math.MaxInt64))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxInt64 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewIntPointerValue(pb)
	isIntValid(t, v3, "NewIntPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestIntScan(t *testing.T) {
	v1 := NewInt()
	errorPanic(v1.Scan(int(math.MaxInt64)))
	isIntValid(t, v1, "Scan()")

	v2 := NewInt()
	errorPanic(v2.Scan("9223372036854775807"))
	isIntValid(t, v2, "Scan()")

	v3 := NewInt()
	errorPanic(v3.Scan(nil))
	isIntNull(t, v3, "Scan()")

	v4 := NewInt()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v5 := NewInt()
	errorPanic(v5.Scan(int64(math.MaxInt64)))
	isIntValid(t, v5, "Scan()")
}

func TestIntValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item int64

	v1 := NewIntValue(int(math.MaxInt64))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(int64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "int64")
	}
	if item != int64(math.MaxInt64) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewInt()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestIntUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewInt()
	err = json.Unmarshal(intJSON, &v1)
	errorPanic(err)
	isIntValid(t, v1, "UnmarshalJSON()")

	v2 := NewInt()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isIntNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewInt()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewInt()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewInt()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewInt()
	err = json.Unmarshal(intStringJSON, &v6)
	errorPanic(err)
	isIntValid(t, v6, "UnmarshalJSON()")

	v7 := NewInt()
	err = json.Unmarshal(intBlankJSON, &v7)
	errorPanic(err)
	isIntNull(t, v7, "UnmarshalJSON()")

	v8 := NewInt()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isIntNull(t, v8, "UnmarshalJSON()")
}

func TestIntMarshalJSON(t *testing.T) {
	v1 := NewIntValue(int(math.MaxInt64))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(intJSON), "non-empty json marshal")

	v2 := NewInt()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestIntUnmarshalText(t *testing.T) {
	var err error

	v1 := NewInt()
	err = v1.UnmarshalText(intJSON)
	errorPanic(err)
	isIntValid(t, v1, "UnmarshalText()")

	v2 := NewInt()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Int != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewInt()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isIntNull(t, v3, "UnmarshalText()")
}

func TestIntMarshalText(t *testing.T) {
	v1 := NewIntValue(int(math.MaxInt64))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(intJSON), "Non-empty text marshal")

	v2 := NewInt()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestIntUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Int{}
	btf, err = hex.DecodeString(intNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isIntNull(t, v1, "UnmarshalBinary() invalid")

	v2 := Int{}
	btf, err = hex.DecodeString(intOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isIntValid(t, v2, "UnmarshalBinary() ok")

	v3 := Int{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestIntMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Int

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewInt()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isIntNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewIntValue(int(math.MaxInt64))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isIntValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
