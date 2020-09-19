package oauth2

import (
	"context"
	"github.com/common-go/auth"
	"strings"
)

type DefaultOAuth2IntegrationService struct {
	OAuth2UserRepositories             map[string]OAuth2UserRepository
	UserRepositories                   map[string]UserRepository
	IntegrationConfigurationRepository IntegrationConfigurationRepository
	IdGenerator                        IdGenerator
	TokenService                       TokenService
	TokenConfig                        auth.TokenConfig
	Status                             auth.StatusConfig
	PrivilegeRepository                auth.PrivilegeService
	TokenStored                        bool
	AccessTimeRepository               auth.AccessTimeService
}

func NewDefaultOAuth2IntegrationService(oauth2UserRepositories map[string]OAuth2UserRepository, userRepositories map[string]UserRepository, integrationConfigurationService IntegrationConfigurationRepository, userIdGenerator IdGenerator, tokenGenerator TokenService, tokenConfig auth.TokenConfig, status auth.StatusConfig, privilegeService auth.PrivilegeService, tokenStored bool, accessTimeService auth.AccessTimeService) *DefaultOAuth2IntegrationService {
	if userIdGenerator == nil {
		panic("IdGenerator cannot be nil")
	}
	return &DefaultOAuth2IntegrationService{
		OAuth2UserRepositories:             oauth2UserRepositories,
		UserRepositories:                   userRepositories,
		IntegrationConfigurationRepository: integrationConfigurationService,
		IdGenerator:                        userIdGenerator,
		TokenService:                       tokenGenerator,
		TokenConfig:                        tokenConfig,
		Status:                             status,
		PrivilegeRepository:                privilegeService,
		TokenStored:                        tokenStored,
		AccessTimeRepository:               accessTimeService,
	}
}
func (s *DefaultOAuth2IntegrationService) GetIntegrationConfiguration(ctx context.Context, sourceType string) (*IntegrationConfiguration, error) {
	model, _, err := s.IntegrationConfigurationRepository.GetIntegrationConfiguration(ctx, sourceType)
	return model, err
}

