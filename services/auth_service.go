package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"newyear-api/cache"
	"newyear-api/dto"
	"newyear-api/models"
	"newyear-api/repositories"
	"newyear-api/utils"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
)

const (
	cacheKeyUserProfile   = "wp:users:profile:%s"
	cacheKeyAccessJTI     = "wp:auth:user:%s:access:%s"
	cacheKeyAccessPattern = "wp:auth:user:%s:access:*"
)

type AuthService struct {
	repo  *repositories.AuthRepository
	cache *cache.RedisCache
}

func NewAuthService(repo *repositories.AuthRepository, redisCache *cache.RedisCache) *AuthService {
	return &AuthService{
		repo:  repo,
		cache: redisCache,
	}
}

func (s *AuthService) Register(input dto.RegisterDTO) (*models.User, error) {
	if err := utils.ValidatePasswordComplexity(input.Password); err != nil {
		return nil, err
	}

	existingUser, err := s.repo.FindUserByEmail(input.Email)
	if err == nil && existingUser != nil && existingUser.ID != "" {
		return nil, errors.New("user already exists")
	}
	if err != nil &&
		!errors.Is(err, mongo.ErrNoDocuments) &&
		!strings.Contains(err.Error(), "no documents in result") {
		return nil, err
	}

	hashedPassword, err := utils.HashPassword(input.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		Email:        input.Email,
		PasswordHash: hashedPassword,
	}

	if err := s.repo.CreateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}

func (s *AuthService) Login(input dto.LoginDTO) (accessToken, refreshToken string, err error) {
	user, err := s.repo.FindUserByEmail(input.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", errors.New("invalid credentials")
		}
		return "", "", err
	}

	if !utils.ComparePassword(user.PasswordHash, input.Password) {
		return "", "", errors.New("invalid credentials")
	}

	return s.issueSession(user)
}

func (s *AuthService) RefreshTokens(oldRefreshToken, oldAccessToken string) (accessToken, refreshToken string, err error) {
	claims, err := utils.ValidateRefreshToken(oldRefreshToken)
	if err != nil {
		return "", "", errors.New("invalid refresh token")
	}

	tokenHash := hashToken(oldRefreshToken)
	tokenModel, err := s.repo.FindRefreshToken(tokenHash)
	if err != nil {
		return "", "", errors.New("token not found or revoked")
	}

	if tokenModel.IsRevoked || time.Now().After(tokenModel.ExpiresAt) {
		return "", "", errors.New("refresh token expired or revoked")
	}

	if claims.UserID != tokenModel.UserID {
		return "", "", errors.New("refresh token does not belong to user")
	}

	if err := s.repo.RevokeRefreshToken(tokenHash); err != nil {
		return "", "", err
	}

	if oldAccessToken != "" {
		_ = s.revokeAccessToken(oldAccessToken)
		_ = s.deleteAccessJTI(oldAccessToken)
	}

	return s.issueSession(&tokenModel.User)
}

func (s *AuthService) ValidateAccessSession(accessToken string) (*utils.Claims, error) {
	claims, err := utils.ValidateAccessToken(accessToken)
	if err != nil {
		return nil, errors.New("invalid access token")
	}

	tokenHash := hashToken(accessToken)
	tokenModel, err := s.repo.FindAccessToken(tokenHash)
	if err != nil {
		return nil, errors.New("access token not found or revoked")
	}

	if tokenModel.IsRevoked || time.Now().After(tokenModel.ExpiresAt) {
		return nil, errors.New("access token expired or revoked")
	}

	if claims.UserID != tokenModel.UserID {
		return nil, errors.New("access token does not belong to user")
	}

	if claims.ID == "" {
		return nil, errors.New("access token missing jti")
	}

	jtiKey := fmt.Sprintf(cacheKeyAccessJTI, claims.UserID, claims.ID)

	exists, err := s.cache.Exists(jtiKey)
	if err != nil {
		log.Printf("Warning: Redis unavailable while validating access session: %v. Falling back to DB validation.", err)
		return claims, nil
	}

	if !exists {
		return nil, errors.New("access token revoked or expired in redis")
	}

	return claims, nil
}

