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

func isTrueBool(t *testing.T, b Bool, from string) {
	const errBool = "ошибка %q, bool: \"%t\" ≠ \"%t\""

	if !b.Bool {
		t.Errorf(errBool, from, b.Bool, true)
	}
	if !b.Valid {
		t.Error(from, errValid)
	}
}

func isFalseBool(t *testing.T, b Bool, from string) {
	const errBool = "ошибка %q, bool: \"%t\" ≠ \"%t\""

	if b.Bool {
		t.Errorf(errBool, from, b.Bool, false)
	}
	if !b.Valid {
		t.Error(from, errValid)
	}
}

func isNullBool(t *testing.T, b Bool, from string) {
	if b.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewBool(t *testing.T) {
	b := NewBool()
	isNullBool(t, b, "NewBool()")
}

func TestNewBoolValue(t *testing.T) {
	fb := NewBoolValue(false)
	isFalseBool(t, fb, "NewBoolValue(false)")
	tb := NewBoolValue(true)
	isTrueBool(t, tb, "NewBoolValue(true)")
}

func TestNewBoolPointerValue(t *testing.T) {
	var bv bool

	fb := NewBoolPointerValue(&bv)
	isFalseBool(t, fb, "NewBoolPointerValue(*false)")

	bv = true
	tb := NewBoolPointerValue(&bv)
	isTrueBool(t, tb, "NewBoolPointerValue(*true)")

	nb := NewBoolPointerValue(nil)
	isNullBool(t, nb, "NewBoolPointerValue(nil)")
}

func TestBoolSetValid(t *testing.T) {
	tb := NewBool()
	tb.Set(false)
	isFalseBool(t, tb, "SetValid(false)")

	fb := NewBool()
	fb.SetValid(true)
	isTrueBool(t, fb, "SetValid(true)")
}

func TestBoolReset(t *testing.T) {
	ib := NewBool()
	isNullBool(t, ib, "NewBool()")

	ib.Set(false)
	isFalseBool(t, ib, "SetValid(false)")
	if !ib.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}

	ib.Reset()
	isNullBool(t, ib, "Invalidate()")
	if ib.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestBoolNullIfDefault(t *testing.T) {
	v1 := NewBoolValue(true)
	isTrueBool(t, v1, "NewBoolValue(true)")
	v1.NullIfDefault()
	isTrueBool(t, v1, "NullIfDefault()")

	v1.Set(false)
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isNullBool(t, v1, "NullIfDefault()")
}

func TestBoolMustValue(t *testing.T) {
	mb := NewBool()
	isNullBool(t, mb, "NewBool()")
	if mb.MustValue() {
		t.Error("MustValue()", errValidTrue)
	}

	mb.Set(false)
	if mb.MustValue() {
		t.Error("MustValue()", errValidTrue)
	}

	mb.SetValid(true)
	if !mb.MustValue() {
		t.Error("MustValue()", errValidFalse)
	}
}

func TestBoolPointer(t *testing.T) {
	nb := NewBoolPointerValue(nil)
	isNullBool(t, nb, "NewBoolPointerValue(nil)")
	if nb.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	fb := NewBoolValue(false)
	pb := fb.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb {
		t.Error("Pointer() value", errValidTrue)
	}

	tb := NewBoolValue(true)
	pb = tb.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	!*pb {
		t.Error("Pointer() value", errValidFalse)
	}
}

func TestBoolScan(t *testing.T) {
	fb := NewBoolValue(false)
	errorPanic(fb.Scan(true))
	isTrueBool(t, fb, "Scan(true)")

	nb := NewBoolValue(false)
	errorPanic(nb.Scan(nil))
	isNullBool(t, nb, "Scan(nil)")

	eb := NewBoolValue(true)
	if err := eb.Scan(-1); err == nil {
		t.Error("Scan()", errReturnsNil)
	}
	isNullBool(t, eb, "Scan(nil)")
}

func TestBoolValue(t *testing.T) {
	var err error
	var dv driver.Value
	var item, ok bool

	b := NewBoolValue(true)
	dv, err = b.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(bool); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "bool")
	}
	if !item {
		t.Error("Value() value", errValidFalse)
	}

	nb := NewBool()
	dv, err = nb.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

