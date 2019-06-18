/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cql

import (
	"fmt"
	"github.com/gocql/gocql"
	"github.com/relops/cqlr"
	"golang.org/x/crypto/bcrypt"
	"reflect"
	"strings"
	"unsafe"
	// "github.com/hedzr/voxr-common/log"
	log "github.com/sirupsen/logrus"
)

func (s *CassandraCRUD) Insert() (rows int, err error) {
	var fieldNames, placeHolders []string
	var values []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		before_tag := f.Tag.Get(CQL_BEFORE_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		// found := false
		// for _, ss := range s.PrimaryKey {
		// 	if strings.EqualFold(ss, fieldName) {
		// 		found = true
		// 		break
		// 	}
		// }

		// if ! found {
		nf := nr.Field(i)
		// nt := nf.Type()
		k := nf.Kind()
		switch k {
		case reflect.Struct, reflect.Array, reflect.Slice:
			tag = f.Tag.Get(CQL_DEFAULT_VALUE_KEY)
			if len(tag) > 0 {
				placeHolders = append(placeHolders, tag)
			} else {
				placeHolders = append(placeHolders, "?")
				values = append(values, nf.Interface())
			}
			break
		default:
			value := nf.String()
			tag = f.Tag.Get(CQL_DEFAULT_VALUE_KEY)
			if len(tag) > 0 && len(value) == 0 {
				placeHolders = append(placeHolders, tag)
			} else {
				placeHolders = append(placeHolders, "?")
				if strings.EqualFold(before_tag, CQL_ENC_BCRYPT) {
					hashedPassword, err := bcrypt.GenerateFromPassword([]byte(nf.String()), bcrypt.DefaultCost)
					if err != nil {
						panic(err)
					}
					values = append(values, hashedPassword)
				} else {
					values = append(values, nf.Interface())
				}
			}
			break
		}

		fieldNames = append(fieldNames, fieldName)
		// }
	}

	cql := fmt.Sprintf("INSERT INTO %s (%s) VALUES (%s) IF NOT EXISTS", s.TableName,
		strings.Join(fieldNames, ","),
		strings.Join(placeHolders, ","))

	if err := s.Session.Instance.Query(cql, values...).Exec(); err != nil {
		log.Errorf("Insert Failed: %v", err)
		return 0, err
	}
	return 1, nil
}

func (s *CassandraCRUD) Update() (rows int, err error) {
	var fieldNames, placeHolders, pkConditions []string
	var values []interface{}
	var pkValues []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		// before_tag := f.Tag.Get(CQL_BEFORE_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		found := false
		for _, pk := range s.PrimaryKey {
			if strings.EqualFold(pk, fieldName) {
				found = true
				pkValues = append(pkValues, nr.Field(i).Interface())
				pkConditions = append(pkConditions, fmt.Sprintf("%s=?", fieldName))
				break
			}
		}

		if !found {
			nf := nr.Field(i)
			// nt := nf.Type()
			k := nf.Kind()
			switch k {
			case reflect.Struct:
				placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
				values = append(values, nf.Interface())
				break
			case reflect.Array, reflect.Slice:
				placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
				values = append(values, nf.Interface())
				break
			default:
				placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
				values = append(values, nf.Interface())
				break
			}

			fieldNames = append(fieldNames, fieldName)
		}
	}

	cql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", s.TableName,
		strings.Join(placeHolders, ","),
		strings.Join(pkConditions, " and "))

	values = append(values, pkValues...)
	if err := s.Session.Instance.Query(cql, values...).Exec(); err != nil {
		log.Errorf("Update Failed: %v", err)
		return 0, err
	}
	return 1, nil
}

