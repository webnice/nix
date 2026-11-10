package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"reflect"
	"regexp"
	"strconv"

	"github.com/webnice/nix/wrapper"
)

// Uint32 Объект uint32, допускающий значение nil.
type Uint32 struct {
	Uint32 uint32 // Значение объекта.
	Valid  bool   // Valid равен true если значение не nil.
}

// NewUint32 Создание нового объекта Uint32.
func NewUint32() Uint32 {
	return Uint32{
		Uint32: 0,
		Valid:  false,
	}
}

// NewUint32Value Создание нового действительного объекта Uint32 из значения.
func NewUint32Value(value uint32) Uint32 {
	return Uint32{
		Uint32: value,
		Valid:  true,
	}
}

// NewUint32PointerValue Создание нового действительного объекта Uint32 из ссылки на значение.
func NewUint32PointerValue(ptr *uint32) Uint32 {
	if ptr == nil {
		return NewUint32()
	}
	return NewUint32Value(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *Uint32) SetValid(value uint32) { u.Uint32, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *Uint32) Set(value uint32) { u.Uint32, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *Uint32) Reset() { u.Uint32, u.Valid = 0, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *Uint32) NullIfDefault() Uint32 {
	if u.Uint32 == 0 {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *Uint32) MustValue() uint32 {
	if !u.Valid {
		return 0
	}
	return u.Uint32
}

// Pointer Возвращает ссылку на значение.
func (u *Uint32) Pointer() *uint32 {
	if !u.Valid {
		return nil
	}
	return &u.Uint32
}

// Scan Реализация интерфейса Scanner.
func (u *Uint32) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.Uint32, u.Valid = 0, false
		return
	case uint32:
		u.Uint32 = value.(uint32)
	default:
		buf := asString(x)
		var n uint64
		n, err = strconv.ParseUint(buf, 10, 32)
		u.Uint32 = uint32(n)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint32) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return []byte(strconv.FormatUint(uint64(u.Uint32), 10)), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *Uint32) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		u.Uint32, u.Valid = 0, false
		return
	case float64:
		err = json.Unmarshal(data, &u.Uint32)
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		var n uint64
		n, err = strconv.ParseUint(str, 10, 32)
		u.Uint32 = uint32(n)
	case map[string]any:
		var rex *regexp.Regexp
		_, uiOK := x["Uint32"].(float64)
		valid, validOK := x["Valid"].(bool)
		if !uiOK || !validOK {
			return fmt.Errorf(`unmarshalling object into Go value of type nix.Uint32 requires key "Uint32" to be of type `+
				`string and key "Valid" to be of type bool; `+
				`found %T and %T, respectively`, x["Uint32"], x["Valid"])
		}
		rex = regexp.MustCompile(`"Uint32"\s*:\s*([0-9]+)`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	default:
		err = fmt.Errorf("Can't unmarshal %v into Go value of type nix.Uint32", reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint32) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data = []byte(strconv.FormatUint(uint64(u.Uint32), 10))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *Uint32) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.Uint32, u.Valid = 0, false
		return
	case emptyString:
		u.Uint32, u.Valid = 0, true
		return
	default:
		var n uint64
		n, err = strconv.ParseUint(str, 10, 32)
		u.Uint32 = uint32(n)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint32) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(strconv.FormatUint(uint64(u.Uint32), 10))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *Uint32) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.Uint32Wrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.Uint32Wrapper)
	if err = dec.Decode(item); err == nil {
		u.Uint32, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u Uint32) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.Uint32Wrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.Uint32Wrapper{Value: u.Uint32, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
