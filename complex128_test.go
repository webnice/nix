package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"
)

const complex128TestString = `(1.5+2.5i)`

var (
	complex128JSON              = []byte(`(1.5+2.5i)`)
	complex128StringJSON        = []byte(`"(1.5+2.5i)"`)
	complex128BlankJSON         = []byte(`""`)
	complex128MaxValueValidJSON = []byte(`{"Complex128":(1.5+2.5i),"Valid":true}`)
	complex128NullInvalidGob    = `33ff9503010111436f6d706c65783132385772617070657201ff96000102010556616c7565010e00010556616c6964010200000003ff9600`
	complex128OkValidGob        = `33ff9503010111436f6d706c65783132385772617070657201ff96000102010556616c7565010e00010556616c696401020000000cff9601fef83ffe0440010100`
)

func isComplex128Valid(t *testing.T, c Complex128, from string) {
	const (
		errComplex128 = "ошибка %q, complex128: %v ≠ %v"
		errValid      = "является недействительным, ожидалось действительное"
	)

	if c.Complex128 != complex128(1.5+2.5i) {
		t.Errorf(errComplex128, from, c.Complex128, complex128(1.5+2.5i))
	}
	if !c.Valid {
		t.Error(from, errValid)
	}
}

func isComplex128Null(t *testing.T, i Complex128, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewComplex128(t *testing.T) {
	v1 := NewComplex128()
	isComplex128Null(t, v1, "NewComplex128()")
}

func TestNewComplex128Value(t *testing.T) {
	v1 := NewComplex128Value(complex128(1.5 + 2.5i))
	isComplex128Valid(t, v1, "NewComplex128Value()")
}

func TestNewComplex128PointerValue(t *testing.T) {
	var bv = complex128(1.5 + 2.5i)

	v1 := NewComplex128PointerValue(&bv)
	isComplex128Valid(t, v1, "NewComplex128PointerValue()")

	v2 := NewComplex128PointerValue(nil)
	isComplex128Null(t, v2, "NewComplex128PointerValue()")
}

func TestComplex128SetValid(t *testing.T) {
	v1 := NewComplex128()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(complex128(1.5 + 2.5i))
	isComplex128Valid(t, v1, "SetValid()")
}

func TestComplex128Reset(t *testing.T) {
	v1 := NewComplex128()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestComplex128NullIfDefault(t *testing.T) {
	v1 := NewComplex128Value(complex128(1.5 + 2.5i))
	isComplex128Valid(t, v1, "NewComplex128Value()")
	v1.NullIfDefault()
	isComplex128Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isComplex128Null(t, v1, "NullIfDefault()")
}

func TestComplex128MustValue(t *testing.T) {
	var buf any

	v1 := NewComplex128()
	buf = v1.MustValue()
	if _, ok := buf.(complex128); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(complex128(1.5 + 2.5i))
	bf2 := v1.MustValue()
	if bf2 != complex128(1.5+2.5i) {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex128Pointer(t *testing.T) {
	v1 := NewComplex128PointerValue(nil)
	isComplex128Null(t, v1, "NewComplex128PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewComplex128Value(complex128(1.5 + 2.5i))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != complex128(1.5+2.5i) {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewComplex128PointerValue(pb)
	isComplex128Valid(t, v3, "NewComplex128PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestComplex128Scan(t *testing.T) {
	v1 := NewComplex128()
	errorPanic(v1.Scan(complex128(1.5 + 2.5i)))
	isComplex128Valid(t, v1, "Scan()")

	v2 := NewComplex128()
	errorPanic(v2.Scan("(1.5+2.5i)"))
	isComplex128Valid(t, v2, "Scan()")

	v3 := NewComplex128()
	errorPanic(v3.Scan(nil))
	isComplex128Null(t, v3, "Scan()")

	v4 := NewComplex128()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestComplex128Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	v1 := NewComplex128Value(complex128(1.5 + 2.5i))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != complex128TestString {
		t.Error("Value() value", errWrong)
	}

	v2 := NewComplex128()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex128UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewComplex128()
	err = json.Unmarshal(complex128StringJSON, &v1)
	errorPanic(err)
	isComplex128Valid(t, v1, "UnmarshalJSON()")

	v2 := NewComplex128()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isComplex128Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewComplex128()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewComplex128()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewComplex128()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewComplex128()
	err = json.Unmarshal(complex128StringJSON, &v6)
	errorPanic(err)
	isComplex128Valid(t, v6, "UnmarshalJSON()")

	v7 := NewComplex128()
	err = json.Unmarshal(complex128BlankJSON, &v7)
	errorPanic(err)
	isComplex128Null(t, v7, "UnmarshalJSON()")

	v8 := NewComplex128()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isComplex128Null(t, v8, "UnmarshalJSON()")
}

func TestComplex128MarshalJSON(t *testing.T) {
	v1 := NewComplex128Value(complex128(1.5 + 2.5i))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(complex128StringJSON), "non-empty json marshal")

	v2 := NewComplex128()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestComplex128UnmarshalText(t *testing.T) {
	var err error

	v1 := NewComplex128()
	err = v1.UnmarshalText(complex128JSON)
	errorPanic(err)
	isComplex128Valid(t, v1, "UnmarshalText()")

	v2 := NewComplex128()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Complex128 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewComplex128()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isComplex128Null(t, v3, "UnmarshalText()")
}

func TestComplex128MarshalText(t *testing.T) {
	v1 := NewComplex128Value(complex128(1.5 + 2.5i))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, complex128TestString, "Non-empty text marshal")

	v2 := NewComplex128()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestComplex128UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Complex128{}
	btf, err = hex.DecodeString(complex128NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isComplex128Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Complex128{}
	btf, err = hex.DecodeString(complex128OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isComplex128Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Complex128{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex128MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Complex128

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewComplex128()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isComplex128Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewComplex128Value(complex128(1.5 + 2.5i))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isComplex128Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestComplex128SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewComplex128()
	v1.SetValid(0)
	isComplex128Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewComplex128()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