func (s *CassandraCRUD) Get() (err error) {
	var pkConditions []string
	var pkValues []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		for _, pk := range s.PrimaryKey {
			if strings.EqualFold(pk, fieldName) {
				pkValues = append(pkValues, nr.Field(i).Interface())
				pkConditions = append(pkConditions, fmt.Sprintf("%s=?", fieldName))
				break
			}
		}
	}

	cql := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", s.TableName,
		strings.Join(pkConditions, " and "))

	// if err := s.Session.Instance.Query(cql, pkValues...).Consistency(gocql.One).Scan(&s.Model); err != nil {
	// 	log.Errorf("GetOne: SELECT Failed: %v", err)
	// 	return err
	// }
	// return nil

	q := s.Session.Instance.Query(cql, pkValues...).Consistency(gocql.One)
	// b := s.BindQuery(q)
	b := cqlr.BindQuery(q)
	for b.Scan(s.Model) {
		return nil
	}
	return fmt.Errorf("Get() Failed.")
}

func (s *CassandraCRUD) findGoName(fieldName string) (string, reflect.StructField, reflect.Value) {
	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		v := strings.Split(tag, ",")
		dbFieldName := f.Name
		if len(v) > 0 {
			dbFieldName = v[0]
		}
		if strings.EqualFold(dbFieldName, fieldName) {
			return f.Name, f, nr.Field(i)
		}
	}
	return "", reflect.StructField{}, reflect.Value{}
}

func (s *CassandraCRUD) Delete() bool {
	// var pkConditions, values []string
	// vl := reflect.ValueOf(s.Model).Elem()
	var pkConditions []string
	var values []interface{}
	for i := 0; i < len(s.PrimaryKey); i++ {
		pk := s.PrimaryKey[i]
		goName, _, goValue := s.findGoName(pk)
		if len(goName) == 0 {
			panic("primary key name cannot be found.")
		}
		pkConditions = append(pkConditions, fmt.Sprintf("%s=?", goName))
		value := goValue.Interface() // vl.String()
		values = append(values, value)
	}

	cql := fmt.Sprintf("DELETE FROM %s WHERE %s", s.TableName,
		strings.Join(pkConditions, " and "))

	if err := s.Session.Instance.Query(cql, values).Exec(); err != nil {
		log.Errorf("DELETE Failed: %v", err)
		return false
	}
	return true
}

func (s *CassandraCRUD) Exists() bool {
	var pkConditions []string
	var values []interface{}
	// //// vl := reflect.ValueOf(s.Model).Elem()
	// nr := reflect.Indirect(reflect.ValueOf(s.Model))
	// el := nr.Type()
	for i := 0; i < len(s.PrimaryKey); i++ {
		pk := s.PrimaryKey[i]

		goName, _, goValue := s.findGoName(pk)
		// goName := ""
		// var goValue reflect.Value
		// for i := 0; i < el.NumField(); i++ {
		// 	f := el.Field(i)
		// 	tag := f.Tag.Get(CQL_KEY)
		// 	v := strings.Split(tag, ",")
		// 	dbFieldName := f.Name
		// 	if len(v) > 0 {
		// 		dbFieldName = v[0]
		// 	}
		// 	if strings.EqualFold(dbFieldName, pk) {
		// 		goName = f.Name
		// 		goValue = nr.Field(i)
		// 		break
		// 	}
		// }

		if len(goName) == 0 {
			panic("primary key name cannot be found.")
		}

		pkConditions = append(pkConditions, fmt.Sprintf("%s=?", goName))
		value := goValue.Interface()
		values = append(values, value)
	}

	return s.ExistsBy(strings.Join(pkConditions, " and "), values...)
}

func (s *CassandraCRUD) ExistsBy(conditions string, params ...interface{}) bool {
	var pkConditions []string
	var values []interface{}

	if conditions != "" {
		pkConditions = append(pkConditions, conditions)
		for _, p := range params {
			values = append(values, p)
		}
	}

	cql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", s.TableName,
		strings.Join(pkConditions, " and "))

	i := 0
	if s.Session.Instance.Query(cql, values...).Iter().Scan(&i) {
		// log.Errorf("ExistsWith() Failed: %v", err)
		return i > 0
	}
	return false
}