func (s *DefaultOAuth2IntegrationService) Authenticate(ctx context.Context, info OAuth2Info, authorization string) (auth.AuthResult, error) {
	result := auth.AuthResult{Status: auth.StatusFail}
	var linkUserId = ""
	if info.Link {
		if len(authorization) == 0 {
			linkUserId = ""
		} else {
			if strings.HasPrefix(authorization, "Bearer ") != true {
				return result, nil
			}
			token := authorization[7:]
			_, _, _, er0 := s.TokenService.VerifyToken(token, s.TokenConfig.Secret)
			if er0 != nil {
				result.Status = auth.StatusSystemError
				return result, er0
			}
			linkUserId = s.getStringValue(token, "userId") // TODO
		}
	}
	integrations, clientId, er1 := s.IntegrationConfigurationRepository.GetIntegrationConfiguration(ctx, info.SourceType)
	if er1 != nil {
		return result, er1
	}

	if len(integrations.ClientId) > 0 {
		if len(info.SourceType) == 0 {
			return result, nil
		}
		integrations.ClientId = clientId
		return s.processAccount(ctx, info, *integrations, linkUserId)
	}
	return result, nil
}
func (s *DefaultOAuth2IntegrationService) getStringValue(tokenData interface{}, field string) string {
	if authorizationToken, ok := tokenData.(map[string]interface{}); ok {
		value, _ := authorizationToken[field].(string)
		return value
	}
	return ""
}
func (s *DefaultOAuth2IntegrationService) buildResult(ctx context.Context, id, email, displayName string, sourceType string, accessToken string, newUser bool) (auth.AuthResult, error) {
	user := auth.AccessTime{}
	result := auth.AuthResult{Status: auth.StatusSystemError}
	if s.AccessTimeRepository != nil {
		accessTime, er1 := s.AccessTimeRepository.Load(ctx, id)
		if er1 != nil {
			return result, er1
		}
		if accessTime != nil {
			user = *accessTime
			if !auth.IsAccessDateValid(accessTime.AccessDateFrom, accessTime.AccessDateTo) {
				result := auth.AuthResult{Status: auth.StatusDisabled}
				return result, nil
			}
			if !auth.IsAccessTimeValid(accessTime.AccessTimeFrom, accessTime.AccessTimeTo) {
				result := auth.AuthResult{Status: auth.StatusAccessTimeLocked}
				return result, nil
			}
		}
	}

	tokenExpiredTime, jwtTokenExpires := auth.SetTokenExpiredTime(user.AccessTimeFrom, user.AccessTimeTo, s.TokenConfig.Expires)
	var tokens map[string]string
	if s.TokenStored {
		tokens = make(map[string]string)
		tokens[sourceType] = accessToken
	}
	storedUser := auth.StoredUser{
		UserId:   id,
		Username: email,
		Contact:  email,
		Tokens:   tokens,
	}
	token, er2 := s.TokenService.GenerateToken(storedUser, s.TokenConfig.Secret, jwtTokenExpires)

	if er2 != nil {
		return result, er2
	}
	var account auth.UserAccount
	account.Username = email
	account.UserId = id
	account.Contact = email
	account.DisplayName = displayName
	account.Token = token
	account.NewUser = false
	account.TokenExpiredTime = &tokenExpiredTime

	if s.PrivilegeRepository != nil {
		privileges, er1 := s.PrivilegeRepository.GetPrivileges(ctx, id)
		if er1 != nil {
			return result, er1
		}
		account.Privileges = &privileges
	}
	result.Status = auth.StatusSuccess
	result.User = &account
	return result, nil
}
func (s *DefaultOAuth2IntegrationService) processAccount(ctx context.Context, data OAuth2Info, integration IntegrationConfiguration, linkUserId string) (auth.AuthResult, error) {
	code := data.Code
	urlRedirect := data.RedirectUri
	clientSecret := integration.ClientSecret
	clientId := integration.ClientId
	repository := s.OAuth2UserRepositories[data.SourceType]
	user, accessToken, err := repository.GetUserFromOAuth2(ctx, urlRedirect, clientId, clientSecret, code)
	if err != nil || user == nil {
		result := auth.AuthResult{Status: auth.StatusSystemError}
		return result, err
	}
	return s.checkAccount(ctx, *user, accessToken, linkUserId, data.SourceType)
}

func (s *DefaultOAuth2IntegrationService) checkAccount(ctx context.Context, user User, accessToken string, linkUserId string, types string) (auth.AuthResult, error) {
	personRepository := s.UserRepositories[types]
	eId, disable, suspended, er0 := personRepository.GetUser(ctx, user.Email) //i
	result := auth.AuthResult{Status: auth.StatusSystemError}
	if er0 != nil {
		return result, er0
	}
	if len(linkUserId) > 0 {
		if eId != linkUserId {
			result := auth.AuthResult{Status: auth.StatusFail}
			return result, nil
		}
		ok1, er2 := personRepository.Update(ctx, linkUserId, user.Email, user.Account)
		if ok1 && er2 == nil {
			return s.buildResult(ctx, eId, user.Email, user.DisplayName, types, accessToken, false)
		}
	}
	if len(eId) == 0 {
		userId, er3 := s.IdGenerator.Generate(ctx)
		if er3 != nil {
			return result, er3
		}
		duplicate, er4 := personRepository.Insert(ctx, userId, user)
		if duplicate {
			i := 1
			for duplicate && i <= 5 {
				i++
				userId, er3 = s.IdGenerator.Generate(ctx)
				if er3 != nil {
					return result, er3
				}
				duplicate, er4 = personRepository.Insert(ctx, userId, user)
				if er4 != nil {
					return result, er4
				}
			}
			if duplicate {
				return result, nil
			}
		}
		if er4 == nil && !duplicate {
			return s.buildResult(ctx, eId, user.Email, user.DisplayName, types, accessToken, true)
		}
		return result, er4
	}
	if disable {
		result.Status = auth.StatusDisabled
		return result, nil
	}
	if suspended {
		result.Status = auth.StatusSuspended
		return result, nil
	}

	ok3, er5 := personRepository.Update(ctx, eId, user.Email, user.Account)
	if ok3 && er5 == nil {
		return s.buildResult(ctx, eId, user.Email, user.Account, types, accessToken, false)
	}

	return result, nil
}
