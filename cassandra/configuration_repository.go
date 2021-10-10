package cassandra

import (
	"context"
	"fmt"
	"github.com/core-go/oauth2"
	"github.com/gocql/gocql"
	"reflect"
)

type ConfigurationRepository struct {
	Session                *gocql.Session
	TableName              string
	OAuth2UserRepositories map[string]oauth2.OAuth2UserRepository
	Status                 string
	Active                 string
	configurationFields    map[string]int
}

func NewConfigurationRepository(session *gocql.Session, tableName string, oauth2UserRepositories map[string]oauth2.OAuth2UserRepository, status string, active string) (*ConfigurationRepository, error) {
	if len(status) == 0 {
		status = "status"
	}
	if len(active) == 0 {
		active = "A"
	}
	var configuration oauth2.Configuration
	configurationType := reflect.TypeOf(configuration)
	configurationFields, err := getColumnIndexes(configurationType)
	if err != nil {
		return nil, err
	}
	return &ConfigurationRepository{Session: session, TableName: tableName, OAuth2UserRepositories: oauth2UserRepositories, Status: status, Active: active, configurationFields: configurationFields}, nil
}

func (s *ConfigurationRepository) GetConfiguration(ctx context.Context, id string) (*oauth2.Configuration, string, error) {
	var configurations []oauth2.Configuration
	query := fmt.Sprintf(`select %s from %s where %s = ? ALLOW FILTERING`, "clientid, clientsecret ", s.TableName, "sourcetype")
	err := query(s.Session, s.configurationFields, &configurations, query, id)
	if err != nil {
		return nil, "", err
	}
	if len(configurations) == 0 {
		return nil, "", nil
	}
	model := configurations[0]
	clientId := model.ClientId
	clientId, err = s.OAuth2UserRepositories[id].GetRequestTokenOAuth(ctx, model.ClientId, model.ClientSecret)
	return &model, clientId, err
}
func (s *ConfigurationRepository) GetConfigurations(ctx context.Context) (*[]oauth2.Configuration, error) {
	var configurations []oauth2.Configuration
	query := fmt.Sprintf(`select * from %s where %s = ? `, s.TableName, s.Status)
	err := query(s.Session, s.configurationFields, &configurations, query, s.Active)
	if err != nil {
		return nil, err
	}
	return &configurations, nil
}