func (s *CassandraCRUD) ExistsWith(conditionFields ...string) bool {
	var conditions []string
	var condValues []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		// before_tag := f.Tag.Get(CQL_BEFORE_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		// found := false
		if len(conditionFields) == 0 {
			for _, pk := range s.PrimaryKey {
				if strings.EqualFold(pk, fieldName) {
					// found = true
					condValues = append(condValues, nr.Field(i).Interface())
					conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
					break
				}
			}
		}
		for _, condition := range conditionFields {
			if strings.EqualFold(condition, fieldName) {
				// found = true
				condValues = append(condValues, nr.Field(i).Interface())
				conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
			}
		}
	}

	cql := fmt.Sprintf("SELECT COUNT(*) FROM %s WHERE %s", s.TableName,
		strings.Join(conditions, " and "))

	i := 0
	if s.Session.Instance.Query(cql, condValues...).Iter().Scan(&i) {
		return i > 0
	}

	return false
}

func (s *CassandraCRUD) UpdateWith(conditionFields ...string) bool {
	var placeHolders, conditions []string
	var values, condValues []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		// before_tag := f.Tag.Get(CQL_BEFORE_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		found := false
		pkFound := false
		if len(conditionFields) == 0 {
			for _, pk := range s.PrimaryKey {
				if strings.EqualFold(pk, fieldName) {
					found = true
					pkFound = true
					condValues = append(condValues, nr.Field(i).Interface())
					conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
					break
				}
			}
		} else {
			for _, pk := range s.PrimaryKey {
				if strings.EqualFold(pk, fieldName) {
					pkFound = true
					break
				}
			}
			for _, condition := range conditionFields {
				if strings.EqualFold(condition, fieldName) {
					found = true
					condValues = append(condValues, nr.Field(i).Interface())
					conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
				}
			}
		}

		if !found {
			if len(conditionFields) != 0 && pkFound {
				found = false
			} else {
				nf := nr.Field(i)
				// nt := nf.Type()
				k := nf.Kind()
				switch k {
				case reflect.Struct:
					placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
					values = append(values, nf.Interface())
					break
				case reflect.Array, reflect.Slice:
					placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
					values = append(values, nf.Interface())
					break
				default:
					placeHolders = append(placeHolders, fmt.Sprintf("%s=?", fieldName))
					values = append(values, nf.Interface())
					break
				}
			}
		}
	}

	cql := fmt.Sprintf("UPDATE %s SET %s WHERE %s", s.TableName,
		strings.Join(placeHolders, ","),
		strings.Join(conditions, " and "))

	for _, v := range condValues {
		values = append(values, v)
	}

	if err := s.Session.Instance.Query(cql, values...).Exec(); err != nil {
		log.Errorf("UpdateWith() Failed: %v", err)
		return false
	}

	return s.GetOne(strings.Join(conditions, " and "), condValues...)
	// return true
}

func (s *CassandraCRUD) DeleteWith(conditionFields ...string) bool {
	var conditions []string
	var condValues []interface{}

	nr := reflect.Indirect(reflect.ValueOf(s.Model))
	el := nr.Type()
	for i := 0; i < el.NumField(); i++ {
		f := el.Field(i)
		tag := f.Tag.Get(CQL_KEY)
		// before_tag := f.Tag.Get(CQL_BEFORE_KEY)
		v := strings.Split(tag, ",")
		fieldName := f.Name
		if len(v) > 0 {
			fieldName = v[0]
		}

		// found := false
		if len(conditionFields) == 0 {
			for _, pk := range s.PrimaryKey {
				if strings.EqualFold(pk, fieldName) {
					// found = true
					condValues = append(condValues, nr.Field(i).Interface())
					conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
					break
				}
			}
		}
		for _, condition := range conditionFields {
			if strings.EqualFold(condition, fieldName) {
				// found = true
				condValues = append(condValues, nr.Field(i).Interface())
				conditions = append(conditions, fmt.Sprintf("%s=?", fieldName))
			}
		}
	}

	cql := fmt.Sprintf("DELETE FROM %s WHERE %s", s.TableName,
		strings.Join(conditions, " and "))

	if err := s.Session.Instance.Query(cql, condValues...).Exec(); err != nil {
		log.Errorf("DeleteWith() Failed: %v", err)
		return false
	}

	return true
}

func (s *CassandraCRUD) InsertOrUpdate(conditions ...string) bool {
	if s.ExistsWith(conditions...) {
		return s.UpdateWith(conditions...)
	}

	rows, _ := s.Insert()
	return rows == 1
}

