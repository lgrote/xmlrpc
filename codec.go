package xmlrpc

import (
	"encoding/base64"
	"encoding/xml"
	"fmt"
	"io"
	"reflect"
	"strconv"
	"strings"
	"time"
)

type tag string

const (
	methodCallTag     tag = "methodCall"
	methodResponseTag tag = "methodResponse"
	methodNameTag     tag = "methodName"
	paramsTag         tag = "params"
	paramTag          tag = "param"
	nameTag           tag = "name"
	valueTag          tag = "value"
	arrayTag          tag = "array"
	dataTag           tag = "data"
	base64Tag         tag = "base64"
	booleanTag        tag = "boolean"
	dateTimeTag       tag = "dateTime.iso8601"
	doubleTag         tag = "double"
	integerTag        tag = "int"
	integerTag2       tag = "i4"
	stringTag         tag = "string"
	structTag         tag = "struct"
	memberTag         tag = "member"
	nilTag            tag = "nil"
	faultTag          tag = "fault"
	iso8601Format         = "20060102T15:04:5"
)

func Marshal(w io.Writer, method string, args ...interface{}) error {

	c := encoder{w: w}
	io.WriteString(w, xml.Header)
	openTag(w, methodCallTag)
	openTag(w, methodNameTag)
	io.WriteString(w, method)
	closeTag(w, methodNameTag)
	openTag(w, paramsTag)
	for _, o := range args {
		openTag(w, paramTag)
		c.write(o)
		closeTag(w, paramTag)
	}
	closeTag(w, paramsTag)
	closeTag(w, methodCallTag)
	return nil
}

type decoder struct {
	d *xml.Decoder
}

func Unmarshal(r io.Reader, o interface{}) error {
	d := newDecoder(r)
	if err := d.read(o); err != nil {
		return err
	}
	return nil
}
func newDecoder(r io.Reader) decoder {
	return decoder{d: xml.NewDecoder(r)}
}

