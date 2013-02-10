package xmlrpc

import (
	"bytes"
	"testing"
	"time"
)

func TestMarshal(t *testing.T) {
	buf := new(bytes.Buffer)
	a := []struct {
		Title string
		Val   float64
		Arr   []string
	}{
		{"hallo", 12.0, []string{"one", "two", "three"}},
		{"asdf", 132.0, []string{"one", "two", "three"}},
		{"harthhllo", 12.1, []string{"one", "two", "three"}},
	}
	for _, v := range a {
		Marshal(buf, "test.method", 1, "2", 3.0, v, time.Now())
		//		t.Log(buf.String())
		buf.Reset()
	}
	Marshal(buf, "test.method", 1, "2", 3.0, map[string]int{"one": 1, "two": 2, "three": 3})
	//t.Log(buf.String())
	buf.Reset()
	Marshal(buf, "test.method", 1, "2", 3.0, map[int]int{1: 1, 2: 2, 3: 3})
	//t.Log(buf.String())
	buf.Reset()
}

func TestUnmarshalFault(t *testing.T) {
	s := `<?xml version="1.0"?>
			<methodResponse>
			  <fault>
			    <value>
			      <struct>
			        <member>
			          <name>faultCode</name>
			          <value><int>4</int></value>
			        </member>
			        <member>
			          <name>faultString</name>
			          <value><string>Too many parameters.</string></value>
			        </member>
			        <member>
			          <name>faultBla</name>
			          <value><double>12.0</double></value>
			        </member>
			        <member>
			          <name>faultI4</name>
			          <value><i4>12</i4></value>
			        </member>
			        <member>
			          <name>faultBool</name>
			          <value><boolean>1</boolean></value>
			        </member>
			      </struct>
			    </value>
			  </fault>
			</methodResponse>`
	buf := bytes.NewBufferString(s)
	var o interface{}
	if err := Unmarshal(buf, &o); err != nil {
		t.Fatalf("error unmarshaling err:%v", err)
	}
	t.Logf("unmarshalled: %+v\n", o)
	m := o.(map[string]interface{})
	if m1, ok := m["fault"].(map[string]interface{}); ok {
		if m1["faultCode"].(int64) != 4 {
			t.Errorf("expected %d but got %d\n", 4, m1["faultCode"])
		}
	} else {
		t.Fatalf("cannot cast %T\n", m["fault"])
	}
}

func TestUnmarshalResp(t *testing.T) {
	s := `<?xml version="1.0"?>
		<methodResponse>
		  <params>
		    <param>
		        <value><string>South Dakota</string></value>
		    </param>
		    <param>
		    	<value><i4>7</i4></value>
		    </param>
		    <param>
		    	<value><dateTime.iso8601>19980717T14:08:55</dateTime.iso8601></value>
		    </param>
		    <param>
		    	<value><base64>eW91IGNhbid0IHJlYWQgdGhpcyE=</base64></value>
		    </param>
		    <param>
		    	<value><nil/></value>
		    </param>
		    <param>
		    	<value>
					<array>
					  <data>
					    <value><i4>1404</i4></value>
					    <value><string>Something here</string></value>
					    <value><i4>1</i4></value>
					  </data>
					</array>
				</value>
		    </param>
		  </params>
		</methodResponse>`

	buf := bytes.NewBufferString(s)
	var o interface{}
	if err := Unmarshal(buf, &o); err != nil {
		t.Fatalf("error unmarshaling err:%v", err)
	}

	m := o.(map[string]interface{})
	t.Logf("unmarshalled: %+v\n", m)
	if m1, ok := m["fault"].(map[string]interface{}); ok {
		if m1["faultCode"].(int64) != 4 {
			t.Errorf("expected %d but got %d\n", 4, m1["faultCode"])
		}
	} else {
		t.Fatalf("cannot cast %T\n", m["fault"])
	}

}
