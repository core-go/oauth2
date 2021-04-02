package oauth2

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/common-go/auth"
	"strconv"
	"strings"
	"time"
)

const (
	DriverPostgres   = "postgres"
	DriverMysql      = "mysql"
	DriverMssql      = "mssql"
	DriverOracle     = "oracle"
	DriverNotSupport = "no support"
)

type SqlUserRepository struct {
	DB              *sql.DB
	Driver          string
	TableName       string
	Prefix          string
	ActivatedStatus string
	Services        []string
	StatusName      string
	UserIdName      string
	UserName        string
	EmailName       string
	OAuth2EmailName string
	AccountName     string
	ActiveName      string

	updatedTimeName string
	updatedByName   string
	Status          *auth.UserStatusConfig
	GenderMapper    OAuth2GenderMapper
	Schema          *OAuth2SchemaConfig
}

func NewSqlUserRepositoryByConfig(db *sql.DB, tableName, prefix string, activatedStatus string, services []string, c OAuth2SchemaConfig, driver string, status *auth.UserStatusConfig, options ...OAuth2GenderMapper) *SqlUserRepository {
	var genderMapper OAuth2GenderMapper
	if len(options) >= 1 {
		genderMapper = options[0]
	}
	c.UserId = strings.ToLower(c.UserId)
	c.UserName = strings.ToLower(c.UserName)
	c.Email = strings.ToLower(c.Email)
	c.Status = strings.ToLower(c.Status)
	c.OAuth2Email = strings.ToLower(c.OAuth2Email)
	c.Account = strings.ToLower(c.Account)
	c.Active = strings.ToLower(c.Active)
	c.DisplayName = strings.ToLower(c.DisplayName)
	c.Picture = strings.ToLower(c.Picture)
	c.Locale = strings.ToLower(c.Locale)
	c.Gender = strings.ToLower(c.Gender)
	c.DateOfBirth = strings.ToLower(c.DateOfBirth)
	c.GivenName = strings.ToLower(c.GivenName)
	c.MiddleName = strings.ToLower(c.MiddleName)
	c.FamilyName = strings.ToLower(c.FamilyName)
	c.CreatedTime = strings.ToLower(c.CreatedTime)
	c.CreatedBy = strings.ToLower(c.CreatedBy)
	c.UpdatedTime = strings.ToLower(c.UpdatedTime)
	c.UpdatedBy = strings.ToLower(c.UpdatedBy)
	c.Version = strings.ToLower(c.Version)
	s := make([]string, 0)
	for _, sv := range services {
		s = append(s, strings.ToLower(sv))
	}

	if len(c.UserName) == 0 {
		c.UserName = "username"
	}
	if len(c.Email) == 0 {
		c.Email = "email"
	}
	if len(c.Status) == 0 {
		c.Status = "status"
	}
	if len(c.OAuth2Email) == 0 {
		c.OAuth2Email = "email"
	}
	if len(c.Account) == 0 {
		c.Account = "account"
	}
	if len(c.Active) == 0 {
		c.Active = "active"
	}
	m := &SqlUserRepository{
		DB:              db,
		Driver:          driver,
		TableName:       tableName,
		Prefix:          prefix,
		ActivatedStatus: activatedStatus,
		Services:        s,
		GenderMapper:    genderMapper,
		Schema:          &c,
		updatedTimeName: c.UpdatedTime,
		updatedByName:   c.UpdatedBy,
		Status:          status,
	}
	return m
}