func (s *AuthService) GetUserByID(userID string) (*models.User, error) {
	cacheKey := fmt.Sprintf(cacheKeyUserProfile, userID)

	var cachedUser models.User
	if err := s.cache.Get(cacheKey, &cachedUser); err == nil {
		log.Printf("Cache HIT: %s", cacheKey)
		return &cachedUser, nil
	}

	log.Printf("Cache MISS: %s", cacheKey)

	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	_ = s.cache.Set(cacheKey, user, getDefaultCacheTTL())

	return user, nil
}

func (s *AuthService) Logout(accessToken, refreshToken string) error {
	if accessToken != "" {
		_ = s.revokeAccessToken(accessToken)
		_ = s.deleteAccessJTI(accessToken)

		if claims, err := utils.ValidateAccessToken(accessToken); err == nil {
			_ = s.cache.Delete(fmt.Sprintf(cacheKeyUserProfile, claims.UserID))
		}
	}

	if refreshToken != "" {
		_ = s.revokeRefreshToken(refreshToken)
	}

	return nil
}

func (s *AuthService) LogoutAll(userID string) error {
	if err := s.repo.RevokeAllUserAccessTokens(userID); err != nil {
		return err
	}
	if err := s.repo.RevokeAllUserRefreshTokens(userID); err != nil {
		return err
	}

	_ = s.cache.DeleteByPattern(fmt.Sprintf(cacheKeyAccessPattern, userID))
	_ = s.cache.Delete(fmt.Sprintf(cacheKeyUserProfile, userID))

	return nil
}