func TestBoolUnmarshalJSON(t *testing.T) {
	var err error
	var b, fb, nb, badType, valid, invalid Bool

	b = NewBool()
	err = json.Unmarshal(boolTrueJSON, &b)
	errorPanic(err)
	isTrueBool(t, b, "UnmarshalJSON(true)")

	fb = NewBool()
	err = json.Unmarshal(boolFalseJSON, &fb)
	errorPanic(err)
	isFalseBool(t, fb, "UnmarshalJSON(false)")

	nb = NewBool()
	err = json.Unmarshal(boolNullJSON, &nb)
	errorPanic(err)
	isNullBool(t, nb, "UnmarshalJSON(null)")

	valid = NewBool()
	err = json.Unmarshal(boolFalseValidJSON, &valid)
	errorPanic(err)

	invalid = NewBool()
	err = invalid.UnmarshalJSON(invalidJSON)
	if //goland:noinspection ALL
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	badType = NewBool()
	err = badType.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	badType = NewBool()
	err = badType.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}
}

func TestBoolMarshalJSON(t *testing.T) {
	b := NewBoolValue(true)
	data, err := b.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "true", "non-empty json marshal")

	empty := NewBoolValue(false)
	data, err = empty.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "false", "empty json marshal")

	null := NewBool()
	data, err = null.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")

}

func TestBoolUnmarshalText(t *testing.T) {
	var err error
	var b, zero, blank, null, invalid Bool

	b = NewBool()
	err = b.UnmarshalText([]byte("true"))
	errorPanic(err)
	isTrueBool(t, b, "UnmarshalText(`true`) bool")

	zero = NewBool()
	err = zero.UnmarshalText([]byte("false"))
	errorPanic(err)
	isFalseBool(t, zero, "UnmarshalText(`false`) bool")

	blank = NewBool()
	err = blank.UnmarshalText([]byte(""))
	errorPanic(err)
	isNullBool(t, blank, "UnmarshalText(``) bool")

	null = NewBool()
	err = null.UnmarshalText([]byte("null"))
	errorPanic(err)
	isNullBool(t, null, "UnmarshalText(`null`) bool")

	invalid = NewBool()
	err = invalid.UnmarshalText([]byte(":D"))
	if err == nil {
		panic(errErrorNotNil)
	}
	isNullBool(t, invalid, "Invalid json")
}

func TestBoolMarshalText(t *testing.T) {
	b := NewBoolValue(true)
	data, err := b.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "true", "Non-empty text marshal")

	zero := NewBoolValue(false)
	data, err = zero.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "false", "Zero text marshal")

	null := NewBool()
	data, err = null.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

func TestBoolUnmarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var dec *gob.Decoder
	var btf []byte
	var trueBool, falseBool, nullBool Bool

	btf, err = hex.DecodeString(boolTrueValidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&trueBool)
	errorPanic(err)
	isTrueBool(t, trueBool, "UnmarshalBinary() -> Bool(true)")

	btf, err = hex.DecodeString(boolFalseValidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&falseBool)
	errorPanic(err)
	isFalseBool(t, falseBool, "UnmarshalBinary() -> Bool(false)")

	btf, err = hex.DecodeString(boolNullInvalidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&nullBool)
	errorPanic(err)
	isNullBool(t, nullBool, "UnmarshalBinary() -> Bool(null)")
}

//goland:noinspection DuplicatedCode
func TestBoolMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Bool

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	trueBool := NewBoolValue(true)
	err = enc.Encode(&trueBool)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isTrueBool(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	falseBool := NewBoolValue(false)
	err = enc.Encode(&falseBool)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isFalseBool(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	nullBool := NewBool()
	err = enc.Encode(&nullBool)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isNullBool(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestBoolSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewBool()
	v1.SetValid(false)
	isNullBool(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewBool()
	v2.Set(false)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
