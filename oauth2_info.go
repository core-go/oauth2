package oauth2

type OAuth2Info struct {
	SourceType     string `json:"sourceType,omitempty" gorm:"column:sourceType" bson:"sourceType,omitempty" dynamodbav:"sourceType,omitempty" firestore:"sourceType,omitempty"`
	Code           string `json:"code,omitempty" gorm:"column:code" bson:"code,omitempty" dynamodbav:"code,omitempty" firestore:"code,omitempty"`
	RedirectUri    string `json:"redirectUri,omitempty" gorm:"column:redirecturi" bson:"redirectUri,omitempty" dynamodbav:"redirectUri,omitempty" firestore:"redirectUri,omitempty"`
	InvitationMail string `json:"invitationMail,omitempty" gorm:"column:invitationmail" bson:"invitationMail,omitempty" dynamodbav:"invitationMail,omitempty" firestore:"invitationMail,omitempty"`
	Link           bool   `json:"link,omitempty" gorm:"column:link" bson:"link,omitempty" dynamodbav:"link,omitempty" firestore:"link,omitempty"`
}
