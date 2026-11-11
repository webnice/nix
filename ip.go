package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"net"
	"reflect"
	"regexp"

	"github.com/webnice/nix/wrapper"
)

// IP Объект net.IP, допускающий значение nil.
type IP struct {
	IP    net.IP // Значение объекта.
	Valid bool   // Valid равен true если значение не nil.
}

// NewIP Создание нового объекта IP.
func NewIP() IP {
	return IP{
		IP:    nil,
		Valid: false,
	}
}

// NewIPValue Создание нового действительного объекта IP из значения.
func NewIPValue(value net.IP) IP {
	return IP{
		IP:    append(net.IP(nil), value...),
		Valid: true,
	}
}

// NewIPPointerValue Создание нового действительного объекта IP из ссылки на значение.
func NewIPPointerValue(ptr *net.IP) IP {
	if ptr == nil {
		return NewIP()
	}
	return NewIPValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (i *IP) SetValid(value net.IP) {
	i.IP, i.Valid = append(i.IP[:0], value...), true
	i.NullIfDefault()
}

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (i *IP) Set(value net.IP) { i.IP, i.Valid = append(i.IP[:0], value...), true }

// Reset Сброс значения и установка флага не действительного значения.
func (i *IP) Reset() { i.IP, i.Valid = nil, false }

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (i *IP) NullIfDefault() IP {
	if len(i.IP) == 0 {
		i.Reset()
	}
	return *i
}

// MustValue Возвращает значение в любом случае.
func (i *IP) MustValue() net.IP {
	if !i.Valid {
		return nil
	}
	return i.IP
}

// Pointer Возвращает ссылку на значение.
func (i *IP) Pointer() *net.IP {
	if !i.Valid {
		return nil
	}
	return &i.IP
}

// Scan Реализация интерфейса Scanner.
func (i *IP) Scan(value any) (err error) {
	switch x := value.(type) {
	case nil:
		i.IP, i.Valid = nil, false
		return
	case net.IP:
		i.IP = append(net.IP(nil), x...)
	case string:
		i.IP = net.ParseIP(x)
		if i.IP == nil {
			err = fmt.Errorf(errIPParse, x)
		}
	case []byte:
		i.IP = append(net.IP(nil), x...)
	default:
		err = fmt.Errorf(errIPScan, value, value)
	}
	i.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (i IP) Value() (driver.Value, error) {
	if !i.Valid {
		return nil, nil
	}
	return i.IP.String(), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (i *IP) UnmarshalJSON(data []byte) (err error) {
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		i.IP, i.Valid = nil, false
		return
	case string:
		var str = x
		if len(str) == 0 {
			i.Valid = false
			return
		}
		i.IP = net.ParseIP(str)
		if i.IP == nil {
			err = fmt.Errorf(errIPParse, str)
		}
	case map[string]any:
		var rex *regexp.Regexp
		_, ipOK := x["IP"].(string)
		valid, validOK := x["Valid"].(bool)
		if !ipOK || !validOK {
			return fmt.Errorf(errIPObject, x["IP"], x["Valid"])
		}
		rex = regexp.MustCompile(`"IP"\s*:\s*"([^"]+)"`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = i.UnmarshalText([]byte(arr[1]))
			i.Valid = valid
		}
		return
	default:
		err = fmt.Errorf(errIPUnmarshal, reflect.TypeOf(v).Name())
	}
	i.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i IP) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(i.IP.String())

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (i *IP) UnmarshalText(text []byte) (err error) {
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		i.IP, i.Valid = nil, false
		return
	case emptyString:
		i.IP, i.Valid = nil, true
		return
	default:
		i.IP = net.ParseIP(str)
		if i.IP == nil {
			err = fmt.Errorf(errIPParse, str)
		}
	}
	i.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i IP) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !i.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(i.IP.String())

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (i *IP) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.IPWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.IPWrapper)
	if err = dec.Decode(item); err == nil {
		i.IP, i.Valid = append(i.IP[:0], item.Value...), item.Valid
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (i IP) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.IPWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.IPWrapper{Value: i.IP, Valid: i.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
