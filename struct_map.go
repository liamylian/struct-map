package struct_map

import (
	"errors"
	"reflect"
	"sync"
)

const (
	tagName        = "struct_map"
	structMapPk    = "pk"
	structMapIndex = "index"
)

var (
	ErrPk               = errors.New("pk must only be one")
	ErrPkIndexNotString = errors.New("pk or index must be string")
	ErrNotStruct        = errors.New("not struct or pointer to struct")
	ErrDifferentType    = errors.New("different type error")
)

type pk struct {
	mu sync.RWMutex
	m  map[string]interface{} // pk -> struct
}

type index struct {
	mu sync.RWMutex
	m  map[string]map[string]struct{} // index -> pks
}

type Mapper struct {
	typ    reflect.Type
	eleTyp reflect.Type

	pkName  string
	pk      *pk
	indexes map[string]*index // index name -> index
}

func NewMapper(obj interface{}) (*Mapper, error) {
	typ := reflect.TypeOf(obj)
	eleTyp := typ
	if typ.Kind() == reflect.Ptr {
		eleTyp = typ.Elem()
	}
	if eleTyp.Kind() != reflect.Struct {
		return nil, ErrNotStruct
	}

	mapper := &Mapper{
		typ:    typ,
		eleTyp: eleTyp,
		pk: &pk{
			m: make(map[string]interface{}),
		},
		indexes: make(map[string]*index),
	}

	pkCount := 0
	fieldNum := eleTyp.NumField()
	for i := 0; i < fieldNum; i ++ {
		field := eleTyp.Field(i)
		tagVal := field.Tag.Get(tagName)
		if !(tagVal == structMapPk || tagVal == structMapIndex) {
			continue
		}
		if field.Type.Kind() != reflect.String {
			return nil, ErrPkIndexNotString
		}

		switch tagVal {
		case structMapPk:
			mapper.pkName = field.Name
			pkCount++
		case structMapIndex:
			mapper.indexes[field.Name] = &index{
				m: make(map[string]map[string]struct{}),
			}
		}
	}

	if pkCount != 1 {
		return nil, ErrPk
	}

	return mapper, nil
}

func (m *Mapper) Add(obj interface{}) error {
	typ := reflect.TypeOf(obj)
	if typ != m.typ {
		return ErrDifferentType
	}
	val := reflect.ValueOf(obj)
	if m.typ != m.eleTyp {
		val = val.Elem()
	}

	pkVal := val.FieldByName(m.pkName)
	pk := pkVal.String()
	m.addByPk(pk, obj)

	for indexName := range m.indexes {
		indexVal := val.FieldByName(indexName)
		m.addByIndex(indexName, indexVal.String(), pk)
	}

	return nil
}

func (m *Mapper) Remove(obj interface{}) error {
	typ := reflect.TypeOf(obj)
	if typ != m.typ {
		return ErrDifferentType
	}
	val := reflect.ValueOf(obj)
	if m.typ != m.eleTyp {
		val = val.Elem()
	}

	pkVal := val.FieldByName(m.pkName)
	pk := pkVal.String()
	m.removeByPk(pk)

	for indexName := range m.indexes {
		indexVal := val.FieldByName(indexName)
		m.removeByIndex(indexName, indexVal.String(), pk)
	}

	return nil
}

func (m *Mapper) Get(fieldName, val string) []interface{} {
	if fieldName == m.pkName {
		if obj, ok := m.getByPk(val); ok {
			return []interface{}{obj}
		} else {
			return nil
		}
	}

	for indexName := range m.indexes {
		if fieldName == indexName {
			return m.getByIndex(indexName, val)
		}
	}

	return nil
}

func (m *Mapper) getByPk(pk string) (interface{}, bool) {
	m.pk.mu.RLock()
	defer m.pk.mu.RUnlock()

	obj, ok := m.pk.m[pk]
	return obj, ok
}

func (m *Mapper) addByPk(pk string, obj interface{}) {
	m.pk.mu.Lock()
	defer m.pk.mu.Unlock()
	m.pk.m[pk] = obj
}

func (m *Mapper) removeByPk(pk string) {
	m.pk.mu.Lock()
	defer m.pk.mu.Unlock()
	delete(m.pk.m, pk)
}

func (m *Mapper) getByIndex(indexName, index string) (objs []interface{}) {
	idx := m.indexes[indexName]

	idx.mu.RLock()
	pks, ok := idx.m[index]
	idx.mu.RUnlock()
	if !ok {
		return
	}

	for pk := range pks {
		obj, ok := m.getByPk(pk)
		if ok {
			objs = append(objs, obj)
		}
	}

	return
}

func (m *Mapper) addByIndex(indexName, index, pk string) {
	idx := m.indexes[indexName]

	idx.mu.Lock()
	defer idx.mu.Unlock()
	if _, ok := idx.m[index]; !ok {
		idx.m[index] = make(map[string]struct{})
	}
	idx.m[index][pk] = struct{}{}
}

func (m *Mapper) removeByIndex(indexName, index, pk string) {
	idx := m.indexes[indexName]

	idx.mu.Lock()
	defer idx.mu.Unlock()
	if _, ok := idx.m[index]; !ok {
		return
	}
	delete(idx.m[index], pk)

	if len(idx.m[index]) == 0 {
		delete(idx.m, index)
	}
}
