/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cql_test

import (
	"fmt"
	"github.com/hedzr/voxr-common/db"
	"github.com/satori/go.uuid"
	"testing"
	// _ set "github.com/deckarep/golang-set"
	"reflect"
	"unsafe"
)

func TestReflect(t *testing.T) {
	phrase1()
	setString()
	setUUID()
	setUUIDInStruct()
	setUUIDInStructEnh()
	// phrase2()
}

func phrase1() {
	u := db.Users2{
		Name: "dsds",
	}

	nr := reflect.TypeOf(u)
	// el := nr.Elem()
	fmt.Println("Type:", nr.Name(), "Kind:", nr.Kind()) // , "el", el)

	v := reflect.ValueOf(u) // 获取接口的值类型
	fmt.Println("Fields:")

	for i := 0; i < nr.NumField(); i++ { // NumField取出这个接口所有的字段数量
		f := nr.Field(i)                                   // 取得结构体的第i个字段
		val := v.Field(i).Interface()                      // 取得字段的值
		fmt.Printf("%11s: %v = %v\n", f.Name, f.Type, val) // 第i个字段的名称,类型,值
	}
	fmt.Println()
}

func setString() {
	s := "xxx * xxx"
	var sk string
	var i interface{} = &sk

	// svi: Value of *i (Type: stringKind)
	svi := reflect.Indirect(reflect.ValueOf(i))
	// ss: Value of *(*stringKind(&s))
	ss := reflect.Indirect(reflect.NewAt(svi.Type(),
		unsafe.Pointer(reflect.ValueOf(&s).Pointer())))
	// Now assign ss to svi
	svi.Set(ss)
	fmt.Printf("sk = %s\n", sk)
}

func setUUID() {
	s := uuid.NewV4()
	var sk uuid.UUID
	var i interface{} = &sk

	// svi: Value of *i (Type: stringKind)
	svi := reflect.Indirect(reflect.ValueOf(i))
	// ss: Value of *(*stringKind(&s))
	ss := reflect.Indirect(reflect.NewAt(svi.Type(),
		unsafe.Pointer(reflect.ValueOf(&s).Pointer())))
	// Now assign ss to svi
	svi.Set(ss)
	fmt.Printf("sk = %s\n", sk)
}

type UX struct {
	ID uuid.UUID
}

func setUUIDInStruct() {
	s := uuid.NewV4()
	ux := UX{}
	var i interface{} = &ux.ID

	// svi: Value of *i (Type: stringKind)
	svi := reflect.Indirect(reflect.ValueOf(i))
	// ss: Value of *(*stringKind(&s))
	ss := reflect.Indirect(reflect.NewAt(svi.Type(),
		unsafe.Pointer(reflect.ValueOf(&s).Pointer())))
	// Now assign ss to svi
	svi.Set(ss)
	fmt.Printf("sk = %s\n", ux.ID)
}

func setUUIDInStructEnh() {
	s := uuid.NewV4()

	ux := UX{}
	var i interface{} = &ux

	nr := reflect.Indirect(reflect.ValueOf(i))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		// f := el.Field(i)
		vf := reflect.Indirect(nr.FieldByName("ID"))
		ss := reflect.Indirect(reflect.NewAt(vf.Type(),
			unsafe.Pointer(reflect.ValueOf(&s).Pointer())))
		vf.Set(ss)
	}

	// // svi: Value of *i (Type: stringKind)
	// svi := reflect.Indirect(reflect.ValueOf(i))
	// // ss: Value of *(*stringKind(&s))
	// ss := reflect.Indirect(reflect.NewAt(svi.Type(),
	// 	unsafe.Pointer(reflect.ValueOf(&s).Pointer())))
	// // Now assign ss to svi
	// svi.Set(ss)
	fmt.Printf("ENH sk = %s\n", ux.ID)
}

func phrase2() {
	u := db.Users2{
		Name: "dsds",
	}

	nr := reflect.TypeOf(u)
	// el := nr.Elem()
	fmt.Println("Type:", nr.Name(), "Kind:", nr.Kind()) // , "el", el)

	v := reflect.ValueOf(u) // 获取接口的值类型
	vx := v.FieldByName("ID")
	fmt.Printf("vx -> ID = %v\n", vx.Interface())

	uuid := uuid.NewV4()
	var ui interface{}
	ui = uuid
	// vx.Set("00000000-1234-0000-1234-000000000000")
	fmt.Printf("using %v\n", ui)

	v.FieldByName("Name").Set(reflect.ValueOf("xxx"))
	// v.FieldByName("Name").SetString("yyy")
	fmt.Printf("set: %v", u.Name)

	// n := reflect.ValueOf(ui)
	// vx.Set(n.Interface())
	// fmt.Printf("set: %v", u.ID)
}
