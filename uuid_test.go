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

const uuidTestString = `00112233-4455-6677-8899-aabbccddeeff`

const uuidTestStringCompact = `00112233445566778899aabbccddeeff`

var (
	uuidJSON              = []byte(`"00112233-4455-6677-8899-aabbccddeeff"`)
	uuidBlankJSON         = []byte(`""`)
	uuidMaxValueValidJSON = []byte(`{"UUID":"00112233-4455-6677-8899-aabbccddeeff","Valid":true}`)
	uuidNullInvalidGob    = `2eff8b0301010b555549445772617070657201ff8c000102010556616c756501ff8e00010556616c6964010200000019ff8d010101095b31365d75696e743801ff8e0001060120000015ff8c01100000000000000000000000000000000000`
	uuidOkValidGob        = `2eff8b0301010b555549445772617070657201ff8c000102010556616c756501ff8e00010556616c6964010200000019ff8d010101095b31365d75696e743801ff8e000106012000001fff8c01100011223344556677ff88ff99ffaaffbbffccffddffeeffff010100`
)

func isUUIDValid(t *testing.T, u UUID, from string) {
	const (
		errUUID  = "ошибка %q, UUID: %q ≠ %q"
		errValid = "является недействительным, ожидалось действительное"
	)

	if uuidString(u.UUID) != uuidTestString {
		t.Errorf(errUUID, from, uuidString(u.UUID), uuidTestString)
	}
	if !u.Valid {
		t.Error(from, errValid)
	}
}

func isUUIDNull(t *testing.T, u UUID, from string) {
	const errInvalid = "является действительным, ожидалось недействительное"
	if u.Valid {
		t.Error(from, errInvalid)
	}
}

func TestNewUUID(t *testing.T) {
	v1 := NewUUID()
	isUUIDNull(t, v1, "NewUUID()")
}

func TestNewUUIDValue(t *testing.T) {
	v1, err := uuidParse(uuidTestString)
	errorPanic(err)
	v2 := NewUUIDValue(v1)
	isUUIDValid(t, v2, "NewUUIDValue()")
}

func TestNewUUIDPointerValue(t *testing.T) {
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)

	v1 := NewUUIDPointerValue(&bv)
	isUUIDValid(t, v1, "NewUUIDPointerValue()")

	v2 := NewUUIDPointerValue(nil)
	isUUIDNull(t, v2, "NewUUIDPointerValue()")
}

func TestUUIDSetValid(t *testing.T) {
	v1 := NewUUID()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1.SetValid(bv)
	isUUIDValid(t, v1, "SetValid()")
}

func TestUUIDReset(t *testing.T) {
	v1 := NewUUID()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1.SetValid(bv)
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.Reset()
	if v1.Valid {
		t.Error("Свойство Valid", errValidTrue)
	}
}

func TestUUIDNullIfDefault(t *testing.T) {
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1 := NewUUIDValue(bv)
	isUUIDValid(t, v1, "NewUUIDValue()")
	v1.NullIfDefault()
	isUUIDValid(t, v1, "NullIfDefault()")

	v1.Set([16]byte{})
	if !v1.Valid {
		t.Error("Свойство Valid", errValidFalse)
	}
	v1.NullIfDefault()
	isUUIDNull(t, v1, "NullIfDefault()")
}

