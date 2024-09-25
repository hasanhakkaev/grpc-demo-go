package service

import (
	"context"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	v1 "github.com/hasanhakkaev/yqapp-demo/api/tasks/v1"
	"github.com/hasanhakkaev/yqapp-demo/internal/database"
	"github.com/hasanhakkaev/yqapp-demo/internal/domain"
	_ "github.com/lib/pq"

	"github.com/stretchr/testify/suite"
	"go.opentelemetry.io/otel/metric/noop"
	"go.uber.org/zap"
	"testing"
)

func TestTasksServiceSuite(t *testing.T) {
	suite.Run(t, new(TasksServiceTestSuite))
}

type TasksServiceTestSuite struct {
	suite.Suite
	service v1.TaskServiceServer
	logger  *zap.Logger
	db      *database.Postgres
}

func (suite *TasksServiceTestSuite) SetupSuite() {

}

func (suite *TasksServiceTestSuite) SetupTest() {
	var err error

	suite.logger, err = zap.NewDevelopment()
	suite.Require().NoError(err)

	err = database.MigrateModels("postgres:postgres@localhost:5432/postgres")
	suite.Require().NoError(err)

	suite.db, err = database.NewPostgres("postgres:postgres@localhost:5432/postgres")
	suite.Require().NoError(err)

	suite.Require().NoError(err)
	queries := database.New(suite.db.DB)

	suite.service = NewTaskService(suite.logger, queries, noop.NewMeterProvider().Meter(""), nil, nil)
}

func (suite *TasksServiceTestSuite) TearDownTest() {
	db := suite.db.DB
	db.Close()
	err := db.Ping(context.Background())
	suite.Require().Errorf(err, "closed pool")
}

func (suite *TasksServiceTestSuite) TearDownSuite() {

}

func (suite *TasksServiceTestSuite) TestCreate_Success() {

	taskType := uint32(domain.RandomInt(0, 9))
	taskValue := uint32(domain.RandomInt(0, 99))

	res, err := suite.service.CreateTask(context.Background(), &v1.CreateTaskRequest{
		Task: &v1.Task{
			Type:  taskType,
			Value: taskValue,
			State: v1.TaskState_RECEIVED,
		},
	})
	suite.Assert().NoError(err)
	suite.Assert().NotNil(res)

}
