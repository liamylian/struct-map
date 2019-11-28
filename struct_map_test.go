package struct_map

import (
	"github.com/stretchr/testify/assert"
	"strconv"
	"sync"
	"testing"
)

type Book struct {
	Id     string `struct_map:"pk"`
	Title  string `struct_map:"index"`
	Desc   string
	Author string `struct_map:"index"`
}

var (
	book1 = &Book{
		Id:     "1",
		Title:  "Jane Eyre",
		Author: "Charlotte Brontë",
	}

	book2 = &Book{
		Id:     "2",
		Title:  "Gone With the Wind",
		Author: "Margaret Mitchell",
	}
	book3 = &Book{
		Id:     "3",
		Title:  "Pride and Prejudice",
		Author: "Jane Austen",
	}
)

func TestMapper_NewMapper(t *testing.T) {
	m, err := NewMapper(&Book{})
	assert.Nil(t, err)

	assert.Equal(t, "Id", m.pkName)
	assert.Equal(t, map[string]interface{}{}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title":  {m: map[string]map[string]struct{}{}},
		"Author": {m: map[string]map[string]struct{}{}},
	}, m.indexes)
}

func TestMapper_NewMapper_Errors(t *testing.T) {
	t.Log("not struct")
	notStruct := make(map[string]string)
	_, err := NewMapper(notStruct)
	assert.Equal(t, ErrNotStruct, err)

	t.Log("no pk")
	noPk := struct {
		Id    string
		Title string `struct_map:"index"`
	}{}
	_, err = NewMapper(noPk)
	assert.Equal(t, ErrPk, err)

	t.Log("many pk")
	manyPk := struct {
		Id    string `struct_map:"pk"`
		Title string `struct_map:"pk"`
	}{}
	_, err = NewMapper(manyPk)
	assert.Equal(t, ErrPk, err)

	t.Log("pk not string")
	pkNotString := struct {
		Id    int    `struct_map:"pk"`
		Title string `struct_map:"index"`
	}{}
	_, err = NewMapper(pkNotString)
	assert.Equal(t, ErrPkIndexNotString, err)

	t.Log("index not string")
	indexNotString := struct {
		Id    string  `struct_map:"pk"`
		Title *string `struct_map:"index"`
	}{}
	_, err = NewMapper(indexNotString)
	assert.Equal(t, ErrPkIndexNotString, err)
}

func TestMapper_Add(t *testing.T) {
	m, err := NewMapper(&Book{})
	assert.Nil(t, err)

	m.Add(book1)
	assert.Equal(t, map[string]interface{}{
		book1.Id: book1,
	}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title": {m: map[string]map[string]struct{}{
			book1.Title: {book1.Id: struct{}{}},
		}},
		"Author": {m: map[string]map[string]struct{}{
			book1.Author: {book1.Id: struct{}{}},
		}},
	}, m.indexes)

	m.Add(book2)
	assert.Equal(t, map[string]interface{}{
		book1.Id: book1,
		book2.Id: book2,
	}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title": {m: map[string]map[string]struct{}{
			book1.Title: {book1.Id: struct{}{}},
			book2.Title: {book2.Id: struct{}{}},
		}},
		"Author": {m: map[string]map[string]struct{}{
			book1.Author: {book1.Id: struct{}{}},
			book2.Author: {book2.Id: struct{}{}},
		}},
	}, m.indexes)

	m.Add(book3)
	assert.Equal(t, map[string]interface{}{
		book1.Id: book1,
		book2.Id: book2,
		book3.Id: book3,
	}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title": {m: map[string]map[string]struct{}{
			book1.Title: {book1.Id: struct{}{}},
			book2.Title: {book2.Id: struct{}{}},
			book3.Title: {book3.Id: struct{}{}},
		}},
		"Author": {m: map[string]map[string]struct{}{
			book1.Author: {book1.Id: struct{}{}},
			book2.Author: {book2.Id: struct{}{}},
			book3.Author: {book3.Id: struct{}{}},
		}},
	}, m.indexes)
}

func TestMapper_Add_Error(t *testing.T) {
	book4 := &struct {
		Id    string  `struct_map:"pk"`
		Title *string `struct_map:"index"`
	}{}
	mapper, err := NewMapper(&Book{})
	assert.Nil(t, err)

	err = mapper.Add(book4)
	assert.Equal(t, ErrDifferentType, err)
}

func TestMapper_Remove(t *testing.T) {
	m, err := NewMapper(&Book{})
	assert.Nil(t, err)

	m.Add(book1)
	m.Add(book2)
	m.Add(book3)

	m.Remove(book3)
	assert.Equal(t, map[string]interface{}{
		book1.Id: book1,
		book2.Id: book2,
	}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title": {m: map[string]map[string]struct{}{
			book1.Title: {book1.Id: struct{}{}},
			book2.Title: {book2.Id: struct{}{}},
		}},
		"Author": {m: map[string]map[string]struct{}{
			book1.Author: {book1.Id: struct{}{}},
			book2.Author: {book2.Id: struct{}{}},
		}},
	}, m.indexes)

	m.Remove(book2)
	assert.Equal(t, map[string]interface{}{
		book1.Id: book1,
	}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title": {m: map[string]map[string]struct{}{
			book1.Title: {book1.Id: struct{}{}},
		}},
		"Author": {m: map[string]map[string]struct{}{
			book1.Author: {book1.Id: struct{}{}},
		}},
	}, m.indexes)

	m.Remove(book1)
	assert.Equal(t, map[string]interface{}{}, m.pk.m)
	assert.Equal(t, map[string]*index{
		"Title":  {m: map[string]map[string]struct{}{}},
		"Author": {m: map[string]map[string]struct{}{}},
	}, m.indexes)
}

func TestMapper_Remove_Error(t *testing.T) {
	book4 := &struct {
		Id    string  `struct_map:"pk"`
		Title *string `struct_map:"index"`
	}{}
	mapper, err := NewMapper(&Book{})
	assert.Nil(t, err)

	err = mapper.Remove(book4)
	assert.Equal(t, ErrDifferentType, err)
}

func TestMapper_Get(t *testing.T) {
	m, err := NewMapper(&Book{})
	assert.Nil(t, err)

	m.Add(book1)
	m.Add(book2)
	m.Add(book3)
	books := m.Get("Id", book1.Id)
	assert.Equal(t, []interface{}{book1}, books)
	books = m.Get("Title", book2.Title)
	assert.Equal(t, []interface{}{book2}, books)
	books = m.Get("Author", book3.Author)
	assert.Equal(t, []interface{}{book3}, books)
}

func TestMapper_Concurrent(t *testing.T) {
	m, err := NewMapper(&Book{})
	assert.Nil(t, err)

	wg := sync.WaitGroup{}
	ch := make(chan *Book, 10)
	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				book := &Book{
					Id: strconv.Itoa(i*100 + j),
				}

				err := m.Add(book)
				assert.Nil(t, err)
				ch <- book

				wg.Add(1)
			}
		}()
	}

	for i := 0; i < 10; i++ {
		go func() {
			for j := 0; j < 100; j++ {
				book := <-ch
				m.Get("Id", book.Id)
				m.Get("Title", book.Title)
				m.Get("Author", book.Author)
				err := m.Remove(book)
				assert.Nil(t, err)

				wg.Add(-1)
			}
		}()
	}

	wg.Wait()
}
