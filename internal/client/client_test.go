package client

import (
	conf "github.com/hasanhakkaev/yqapp-demo/internal/config"
	_ "github.com/lib/pq"
	"github.com/stretchr/testify/suite"
	"testing"
)

func TestServerSuite(t *testing.T) {
	suite.Run(t, new(ServerTestSuite))
}

type ServerTestSuite struct {
	suite.Suite
}

func (suite *ServerTestSuite) SetupSuite() {

}

func (suite *ServerTestSuite) SetupTest() {

}

func (suite *ServerTestSuite) TearDownTest() {

}

func (suite *ServerTestSuite) TearDownSuite() {

}

func (suite *ServerTestSuite) TestSetup() {
	cfg, err := conf.Read()
	suite.Require().NoError(err)
	suite.Require().NotZero(cfg)

	app, err := Setup(*cfg)
	suite.Assert().NoError(err)
	suite.Assert().NotZero(app)
	suite.Assert().NotNil(app.logger)
}