func TestUUIDMustValue(t *testing.T) {
	v1 := NewUUID()
	if v1.MustValue() != [16]byte{} {
		t.Error("MustValue()", errWrong)
	}
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1.SetValid(bv)
	if v1.MustValue() != bv {
		t.Error("MustValue()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUUIDPointer(t *testing.T) {
	v1 := NewUUIDPointerValue(nil)
	isUUIDNull(t, v1, "NewUUIDPointerValue(nil)")
	if v1.Pointer() != nil {
		t.Error("Pointer()", errNotNil)
	}

	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v2 := NewUUIDValue(bv)
	pb := v2.Pointer()
	if pb == nil {
		t.Error("Pointer()", errNil)
	}
	if //goland:noinspection GoMaybeNil
	*pb != bv {
		t.Error("Pointer() value", errWrong)
	}

	v3 := NewUUIDPointerValue(pb)
	isUUIDValid(t, v3, "NewUUIDPointerValue(max)")
	if v3.Pointer() == nil {
		t.Error("Pointer()", errNil)
	}
}

func TestUUIDScan(t *testing.T) {
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)

	v1 := NewUUID()
	errorPanic(v1.Scan(bv))
	isUUIDValid(t, v1, "Scan()")

	v2 := NewUUID()
	errorPanic(v2.Scan(uuidTestString))
	isUUIDValid(t, v2, "Scan()")

	v3 := NewUUID()
	errorPanic(v3.Scan(uuidTestStringCompact))
	isUUIDValid(t, v3, "Scan()")

	v4 := NewUUID()
	errorPanic(v4.Scan([]byte(uuidTestString)))
	isUUIDValid(t, v4, "Scan()")

	v5 := NewUUID()
	errorPanic(v5.Scan(nil))
	isUUIDNull(t, v5, "Scan()")

	v6 := NewUUID()
	err = v6.Scan(false)
	if err == nil {
		t.Error("Scan()", errNil)
	}

	v7 := NewUUID()
	err = v7.Scan("not-an-uuid")
	if err == nil {
		t.Error("Scan()", errNil)
	}
}

func TestUUIDValue(t *testing.T) {
	var err error
	var dv driver.Value
	var ok bool
	var item string

	bv, err2 := uuidParse(uuidTestString)
	errorPanic(err2)
	v1 := NewUUIDValue(bv)
	dv, err = v1.Value()
	errorPanic(err)
	if dv == nil {
		t.Error("Value()", errReturnsNil)
	}
	if item, ok = dv.(string); !ok {
		t.Errorf(errReturnsType, "Value()", reflect.TypeOf(dv).Name(), "string")
	}
	if item != uuidTestString {
		t.Error("Value() value", errWrong)
	}

	v2 := NewUUID()
	dv, err = v2.Value()
	errorPanic(err)
	if dv != nil {
		t.Error("Value()", errReturnsNotNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUUIDUnmarshalJSON(t *testing.T) {
	var err error

	v1 := NewUUID()
	err = json.Unmarshal(uuidJSON, &v1)
	errorPanic(err)
	isUUIDValid(t, v1, "UnmarshalJSON()")

	v2 := NewUUID()
	err = json.Unmarshal(bytesNullJSON, &v2)
	errorPanic(err)
	isUUIDNull(t, v2, "UnmarshalJSON(null)")

	v3 := NewUUID()
	err = v3.UnmarshalJSON(invalidJSON)
	if //goland:noinspection GoTypeAssertionOnErrors
	_, ok := err.(*json.SyntaxError); !ok {
		t.Errorf(errSyntaxError, err)
	}

	v4 := NewUUID()
	err = v4.UnmarshalJSON(uuidMaxValueValidJSON)
	errorPanic(err)
	isUUIDValid(t, v4, "UnmarshalJSON()")

	v5 := NewUUID()
	err = v5.UnmarshalJSON(blankJSON)
	if err == nil {
		t.Error(errErrorNotNil)
	}

	v6 := NewUUID()
	err = json.Unmarshal(uuidBlankJSON, &v6)
	errorPanic(err)
	isUUIDNull(t, v6, "UnmarshalJSON()")

	v7 := NewUUID()
	err = json.Unmarshal(boolFalseJSON, &v7)
	if err == nil {
		panic(errErrorNotNil)
	}
	isUUIDNull(t, v7, "UnmarshalJSON()")
}

func TestUUIDMarshalJSON(t *testing.T) {
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1 := NewUUIDValue(bv)
	data, err := v1.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, string(uuidJSON), "non-empty json marshal")

	v2 := NewUUID()
	data, err = v2.MarshalJSON()
	errorPanic(err)
	jsonEquals(t, data, "null", "null json marshal")
}

//goland:noinspection ALL
func TestUUIDUnmarshalText(t *testing.T) {
	var err error

	v1 := NewUUID()
	err = v1.UnmarshalText([]byte(uuidTestString))
	errorPanic(err)
	isUUIDValid(t, v1, "UnmarshalText()")

	v2 := NewUUID()
	err = v2.UnmarshalText([]byte(""))
	errorPanic(err)
	if v2.UUID != [16]byte{} || !v2.Valid {
		t.Error(errValueValid)
	}

	v3 := NewUUID()
	err = v3.UnmarshalText(boolNullJSON)
	errorPanic(err)
	isUUIDNull(t, v3, "UnmarshalText()")

	v4 := NewUUID()
	err = v4.UnmarshalText([]byte("not-an-uuid"))
	if err == nil {
		t.Error(errErrorNotNil)
	}
}

func TestUUIDMarshalText(t *testing.T) {
	bv, err := uuidParse(uuidTestString)
	errorPanic(err)
	v1 := NewUUIDValue(bv)
	data, err := v1.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, uuidTestString, "Non-empty text marshal")

	v2 := NewUUID()
	data, err = v2.MarshalText()
	errorPanic(err)
	jsonEquals(t, data, "null", "Null text marshal")
}

//goland:noinspection ALL
func TestUUIDParse(t *testing.T) {
	var err error

	// Неверные разделители в 36-символьной форме.
	_, err = uuidParse("00112233-4455-6677-8899Xaabbccddeeff")
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Неверная длина.
	_, err = uuidParse("00112233")
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Неверные шестнадцатеричные символы.
	_, err = uuidParse("zzzzzzzz-4455-6677-8899-aabbccddeeff")
	if err == nil {
		t.Error(errErrorNotNil)
	}

	// Регистр не влияет на результат.
	var lower, upper [16]byte
	lower, err = uuidParse(uuidTestString)
	errorPanic(err)
	upper, err = uuidParse("00112233-4455-6677-8899-AABBCCDDEEFF")
	errorPanic(err)
	if lower != upper {
		t.Error("uuidParse()", errWrong)
	}
}

//goland:noinspection DuplicatedCode
func TestUUIDUnmarshalBinary(t *testing.T) {
	var err error
	var btf []byte

	v1 := UUID{}
	btf, err = hex.DecodeString(uuidNullInvalidGob)
	errorPanic(err)
	errorPanic(v1.UnmarshalBinary(btf))
	isUUIDNull(t, v1, "UnmarshalBinary() invalid")

	v2 := UUID{}
	btf, err = hex.DecodeString(uuidOkValidGob)
	errorPanic(err)
	errorPanic(v2.UnmarshalBinary(btf))
	isUUIDValid(t, v2, "UnmarshalBinary() ok")

	v3 := UUID{}
	err = v3.UnmarshalBinary([]byte("not-a-gob"))
	if err == nil {
		t.Error("UnmarshalBinary()", errNil)
	}
}

//goland:noinspection DuplicatedCode
func TestUUIDMarshalBinary(t *testing.T) {
	var err error
	var buf *bytes.Buffer
	var enc *gob.Encoder
	var item UUID

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)

	null := NewUUID()
	err = enc.Encode(&null)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUUIDNull(t, item, "MarshalBinary() -> UnmarshalBinary()")

	buf.Reset()
	enc = gob.NewEncoder(buf)
	bv, err2 := uuidParse(uuidTestString)
	errorPanic(err2)
	ok := NewUUIDValue(bv)
	err = enc.Encode(&ok)
	errorPanic(err)
	if buf.Len() == 0 {
		t.Error("MarshalBinary()", errEmpty)
	}
	err = gob.NewDecoder(buf).Decode(&item)
	errorPanic(err)
	isUUIDValid(t, item, "MarshalBinary() -> UnmarshalBinary()")
}

//goland:noinspection DuplicatedCode
func TestUUIDSetValidDefault(t *testing.T) {
	// SetValid сбрасывает флаг Valid, если присвоено значение по умолчанию.
	v1 := NewUUID()
	v1.SetValid([16]byte{})
	isUUIDNull(t, v1, "SetValid(0)")

	// Set присваивает значение как есть, без проверки на значение по умолчанию.
	v2 := NewUUID()
	v2.Set([16]byte{})
	if !v2.Valid {
		t.Error("Set()", errValidFalse)
	}
}
