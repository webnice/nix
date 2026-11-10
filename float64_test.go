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

func isFloat64Valid(t *testing.T, f Float64, from string) {
	const errFloat64 = "ошибка %q, float64: \"%f\" ≠ \"%f\""

	dv, _ := f.Value()
	if dv.(float64) != math.MaxFloat64 {
		t.Errorf(errFloat64, from, dv, math.MaxFloat64)
	}
	if !f.Valid {
		t.Error(from, errValid)
	}
}

func isFloat64Null(t *testing.T, f Float64, from string) {
	if f.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewFloat64(t *testing.T) {
	v1 := NewFloat64()
	isFloat64Null(t, v1, "NewFloat64()")
}

func TestNewFloat64Value(t *testing.T) {
	v1 := NewFloat64Value(math.MaxFloat64)
	isFloat64Valid(t, v1, "NewFloat64Value()")
}

func TestNewFloat64PointerValue(t *testing.T) {
	var bv = math.MaxFloat64

	v1 := NewFloat64PointerValue(&bv)
	isFloat64Valid(t, v1, "NewFloat64PointerValue()")

	v2 := NewFloat64PointerValue(nil)
	isFloat64Null(t, v2, "NewFloat64PointerValue()")
}

func TestFloat64SetValid(t *testing.T) {
	v1 := NewFloat64()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(math.MaxFloat64)
	isFloat64Valid(t, v1, "SetValid()")
}

func TestFloat64Reset(t *testing.T) {
	v1 := NewFloat64()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(0)
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestFloat64NullIfDefault(t *testing.T) {
	v1 := NewFloat64Value(math.MaxFloat64)
	isFloat64Valid(t, v1, "NewFloat64Value()")
	v1.NullIfDefault()
	isFloat64Valid(t, v1, "NullIfDefault()")

	v1.SetValid(0)
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isFloat64Null(t, v1, "NullIfDefault()")
}

func TestFloat64MustValue(t *testing.T) {
	var buf interface{}

	v1 := NewFloat64()
	buf = v1.MustValue()
	if _, ok := buf.(float64); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(math.MaxFloat64)
	bf2 := v1.MustValue()
	if bf2 != math.MaxFloat64 {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat64Pointer(t *testing.T) {
	v1 := NewFloat64PointerValue(nil)
	isFloat64Null(t, v1, "NewFloat64PointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewFloat64Value(math.MaxFloat64)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != math.MaxFloat64 {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewFloat64PointerValue(pb)
	isFloat64Valid(t, v3, "NewFloat64PointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestFloat64Scan(t *testing.T) {
	v1 := NewFloat64()
	errorPanic(v1.Scan(math.MaxFloat64))
	isFloat64Valid(t, v1, "Scan()")

	v2 := NewFloat64()
	errorPanic(v2.Scan("179769313486231570814527423731704356798070567525844996598917476803157260780028538760589558632766878171540458953514382464234321326889464182768467546703537516986049910576551282076245490090389328944075868508455133942304583236903222948165808559332123348274797826204144723168738177180919299881250404026184124858368.000000"))
	isFloat64Valid(t, v2, "Scan()")

	v3 := NewFloat64()
	errorPanic(v3.Scan(nil))
	isFloat64Null(t, v3, "Scan()")

	v4 := NewFloat64()
	err := v4.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestFloat64Value(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item float64

	v1 := NewFloat64Value(math.MaxFloat64)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(float64); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "float64")
	}
	if item != math.MaxFloat64 {
		t.Error("Value() value", errWrong)
	}

	v2 := NewFloat64()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat64UnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewFloat64()
	err = json.Unmarshal(float64JSON, &v1)
	errorPanic(err)
	isFloat64Valid(t, v1, "UnmarshalJSON()")

	v2 := NewFloat64()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isFloat64Null(t, v2, "UnmarshalJSON(null)")

	v3 := NewFloat64()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewFloat64()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewFloat64()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewFloat64()
	err = json.Unmarshal(float64StringJSON, &v6)
	errorPanic(err)
	isFloat64Valid(t, v6, "UnmarshalJSON()")

	v7 := NewFloat64()
	err = json.Unmarshal(float64BlankJSON, &v7)
	errorPanic(err)
	isFloat64Null(t, v7, "UnmarshalJSON()")

	v8 := NewFloat64()
	err = json.Unmarshal(boolFalseJSON, &v8)
	if err == nil {
		panic(errErrorNotNil)
	}
	isFloat64Null(t, v8, "UnmarshalJSON()")
}

func TestFloat64MarshalJSON(t *testing.T) {
	v1 := NewFloat64Value(math.MaxFloat64)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(float64JSON), "non-empty json marshal")

	v2 := NewFloat64()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")

	v3 := NewFloat64Value(math.NaN())
	data, err = v3.MarshalJSON()
	if err == nil || len(data) > 0 {
		panic(errErrorNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestFloat64UnmarshalText(t *testing.T) {
	var err error

	v1 := NewFloat64()
	err = v1.UnmarshalText(float64JSON)
	errorPanic(err)
	isFloat64Valid(t, v1, "UnmarshalText()")

	v2 := NewFloat64()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.Float64 != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewFloat64()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isFloat64Null(t, v3, "UnmarshalText()")
}

func TestFloat64MarshalText(t *testing.T) {
	v1 := NewFloat64Value(math.MaxFloat64)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, string(float64JSON), "Non-empty text marshal")

	v2 := NewFloat64()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestFloat64UnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := Float64{}
	btf, err = hex.DecodeString(floatNullInvalidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v1))
	isFloat64Null(t, v1, "UnmarshalBinary() invalid")

	v2 := Float64{}
	btf, err = hex.DecodeString(floatOkValidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v2))
	isFloat64Valid(t, v2, "UnmarshalBinary() ok")
}

//goland:noinspection DuplicatedCode
func TestFloat64MarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Float64

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewFloat64()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isFloat64Null(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewFloat64Value(math.MaxFloat64)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isFloat64Valid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