func (this *decoder) read(o interface{}) error {

	m := reflect.ValueOf(make(map[string]interface{}))
	value := reflect.ValueOf(o)
	value.Elem().Set(m)
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(methodResponseTag):
				if err = this.decodeMethodResponse(m); err != nil {
					return err
				} else {
					return nil
				}
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeMethodResponse(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		var n interface{}
		nVal := reflect.ValueOf(&n).Elem()

		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(paramsTag):
				if err := this.decodeParams(nVal); err != nil {
					return err
				}
				o.SetMapIndex(reflect.ValueOf(string(paramsTag)), nVal)

			case string(faultTag):
				if err := this.decodeFault(nVal); err != nil {
					return err
				}
				o.SetMapIndex(reflect.ValueOf(string(faultTag)), nVal)
			}
		case xml.EndElement:
			if v.Name.Local == string(methodResponseTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, methodResponseTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeParams(o reflect.Value) error {
	arr := make([]interface{}, 0)
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(paramTag):
				var n interface{}
				nVal := reflect.ValueOf(&n).Elem()
				if err := this.decodeParam(nVal); err != nil {
					return err
				}
				arr = append(arr, n)
			}
		case xml.EndElement:
			switch v.Name.Local {
			case string(paramsTag):
				o.Set(reflect.ValueOf(arr))
				return nil
			case string(paramTag):
				// continue
			default:
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s or %s", v.Name.Local, paramsTag, paramTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeArray(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(dataTag):
				if err := this.decodeData(o); err != nil {
					return err
				}
			}
		case xml.EndElement:
			switch v.Name.Local {
			case string(arrayTag):
				return nil
			default:
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, arrayTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeData(o reflect.Value) error {
	arr := make([]interface{}, 0)
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(valueTag):
				var n interface{}
				nVal := reflect.ValueOf(&n).Elem()
				if err := this.decodeValue(nVal); err != nil {
					return err
				}
				arr = append(arr, n)
			}
		case xml.EndElement:
			switch v.Name.Local {
			case string(dataTag):
				o.Set(reflect.ValueOf(arr))
				return nil
			default:
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, dataTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeParam(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(valueTag):
				if err := this.decodeValue(o); err != nil {
					return err
				}
			}
		case xml.EndElement:
			switch v.Name.Local {
			case string(paramTag):
				return nil
			default:
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, paramTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeFault(o reflect.Value) error {
	var n interface{}
	nVal := reflect.ValueOf(&n).Elem()

	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}

		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(valueTag):
				if err := this.decodeValue(nVal); err != nil {
					return err
				}
			}
		case xml.EndElement:
			if v.Name.Local == string(faultTag) {
				o.Set(nVal)
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, faultTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeValue(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(structTag):
				err = this.decodeStruct(o)
			case string(integerTag):
				err = this.decodeInt(o)
			case string(integerTag2):
				err = this.decodeInt(o)
			case string(stringTag):
				err = this.decodeString(o)
			case string(doubleTag):
				err = this.decodeDouble(o)
			case string(booleanTag):
				err = this.decodeBoolean(o)
			case string(dateTimeTag):
				err = this.decodeDate(o)
			case string(base64Tag):
				err = this.decodeBase64(o)
			case string(nilTag):
				err = this.decodeNil(o)
			case string(arrayTag):
				err = this.decodeArray(o)
			}
		case xml.EndElement:
			if v.Name.Local == string(valueTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, valueTag)
			}
		}
		if err != nil {
			return fmt.Errorf("decodeValue got error: %v", err)
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeNil(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.EndElement:
			if v.Name.Local == string(nilTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, nil)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeBase64(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			if data, err := base64.StdEncoding.DecodeString(string(v)); err != nil {
				return fmt.Errorf("errr parsing base64: %s", err)
			} else {
				o.Set(reflect.ValueOf(data))
			}
		case xml.EndElement:
			if v.Name.Local == string(base64Tag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, base64Tag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeDate(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			if date, err := time.Parse(iso8601Format, string(v)); err != nil {
				return fmt.Errorf("errr parsing date: %s", err)
			} else {
				o.Set(reflect.ValueOf(date))
			}
		case xml.EndElement:
			if v.Name.Local == string(dateTimeTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, dateTimeTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeBoolean(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			if string(v) == "1" {
				o.Set(reflect.ValueOf(true))
			} else {
				o.Set(reflect.ValueOf(false))
			}
		case xml.EndElement:
			if v.Name.Local == string(booleanTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, booleanTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeDouble(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			if flo, err := strconv.ParseFloat(string(v), 64); err != nil {
				return err
			} else {
				o.Set(reflect.ValueOf(flo))
			}
		case xml.EndElement:
			if v.Name.Local == string(doubleTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, doubleTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeInt(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			if intV, err := strconv.ParseInt(string(v), 10, 64); err != nil {
				return err
			} else {
				o.Set(reflect.ValueOf(intV))
			}
		case xml.EndElement:
			if v.Name.Local == string(integerTag) || v.Name.Local == string(integerTag2) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, integerTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeString(o reflect.Value) error {
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.CharData:
			o.Set(reflect.ValueOf(string(v)))
		case xml.EndElement:
			if v.Name.Local == string(stringTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, stringTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeStruct(o reflect.Value) error {
	m := reflect.ValueOf(make(map[string]interface{}))
	o.Set(m)
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(memberTag):
				this.decodeMember(m)
			}
		case xml.EndElement:
			if v.Name.Local == string(structTag) {
				return nil
			} else {
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s", v.Name.Local, structTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

func (this *decoder) decodeMember(o reflect.Value) error {
	var name string
	for {
		t, err := this.d.Token()
		if err != nil {
			return err
		}
		switch v := t.(type) {
		case xml.StartElement:
			switch v.Name.Local {
			case string(nameTag):
				if name, err = this.readNextCharData(); err != nil {
					return err
				}
			case string(valueTag):
				if name != "" {
					var n interface{}
					nVal := reflect.ValueOf(&n).Elem()
					this.decodeValue(nVal)
					o.SetMapIndex(reflect.ValueOf(name), nVal)
				} else {
					return fmt.Errorf("got value element without name element before")
				}
			}
		case xml.EndElement:
			switch v.Name.Local {
			case string(memberTag):
				return nil
			case string(nameTag):
				// ignore
			default:
				return fmt.Errorf("got xml.EndElement %s expected xml.EndElement %s ot %s", v.Name.Local, memberTag, nameTag)
			}
		}
	}
	return fmt.Errorf("this point shouldn't be reached")
}

// doesn't close the current element
func (this *decoder) readNextCharData() (string, error) {
	for {
		t, err := this.d.Token()
		if err != nil {
			return "", err
		}
		switch v := t.(type) {
		case xml.CharData:
			return string(v), nil
		case xml.EndElement:
			return "", fmt.Errorf("expected xml.CharData but go xml.EndElement %s", v)
		}
	}
	return "", fmt.Errorf("this point shouldn't be reached")
}

type encoder struct {
	w io.Writer
}

func (this *encoder) write(o interface{}) {
	if o == nil {
		this.writeNil()
	}

	// use simple type switch if possible and use the refelction switch only as fallback
	switch f := reflect.ValueOf(o); f.Kind() {
	case reflect.Bool:
		this.writeBoolean(f.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		this.writeInt(f.Int())
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		this.writeUint(f.Uint())
	case reflect.Float32, reflect.Float64:
		this.writeFloat(f.Float())
	case reflect.String:
		this.writeString(f.String())
	case reflect.Array, reflect.Slice:
		switch o.(type) {
		case []byte:
			// byte arrays are special
			this.writeBytes(o.([]byte))
		default:
			openTag(this.w, arrayTag)
			openTag(this.w, dataTag)

			for i := 0; i < f.Len(); i++ {
				openTag(this.w, valueTag)
				this.write(f.Index(i).Interface())
				closeTag(this.w, valueTag)
			}
			closeTag(this.w, dataTag)
			closeTag(this.w, arrayTag)
		}
	case reflect.Struct:
		// time is special
		if date, ok := o.(time.Time); ok {
			this.writeTime(date)
			break
		}

		openTag(this.w, structTag)
		for i := 0; i < f.NumField(); i++ {
			field := f.Type().Field(i)
			// check if field is exported
			if field.PkgPath == "" {
				openTag(this.w, memberTag)
				openTag(this.w, nameTag)
				xml.Escape(this.w, []byte(strings.ToLower(field.Name)))
				closeTag(this.w, nameTag)
				openTag(this.w, valueTag)
				this.write(getField(f, i).Interface())
				closeTag(this.w, valueTag)
				closeTag(this.w, memberTag)
			}
		}
		closeTag(this.w, structTag)
	case reflect.Map:
		if len(f.MapKeys()) == 0 || f.MapKeys()[0].Kind() != reflect.String {
			break
		}
		for _, key := range f.MapKeys() {
			openTag(this.w, memberTag)
			openTag(this.w, nameTag)
			xml.Escape(this.w, []byte(key.String()))
			closeTag(this.w, nameTag)
			openTag(this.w, valueTag)
			this.write(f.MapIndex(key).Interface())
			closeTag(this.w, valueTag)
			closeTag(this.w, memberTag)
		}
	}
}
func (this *encoder) writeTime(time time.Time) {
	openTag(this.w, dateTimeTag)
	io.WriteString(this.w, time.Format(iso8601Format))
	closeTag(this.w, dateTimeTag)
}
func (this *encoder) writeBytes(b []byte) {
	openTag(this.w, base64Tag)
	io.WriteString(this.w, base64.StdEncoding.EncodeToString(b))
	closeTag(this.w, base64Tag)
}
func (this *encoder) writeNil() {
	openCloseTag(this.w, nilTag)
}

func (this *encoder) writeString(s string) {
	openTag(this.w, stringTag)
	xml.Escape(this.w, []byte(s))
	closeTag(this.w, stringTag)
}

func (this *encoder) writeFloat(f float64) {
	openTag(this.w, doubleTag)
	io.WriteString(this.w, strconv.FormatFloat(f, 'f', 10, 64))
	closeTag(this.w, doubleTag)
}
func (this *encoder) writeBoolean(b bool) {
	openTag(this.w, booleanTag)
	if b {
		io.WriteString(this.w, "0")
	} else {
		io.WriteString(this.w, "1")
	}
	closeTag(this.w, booleanTag)
}
func (this *encoder) writeUint(i uint64) {
	openTag(this.w, integerTag)
	io.WriteString(this.w, strconv.FormatUint(i, 10))
	closeTag(this.w, integerTag)
}
func (this *encoder) writeInt(i int64) {
	openTag(this.w, integerTag)
	io.WriteString(this.w, strconv.FormatInt(i, 10))
	closeTag(this.w, integerTag)
}

// Get the i'th arg of the struct value.
// If the arg itself is an interface, return a value for
// the thing inside the interface, not the interface itself.
// (stolen from the fmt package)
func getField(v reflect.Value, i int) reflect.Value {
	val := v.Field(i)
	if val.Kind() == reflect.Interface && !val.IsNil() {
		val = val.Elem()
	}
	return val
}
func openTag(w io.Writer, t tag) {
	io.WriteString(w, "<")
	io.WriteString(w, string(t))
	io.WriteString(w, ">")
}

func closeTag(w io.Writer, t tag) {
	io.WriteString(w, "</")
	io.WriteString(w, string(t))
	io.WriteString(w, ">")
}

func openCloseTag(w io.Writer, t tag) {
	io.WriteString(w, "<")
	io.WriteString(w, string(t))
	io.WriteString(w, "/>")
}
