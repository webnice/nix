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

const rawMessageTestBody = `{"key":"value"}`

var (
	rawMessageNullInvalidGob = `33ff81030101115261774d6573736167655772617070657201ff82000102010556616c7565010a00010556616c6964010200000003ff8200`
	rawMessageOkValidGob     = `33ff81030101115261774d6573736167655772617070657201ff82000102010556616c7565010a00010556616c6964010200000016ff82010f7b226b6579223a2276616c7565227d010100`
)

func isRawMessageValid(t *testing.T, s RawMessage, from string) {
	const (
		errRawMessage = "ошибка %q, json.RawMessage: %q ≠ %q"
		errValid      = "является недействительным, ожидалось действительное"
	)

	if !bytes.Equal(s.RawMessage, []byte(rawMessageTestBody)) {
		t.Errorf(errRawMessage, from, s.RawMessage, rawMessageTestBody)
	}
	if !s.Valid {
		t.Error(from, errValid)
	}
}

func isRawMessageNull(t *testing.T, s RawMessage, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if s.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewRawMessage(t *testing.T) {
	v1 := NewRawMessage()
	isRawMessageNull(t, v1, "NewRawMessage()")
}

func TestNewRawMessageValue(t *testing.T) {
	v1 := NewRawMessageValue([]byte(rawMessageTestBody))
	isRawMessageValid(t, v1, "NewRawMessageValue()")
}

func TestNewRawMessagePointerValue(t *testing.T) {
	var bv = json.RawMessage(rawMessageTestBody)

	v1 := NewRawMessagePointerValue(&bv)
	isRawMessageValid(t, v1, "NewRawMessagePointerValue()")

	v2 := NewRawMessagePointerValue(nil)
	isRawMessageNull(t, v2, "NewRawMessagePointerValue()")
}

func TestRawMessageSetValid(t *testing.T) {
	v1 := NewRawMessage()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid([]byte(rawMessageTestBody))
	isRawMessageValid(t, v1, "SetValid()")
}

func TestRawMessageReset(t *testing.T) {
	v1 := NewRawMessage()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	v1.SetValid([]byte("test"))
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestRawMessageNullIfDefault(t *testing.T) {
	v1 := NewRawMessageValue([]byte(rawMessageTestBody))
	isRawMessageValid(t, v1, "NewRawMessageValue()")
	v1.NullIfDefault()
	isRawMessageValid(t, v1, "NullIfDefault()")

	v1.Set(nil)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isRawMessageNull(t, v1, "NullIfDefault()")
}

func TestRawMessageMustValue(t *testing.T) {
	v1 := NewRawMessage()
	if v1.MustValue() != nil {
		t.Error("MustValue()", errNotNil)
	}
	v1.SetValid([]byte(rawMessageTestBody))
	if !bytes.Equal(v1.MustValue(), []byte(rawMessageTestBody)) {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestRawMessagePointer(t *testing.T) {
	v1 := NewRawMessagePointerValue(nil)
	isRawMessageNull(t, v1, "NewRawMessagePointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	v2 := NewRawMessageValue([]byte(rawMessageTestBody))
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if !bytes.Equal(*pb, []byte(rawMessageTestBody)) {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewRawMessagePointerValue(pb)
	isRawMessageValid(t, v3, "NewRawMessagePointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestRawMessageScan(t *testing.T) {
	v1 := NewRawMessage()
	errorPanic(v1.Scan([]byte(rawMessageTestBody)))
	isRawMessageValid(t, v1, "Scan()")

	v2 := NewRawMessage()
	errorPanic(v2.Scan(rawMessageTestBody))
	isRawMessageValid(t, v2, "Scan()")

	v3 := NewRawMessage()
	errorPanic(v3.Scan(nil))
	isRawMessageNull(t, v3, "Scan()")

	v4 := NewRawMessage()
	errorPanic(v4.Scan(false))
	if !v4.Valid || string(v4.RawMessage) != "false" {
		t.Error("Scan()", errWrong)
	}
}

func TestRawMessageValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item []byte

	v1 := NewRawMessageValue([]byte(rawMessageTestBody))
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.([]byte); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "[]byte")
	}
	if !bytes.Equal(item, []byte(rawMessageTestBody)) {
		t.Error("Value() value", errWrong)
	}

	v2 := NewRawMessage()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection ALL
func TestRawMessageUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewRawMessage()
	err = json.Unmarshal([]byte(rawMessageTestBody), &v1)
	errorPanic(err)
	isRawMessageValid(t, v1, "UnmarshalJSON()")

	v2 := NewRawMessage()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isRawMessageNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewRawMessage()
	err = v3.UnmarshalJSON(invalidJSON)
	errorPanic(err)
	if !bytes.Equal(v3.RawMessage, invalidJSON) {
		t.Error("UnmarshalJSON()", errWrong)
	}

	v4 := NewRawMessage()
	err = v4.UnmarshalJSON([]byte(""))
	errorPanic(err)
	if v4.Valid || len(v4.RawMessage) != 0 {
		t.Error("UnmarshalJSON()", errWrong)
	}
}

func TestRawMessageMarshalJSON(t *testing.T) {
	v1 := NewRawMessageValue([]byte(rawMessageTestBody))
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, rawMessageTestBody, "non-empty json marshal")

	v2 := NewRawMessage()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")

	v3 := NewRawMessageValue(nil)
	data, err = v3.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null raw message json marshal")
}

func TestRawMessageUnmarshalText(t *testing.T) {
	var err error

	v1 := NewRawMessage()
	err = v1.UnmarshalText([]byte(rawMessageTestBody))
	errorPanic(err)
	isRawMessageValid(t, v1, "UnmarshalText()")

	v2 := NewRawMessage()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if len(v2.RawMessage) != 0 || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewRawMessage()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isRawMessageNull(t, v3, "UnmarshalText()")
}

func TestRawMessageMarshalText(t *testing.T) {
	v1 := NewRawMessageValue([]byte(rawMessageTestBody))
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, rawMessageTestBody, "Non-empty text marshal")

	v2 := NewRawMessage()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection DuplicatedCode
func TestRawMessageUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := RawMessage{}
	btf, err = hex.DecodeString(rawMessageNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isRawMessageNull(t, v1, "UnmarshalBinary() invalid")

	v2 := RawMessage{}
	btf, err = hex.DecodeString(rawMessageOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isRawMessageValid(t, v2, "UnmarshalBinary() ok")

	v3 := RawMessage{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestRawMessageMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item RawMessage

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewRawMessage()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isRawMessageNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	ok := NewRawMessageValue([]byte(rawMessageTestBody))
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isRawMessageValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestRawMessageSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewRawMessage()
	v1.SetValid(nil)
	isRawMessageNull(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewRawMessage()
	v2.Set(nil)
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
