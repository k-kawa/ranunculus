// Code generated by "stringer -type=ContextKey"; DO NOT EDIT

package constants

import "fmt"

const _ContextKey_name = "CtxCliContextCtxConfigCtxRedis"

var _ContextKey_index = [...]uint8{0, 13, 22, 30}

func (i ContextKey) String() string {
	if i < 0 || i >= ContextKey(len(_ContextKey_index)-1) {
		return fmt.Sprintf("ContextKey(%d)", i)
	}
	return _ContextKey_name[_ContextKey_index[i]:_ContextKey_index[i+1]]
}
