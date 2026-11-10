package nix

import (
	"bytes"
	"database/sql/driver"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"math/big"
	"reflect"
	"regexp"

	"github.com/webnice/nix/wrapper"
)

// BigInt Объект big.Int, допускающий значение nil.
type BigInt struct {
	BigInt *big.Int // Значение объекта.
	Valid  bool     // Valid равен true если значение не nil.
}

// NewBigInt Создание нового объекта BigInt.
func NewBigInt() BigInt {
	return BigInt{
		BigInt: new(big.Int),
		Valid:  false,
	}
}

// NewBigIntValue Создание нового действительного объекта BigInt из значения.
func NewBigIntValue(value *big.Int) BigInt {
	if value == nil {
		return NewBigInt()
	}
	return BigInt{
		BigInt: new(big.Int).Set(value),
		Valid:  true,
	}
}

// NewBigIntPointerValue Создание нового действительного объекта BigInt из ссылки на значение.
func NewBigIntPointerValue(ptr **big.Int) BigInt {
	if ptr == nil || *ptr == nil {
		return NewBigInt()
	}
	return NewBigIntValue(*ptr)
}

// SetValid Изменение значения и установка флага действительного значения.
// Если присвоенное значение является значением по умолчанию, флаг действительного значения сбрасывается.
func (b *BigInt) SetValid(value *big.Int) {
	if b.BigInt == nil {
		b.BigInt = new(big.Int)
	}
	if value == nil {
		b.BigInt.SetInt64(0)
	} else {
		b.BigInt.Set(value)
	}
	b.Valid = true
	b.NullIfDefault()
}

// Set Присваивает значение как есть и выставляет флаг действительного значения без проверки значения.
func (b *BigInt) Set(value *big.Int) {
	if b.BigInt == nil {
		b.BigInt = new(big.Int)
	}
	if value == nil {
		b.BigInt.SetInt64(0)
	} else {
		b.BigInt.Set(value)
	}
	b.Valid = true
}

// Reset Сброс значения и установка флага не действительного значения.
func (b *BigInt) Reset() {
	if b.BigInt == nil {
		b.BigInt = new(big.Int)
	} else {
		b.BigInt.SetInt64(0)
	}
	b.Valid = false
}

// NullIfDefault Выполняет сброс значения до nil, если значение переменной является значением по умолчанию.
func (b *BigInt) NullIfDefault() BigInt {
	if b.BigInt == nil || b.BigInt.Sign() == 0 {
		b.Reset()
	}
	return *b
}

// MustValue Возвращает значение в любом случае.
func (b *BigInt) MustValue() *big.Int {
	if b.BigInt == nil {
		return new(big.Int)
	}
	return b.BigInt
}

// Pointer Возвращает ссылку на значение.
func (b *BigInt) Pointer() *big.Int {
	if !b.Valid {
		return nil
	}
	return b.BigInt
}

// ensureBigInt Гарантирует наличие выделенного значения для работы с zero-value объектом.
func (b *BigInt) ensureBigInt() {
	if b.BigInt == nil {
		b.BigInt = new(big.Int)
	}
}

// Scan Реализация интерфейса Scanner.
func (b *BigInt) Scan(value any) (err error) {
	b.ensureBigInt()
	switch x := value.(type) {
	case nil:
		b.Reset()
		return
	case *big.Int:
		b.Set(x)
	case string:
		if _, ok := b.BigInt.SetString(x, 10); !ok {
			err = fmt.Errorf("can't parse %q as big.Int", x)
		}
		b.Valid = err == nil
		return
	case []byte:
		if _, ok := b.BigInt.SetString(string(x), 10); !ok {
			err = fmt.Errorf("can't parse %q as big.Int", string(x))
		}
		b.Valid = err == nil
		return
	case int64:
		b.BigInt.SetInt64(x)
		b.Valid = true
		return
	default:
		err = fmt.Errorf("can't scan type %T into nix.BigInt: %v", value, value)
	}
	b.Valid = err == nil

	return
}

// Value Реализация интерфейса driver.Valuer.
//
//goland:noinspection GoMixedReceiverTypes
func (b BigInt) Value() (driver.Value, error) {
	if !b.Valid {
		return nil, nil
	}
	return b.BigInt.String(), nil
}

