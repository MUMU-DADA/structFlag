// Package structflag 将配置结构体暴露为命令行标志。
package structflag

import (
	"flag"
	"reflect"
	"strconv"
	"time"
)

// Load 为结构体的每个字段创建一个命令行标志。详见 LoadTo 有关如何配置标志命名、用法和默认值的详细说明。
//
// 这些标志将创建在 flag.CommandLine 上，这是默认（全局）的 FlagSet。标志名称不带前缀。
func Load(v interface{}) {
	LoadTo(flag.CommandLine, "", v)
}

// LoadTo 为给定 FlagSet 的结构体的每个字段创建一个命令行标志。
//
// 每个创建的标志将加上给定的前缀和一个破折号("-")，
// 除非前缀为空，在这种情况下不会在标志名称前加上破折号。
//
// 创建的标志将设置为更新 v 的字段；调用 fs.Parse 后，v 的字段可能被 flag 包更新。
//
// v 的字段值将作为传递给 flag 包的默认值。
//
// 默认情况下，标志将按照给定结构体中的字段名称命名。要设置自定义名称，请使用名为 "flag" 的标签。
// 要禁用某个字段生成任何标志，请使用名称 "-"。
//
// 默认情况下，标志不会有任何用法信息。要设置用法信息，请使用名为 "usage" 的标签。
//
// 结构体字段标签及其含义的示例：
//
//	// Field 会作为一个名为 "Field" 的标志出现，没有用法信息。
//	Field int
//
//	// Field 会作为一个名为 "foo" 的标志出现，没有用法信息。
//	Field int `flag:"foo"`
//
//	// Field 会作为一个名为 "foo" 的标志出现，用法信息为 "bar"。
//	Field int `flag:"foo" usage:"bar"`
//
//	// Field 将被此包忽略。
//	Field int `flag:"-"`
//
// 此包支持以下字段类型，其他类型将被忽略：
//
//	bool
//	float64
//	int
//	uint
//	int64
//	uint64
//	time.Duration
//
// 这些类型对应于 flag 包原生支持的类型。
//
// 如果字段的值是一个结构体，则该嵌套结构体将递归加载。匿名结构体字段将按照其类型的名称加载，除非通过 "flag" 标签重命名。
//
// 例如，给定以下 "config" 结构体：
//
//	type config struct {
//	  Foo string `flag:"foo"`
//	  Bar struct {
//	    Baz string `flag:"baz"`
//	  } `flag:"baaar"`
//	  embedded `flag:"embezzled"`
//	}
//
//	type embedded struct {
//	  Quux string `flag:"quux"`
//	}
//
// 如果将 config 的实例传递给 LoadTo 并且前缀为空，则会生成以下标志：
//
//	foo
//	baaar-baz
//	embezzled-quux
//
// LoadTo 遵循 Go 的常规可见性规则。如果字段未导出，则不会为此字段创建标志。
//
// 循环数据结构会导致栈溢出。
//
// 如果 v 不是指向结构体的指针，则会引发 panic。
//
// 新增特性：
//   - 支持设置短选项(short option)，可以通过 "short" 标签指定。例如：
//     Field int `flag:"foo" short:"-f"`
//   - 支持设置默认值，默认值可以通过 "default" 标签指定。例如：
//     Field int `flag:"foo" default:"42"`
func LoadTo(fs *flag.FlagSet, prefix string, v interface{}) {
	val := reflect.ValueOf(v).Elem()
	load(fs, prefix, val)
}

func load(fs *flag.FlagSet, prefix string, val reflect.Value) {
	for i := 0; i < val.NumField(); i++ {
		field := val.Type().Field(i)
		usage := field.Tag.Get("usage")
		flagValue := field.Tag.Get("flag")
		defaultValue := field.Tag.Get("default")
		short := field.Tag.Get("short")

		// 跳过标记为 `flag-"` 的结构体字段
		if flagValue == "-" {
			continue
		}

		// 标志名称按照 `flag:"xxx"` 标签的值命名。如果未提供，则默认使用字段名称。
		//
		// 这类似于 encoding/json 包的默认行为。
		name := field.Name
		if flagValue != "" {
			name = flagValue
		}

		// 假设前缀为 "prefix-"，则标志名称为 "prefix-name"。
		//
		// 然而，如果前缀为空，则标志名称仅为 "name"，没有额外的破折号。
		if prefix != "" {
			name = prefix + "-" + name
		}

		switch val.Field(i).Kind() {
		case reflect.Struct:
			load(fs, name, val.Field(i))
		case reflect.Bool, reflect.Int64, reflect.Float64, reflect.Int, reflect.Uint, reflect.Uint64, reflect.String:
			switch f := val.Field(i).Addr().Interface().(type) {
			case *bool:
				defaultBool := defaultValue == "true"
				fs.BoolVar(f, name, defaultBool, usage)
				if short != "" {
					fs.BoolVar(f, short, defaultBool, usage)
				}
			case *time.Duration:
				defaultDuration, _ := time.ParseDuration(defaultValue)
				fs.DurationVar(f, name, defaultDuration, usage)
				if short != "" {
					fs.DurationVar(f, short, defaultDuration, usage)
				}
			case *float64:
				defaultFloat64, _ := strconv.ParseFloat(defaultValue, 64)
				fs.Float64Var(f, name, defaultFloat64, usage)
				if short != "" {
					fs.Float64Var(f, short, defaultFloat64, usage)
				}
			case *int:
				defaultInt, _ := strconv.Atoi(defaultValue)
				fs.IntVar(f, name, defaultInt, usage)
				if short != "" {
					fs.IntVar(f, short, defaultInt, usage)
				}
			case *int64:
				defaultInt64, _ := strconv.ParseInt(defaultValue, 10, 64)
				fs.Int64Var(f, name, defaultInt64, usage)
				if short != "" {
					fs.Int64Var(f, short, defaultInt64, usage)
				}
			case *string:
				fs.StringVar(f, name, defaultValue, usage)
				if short != "" {
					fs.StringVar(f, short, defaultValue, usage)
				}
			case *uint:
				defaultUint, _ := strconv.ParseUint(defaultValue, 10, 32)
				fs.UintVar(f, name, uint(defaultUint), usage)
				if short != "" {
					fs.UintVar(f, short, uint(defaultUint), usage)
				}
			case *uint64:
				defaultUint64, _ := strconv.ParseUint(defaultValue, 10, 64)
				fs.Uint64Var(f, name, defaultUint64, usage)
				if short != "" {
					fs.Uint64Var(f, short, defaultUint64, usage)
				}
			}
		default:
			return
		}
	}
}