func (s *AuthService) YandexLogin(code string) (accessToken, refreshToken string, err error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("client_id", os.Getenv("YANDEX_CLIENT_ID"))
	data.Set("client_secret", os.Getenv("YANDEX_CLIENT_SECRET"))

	req, err := http.NewRequest("POST", "https://oauth.yandex.ru/token", strings.NewReader(data.Encode()))
	if err != nil {
		return "", "", err
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{Timeout: 15 * time.Second}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	var tokenRes struct {
		AccessToken string `json:"access_token"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&tokenRes); err != nil {
		return "", "", err
	}

	if resp.StatusCode >= 400 || tokenRes.AccessToken == "" {
		return "", "", errors.New("failed to exchange authorization code")
	}

	reqInfo, err := http.NewRequest("GET", "https://login.yandex.ru/info?format=json", nil)
	if err != nil {
		return "", "", err
	}
	reqInfo.Header.Add("Authorization", "OAuth "+tokenRes.AccessToken)

	respInfo, err := client.Do(reqInfo)
	if err != nil {
		return "", "", err
	}
	defer respInfo.Body.Close()

	var userInfo struct {
		ID           string `json:"id"`
		DefaultEmail string `json:"default_email"`
	}

	if err := json.NewDecoder(respInfo.Body).Decode(&userInfo); err != nil {
		return "", "", err
	}

	if respInfo.StatusCode >= 400 || userInfo.ID == "" {
		return "", "", errors.New("failed to fetch user profile from yandex")
	}

	user, err := s.repo.FindUserByYandexID(userInfo.ID)
	if err == nil {
		return s.issueSession(user)
	}
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", "", err
	}

	if userInfo.DefaultEmail == "" {
		return "", "", errors.New("yandex did not return email")
	}

	user, err = s.repo.FindUserByEmail(userInfo.DefaultEmail)
	if err != nil {
		if !errors.Is(err, mongo.ErrNoDocuments) {
			return "", "", err
		}

		randomPasswordHash, err := utils.HashPassword(uuid.New().String())
		if err != nil {
			return "", "", err
		}

		user = &models.User{
			Email:        userInfo.DefaultEmail,
			PasswordHash: randomPasswordHash,
			YandexID:     &userInfo.ID,
		}

		if err := s.repo.CreateUser(user); err != nil {
			return "", "", err
		}
	} else if user.YandexID == nil {
		user.YandexID = &userInfo.ID
		if err := s.repo.UpdateUser(user); err != nil {
			return "", "", err
		}
	}

	return s.issueSession(user)
}

func (s *AuthService) ForgotPassword(input dto.ForgotPasswordDTO) (string, error) {
	user, err := s.repo.FindUserByEmail(input.Email)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", err
	}

	rawToken := uuid.New().String()
	tokenHash := hashToken(rawToken)

	resetToken := &models.PasswordResetToken{
		UserID:    user.ID,
		TokenHash: tokenHash,
		ExpiresAt: time.Now().Add(15 * time.Minute),
		Used:      false,
	}

	if err := s.repo.SavePasswordResetToken(resetToken); err != nil {
		return "", err
	}

	return rawToken, nil
}

func (s *AuthService) ResetPassword(input dto.ResetPasswordDTO) error {
	if err := utils.ValidatePasswordComplexity(input.NewPassword); err != nil {
		return err
	}

	tokenHash := hashToken(input.Token)

	resetToken, err := s.repo.FindPasswordResetToken(tokenHash)
	if err != nil {
		return errors.New("invalid or expired reset token")
	}

	if resetToken.Used || time.Now().After(resetToken.ExpiresAt) {
		return errors.New("invalid or expired reset token")
	}

	newPasswordHash, err := utils.HashPassword(input.NewPassword)
	if err != nil {
		return err
	}

	if err := s.repo.UpdatePassword(resetToken.UserID, newPasswordHash); err != nil {
		return err
	}

	if err := s.repo.MarkPasswordResetTokenUsed(tokenHash); err != nil {
		return err
	}

	if err := s.LogoutAll(resetToken.UserID); err != nil {
		return err
	}

	return nil
}

func (s *AuthService) issueSession(user *models.User) (accessToken, refreshToken string, err error) {
	accessToken, err = utils.GenerateAccessToken(user.ID, user.Email)
	if err != nil {
		return "", "", err
	}

	accessClaims, err := utils.ValidateAccessToken(accessToken)
	if err != nil {
		return "", "", err
	}

	refreshToken, err = utils.GenerateRefreshToken(user.ID)
	if err != nil {
		return "", "", err
	}

	accessTokenModel := &models.AccessToken{
		UserID:    user.ID,
		TokenHash: hashToken(accessToken),
		ExpiresAt: time.Now().Add(utils.GetAccessTokenDuration()),
		IsRevoked: false,
	}

	if err := s.repo.SaveAccessToken(accessTokenModel); err != nil {
		return "", "", err
	}

	refreshTokenModel := &models.RefreshToken{
		UserID:    user.ID,
		TokenHash: hashToken(refreshToken),
		ExpiresAt: time.Now().Add(utils.GetRefreshTokenDuration()),
		IsRevoked: false,
	}

	if err := s.repo.SaveRefreshToken(refreshTokenModel); err != nil {
		return "", "", err
	}

	if accessClaims.ID != "" {
		jtiKey := fmt.Sprintf(cacheKeyAccessJTI, user.ID, accessClaims.ID)
		_ = s.cache.Set(jtiKey, user.ID, utils.GetAccessTokenDuration())
	}

	return accessToken, refreshToken, nil
}

func (s *AuthService) revokeAccessToken(token string) error {
	return s.repo.RevokeAccessToken(hashToken(token))
}

func (s *AuthService) revokeRefreshToken(token string) error {
	return s.repo.RevokeRefreshToken(hashToken(token))
}

func (s *AuthService) deleteAccessJTI(accessToken string) error {
	claims, err := utils.ValidateAccessToken(accessToken)
	if err != nil {
		return err
	}

	if claims.ID == "" {
		return nil
	}

	jtiKey := fmt.Sprintf(cacheKeyAccessJTI, claims.UserID, claims.ID)
	return s.cache.Delete(jtiKey)
}

func hashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}
