package oauth2

type OAuth2Config struct {
	Services string             `mapstructure:"services" json:"services,omitempty" gorm:"column:services" bson:"services,omitempty" dynamodbav:"services,omitempty" firestore:"services,omitempty"`
	Schema   OAuth2SchemaConfig `mapstructure:"schema" json:"schema,omitempty" gorm:"column:schema" bson:"schema,omitempty" dynamodbav:"schema,omitempty" firestore:"schema,omitempty"`
}

type CallbackURL struct {
	Microsoft string `mapstructure:"microsoft" json:"microsoft,omitempty" gorm:"column:microsoft" bson:"microsoft,omitempty" dynamodbav:"microsoft,omitempty" firestore:"microsoft,omitempty"`
	Amazon    string `mapstructure:"amazon" json:"amazon,omitempty" gorm:"column:amazon" bson:"amazon,omitempty" dynamodbav:"amazon,omitempty" firestore:"amazon,omitempty"`
	Twitter   string `mapstructure:"twitter" json:"twitter,omitempty" gorm:"column:twitter" bson:"twitter,omitempty" dynamodbav:"twitter,omitempty" firestore:"twitter,omitempty"`
}
