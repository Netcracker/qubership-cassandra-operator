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

package cassandra_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Netcracker/qubership-cassandra-dbaas-adapter/go/impl/cassandra"
	"github.com/Netcracker/qubership-cassandra-dbaas-adapter/go/impl/cassandra/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

const dropStmt = "drop keyspace if exists ks"

func newService() *cassandra.CassandraServiceImpl {
	return &cassandra.CassandraServiceImpl{Logger: zap.NewNop()}
}

func TestDropResourceAwaitsSchemaAgreementOnSuccess(t *testing.T) {
	cassandra.SetDropRetryForTest(3, time.Millisecond)
	session := mocks.NewSession(t)
	query := mocks.NewQueryInterface(t)
	session.On("Query", dropStmt).Return(query).Once()
	query.On("Exec", false).Return(nil).Once()
	session.On("AwaitSchemaAgreement", mock.Anything).Return(nil).Once()

	assert.NoError(t, newService().DropResource(context.Background(), session, "keyspace", "ks"))
}

func TestDropResourceRetriesAfterFailure(t *testing.T) {
	cassandra.SetDropRetryForTest(3, time.Millisecond)
	session := mocks.NewSession(t)
	query := mocks.NewQueryInterface(t)
	session.On("Query", dropStmt).Return(query).Twice()
	query.On("Exec", false).Return(errors.New("java.lang.IllegalArgumentException: Unknown CF")).Once()
	query.On("Exec", false).Return(nil).Once()
	// once between attempts, once after the successful drop
	session.On("AwaitSchemaAgreement", mock.Anything).Return(nil).Twice()

	assert.NoError(t, newService().DropResource(context.Background(), session, "keyspace", "ks"))
}

func TestDropResourceReturnsLastErrorAfterAllAttempts(t *testing.T) {
	cassandra.SetDropRetryForTest(3, time.Millisecond)
	session := mocks.NewSession(t)
	query := mocks.NewQueryInterface(t)
	session.On("Query", dropStmt).Return(query).Times(3)
	query.On("Exec", false).Return(errors.New("boom")).Times(3)
	// only between attempts, not after the last one
	session.On("AwaitSchemaAgreement", mock.Anything).Return(nil).Twice()

	assert.EqualError(t, newService().DropResource(context.Background(), session, "keyspace", "ks"), "boom")
}

func TestDropResourceSucceedsWhenSchemaAgreementTimesOut(t *testing.T) {
	cassandra.SetDropRetryForTest(3, time.Millisecond)
	session := mocks.NewSession(t)
	query := mocks.NewQueryInterface(t)
	session.On("Query", dropStmt).Return(query).Once()
	query.On("Exec", false).Return(nil).Once()
	session.On("AwaitSchemaAgreement", mock.Anything).Return(errors.New("schema versions not consistent")).Once()

	assert.NoError(t, newService().DropResource(context.Background(), session, "keyspace", "ks"))
}
