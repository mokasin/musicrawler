package encoding

import (
	"fmt"
	"musicrawler/lib/database"
	"testing"
)

type aStruct struct {
	Ignore     int
	I          int    `column:"myint"`
	S          string `column:"mystring"`
	unexported string `column:"myunexported"`
	NoSet      int    `column:"noset" set:"0"`
}

var str *aStruct = &aStruct{
	Ignore:     1,
	I:          2,
	S:          "a string",
	unexported: "not exported",
	NoSet:      3,
}

var res database.Result = database.Result{
	"myint":    int64(2),
	"mystring": "a string",
	"noset":    int64(3),
}

func TestEncode(t *testing.T) {
	ent, err := Encode(str)
	if err != nil {
		t.Error(err)
	}

	for i, v := range ent {
		if fmt.Sprintf("%v", v.Value) != fmt.Sprintf("%v", res[v.Column]) {
			t.Errorf("%d. Want: %v, Got: %v", i, res[v.Column], v.Value)
		}
	}
}

func BenchmarkEncode(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Encode(str)
	}
}

func TestDecode(t *testing.T) {
	str := &aStruct{}
	err := Decode(res, str)
	if err != nil {
		t.Error(err)
	}

	wantgot := []struct {
		want, got interface{}
	}{
		{res["myint"], str.I},
		{res["mystring"], str.S},
	}

	for i, v := range wantgot {
		if fmt.Sprintf("%v", v.want) != fmt.Sprintf("%v", v.got) {
			t.Errorf("%d. Want: %v, Got: %v", i, v.want, v.got)
		}
	}
}

func BenchmarkDecode(b *testing.B) {
	b.StopTimer()
	s := &aStruct{}
	b.StartTimer()

	for i := 0; i < b.N; i++ {
		Decode(res, s)
	}
}
