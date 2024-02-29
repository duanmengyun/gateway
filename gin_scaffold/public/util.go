package public

import (
	"crypto/md5"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
)

func GenSaltpsw(info string, salt string) string {
	s1 := sha256.New()
	s1.Write([]byte(info))
	str1 := fmt.Sprintf("%x", s1.Sum(nil))
	s2 := sha256.New()
	s2.Write([]byte(str1 + salt))
	return fmt.Sprintf("%x", s1.Sum(nil))
}
func Obj2json(x any) string {
	bts, _ := json.Marshal(x)
	return string(bts)
}
func MD5(x string) string {
	h := md5.New()
	io.WriteString(h, x)
	return fmt.Sprintf("%x", h.Sum(nil))
}
