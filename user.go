package oauth2

import "time"

type User struct {
	Id          string     `mapstructure:"id" json:"id,omitempty" gorm:"column:id" bson:"_id,omitempty" dynamodbav:"id,omitempty" firestore:"id,omitempty"`
	Email       string     `mapstructure:"email" json:"email,omitempty" gorm:"column:email" bson:"email,omitempty" dynamodbav:"email,omitempty" firestore:"email,omitempty"`
	Account     string     `mapstructure:"account" json:"account,omitempty" gorm:"column:account" bson:"account,omitempty" dynamodbav:"account,omitempty" firestore:"account,omitempty"`
	Active      bool       `mapstructure:"active" json:"active" gorm:"column:active" bson:"active" dynamodbav:"active" firestore:"active"`
	Picture     string     `mapstructure:"picture" json:"picture,omitempty" gorm:"column:picture" bson:"picture,omitempty" dynamodbav:"picture,omitempty" firestore:"picture,omitempty"`
	DisplayName string     `mapstructure:"display_name" json:"displayName,omitempty" gorm:"column:displayname" bson:"displayName,omitempty" dynamodbav:"displayName,omitempty" firestore:"displayName,omitempty"`
	GivenName   string     `mapstructure:"given_name" json:"givenName,omitempty" gorm:"column:givenname" bson:"givenName,omitempty" dynamodbav:"givenName,omitempty" firestore:"givenName,omitempty"`
	FamilyName  string     `mapstructure:"family_name" json:"familyName,omitempty" gorm:"column:familyname" bson:"familyName,omitempty" dynamodbav:"familyName,omitempty" firestore:"familyName,omitempty"`
	DateOfBirth *time.Time `mapstructure:"date_of_birth" json:"dateOfBirth,omitempty" gorm:"column:dateofbirth" bson:"dateOfBirth,omitempty" dynamodbav:"dateOfBirth,omitempty" firestore:"dateOfBirth,omitempty"`
	Gender      Gender     `mapstructure:"gender" json:"gender,omitempty" gorm:"column:gender" bson:"gender,omitempty" dynamodbav:"gender,omitempty" firestore:"gender,omitempty"`
	MiddleName  string     `mapstructure:"middle_name" json:"middleName,omitempty" gorm:"column:middlename" bson:"middleName,omitempty" dynamodbav:"middleName,omitempty" firestore:"middleName,omitempty"`
}
