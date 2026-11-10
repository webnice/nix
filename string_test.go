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

func isStringValid(t *testing.T, s String, from string) {
	const errString = "ошибка %q, string: \"%s\" ≠ \"%s\""

	dv, _ := s.Value()
	if dv.(string) != stringTestBody {
		t.Errorf(errString, from, dv, stringTestBody)
	}
	if !s.Valid {
		t.Error(from, errValid)
	}
}

func isStringNull(t *testing.T, s String, from string) {
	if s.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewString(t *testing.T) {
	v1 := NewString()
	isStringNull(t, v1, "NewString()")
}

func TestNewStringValue(t *testing.T) {
	v1 := NewStringValue(stringTestBody)
	isStringValid(t, v1, "NewStringValue()")
}

func TestNewStringPointerValue(t *testing.T) {
	var bv = stringTestBody

	v1 := NewStringPointerValue(&bv)
	isStringValid(t, v1, "NewStringPointerValue()")

	v2 := NewStringPointerValue(nil)
	isStringNull(t, v2, "NewStringPointerValue()")
}

func TestStringSetValid(t *testing.T) {
	v1 := NewString()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(stringTestBody)
	isStringValid(t, v1, "SetValid()")
}

func TestStringReset(t *testing.T) {
	v1 := NewString()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	v1.SetValid(``)
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestStringNullIfDefault(t *testing.T) {
	v1 := NewStringValue(stringTestBody)
	isStringValid(t, v1, "NewStringValue()")
	v1.NullIfDefault()
	isStringValid(t, v1, "NullIfDefault()")

	v1.SetValid(``)
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isStringNull(t, v1, "NullIfDefault()")
}

func TestStringMustValue(t *testing.T) {
	var buf interface{}

	v1 := NewString()
	buf = v1.MustValue()
	if _, ok := buf.(string); !ok {
		t.Error("MustValue()", errNil)
	}
	v1.SetValid(stringTestBody)
	bf2 := v1.MustValue()
	if bf2 != stringTestBody {
		t.Error("MustValue()", errWrong)
	}
}

func TestStringPointer(t *testing.T) {
	v1 := NewStringPointerValue(nil)
	isStringNull(t, v1, "NewStringPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewStringValue(stringTestBody)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != stringTestBody {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewStringPointerValue(pb)
	isStringValid(t, v3, "NewStringPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestStringScan(t *testing.T) {
	v1 := NewString()
	errorPanic(v1.Scan(stringTestBody))
	isStringValid(t, v1, "Scan()")

	v2 := NewString()
	errorPanic(v2.Scan(nil))
	isStringNull(t, v2, "Scan()")
}

func TestStringValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	v1 := NewStringValue(stringTestBody)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != stringTestBody {
		t.Error("Value() value", errWrong)
	}

	v2 := NewString()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestStringUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewString()
	err = json.Unmarshal(stringJSON, &v1)
	errorPanic(err)
	isStringValid(t, v1, "UnmarshalJSON()")

	v2 := NewString()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isStringNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewString()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewString()
	err = v4.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v5 := NewString()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewString()
	err = json.Unmarshal(int64JSON, &v6)
	if err == nil {
		panic(errErrorNotNil)
	}
	isStringNull(t, v6, "UnmarshalJSON()")
}

func TestStringMarshalJSON(t *testing.T) {
	v1 := NewStringValue(stringTestBody)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, `"3LbOVMltCjj1Mg6sSRYLzS5j64DDNEVax29ypIGxwEx9mnbFnT9FY0sZqP11"`, "non-empty json marshal")

	v2 := NewString()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

func TestStringUnmarshalText(t *testing.T) {
	var err error

	v1 := NewString()
	err = v1.UnmarshalText([]byte(stringTestBody))
	errorPanic(err)
	isStringValid(t, v1, "UnmarshalText()")

	v2 := NewString()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.String != "" || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewString()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isStringNull(t, v3, "UnmarshalText()")
}

func TestStringMarshalText(t *testing.T) {
	v1 := NewStringValue(stringTestBody)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, stringTestBody, "Non-empty text marshal")

	v2 := NewString()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestStringUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := String{}
	btf, err = hex.DecodeString(stringNullInvalidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v1))
	isStringNull(t, v1, "UnmarshalBinary() invalid")

	v2 := String{}
	btf, err = hex.DecodeString(stringOkValidGob)
	errorPanic(err)
	errorPanic(gob.NewDecoder(bytes.NewBuffer(btf)).Decode(&v2))
	isStringValid(t, v2, "UnmarshalBinary() ok")
}

//goland:noinspection DuplicatedCode
func TestStringMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item String

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewString()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isStringNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewStringValue(stringTestBody)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isStringValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}
