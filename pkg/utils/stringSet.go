package utils

import (
	"sort"
	"sync"
)

// StringSet 定义一个string的set
// 用来做字符串的"交并补"操作
type StringSet struct {
	m    map[string]struct{}
	lock sync.RWMutex
}

// NewStringSet 新建集合对象
func NewStringSet(items ...string) *StringSet {
	s := &StringSet{
		m: make(map[string]struct{}, len(items)),
	}
	s.Add(items...)
	return s
}

// Add 添加元素
func (s *StringSet) Add(items ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range items {
		s.m[v] = struct{}{}
	}
}

// Remove 删除元素
func (s *StringSet) Remove(items ...string) {
	s.lock.Lock()
	defer s.lock.Unlock()
	for _, v := range items {
		delete(s.m, v)
	}
}

// Has 判断元素是否存在
func (s *StringSet) Has(items ...string) bool {
	s.lock.RLock()
	defer s.lock.RUnlock()
	for _, v := range items {
		if _, ok := s.m[v]; !ok {
			return false
		}
	}
	return true
}

// Count 元素个数
func (s *StringSet) Count() int {
	return len(s.m)
}

// Clear 清空集合
func (s *StringSet) Clear() {
	s.lock.Lock()
	defer s.lock.Unlock()
	s.m = map[string]struct{}{}
}

// Empty 空集合判断
func (s *StringSet) Empty() bool {
	return len(s.m) == 0
}

// List 无序列表
func (s *StringSet) List() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	list := make([]string, 0, len(s.m))
	for item := range s.m {
		if item != "" {
			list = append(list, item)
		}
	}
	return list
}

// SortList 排序列表
func (s *StringSet) SortList() []string {
	s.lock.RLock()
	defer s.lock.RUnlock()
	list := make([]string, 0, len(s.m))
	for item := range s.m {
		if item != "" {
			list = append(list, item)
		}
	}
	sort.Strings(list)
	return list
}

// Union 并集
func (s *StringSet) Union(sets ...*StringSet) *StringSet {
	r := NewStringSet(s.List()...)
	for _, set := range sets {
		for e := range set.m {
			r.m[e] = struct{}{}
		}
	}
	return r
}

// Minus 差集
func (s *StringSet) Minus(sets ...*StringSet) *StringSet {
	r := NewStringSet(s.List()...)
	for _, set := range sets {
		for e := range set.m {
			if _, ok := s.m[e]; ok {
				delete(r.m, e)
			}
		}
	}
	return r
}

// Intersect 交集
func (s *StringSet) Intersect(sets ...*StringSet) *StringSet {
	r := NewStringSet(s.List()...)
	for _, set := range sets {
		for e := range s.m {
			if _, ok := set.m[e]; !ok {
				delete(r.m, e)
			}
		}
	}
	return r
}

// Complement 两个stringSet补集
func (s *StringSet) Complement(full *StringSet) *StringSet {
	r := NewStringSet()
	for e := range full.m {
		if _, ok := s.m[e]; !ok {
			r.Add(e)
		}
	}
	return r
}
