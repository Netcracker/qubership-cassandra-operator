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
