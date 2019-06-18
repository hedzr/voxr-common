/*
 * Copyright © 2019 Hedzr Yeh.
 */

package cql

import "github.com/hedzr/voxr-common/db/dbi"

func (s *CassandraSession) GetCRUD(model dbi.DbModel, tableName string, pkColumns ...string) (crud dbi.DbCRUD) {
	x := &CassandraCRUD{tableName, pkColumns, model, s}
	return x
}

func (s *CassandraSession) Close() {
	if s.Instance != nil && !s.Instance.Closed() {
		s.Instance.Close()
	}
}