// UnmarshalJSON Реализация интерфейса json.Unmarshaler.
func (b *BigInt) UnmarshalJSON(data []byte) (err error) {
	b.ensureBigInt()
	var v any

	if err = json.Unmarshal(data, &v); err != nil {
		return
	}
	switch x := v.(type) {
	case nil:
		b.Reset()
		return
	case string:
		var str = x
		if len(str) == 0 {
			b.Valid = false
			return
		}
		if _, ok := b.BigInt.SetString(str, 10); !ok {
			err = fmt.Errorf("can't parse %q as big.Int", str)
		}
	case float64:
		if _, ok := b.BigInt.SetString(string(data), 10); !ok {
			err = fmt.Errorf("can't parse %q as big.Int", string(data))
		}
	case map[string]any:
		var rex *regexp.Regexp
		_, biOK := x["BigInt"].(string)
		valid, validOK := x["Valid"].(bool)
		if !biOK || !validOK {
			return fmt.Errorf(`unmarshalling object into Go value of type nix.BigInt requires key "BigInt" to be of type `+
				`string and key "Valid" to be of type bool; `+
				`found %T and %T, respectively`, x["BigInt"], x["Valid"])
		}
		rex = regexp.MustCompile(`"BigInt"\s*:\s*"([^"]+)"`)
		if arr := rex.FindStringSubmatch(string(data)); len(arr) == 2 {
			err = b.UnmarshalText([]byte(arr[1]))
			b.Valid = valid
		}
		return
	default:
		err = fmt.Errorf("Can't unmarshal %v into Go value of type nix.BigInt", reflect.TypeOf(v).Name())
	}
	b.Valid = err == nil

	return
}

// MarshalJSON Реализация интерфейса json.Marshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b BigInt) MarshalJSON() (data []byte, err error) {
	const nullString = "null"

	if !b.Valid {
		data = []byte(nullString)
		return
	}
	data, err = json.Marshal(json.Number(b.BigInt.String()))

	return
}

// UnmarshalText Реализация интерфейса encoding.TextUnmarshaler.
func (b *BigInt) UnmarshalText(text []byte) (err error) {
	b.ensureBigInt()
	const (
		emptyString = ""
		nullString  = "null"
	)
	var str string

	switch str = string(text); str {
	case nullString:
		b.Reset()
		return
	case emptyString:
		b.BigInt.SetInt64(0)
		b.Valid = true
		return
	default:
		if _, ok := b.BigInt.SetString(str, 10); !ok {
			err = fmt.Errorf("can't parse %q as big.Int", str)
		}
	}
	b.Valid = err == nil

	return
}

// MarshalText Реализация интерфейса encoding.TextMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b BigInt) MarshalText() (text []byte, err error) {
	const nullString = "null"

	if !b.Valid {
		text = []byte(nullString)
		return
	}
	text = []byte(b.BigInt.String())

	return
}

// UnmarshalBinary Реализация интерфейса encoding.BinaryUnmarshaler.
func (b *BigInt) UnmarshalBinary(data []byte) (err error) {
	var (
		reader *bytes.Reader
		dec    *gob.Decoder
		item   *wrapper.BigIntWrapper
	)

	reader = bytes.NewReader(data)
	dec = gob.NewDecoder(reader)
	item = new(wrapper.BigIntWrapper)
	if err = dec.Decode(item); err == nil {
		if item.Value == nil {
			b.Reset()
		} else {
			b.ensureBigInt()
			b.BigInt.Set(item.Value)
			b.Valid = item.Valid
		}
	}

	return
}

// MarshalBinary Реализация интерфейса encoding.BinaryMarshaler.
//
//goland:noinspection GoMixedReceiverTypes
func (b BigInt) MarshalBinary() (data []byte, err error) {
	var (
		buf  *bytes.Buffer
		enc  *gob.Encoder
		item *wrapper.BigIntWrapper
	)

	buf = &bytes.Buffer{}
	enc = gob.NewEncoder(buf)
	item = &wrapper.BigIntWrapper{Value: b.BigInt, Valid: b.Valid}
	err = enc.Encode(item)
	data = buf.Bytes()

	return
}