func NewSqlUserRepository(db *sql.DB, tableName, prefix, activatedStatus string, services []string, pictureName, displayName, givenName, familyName, middleName, genderName string, status *auth.UserStatusConfig, options ...OAuth2GenderMapper) *SqlUserRepository {
	var genderMapper OAuth2GenderMapper
	if len(options) >= 1 {
		genderMapper = options[0]
	}

	pictureName = strings.ToLower(pictureName)
	displayName = strings.ToLower(displayName)
	givenName = strings.ToLower(givenName)
	familyName = strings.ToLower(familyName)
	middleName = strings.ToLower(middleName)
	genderName = strings.ToLower(genderName)

	m := &SqlUserRepository{
		DB:              db,
		TableName:       tableName,
		Prefix:          prefix,
		ActivatedStatus: activatedStatus,
		StatusName:      "status",
		Services:        services,
		UserName:        "username",
		EmailName:       "email",
		OAuth2EmailName: "email",
		AccountName:     "account",
		ActiveName:      "active",
		Status:          status,
		GenderMapper:    genderMapper,
	}
	if len(pictureName) > 0 || len(displayName) > 0 || len(givenName) > 0 || len(middleName) > 0 || len(familyName) > 0 || len(genderName) > 0 {
		c := &OAuth2SchemaConfig{}
		c.Picture = pictureName
		c.DisplayName = displayName
		c.GivenName = givenName
		c.MiddleName = middleName
		c.FamilyName = familyName
		c.Gender = genderName
		m.Schema = c
	}
	return m
}

func (s *SqlUserRepository) GetUser(ctx context.Context, email string) (string, bool, bool, error) {
	arr := make(map[string]interface{})
	columns := make([]interface{}, 0)
	values := make([]interface{}, 0)
	s.Driver = DriverOracle
	i := 0
	columns = append(columns, s.Schema.UserId, s.Schema.Status, s.TableName,
		s.Schema.UserName, BuildParam(i, s.Driver),
		s.Schema.Email, BuildParam(i+1, s.Driver), s.Prefix+s.Schema.OAuth2Email, BuildParam(i+2, s.Driver))
	values = append(values, email, email, email)
	var where strings.Builder
	where.WriteString(`%s = %s OR %s = %s OR %s = %s`)
	var sel strings.Builder
	sel.WriteString(`SELECT %s, %s FROM %s WHERE `)
	i = 3
	for _, sv := range s.Services {
		if sv != s.Prefix {
			where.WriteString(` OR %s = `)
			where.WriteString(BuildParam(i, s.Driver))
			i++
			columns = append(columns, sv+s.Schema.OAuth2Email)
			values = append(values, email)
		}
	}
	sel.WriteString(where.String())
	query := fmt.Sprintf(sel.String(), columns...)
	rows, err := s.DB.Query(query, values...)
	disable := false
	suspended := false
	if err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return "", disable, suspended, nil
		}
		return "", disable, suspended, err
	}
	defer rows.Close()
	cols, _ := rows.Columns()
	for rows.Next() {
		columns := make([]interface{}, len(cols))
		columnPointers := make([]interface{}, len(cols))
		for i, _ := range columns {
			columnPointers[i] = &columns[i]
		}

		if err1 := rows.Scan(columnPointers...); err1 != nil {
			return "", disable, suspended, err1
		}

		for i, colName := range cols {
			val := columnPointers[i].(*interface{})
			arr[colName] = *val
		}
	}
	err2 := rows.Err()
	if err2 != nil {
		return "", disable, suspended, err2
	}

	if len(arr) == 0 {
		return "", disable, suspended, nil
	}
	if s.Status != nil {
		status := string(arr[s.Schema.Status].([]byte))
		if status == s.Status.Disable {
			disable = true
		}
		if status == s.Status.Suspended {
			suspended = true
		}
	}
	return string(arr[s.Schema.UserId].([]byte)), disable, suspended, nil
}

func (s *SqlUserRepository) Update(ctx context.Context, id, email, account string) (bool, error) {
	user := make(map[string]interface{})

	user[s.Prefix+s.Schema.OAuth2Email] = email
	user[s.Prefix+s.Schema.Account] = account
	user[s.Prefix+s.Schema.Active] = true

	if len(s.updatedTimeName) > 0 {
		user[s.updatedTimeName] = time.Now()
	}
	if len(s.updatedByName) > 0 {
		user[s.updatedByName] = id
	}

	query, values := s.buildQueryUpdate(user, s.TableName, id, s.Schema.UserId)
	result, err1 := s.DB.Exec(query, values...)
	if err1 != nil {
		return false, err1
	}
	r, err2 := result.RowsAffected()
	if err2 != nil {
		return false, err2
	}
	return r > 0, err2
}

