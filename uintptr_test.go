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
	uintptrJSON              = []byte(`18446744073709551615`)
	uintptrStringJSON        = []byte(`"18446744073709551615"`)
	uintptrBlankJSON         = []byte(`""`)
	uintptrMaxValueValidJSON = []byte(`{"Uintptr":18446744073709551615,"Valid":true}`)
	uintptrNullInvalidGob    = `30ff8f0301010e55696e747074725772617070657201ff90000102010556616c7565010600010556616c6964010200000003ff9000`
	uintptrOkValidGob        = `30ff8f0301010e55696e747074725772617070657201ff90000102010556616c7565010600010556616c696401020000000fff9001f8ffffffffffffffff010100`
)

func isUintptrValid(t *testing.T, u Uintptr, from string) {
	const (
		errUintptr = "ошибка %q, uintptr: \"%d\" ≠ \"%d\""
		errValid   = "является недействительным, ожидалось действительное"
	)

	dv := interface{}(u.Uintptr)
	if dv.(uintptr) != uintptr(math.MaxUint64) {
		t.Errorf(errUintptr, from, dv, uintptr(math.MaxUint64))
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUintptrNull(t *testing.T, i Uintptr, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUintptr(t *testing.T) {
	v1 := NewUintptr()
	isUintptrNull(t, v1, "NewUintptr()")
}

func TestNewUintptrValue(t *testing.T) {
	v1 := NewUintptrValue(uintptr(math.MaxUint64))
	isUintptrValid(t, v1, "NewUintptrValue()")
}

func TestNewUintptrPointerValue(t *testing.T) {
	var bv = uintptr(math.MaxUint64)

	v1 := NewUintptrPointerValue(&bv)
	isUintptrValid(t, v1, "NewUintptrPointerValue()")

	v2 := NewUintptrPointerValue(nil)
	isUintptrNull(t, v2, "NewUintptrPointerValue()")
}

func TestUintptrSetValid(t *testing.T) {
	v1 := NewUintptr()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(uintptr(math.MaxUint64))
	isUintptrValid(t, v1, "SetValid()")
}

func TestUintptrReset(t *testing.T) {
	v1 := NewUintptr()
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

func TestUintptrNullIfDefault(t *testing.T) {
	v1 := NewUintptrValue(uintptr(math.MaxUint64))
	isUintptrValid(t, v1, "NewUintptrValue()")
	v1.NullIfDefault()
	isUintptrValid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUintptrNull(t, v1, "NullIfDefault()")
}

func TestUintptrMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewUintptr()
	buf = v1.MustValue()
	if _, ok := buf.(uintptr); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(uintptr(math.MaxUint64))
	bf2 := v1.MustValue()
	if bf2 != math.MaxUint64 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUintptrPointer(t *testing.T) {
	v1 := NewUintptrPointerValue(nil)
	isUintptrNull(t, v1, "NewUintptrPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewUintptrValue(uintptr(math.MaxUint64))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxUint64 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUintptrPointerValue(pb)
	isUintptrValid(t, v3, "NewUintptrPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUintptrScan(t *testing.T) {
	v1 := NewUintptr()
	errorPanic(v1.Scan(uintptr(math.MaxUint64)))
	isUintptrValid(t, v1, "Scan()")

	v2 := NewUintptr()
	errorPanic(v2.Scan("18446744073709551615"))
	isUintptrValid(t, v2, "Scan()")

	v3 := NewUintptr()
	errorPanic(v3.Scan(nil))
	isUintptrNull(t, v3, "Scan()")

	v4 := NewUintptr()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestUintptrValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewUintptrValue(uintptr(math.MaxUint64))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if !bytes.Equal(item, uintptrJSON) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUintptr()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUintptrUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUintptr()
	err = json.Unmarshal(uintptrJSON, &v1)
	errorPanic(err)
	isUintptrValid(t, v1, "UnmarshalJSON()")

	v2 := NewUintptr()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUintptrNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewUintptr()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUintptr()
	err = v4.UnmarshalJSON(uintptrMaxValueValidJSON)
	errorPanic(err)
	isUintptrValid(t, v4, "UnmarshalJSON()")

	v5 := NewUintptr()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUintptr()
	err = json.Unmarshal(uintptrStringJSON, &v6)
	errorPanic(err)
	isUintptrValid(t, v6, "UnmarshalJSON()")

	v7 := NewUintptr()
	err = json.Unmarshal(uintptrBlankJSON, &v7)
	errorPanic(err)
	isUintptrNull(t, v7, "UnmarshalJSON()")

	v8 := NewUintptr()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUintptrNull(t, v8, "UnmarshalJSON()")
}

func TestUintptrMarshalJSON(t *testing.T) {
	v1 := NewUintptrValue(uintptr(math.MaxUint64))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uintptrJSON), "non-empty json marshal")

	v2 := NewUintptr()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestUintptrUnmarshalText(t *testing.T) {
	var err error

	v1 := NewUintptr()
	err = v1.UnmarshalText(uintptrJSON)
	errorPanic(err)
	isUintptrValid(t, v1, "UnmarshalText()")

	v2 := NewUintptr()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Uintptr != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUintptr()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUintptrNull(t, v3, "UnmarshalText()")
}

func TestUintptrMarshalText(t *testing.T) {
	v1 := NewUintptrValue(uintptr(math.MaxUint64))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(uintptrJSON), "Non-empty text marshal")

	v2 := NewUintptr()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestUintptrUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Uintptr{}
	btf, err = hex.DecodeString(uintptrNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUintptrNull(t, v1, "UnmarshalBinary() invalid")

	v2 := Uintptr{}
	btf, err = hex.DecodeString(uintptrOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUintptrValid(t, v2, "UnmarshalBinary() ok")

	v3 := Uintptr{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUintptrMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Uintptr

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUintptr()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUintptrNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewUintptrValue(uintptr(math.MaxUint64))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUintptrValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
