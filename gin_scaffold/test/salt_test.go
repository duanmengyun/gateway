package test

import (
	"fmt"
	"gin_scaffold/public"
	"testing"
)

// 修改后的测试函数
func TestSalt(t *testing.T) {
	// 测试逻辑，确保将 *testing.T 作为参数传递
	result := public.GenSaltpsw("123456", "admin")
	fmt.Print(result)
	if result != "8d969eef6ecad3c29a3a629280e686cf0c3f5d5a86aff3ca12020c923adc6c92" {
		t.Errorf("wrong")
	}
}
