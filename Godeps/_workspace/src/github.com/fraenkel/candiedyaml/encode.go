package candiedyaml

import (
	"encoding/base64"
	"errors"
	"io"
	"math"
	"reflect"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"time"
)

var timeTimeType = reflect.TypeOf(time.Time{})

// An Encoder writes JSON objects to an output stream.
type Encoder struct {
	w       io.Writer
	emitter yaml_emitter_t
	event   yaml_event_t
	flow    bool
	err     error
}

// NewEncoder returns a new encoder that writes to w.
func NewEncoder(w io.Writer) *Encoder {
	e := &Encoder{w: w}
	yaml_emitter_initialize(&e.emitter)
	yaml_emitter_set_output_writer(&e.emitter, e.w)
	yaml_stream_start_event_initialize(&e.event, yaml_UTF8_ENCODING)
	e.emit()
	yaml_document_start_event_initialize(&e.event, nil, nil, true)
	e.emit()

	return e
}

func (e *Encoder) Encode(v interface{}) (err error) {
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(runtime.Error); ok {
				panic(r)
			}
			switch r := r.(type) {
			case error:
				err = r
			case string:
				err = errors.New(r)
			default:
				err = errors.New("Unknown panic: " + reflect.TypeOf(r).String())
			}

			debug.PrintStack()
		}
	}()

	if e.err != nil {
		return e.err
	}

	e.marshal("", reflect.ValueOf(v))

	yaml_document_end_event_initialize(&e.event, true)
	e.emit()
	e.emitter.open_ended = false
	yaml_stream_end_event_initialize(&e.event)
	e.emit()

	return nil
}

func (e *Encoder) emit() {
	if !yaml_emitter_emit(&e.emitter, &e.event) {
		panic("bad emit")
	}
}

func (e *Encoder) marshal(tag string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Interface:
		if v.IsNil() {
			e.emitNil()
		} else {
			e.marshal(tag, v.Elem())
		}
	case reflect.Map:
		e.emitMap(tag, v)
	case reflect.Ptr:
		if v.IsNil() {
			e.emitNil()
		} else {
			e.marshal(tag, v.Elem())
		}
	case reflect.Struct:
		e.emitStruct(tag, v)
	case reflect.Slice:
		e.emitSlice(tag, v)
	case reflect.String:
		e.emitString(tag, v)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.emitInt(tag, v)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		e.emitUint(tag, v)
	case reflect.Float32, reflect.Float64:
		e.emitFloat(tag, v)
	case reflect.Bool:
		e.emitBool(tag, v)
	default:
		panic("Can't marshal type yet: " + v.Type().String())
	}
}

func (e *Encoder) emitMap(tag string, v reflect.Value) {
	e.mapping(tag, func() {
		var keys stringValues = v.MapKeys()
		sort.Sort(keys)
		for _, k := range keys {
			e.marshal("", k)
			e.marshal("", v.MapIndex(k))
		}
	})
}

func (e *Encoder) emitStruct(tag string, v reflect.Value) {
	if v.Type() == timeTimeType {
		e.emitTime(tag, v)
		return
	}

	fields := cachedTypeFields(v.Type())

	e.mapping(tag, func() {
		for _, f := range fields {
			fv := fieldByIndex(v, f.index)
			if !fv.IsValid() || f.omitEmpty && isEmptyValue(fv) {
				continue
			}

			e.marshal("", reflect.ValueOf(f.name))
			e.flow = f.flow
			e.marshal("", fv)
		}
	})
}

func (e *Encoder) emitTime(tag string, v reflect.Value) {
	t := v.Interface().(time.Time)
	s := t.Format(time.RFC3339)
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func isEmptyValue(v reflect.Value) bool {
	switch v.Kind() {
	case reflect.Array, reflect.Map, reflect.Slice, reflect.String:
		return v.Len() == 0
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func (e *Encoder) mapping(tag string, f func()) {
	implicit := tag == ""
	style := yaml_BLOCK_MAPPING_STYLE
	if e.flow {
		e.flow = false
		style = yaml_FLOW_MAPPING_STYLE
	}
	yaml_mapping_start_event_initialize(&e.event, nil, []byte(tag), implicit, style)
	e.emit()

	f()

	yaml_mapping_end_event_initialize(&e.event)
	e.emit()
}

func (e *Encoder) emitSlice(tag string, v reflect.Value) {
	if v.Type() == byteSliceType {
		e.emitBase64(tag, v)
		return
	}

	implicit := tag == ""
	style := yaml_BLOCK_SEQUENCE_STYLE
	if e.flow {
		e.flow = false
		style = yaml_FLOW_SEQUENCE_STYLE
	}
	yaml_sequence_start_event_initialize(&e.event, nil, []byte(tag), implicit, style)
	e.emit()

	n := v.Len()
	for i := 0; i < n; i++ {
		e.marshal("", v.Index(i))
	}

	yaml_sequence_end_event_initialize(&e.event)
	e.emit()
}

func (e *Encoder) emitBase64(tag string, v reflect.Value) {
	if v.IsNil() {
		e.emitNil()
		return
	}

	s := v.Bytes()

	dst := make([]byte, base64.StdEncoding.EncodedLen(len(s)))

	base64.StdEncoding.Encode(dst, s)
	e.emitScalar(string(dst), "", "!!binary", yaml_DOUBLE_QUOTED_SCALAR_STYLE)
}

func (e *Encoder) emitString(tag string, v reflect.Value) {
	var style yaml_scalar_style_t
	s := v.String()

	style = yaml_DOUBLE_QUOTED_SCALAR_STYLE
	e.emitScalar(s, "", tag, style)
}

func (e *Encoder) emitBool(tag string, v reflect.Value) {
	s := strconv.FormatBool(v.Bool())
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitInt(tag string, v reflect.Value) {
	s := strconv.FormatInt(v.Int(), 10)
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitUint(tag string, v reflect.Value) {
	s := strconv.FormatUint(v.Uint(), 10)
	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitFloat(tag string, v reflect.Value) {
	f := v.Float()

	var s string
	switch {
	case math.IsNaN(f):
		s = ".nan"
	case math.IsInf(f, 1):
		s = "+.inf"
	case math.IsInf(f, -1):
		s = "-.inf"
	default:
		s = strconv.FormatFloat(f, 'g', -1, v.Type().Bits())
	}

	e.emitScalar(s, "", tag, yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitNil() {
	e.emitScalar("null", "", "", yaml_PLAIN_SCALAR_STYLE)
}

func (e *Encoder) emitScalar(value, anchor, tag string, style yaml_scalar_style_t) {
	implicit := tag == ""
	if !implicit {
		style = yaml_PLAIN_SCALAR_STYLE
	}

	yaml_scalar_event_initialize(&e.event, []byte(anchor), []byte(tag), []byte(value), implicit, implicit, style)
	e.emit()
}
