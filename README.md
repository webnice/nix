# nix

[![GoDoc](https://pkg.go.dev/badge/github.com/webnice/nix/v2.svg)](https://pkg.go.dev/github.com/webnice/nix/v2)
[![Coverage Status](https://coveralls.io/repos/github/webnice/nix/badge.svg?branch=v2)](https://coveralls.io/github/webnice/nix?branch=v2)

Библиотека golang

Обработка значений, допускающих nil

## Описание

`nix` — это набор типов данных-обёрток для Go, позволяющих хранить в полях
структур «пустые» (NULL) значения вместе с обычными. Каждый тип реализует
интерфейсы `sql.Scanner` и `driver.Valuer` (совместимость с `database/sql`,
GORM, sqlx и другими ORM), интерфейсы кодеков `encoding/json`,
`encoding.TextMarshaler`/`TextUnmarshaler` и `encoding/gob`.

Пакет является продолжением пакетов `nul` и `lin.v1/nl`: все типы, имена
методов и их семантика сохранены, добавлены новые типы и generic-вариант.

### Особенности

- Полная матрица встроенных скалярных типов: знаковые, беззнаковые, с
  плавающей точкой и комплексные числа;
- Популярные не встроенные типы: `time.Time`, `time.Duration`, `json.RawMessage`,
  `net.IP`, `big.Int`, UUID (без внешних зависимостей);
- Generic-тип `Nullable[T]` — обёртка для любого скалярного типа, включая
  производные (`type MyInt int`);
- Значение, флаг действительности (`Valid`) и удобные методы доступа:
  `Set`, `SetValid`, `MustValue`, `Pointer`, `NullIfDefault`;
- Сериализация в JSON, текстовый и двоичный (gob) форматы;
- Нулевые зависимости, покрытие тестами 100%.

## Требования

- Go 1.18 или новее (используются generics).

## Установка

```bash
go get -u github.com/webnice/nix/v2
```

Подключение:

```go
import (
	"github.com/webnice/nix/v2"
)
```

#### Зависимости

	NONE

## Поддерживаемые типы данных

| Тип пакета   | Базовый тип        | В базе данных             |
|--------------|--------------------|---------------------------|
| `Bool`       | `bool`             | `bool`                    |
| `String`     | `string`           | `string`                  |
| `Bytes`      | `[]byte`           | `[]byte`                  |
| `StringSlice`| `[]string`         | `set`, строка `"a,b,c"`   |
| `Int`        | `int`              | `int64`                   |
| `Int8`       | `int8`             | `int64`                   |
| `Int16`      | `int16`            | `int64`                   |
| `Int32`      | `int32`            | `int64`                   |
| `Int64`      | `int64`            | `int64`                   |
| `Uint`       | `uint`             | десятичная строка         |
| `Uint8`      | `uint8`            | десятичная строка         |
| `Uint16`     | `uint16`           | десятичная строка         |
| `Uint32`     | `uint32`           | десятичная строка         |
| `Uint64`     | `uint64`           | десятичная строка         |
| `Uintptr`    | `uintptr`          | десятичная строка         |
| `Float32`    | `float32`          | `float64`                 |
| `Float64`    | `float64`          | `float64`                 |
| `Complex64`  | `complex64`        | строка `"(1.5+2.5i)"`     |
| `Complex128` | `complex128`       | строка `"(1.5+2.5i)"`     |
| `Time`       | `time.Time`        | `datetime`                |
| `Duration`   | `time.Duration`    | `int64` (наносекунды); чтение из TIME `[-]ЧЧЧ:ММ:СС[.дробная]` |
| `RawMessage` | `json.RawMessage`  | `[]byte`                  |
| `IP`         | `net.IP`           | строка `"192.168.0.1"`    |
| `BigInt`     | `*big.Int`         | десятичная строка         |
| `UUID`       | `[16]byte`         | строка `xxxxxxxx-…`       |
| `Nullable[T]`| скалярный `T`      | как у соответствующего типа |

### Тип Duration и время базы данных

`nix.Duration` читается из колонок типа `time` баз данных MySQL и PostgreSQL,
которые драйверы передают строкой `[-]ЧЧЧ:ММ:СС[.дробная часть]`:

- MySQL `TIME` — длительность в диапазоне `-838:59:59`..`838:59:59`, в том
  числе с дробной частью (`TIME(6)`, микросекунды); отрицательные значения
  передаются со знаком `-`;
- PostgreSQL `time` — время суток `00:00:00`..`24:00:00` с дробной частью
  до микросекунд.

Дробная часть принимается от 1 до 9 цифр и масштабируется до наносекунд.
Наряду с БД-форматом поддерживаются строки формата `time.ParseDuration`
(`"1h30m0s"`, `"300ms"`) и целые числа наносекунд. БД-формат поддерживается
методами `Scan`, `UnmarshalJSON` и `UnmarshalText`; запись в базу данных
выполняется методом `Value` как `int64` (наносекунды).

### Тип StringSlice и формат set базы данных

`nix.StringSlice` — слайс строк `[]string`, совместимый с типом `set` баз
данных MySQL, который драйверы передают строкой с перечислением элементов
через запятую:

- запись в базу данных методом `Value` выполняется строкой `"a,b,c"`;
- пустой слайс записывается пустой строкой — пустой набор `set`;
- чтение разбирает строку `set` по запятым; пустая строка — пустой набор;
- запятая и обратная косая черта в элементах экранируются обратной косой
  чертой: элемент `"a,b"` хранится как `"a\,b"`, элемент `"a\b"` — как
  `"a\\b"`; при чтении экранирование восстанавливается.

JSON-представление значения — массив строк, текстовое — строка `set`.

## Примеры использования

### Поля структуры с NULL-значениями

```go
// Message История запросов и ответов телеграм бота.
type Message struct {
	ID       uint64         `db:"id"       gorm:"column:id;primary_key"`
	CreateAt time.Time      `db:"createAt" gorm:"column:createAt"`
	DeleteAt nix.Time       `db:"deleteAt" gorm:"column:deleteAt"`
	DataRaw  nix.RawMessage `db:"dataRaw"  gorm:"column:dataRaw"`
	Timeout  nix.Duration   `db:"timeout"  gorm:"column:timeout"`
	TraceID  nix.UUID       `db:"traceId"  gorm:"column:traceId"`
}
```

### Запись и чтение в базе данных

Поля передаются в запрос значениями — как это делают GORM и sqlx:

```go
msg := Message{
	DeleteAt: nix.NewTimeValue(time.Now()),
	DataRaw:  nix.NewRawMessageValue([]byte(`{"update_id":1}`)),
	Timeout:  nix.NewDurationValue(90 * time.Minute),
}

_, err := db.Exec(
	"INSERT INTO message(deleteAt, dataRaw, timeout) VALUES(?,?,?)",
	msg.DeleteAt, msg.DataRaw, msg.Timeout,
)
if err != nil {
	log.Fatal(err)
}

// Чтение: NULL корректно попадает в объект с Valid=false.
var sendAt nix.Time
if err = db.QueryRow("SELECT sendAt FROM message WHERE id=?", 1).Scan(&sendAt); err != nil {
	log.Fatal(err)
}
if !sendAt.Valid {
	// Значение в базе — NULL.
}
```

### JSON

```go
data, err := json.Marshal(&msg)
if err != nil {
	log.Fatal(err)
}
if err = json.Unmarshal(data, &msg); err != nil {
	log.Fatal(err)
}
// Действительное значение сериализуется как значение, нулевое — как null:
// {"DeleteAt":"2026-11-11T10:00:00Z","DataRaw":{"update_id":1}, …}
```

### Присвоение значений: Set и SetValid

```go
v := nix.NewInt64()

// Set присваивает значение как есть и выставляет Valid=true без проверок.
v.Set(0)
// v.Valid == true

// SetValid дополнительно сбрасывает Valid, если присвоено значение
// по умолчанию (0, пустая строка, нулевое время и т.п.).
v.SetValid(0)
// v.Valid == false
v.SetValid(42)
// v.Valid == true
```

### Generic-тип Nullable[T]

```go
// Обёртка для любого скалярного типа, включая производные.
type Level int

score := nix.NewNullableValue(9.5)  // Nullable[float64]
stage := nix.NewNullableValue(Level(3))

_ = score
_ = stage
```

### Интерфейсы пакета

Все типы пакета (строго указатели, например `*nix.Time`) удовлетворяют
интерфейсам:

- `nix.Interface` — общее подмножество: `Reset`, `Scan`, `Value`, JSON, Text
  и Binary кодеки. Позволяет работать с гетерогенными коллекциями значений:
  ```go
  values := []nix.Interface{&msg.DeleteAt, &msg.Timeout, &msg.TraceID}
  ```
- `nix.Valuer[T]` — типизированный доступ: `Set(T)`, `SetValid(T)`,
  `MustValue() T`, `Pointer() *T`.
- `nix.Nullifier[W]` — `NullIfDefault() W`.

Полный набор методов типа выражается пересечением `Valuer[T]` и `Nullifier[W]`.

## Миграция с пакетов nul и lin.v1/nl

Типы и имена методов совместимы, меняются только импорт и префикс типов:

```go
// Было:
import nul "gopkg.in/webnice/lin.v1/nl"

// Стало:
import "github.com/webnice/nix/v2"
```

Далее по коду заменяется префикс `nul.` на `nix.`: `nul.Time` → `nix.Time`,
`nul.String` → `nix.String` и т.д. Конструкторы `NewTime()`, `NewTimeValue()`,
методы `Scan`, `Value`, `SetValid`, `NullIfDefault` и кодеки работают без
изменений.

## Тестирование

```bash
# Стандартный прогон с покрытием.
go test -cover ./...

# Прогон через Makefile (агрегированный профиль покрытия в coverage.log).
make test
```

## Лицензия

© WEB NICE, распространяется под [лицензией MIT](LICENSE).