// TODO 这个Scanner在反射uuid时遇到问题，暂时未能完成
type Scanner interface {
	Scan(target interface{}) error
}

type XqlScanner struct {
	Query *gocql.Query
}

func (b *XqlScanner) Scan(target interface{}) error {
	iter := b.Query.Iter()
	row := make(map[string]interface{})
	var err error
	if iter.MapScan(row) {
		nr := reflect.Indirect(reflect.ValueOf(target))
		el := nr.Type()

		for i := 0; i < el.NumField(); i++ {
			f := el.Field(i)
			tag := f.Tag.Get(CQL_KEY)
			v := strings.Split(tag, ",")
			dbFieldName := f.Name
			if len(v) > 0 {
				dbFieldName = v[0]
			}

			if val, ok := row[dbFieldName]; ok {
				k := f.Type.Kind()
				switch k {
				case reflect.Bool:
					nr.Field(i).SetBool(val.(bool))
					break
				case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
					nr.Field(i).SetInt(val.(int64))
					break
				case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
					nr.Field(i).SetUint(val.(uint64))
					break
				case reflect.Uintptr:
					err = fmt.Errorf("[WARN] unknown UintPtr type found. field: %s", f.Name)
					break
				case reflect.Float32, reflect.Float64:
					nr.Field(i).SetFloat(val.(float64))
					break
				case reflect.Complex64, reflect.Complex128:
					nr.Field(i).SetComplex(val.(complex128))
					break
				case reflect.Array:
					// src := reflect.ValueOf(&val)
					dst := reflect.Indirect(nr.FieldByName(f.Name))
					dp := reflect.Indirect(reflect.NewAt(dst.Type(), unsafe.Pointer(dst.Pointer())))
					ss := reflect.Indirect(reflect.NewAt(dst.Type(), unsafe.Pointer(reflect.ValueOf(&val).Pointer())))
					dp.Set(ss)
					// err = fmt.Errorf("[WARN] unknown Chan type found. field: %s", f.Name)
					break
				case reflect.Chan:
					err = fmt.Errorf("[WARN] unknown Chan type found. field: %s", f.Name)
					break
				case reflect.Func:
					err = fmt.Errorf("[WARN] unknown Func type found. field: %s", f.Name)
					break
				case reflect.Interface:
					err = fmt.Errorf("[WARN] unknown Interface type found. field: %s", f.Name)
					break
				case reflect.Map:
					err = fmt.Errorf("[WARN] unknown Map type found. field: %s", f.Name)
					break
				case reflect.Ptr:
					err = fmt.Errorf("[WARN] unknown Ptr type found. field: %s", f.Name)
					break
				case reflect.Slice:
					err = fmt.Errorf("[WARN] unknown Slice type found. field: %s", f.Name)
					break
				case reflect.String:
					nr.Field(i).SetString(val.(string))
					break
				case reflect.Struct:
					err = fmt.Errorf("[WARN] unknown Struct type found. field: %s", f.Name)
					break
				default:
					err = fmt.Errorf("[WARN] unknown type (kind=%d) found. field: %s", k, f.Name)
					break
				}
			}
		}
		return err
	}

	// rowData, _ := iter.RowData()
	//
	// nr := reflect.Indirect(reflect.ValueOf(target))
	// el := nr.Type()
	// var err error
	//
	// for iCol, col := range rowData.Columns {
	// 	for i := 0; i < el.NumField(); i++ {
	// 		f := el.Field(i)
	// 		tag := f.Tag.Get(CQL_KEY)
	// 		v := strings.Split(tag, ",")
	// 		dbFieldName := f.Name
	// 		if len(v) > 0 {
	// 			dbFieldName = v[0]
	// 		}
	// 		if strings.EqualFold(dbFieldName, col) {
	// 			rowData.Values[iCol] = nr.Field(i)
	// 			//k := f.Type.Kind()
	// 			//switch k {
	// 			//case reflect.Bool:
	// 			//	nr.Field(i).SetBool(val.(bool))
	// 			//	break
	// 			//case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
	// 			//	nr.Field(i).SetInt(val.(int64))
	// 			//	break
	// 			//case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
	// 			//	nr.Field(i).SetUint(val.(uint64))
	// 			//	break
	// 			//case reflect.Uintptr:
	// 			//	err = fmt.Errorf("[WARN] unknown UintPtr type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Float32, reflect.Float64:
	// 			//	nr.Field(i).SetFloat(val.(float64))
	// 			//	break
	// 			//case reflect.Complex64, reflect.Complex128:
	// 			//	nr.Field(i).SetComplex(val.(complex128))
	// 			//	break
	// 			//case reflect.Array:
	// 			//	err = fmt.Errorf("[WARN] unknown Chan type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Chan:
	// 			//	err = fmt.Errorf("[WARN] unknown Chan type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Func:
	// 			//	err = fmt.Errorf("[WARN] unknown Func type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Interface:
	// 			//	err = fmt.Errorf("[WARN] unknown Interface type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Map:
	// 			//	err = fmt.Errorf("[WARN] unknown Map type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Ptr:
	// 			//	err = fmt.Errorf("[WARN] unknown Ptr type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.Slice:
	// 			//	err = fmt.Errorf("[WARN] unknown Slice type found. field: %s", f.Name)
	// 			//	break
	// 			//case reflect.String:
	// 			//	nr.Field(i).SetString(val.(string))
	// 			//	break
	// 			//case reflect.Struct:
	// 			//	fmt.Printf("[WARN] unknown Struct type found. field: %s", f.Name)
	// 			//	break
	// 			//default:
	// 			//	fmt.Printf("[WARN] unknown type (kind=%d) found. field: %s", k, f.Name)
	// 			//	break
	// 			//}
	// 		}
	// 	}
	// }
	//
	// if iter.Scan(rowData.Values...) {
	// 	//rowData.rowMap(m)
	// 	return nil
	// }

	// if err != nil {
	// 	return fmt.Errorf("gocql.Iter.MapScan failed to target: %v.", target)
	// }else{
	// 	return nil
	// }
	return fmt.Errorf("gocql.Iter.MapScan failed to target: %v.", target)
}

