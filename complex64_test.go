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

const complex64TestString = `(1.5+2.5i)`

var (
	complex64JSON              = []byte(`(1.5+2.5i)`)
	complex64StringJSON        = []byte(`"(1.5+2.5i)"`)
	complex64BlankJSON         = []byte(`""`)
	complex64MaxValueValidJSON = []byte(`{"Complex64":(1.5+2.5i),"Valid":true}`)
	complex64NullInvalidGob    = `32ff9303010110436f6d706c657836345772617070657201ff94000102010556616c7565010e00010556616c6964010200000003ff9400`
	complex64OkValidGob        = `32ff9303010110436f6d706c657836345772617070657201ff94000102010556616c7565010e00010556616c696401020000000cff9401fef83ffe0440010100`
)

func isComplex64Valid(t *testing.T, c Complex64, from string) {
	const (
		errComplex64 = "ошибка %q, complex64: %v ≠ %v"
		errValid     = "является недействительным, ожидалось действительное"
	)

	if c.Complex64 != complex64(1.5+2.5i) {
		t.Errorf(errComplex64, from, c.Complex64, complex64(1.5+2.5i))
	}
	if !c.Valid {
		t.Error(from, errValid)
	}
}

func isComplex64Null(t *testing.T, i Complex64, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if i.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewComplex64(t *testing.T) {
	v1 := NewComplex64()
	isComplex64Null(t, v1, "NewComplex64()")
}

func TestNewComplex64Value(t *testing.T) {
	v1 := NewComplex64Value(complex64(1.5 + 2.5i))
	isComplex64Valid(t, v1, "NewComplex64Value()")
}

func TestNewComplex64PointerValue(t *testing.T) {
	var bv = complex64(1.5 + 2.5i)

	v1 := NewComplex64PointerValue(&bv)
	isComplex64Valid(t, v1, "NewComplex64PointerValue()")

	v2 := NewComplex64PointerValue(nil)
	isComplex64Null(t, v2, "NewComplex64PointerValue()")
}

func TestComplex64SetValid(t *testing.T) {
	v1 := NewComplex64()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid(complex64(1.5 + 2.5i))
	isComplex64Valid(t, v1, "SetValid()")
}

func TestComplex64Reset(t *testing.T) {
	v1 := NewComplex64()
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

func TestComplex64NullIfDefault(t *testing.T) {
	v1 := NewComplex64Value(complex64(1.5 + 2.5i))
	isComplex64Valid(t, v1, "NewComplex64Value()")
	v1.NullIfDefault()
	isComplex64Valid(t, v1, "NullIfDefault()")

	v1.Set(0)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isComplex64Null(t, v1, "NullIfDefault()")
}

func TestComplex64MustValue(t *testing.T) {
	var buf any

	v1 := NewComplex64()
	buf = v1.MustValue()
	if _, ok := buf.(complex64); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(complex64(1.5 + 2.5i))
	bf2 := v1.MustValue()
	if bf2 != complex64(1.5+2.5i) {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex64Pointer(t *testing.T) {
	v1 := NewComplex64PointerValue(nil)
	isComplex64Null(t, v1, "NewComplex64PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewComplex64Value(complex64(1.5 + 2.5i))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != complex64(1.5+2.5i) {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewComplex64PointerValue(pb)
	isComplex64Valid(t, v3, "NewComplex64PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestComplex64Scan(t *testing.T) {
	v1 := NewComplex64()
	errorPanic(v1.Scan(complex64(1.5 + 2.5i)))
	isComplex64Valid(t, v1, "Scan()")

	v2 := NewComplex64()
	errorPanic(v2.Scan("(1.5+2.5i)"))
	isComplex64Valid(t, v2, "Scan()")

	v3 := NewComplex64()
	errorPanic(v3.Scan(nil))
	isComplex64Null(t, v3, "Scan()")

	v4 := NewComplex64()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v5 := NewComplex64()
	errorPanic(v5.Scan(complex128(1.5 + 2.5i)))
	isComplex64Valid(t, v5, "Scan()")
}

func TestComplex64Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	v1 := NewComplex64Value(complex64(1.5 + 2.5i))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != complex64TestString {
		t.Error("Value() value", errWrong)
	}

	v2 := NewComplex64()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex64UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewComplex64()
	err = json.Unmarshal(complex64StringJSON, &v1)
	errorPanic(err)
	isComplex64Valid(t, v1, "UnmarshalJSON()")

	v2 := NewComplex64()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isComplex64Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewComplex64()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewComplex64()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewComplex64()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewComplex64()
	err = json.Unmarshal(complex64StringJSON, &v6)
	errorPanic(err)
	isComplex64Valid(t, v6, "UnmarshalJSON()")

	v7 := NewComplex64()
	err = json.Unmarshal(complex64BlankJSON, &v7)
	errorPanic(err)
	isComplex64Null(t, v7, "UnmarshalJSON()")

	v8 := NewComplex64()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isComplex64Null(t, v8, "UnmarshalJSON()")
}

func TestComplex64MarshalJSON(t *testing.T) {
	v1 := NewComplex64Value(complex64(1.5 + 2.5i))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(complex64StringJSON), "non-empty json marshal")

	v2 := NewComplex64()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection DuplicatedCode
func TestComplex64UnmarshalText(t *testing.T) {
	var err error

	v1 := NewComplex64()
	err = v1.UnmarshalText(complex64JSON)
	errorPanic(err)
	isComplex64Valid(t, v1, "UnmarshalText()")

	v2 := NewComplex64()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Complex64 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewComplex64()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isComplex64Null(t, v3, "UnmarshalText()")
}

func TestComplex64MarshalText(t *testing.T) {
	v1 := NewComplex64Value(complex64(1.5 + 2.5i))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, complex64TestString, "Non-empty text marshal")

	v2 := NewComplex64()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestComplex64UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Complex64{}
	btf, err = hex.DecodeString(complex64NullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isComplex64Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Complex64{}
	btf, err = hex.DecodeString(complex64OkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isComplex64Valid(t, v2, "UnmarshalBinary() ok")

	v3 := Complex64{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestComplex64MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Complex64

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewComplex64()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isComplex64Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewComplex64Value(complex64(1.5 + 2.5i))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isComplex64Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestComplex64SetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewComplex64()
	v1.SetValid(0)
	isComplex64Null(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewComplex64()
	v2.Set(0)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
