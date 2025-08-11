package help

import (
	"fmt"
	"net/url"
	"testing"
)

func TestStringify(t *testing.T) {
	mp := url.Values{}
	for i := 0; i <= 10; i++ {
		mp.Set(fmt.Sprint(i), fmt.Sprint(i))
	}
	s := AsStringer(mp, StringifyMethodTable)
	t.Log("\n" + s.String())
	s = AsStringer(mp, StringifyMethodList)
	t.Log("\n" + s.String())
}