func (s *CassandraCRUD) BindQuery(q *gocql.Query) Scanner {
	return &XqlScanner{q}
}

func (s *CassandraCRUD) GetOne(conditions string, params ...interface{}) bool {
	var pkConditions []string
	var values []interface{}

	if conditions != "" {
		pkConditions = append(pkConditions, conditions)
		for _, p := range params {
			values = append(values, p)
		}
	}

	cql := fmt.Sprintf("SELECT * FROM %s WHERE %s LIMIT 1", s.TableName,
		strings.Join(pkConditions, " and "))

	q := s.Session.Instance.Query(cql, values...).Consistency(gocql.One)
	// b := s.BindQuery(q)
	b := cqlr.BindQuery(q)
	for b.Scan(s.Model) {
		// log.Errorf("GetOne() Failed: %v", err)
		// return false
		return true
	}
	// b.Close()
	return false
}

func (s *CassandraCRUD) Query(results *[]interface{}, conditions string, params ...interface{}) *[]interface{} {
	var pkConditions []string
	var values []interface{}

	if conditions != "" {
		pkConditions = append(pkConditions, conditions)
		for _, p := range params {
			values = append(values, p)
		}
	}

	cql := fmt.Sprintf("SELECT * FROM %s WHERE %s", s.TableName,
		strings.Join(pkConditions, " and "))

	if s.Session.Instance.Query(cql, values...).Iter().Scan(results) {
		// log.Errorf("Query() Failed: %v", err)
		return results
	}
	return nil
}

func (s *CassandraCRUD) DeleteBy(conditions string, params ...interface{}) bool {
	var pkConditions []string
	var values []interface{}

	if conditions != "" {
		pkConditions = append(pkConditions, conditions)
		for _, p := range params {
			values = append(values, p)
		}
	}

	cql := fmt.Sprintf("DELETE FROM %s WHERE %s", s.TableName,
		strings.Join(pkConditions, " and "))

	if err := s.Session.Instance.Query(cql, values...).Exec(); err != nil {
		log.Errorf("DeleteWith() Failed: %v", err)
		return false
	}
	return true
}