func (s *SqlUserRepository) Insert(ctx context.Context, id string, personInfo User) (bool, error) {
	user := s.userToMap(ctx, id, personInfo)
	query, values := s.buildQueryString(user)
	_, err := s.DB.Exec(query, values...)
	if err != nil {
		return handleDuplicate(s.DB, err, s.Driver)
	}
	return false, err
}

func handleDuplicate(db *sql.DB, err error, driverName string) (bool, error) {
	x := err.Error()
	if driverName == DriverPostgres && strings.Contains(x, "pq: duplicate key value violates unique constraint") {
		return false, nil //pq: duplicate key value violates unique constraint "aa_pkey"
	} else if driverName == DriverMysql && strings.Contains(x, "Error 1062: Duplicate entry") {
		return false, nil //mysql Error 1062: Duplicate entry 'a-1' for key 'PRIMARY'
	} else if driverName == DriverOracle && strings.Contains(x, "ORA-00001: unique constraint") {
		return false, nil //mysql Error 1062: Duplicate entry 'a-1' for key 'PRIMARY'
	} else if driverName == DriverMssql && strings.Contains(x, "Violation of PRIMARY KEY constraint") {
		return false, nil //Violation of PRIMARY KEY constraint 'PK_aa'. Cannot insert duplicate key in object 'dbo.aa'. The duplicate key value is (b, 2).
	}
	return false, err
}

func (s *SqlUserRepository) userToMap(ctx context.Context, id string, user User) map[string]interface{} {

	userMap := UserToMap(ctx, id, user, s.GenderMapper, s.Schema)
	//userMap := User{}
	userMap[s.Schema.UserId] = id
	userMap[s.Schema.UserName] = user.Email
	userMap[s.Schema.Status] = s.ActivatedStatus

	userMap[s.Prefix+s.Schema.OAuth2Email] = user.Email
	userMap[s.Prefix+s.Schema.Account] = user.Account
	userMap[s.Prefix+s.Schema.Active] = true
	return userMap
}

func (s *SqlUserRepository) buildQueryString(user map[string]interface{}) (string, []interface{}) {
	var cols []string
	var values []interface{}
	for col, v := range user {
		cols = append(cols, col)
		values = append(values, v)
	}
	column := fmt.Sprintf("(%v)", strings.Join(cols, ","))
	numCol := len(cols)
	var arrValue []string
	for i := 0; i < numCol; i++ {
		arrValue = append(arrValue, BuildParam(i, s.Driver))
	}
	value := fmt.Sprintf("(%v)", strings.Join(arrValue, ","))
	return fmt.Sprintf("INSERT INTO %v %v VALUES %v", s.TableName, column, value), values
}

func (r *SqlUserRepository) buildQueryUpdate(model map[string]interface{}, table string, id interface{}, idname string) (string, []interface{}) {
	colNumber := 0
	values := []interface{}{}
	querySet := make([]string, 0)
	for colName, v2 := range model {
		values = append(values, v2)
		querySet = append(querySet, fmt.Sprintf("%v="+BuildParam(colNumber, r.Driver), colName))
		colNumber++
	}
	values = append(values, id)
	queryWhere := fmt.Sprintf(" %s = %s",
		idname,
		BuildParam(colNumber, r.Driver),
	)
	query := fmt.Sprintf("update %v set %v where %v", table, strings.Join(querySet, ","), queryWhere)
	return query, values
}

func BuildParam(index int, driver string) string {
	switch driver {
	case DriverPostgres:
		return "$" + strconv.Itoa(index)
	case DriverOracle:
		return ":val" + strconv.Itoa(index)
	default:
		return "?"
	}
}
