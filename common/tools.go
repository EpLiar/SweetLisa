package common

import (
	"crypto/sha1"
	"fmt"
	"net/url"
	"os/user"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	NotSameTypeErr    = fmt.Errorf("cannot fill empty: the two value have different type")
	NeedPassInPointer = fmt.Errorf("the structure passed in should be a pointer")
)

func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func Min(a, b int) int {
	if a > b {
		return b
	}
	return a
}

func Abs(a int) int {
	if a < 0 {
		return -a
	}
	return a
}

func BoolToInt(a bool) int {
	if a {
		return 1
	}
	return 0
}

func BoolToString(a bool) string {
	if a {
		return "true"
	}
	return "false"
}

func Deduplicate(list []string) []string {
	res := make([]string, 0, len(list))
	m := make(map[string]struct{})
	for _, v := range list {
		if _, ok := m[v]; ok {
			continue
		}
		m[v] = struct{}{}
		res = append(res, v)
	}
	return res
}

// UrlEncoded encodes a string like Javascript's encodeURIComponent()
func UrlEncoded(str string) string {
	u, err := url.Parse(str)
	if err != nil {
		return str
	}
	return u.String()
}

func TrimLineContains(parent, sub string) string {
	lines := strings.Split(parent, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		if !strings.Contains(line, sub) {
			result = append(result, line)
		}
	}
	return strings.Join(result, "\n")
}

// FillEmpty fill the empty field of the struct with default value given
func FillEmpty(toFill interface{}, defaultVal interface{}) error {
	ta := reflect.TypeOf(toFill)
	if ta.Kind() != reflect.Ptr {
		return NeedPassInPointer
	}
	tb := reflect.TypeOf(defaultVal)
	va := reflect.ValueOf(toFill)
	vb := reflect.ValueOf(defaultVal)
	for ta.Kind() == reflect.Ptr {
		ta = ta.Elem()
		va = va.Elem()
	}
	for tb.Kind() == reflect.Ptr {
		tb = tb.Elem()
		vb = vb.Elem()
	}
	if ta != tb {
		return NotSameTypeErr
	}
	for i := 0; i < va.NumField(); i++ {
		v := va.Field(i)
		if v.Type().Name() == "bool" {
			continue
		}
		if v.IsZero() {
			v.Set(vb.Field(i))
		}
	}
	return nil
}

func SliceSub(slice []string, toSub []string) []string {
	var res = make([]string, 0, len(slice))
	var m = make(map[string]struct{})
	for _, s := range toSub {
		m[s] = struct{}{}
	}
	for _, s := range slice {
		if _, ok := m[s]; !ok {
			res = append(res, s)
		}
	}
	return res
}

func SliceHas(slice []string, set []string) []string {
	var res = make([]string, 0, len(slice))
	var m = make(map[string]struct{})
	for _, s := range set {
		m[s] = struct{}{}
	}
	for _, s := range slice {
		if _, ok := m[s]; ok {
			res = append(res, s)
		}
	}
	return res
}

func SliceToSet(slice []string) map[string]struct{} {
	var m = make(map[string]struct{})
	for _, s := range slice {
		m[s] = struct{}{}
	}
	return m
}

func BytesCopy(b []byte) []byte {
	var a = make([]byte, len(b))
	copy(a, b)
	return a
}

func Bytes2Sha1(b []byte, salt []byte) []byte {
	h := sha1.New()
	h.Write(b)
	if len(salt) > 0 {
		h.Write(salt)
	}
	return h.Sum(nil)
}

func HomeExpand(path string) (string, error) {
	if !strings.HasPrefix(path, "~") {
		return path, nil
	}

	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	return filepath.Join(usr.HomeDir, path[1:]), nil
}
