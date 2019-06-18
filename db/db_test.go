/*
 * Copyright © 2019 Hedzr Yeh.
 */

package db_test

import (
	"fmt"
	"github.com/hedzr/voxr-common/db"
	"testing"
	// _ set "github.com/deckarep/golang-set"
	// "github.com/satori/go.uuid"
	// "github.com/gocql/gocql"
	"github.com/hedzr/voxr-common/db/cql"
	"github.com/hedzr/voxr-common/db/dbi"
)

func TestCRUD(t *testing.T) {

	backend := cql.New(&dbi.ConfigItem{
		Hosts:                    []string{"127.0.0.1:9042"},
		Database:                 "test",
		DisableInitialHostLookup: true,
	})
	// fmt.Printf("backend is : %v", backend)

	session := backend.Open()
	if session == nil {
		panic("cannot connect with cassandra servers.")
	}
	defer session.Close()

	// uuid0 := uuid.NewV4()
	// uuid00, err := gocql.UUIDFromBytes(uuid0.Bytes())
	// if err != nil {
	// 	panic(err)
	// }
	u0 := db.Users2{
		// ID:        uuid00,
		Name:      "admin",
		Email:     "admin@example.com",
		Mobile:    "13012345678",
		Password:  "admin",
		Blocked:   false,
		Forbidden: false,
		Deleted:   false,
	}

	crud0 := session.GetCRUD(&u0, "users", "name")
	if crud0.InsertOrUpdate("name") {
		fmt.Printf("crud0 insertOrUpdate ok. name=%s, create_at=%v\n", u0.Name, u0.CreatedTime)
	} else {
		t.Fatalf("crud0 insetOtUpdate failed. see app log for more information.")
	}

	// uuid1 := uuid.NewV4()
	// uuid, err := gocql.UUIDFromBytes(uuid1.Bytes())
	// if err != nil {
	// 	t.Fatal(err)
	// 	//panic(err)
	// }
	u := db.Users2{
		// ID:        uuid,
		Name:      "admin2",
		Password:  "dush9,ever!01x",
		Blocked:   false,
		Forbidden: false,
		Deleted:   false,
	}
	// fmt.Printf("using uuid %s\n", uuid1.String())

	crud := session.GetCRUD(&u, "users", "name")

	if crud.InsertOrUpdate("name") {
		fmt.Printf("crud0 insertOrUpdate ok. name=%s\n", u0.Name)
	} else {
		t.Fatalf("crud0 insetOtUpdate failed. see app log for more information.")
	}

	// if ! crud.Exists() {
	// 	crud.Insert()
	// }

	if !crud.GetOne("name=?", u.Name) {
		t.Fatalf("cannot GetOne().")
	}
	// fmt.Printf("Got Users2: id=%s\n", u.ID)

	u.Email = "admin2@example.com"
	if rows, err := crud.Update(); err != nil {
		t.Fatalf("CANNOT UPDATE. ROWS=%d, ERR: %v", rows, err)
	}

	u.Name = ""
	u.Email = ""
	u.Mobile = ""
	if err := crud.Get(); err != nil {
		t.Fatalf("CANNOT GET. ERR: %v", err)
	} else {
		fmt.Printf("Got Users2: name=%s\n", u.Name)
	}

	u2 := db.Users2{}
	crud2 := session.GetCRUD(&u2, "users")
	if crud2.GetOne("name=?", "admin1") {
		crud2.Delete()
	} else {
		fmt.Printf("user '%s' NOT FOUND. Nothing to do.", "admin1")
	}

}
