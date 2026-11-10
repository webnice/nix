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

// uuidLength Длина строкового представления UUID с разделителями.
const uuidLength = 36

// UUID Объект UUID (16 байт), допускающий значение nil.
// Реализация без привязки к внешним библиотекам.
type UUID struct {
	UUID  [16]byte // Значение объекта.
	Valid bool     // Valid равен true если значение не nil.
}

// NewUUID Создание нового объекта UUID.
func NewUUID() UUID {
	return UUID{
		UUID:  [16]byte{},
		Valid: false,
	}
}

// NewUUIDValue Создание нового действительного объекта UUID из значения.
func NewUUIDValue(value [16]byte) UUID {
	return UUID{
		UUID:  value,
		Valid: true,
	}
}

// NewUUIDPointerValue Создание нового действительного объекта UUID из ссылки на значение.
func NewUUIDPointerValue(ptr *[16]byte) UUID {
	if ptr == nil {
		return NewUUID()
	}
	return NewUUIDValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (u *UUID) SetValid(value [16]byte) { u.UUID, u.Valid = value, true; u.NullIfDefault() }

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (u *UUID) Set(value [16]byte) { u.UUID, u.Valid = value, true }

// Reset Сброс значения и установка флага не действительного значения.
func (u *UUID) Reset() { u.UUID, u.Valid = [16]byte{}, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (u *UUID) NullIfDefault() UUID {
	if u.UUID == [16]byte{} {
		u.Reset()
	}
	return *u
}

// MustValue Возвращает значение в любом случае.
func (u *UUID) MustValue() [16]byte {
	if !u.Valid {
		return [16]byte{}
	}
	return u.UUID
}

// Pointer Возвращает ссылку на значение.
func (u *UUID) Pointer() *[16]byte {
	if !u.Valid {
		return nil
	}
	return &u.UUID
}

// Scan Реализация интерфейса Scanner.
func (u *UUID) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		u.UUID, u.Valid = [16]byte{}, false
		return
	case [16]byte:
		u.UUID = x
	case string:
		u.UUID, err = uuidParse(x)
	case []byte:
		u.UUID, err = uuidParse(string(x))
	default:
		err = fmt.Errorf("can't scan type %T into nix.UUID: %v", value, value)
	}
	u.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (u UUID) Value() (driver.Value, error) {
	if !u.Valid {
		return nil, nil
	}
	return uuidString(u.UUID), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (u *UUID) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case string:
		var str = x
		if len(str) == 0 {
			u.Valid = false
			return
		}
		u.UUID, err = uuidParse(str)
	case map[string]any:
		var rex *regexp.Regexp
		_, uuOK := x["UUID"].(string)
		valid, validOK := x["Valid"].(bool)
		if !uuOK || !validOK {
			return fmt.Errorf(`unmarshalling object into Go value of type nix.UUID requires key "UUID" to be of type `+
				`string and key "Valid" to be of type bool; `+
				`found %T and %T, respectively`, x["UUID"], x["Valid"])
		}
		rex = regexp.MustCompile(`"UUID"\s*:\s*"([^"]+)"`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = u.UnmarshalText([]byte(arr[1]))
			u.Valid = valid
		}
		return
	case nil:
		u.Valid = false
		return
	default:
		err = fmt.Errorf("can't unmarshal %q into go value of type nix.UUID", reflect.TypeOf(v).Name())
	}
	u.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u UUID) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(uuidString(u.UUID))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (u *UUID) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		u.UUID, u.Valid = [16]byte{}, false
		return
	case emptyString:
		u.UUID, u.Valid = [16]byte{}, true
		return
	default:
		u.UUID, err = uuidParse(str)
	}
	u.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u UUID) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !u.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(uuidString(u.UUID))

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (u *UUID) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.UUIDWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.UUIDWrapper)
	if err = dec.Decode(item); err == nil {
		u.UUID, u.Valid = item.Value, item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (u UUID) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.UUIDWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.UUIDWrapper{Value: u.UUID, Valid: u.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}

// uuidParse Разбор строкового представления UUID вида xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
// или xxxxxx... (32 шестнадцатеричных символа, регистр любой).
func uuidParse(s string) (u [16]byte, err error) {
	const (
		hexLength       = 32
		hyphenByte byte = '-'
	)

	switch len(s) {
	case uuidLength:
		if s[8] != hyphenByte || s[13] != hyphenByte || s[18] != hyphenByte || s[23] != hyphenByte {
			return u, fmt.Errorf("неверный формат UUID: %q", s)
		}
		s = s[:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
	case hexLength:
	default:
		return u, fmt.Errorf("неверный формат UUID: %q", s)
	}
	for i := 0; i < hexLength; i += 2 {
		var n uint64
		if n, err = strconv.ParseUint(s[i:i+2], 16, 8); err != nil {
			return [16]byte{}, fmt.Errorf("неверный формат UUID: %q", s)
		}
		u[i/2] = byte(n)
	}

	return
}

// uuidString Строковое представление UUID вида xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx.
func uuidString(u [16]byte) string {
	const hexTable = "0123456789abcdef"

	var buf [uuidLength]byte
	var pos int

	for i := 0; i < len(u); i++ {
		switch i {
		case 4, 6, 8, 10:
			buf[pos] = '-'
			pos++
		}
		buf[pos] = hexTable[u[i]>>4]
		buf[pos+1] = hexTable[u[i]&0x0f]
		pos += 2
	}

	return string(buf[:])
}
