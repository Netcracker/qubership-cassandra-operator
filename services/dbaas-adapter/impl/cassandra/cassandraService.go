// Copyright 2024-2025 NetCracker Technology Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package cassandra

import (
	"context"
	"fmt"
	"time"

	"github.com/Netcracker/qubership-dbaas-adapter-core/pkg/utils"
	"go.uber.org/zap"
)

var (
	dropAttempts   = 3
	dropRetryDelay = 2 * time.Second
)

type CassandraService interface {
	GrantPermissions(ctx context.Context, session Session, dbName, role string, permissions []string) error
	CreateRole(ctx context.Context, session Session, role, password string) error
	CreateKeyspace(ctx context.Context, session Session, dbName, settings string) error
	CreateTable(ctx context.Context, session Session, dbName, tableName string) error
	GetRolesForKeyspace(ctx context.Context, session Session, keyspace string) (map[string]bool, error)
	GetAllRoles(ctx context.Context, session Session) (map[string]bool, error)
	DropResource(ctx context.Context, session Session, resourceKind, name string) error
	EnsureMetadataTable(ctx context.Context, session Session, dbName string) error
	UpsertMetadataSetting(ctx context.Context, session Session, dbName, key string, value string) error
}

type CassandraServiceImpl struct {
	Logger *zap.Logger
}

var _ CassandraService = &CassandraServiceImpl{}

func (r *CassandraServiceImpl) executeStatement(ctx context.Context, session Session, stmt string) error {
	logger := utils.AddLoggerContext(r.Logger, ctx)
	logger.Debug(stmt)

	err := session.Query(stmt).Exec(false)
	return err
}

func (r *CassandraServiceImpl) EnsureMetadataTable(ctx context.Context, session Session, dbName string) error {
	cql := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s.metadata (
			id text PRIMARY KEY,
			value text
		)`, dbName)

	return r.executeStatement(ctx, session, cql)
}

func (r *CassandraServiceImpl) UpsertMetadataSetting(ctx context.Context, session Session, dbName, key string, value string) error {
	query := fmt.Sprintf(
		"INSERT INTO %s.metadata (id, value) VALUES (?, ?)",
		dbName,
	)

	return session.Query(query, key, string(value)).Exec(true)
}

func (r *CassandraServiceImpl) DropResource(ctx context.Context, session Session, resourceKind, name string) error {
	logger := utils.AddLoggerContext(r.Logger, ctx)
	stmt := fmt.Sprintf("drop %s if exists %s", resourceKind, name)

	var err error
	for attempt := 1; attempt <= dropAttempts; attempt++ {
		if err = r.executeStatement(ctx, session, stmt); err == nil {
			r.awaitSchemaAgreement(ctx, session, logger)
			return nil
		}
		logger.Warn(fmt.Sprintf("Attempt %d/%d to drop %s %s failed: %v", attempt, dropAttempts, resourceKind, name, err))
		if attempt == dropAttempts {
			break
		}
		r.awaitSchemaAgreement(ctx, session, logger)
		select {
		case <-ctx.Done():
			return err
		case <-time.After(dropRetryDelay):
		}
	}
	return err
}

// awaitSchemaAgreement waits for schema agreement. A timeout is only logged: the statement itself was accepted.
func (r *CassandraServiceImpl) awaitSchemaAgreement(ctx context.Context, session Session, logger *zap.Logger) {
	if err := session.AwaitSchemaAgreement(ctx); err != nil {
		logger.Warn(fmt.Sprintf("Schema agreement was not reached: %v", err))
	}
}

func (r *CassandraServiceImpl) CreateKeyspace(ctx context.Context, session Session, dbName, settings string) error {
	return r.executeStatement(ctx, session, fmt.Sprintf("create KEYSPACE %s WITH REPLICATION = %s",
		dbName, settings))
}

func (r *CassandraServiceImpl) CreateTable(ctx context.Context, session Session, dbName, tableName string) error {
	return r.executeStatement(ctx, session, fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s.%s(id text PRIMARY KEY, value text)", dbName, tableName))
}

func (r *CassandraServiceImpl) CreateRole(ctx context.Context, session Session, role, password string) error {
	return r.executeStatement(ctx, session, fmt.Sprintf("CREATE ROLE IF NOT EXISTS '%s' WITH PASSWORD = '%s' AND LOGIN = true", role, password))
}

func (r *CassandraServiceImpl) GrantPermissions(ctx context.Context, session Session, dbName, role string, permissions []string) error {
	logger := utils.AddLoggerContext(r.Logger, ctx)
	for _, permission := range permissions {
		stmt := fmt.Sprintf("GRANT %s on KEYSPACE %s to '%s'", permission, dbName, role)
		logger.Debug(stmt)
		err := session.Query(stmt).Exec(false)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *CassandraServiceImpl) GetRolesForKeyspace(ctx context.Context, session Session, keyspace string) (map[string]bool, error) {
	stmt := fmt.Sprintf("LIST ALL PERMISSIONS on keyspace %s", keyspace)
	return r.getRoles(ctx, session, stmt)
}
func (r *CassandraServiceImpl) GetAllRoles(ctx context.Context, session Session) (map[string]bool, error) {
	stmt := "LIST ROLES"
	return r.getRoles(ctx, session, stmt)
}

func (r *CassandraServiceImpl) getRoles(ctx context.Context, session Session, stmt string) (map[string]bool, error) {
	logger := utils.AddLoggerContext(r.Logger, ctx)
	logger.Debug(stmt)

	keyspaceRolesIterator := session.Query(stmt).Iter()
	roles := map[string]bool{}
	rd, err := keyspaceRolesIterator.RowData()
	if err != nil {
		return nil, err
	}
	for keyspaceRolesIterator.Scan(rd.RowData.Values...) {
		if iRole := rd.GetValue("role"); iRole != nil {
			roles[*iRole.(*string)] = true
		}
	}
	keyspaceRolesIterator.Close()

	return roles, nil
}
