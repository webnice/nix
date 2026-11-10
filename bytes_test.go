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

const bytesTestString = `Test data 1pHuOxADZkeh8Y9WvL75`

func isBytesValid(t *testing.T, bt Bytes, from string) {
	const errBytes = "ошибка %q, bytes: %q ≠ %q"

	if !bytes.Equal(bt.Bytes.Bytes(), []byte(bytesTestString)) {
		t.Errorf(errBytes, from, bt.Bytes.String(), bytesTestString)
	}
	if !bt.Valid {
		t.Error(from, errValid)
	}
}

func isNullBytes(t *testing.T, bt Bytes, from string) {
	const errBytes = "ошибка %q, bytes: %q ≠ %q"

	if !bytes.Equal(bt.Bytes.Bytes(), []byte{}) {
		t.Errorf(errBytes, from, bt.Bytes.String(), "")
	}
	if bt.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewBytes(t *testing.T) {
	bt := NewBytes()
	isNullBytes(t, bt, "NewBytes()")
}

func TestNewBytesValue(t *testing.T) {
	bt := NewBytesValue([]byte(bytesTestString))
	isBytesValid(t, bt, "NewBytesValue()")
}

func TestNewBytesPointerValue(t *testing.T) {
	var bv = []byte(bytesTestString)

	bt := NewBytesPointerValue(&bv)
	isBytesValid(t, bt, "NewBytesPointerValue()")

	bt = NewBytesPointerValue(nil)
	isNullBytes(t, bt, "NewBytesPointerValue()")
}

func TestBytesSetValid(t *testing.T) {
	bt := NewBytes()
	if bt.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	bt.SetValid([]byte(bytesTestString))
	isBytesValid(t, bt, "SetValid()")
}

func TestBytesReset(t *testing.T) {
	bt := NewBytes()
	if bt.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
	bt.SetValid([]byte("Test"))
	if !bt.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	bt.Reset()
	if bt.Valid {
		t.Error(tplPropertyValid, errValidTrue)
	}
}

func TestBytesNullIfDefault(t *testing.T) {
	v1 := NewBytesValue([]byte(bytesTestString))
	isBytesValid(t, v1, "NewBytesValue()")
	v1.NullIfDefault()
	isBytesValid(t, v1, "NullIfDefault()")

	v1.SetValid([]byte{})
	if !v1.Valid {
		t.Error(tplPropertyValid, errValidFalse)
	}
	v1.NullIfDefault()
	isNullBytes(t, v1, "NullIfDefault()")
}

func TestBytesMustValue(t *testing.T) {
	var buf interface{}

	bt := NewBytes()
	buf = bt.MustValue()
	if _, ok := buf.([]byte); !ok {
		t.Error("MustValue()", errNil)
	}
	bt.SetValid([]byte(bytesTestString))
	bf2 := bt.MustValue()
	if !bytes.Equal(bf2, []byte(bytesTestString)) {
		t.Error("MustValue()", errWrong)
	}
}

func TestBytesPointer(t *testing.T) {
	nb := NewBytesPointerValue(nil)
	isNullBytes(t, nb, "NewBytesPointerValue(nil)")
	if nb.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	fb := NewBytesValue([]byte(bytesTestString))
	pb := fb.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	!bytes.Equal(*pb, []byte(bytesTestString)) {
		t.Error("Pointer() value", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestBytesScan(t *testing.T) {
	bt := NewBytes()
	errorPanic(bt.Scan([]byte(bytesTestString)))
	isBytesValid(t, bt, "Scan()")

	bs := NewBytes()
	errorPanic(bs.Scan(bytesTestString))
	isBytesValid(t, bt, "Scan()")

	bn := NewBytes()
	errorPanic(bn.Scan(nil))
	isNullBytes(t, bn, "Scan()")

	be := NewBytes()
	err := be.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestBytesValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	b := NewBytesValue([]byte(bytesTestString))
	dv, err = b.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "[]byte")
	}
	if !bytes.Equal(item, []byte(bytesTestString)) {
		t.Error("Value() value", errWrong)
	}

	nb := NewBytes()
	dv, err = nb.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection ALL
func TestBytesUnmarshalJSON(t *testing.T) {
	var err error

	b := NewBytes()
	err = json.Unmarshal(bytesTestValidJSON, &b)
	errorPanic(err)
	isBytesValid(t, b, "UnmarshalJSON()")

	nb := NewBytes()
	err = json.Unmarshal(bytesNullJSON, &nb)
	errorPanic(err)
	isNullBytes(t, nb, "UnmarshalJSON(null)")

	invalid := NewBytes()
	err = invalid.UnmarshalJSON(invalidJSON)
	if _, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	badType := NewBytes()
	err = badType.UnmarshalJSON(uint64MaxValueValidJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	badType = NewBytes()
	err = badType.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}
}

func TestBytesMarshalJSON(t *testing.T) {
	bv := NewBytesValue([]byte(bytesTestString))
	data, err := bv.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, `"VGVzdCBkYXRhIDFwSHVPeEFEWmtlaDhZOVd2TDc1"`, "non-empty json marshal")

	null := NewBytes()
	data, err = null.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

func TestBytesUnmarshalText(t *testing.T) {
	var err error

	bt := NewBytes()
	err = bt.UnmarshalText(bytesTestTextBase64)
	errorPanic(err)
	isBytesValid(t, bt, "UnmarshalText()")

	null := NewBytes()
	err = null.UnmarshalText([]byte("null"))
	errorPanic(err)
	isNullBytes(t, null, "UnmarshalText(`null`)")

	zero := NewBytes()
	err = zero.UnmarshalText([]byte(""))
	errorPanic(err)
	if !bytes.Equal(zero.MustValue(), []byte{}) || !zero.Valid {
		t.Error(errValueValid)
	}
}

func TestBytesMarshalText(t *testing.T) {
	bv := NewBytesValue([]byte(bytesTestString))
	data, err := bv.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "VGVzdCBkYXRhIDFwSHVPeEFEWmtlaDhZOVd2TDc1", "Non-empty text marshal")

	zero := NewBytesValue(nil)
	data, err = zero.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "", "Zero text marshal")

	null := NewBytes()
	data, err = null.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

func TestBytesUnmarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var dec *gob.Decoder
	var btf []byte
	var invalidBytes, okBytes, zeroBytes Bytes

	btf, err = hex.DecodeString(bytesNullInvalidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&invalidBytes)
	errorPanic(err)
	isNullBytes(t, invalidBytes, "UnmarshalBinary() -> Bytes(invalid)")

	btf, err = hex.DecodeString(bytesOkValidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&okBytes)
	errorPanic(err)
	isBytesValid(t, okBytes, "UnmarshalBinary() -> Bool(ok)")

	btf, err = hex.DecodeString(bytesZeroValidGob)
	errorPanic(err)
	buf = bytes.NewBuffer(btf)
	dec = gob.NewDecoder(buf)
	err = dec.Decode(&zeroBytes)
	errorPanic(err)
	if !bytes.Equal(zeroBytes.MustValue(), []byte{}) || !zeroBytes.Valid {
		t.Error(errValueValid)
	}
}

//goland:noinspection DuplicatedCode
func TestBytesMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item Bytes

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewBytes()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isNullBytes(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewBytesValue([]byte(bytesTestString))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isBytesValid(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	zero := NewBytesValue(nil)
	err = enc.Encode(&zero)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	if !bytes.Equal(item.MustValue(), []byte{}) || !item.Valid {
		t.Error(errValueValid)
	}
}
