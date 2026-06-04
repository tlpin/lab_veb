package repositories

import (
	"context"
	"time"

	"newyear-api/models"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type AuthRepository struct {
	users               *mongo.Collection
	accessTokens        *mongo.Collection
	refreshTokens       *mongo.Collection
	passwordResetTokens *mongo.Collection
}

func NewAuthRepository(db *mongo.Database) *AuthRepository {
	return &AuthRepository{
		users:               db.Collection("users"),
		accessTokens:        db.Collection("access_tokens"),
		refreshTokens:       db.Collection("refresh_tokens"),
		passwordResetTokens: db.Collection("password_reset_tokens"),
	}
}

func mongoCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

func (r *AuthRepository) CreateUser(user *models.User) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	now := time.Now()

	if user.ID == "" {
		user.ID = uuid.New().String()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	user.UpdatedAt = now

	_, err := r.users.InsertOne(ctx, user)
	return err
}

func (r *AuthRepository) UpdateUser(user *models.User) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	user.UpdatedAt = time.Now()

	filter := bson.M{
		"_id": user.ID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	_, err := r.users.ReplaceOne(ctx, filter, user)
	return err
}

func (r *AuthRepository) UpdatePassword(userID, passwordHash string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"_id": userID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"password_hash": passwordHash,
			"updated_at":    time.Now(),
		},
	}

	_, err := r.users.UpdateOne(ctx, filter, update)
	return err
}

func (r *AuthRepository) FindUserByEmail(email string) (*models.User, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"email": email,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var user models.User
	err := r.users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindUserByID(id string) (*models.User, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"_id": id,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var user models.User
	err := r.users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) FindUserByYandexID(yandexID string) (*models.User, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"yandex_id": yandexID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var user models.User
	err := r.users.FindOne(ctx, filter).Decode(&user)
	if err != nil {
		return nil, err
	}

	return &user, nil
}

func (r *AuthRepository) SaveAccessToken(token *models.AccessToken) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	now := time.Now()

	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	token.UpdatedAt = now

	_, err := r.accessTokens.InsertOne(ctx, token)
	return err
}

func (r *AuthRepository) FindAccessToken(tokenHash string) (*models.AccessToken, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"is_revoked": false,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var token models.AccessToken
	err := r.accessTokens.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	user, err := r.FindUserByID(token.UserID)
	if err != nil {
		return nil, err
	}
	token.User = *user

	return &token, nil
}

func (r *AuthRepository) RevokeAccessToken(tokenHash string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"is_revoked": true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.accessTokens.UpdateOne(ctx, filter, update)
	return err
}

func (r *AuthRepository) RevokeAllUserAccessTokens(userID string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"is_revoked": true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.accessTokens.UpdateMany(ctx, filter, update)
	return err
}

func (r *AuthRepository) SaveRefreshToken(token *models.RefreshToken) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	now := time.Now()

	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	token.UpdatedAt = now

	_, err := r.refreshTokens.InsertOne(ctx, token)
	return err
}

func (r *AuthRepository) FindRefreshToken(tokenHash string) (*models.RefreshToken, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"is_revoked": false,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var token models.RefreshToken
	err := r.refreshTokens.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	user, err := r.FindUserByID(token.UserID)
	if err != nil {
		return nil, err
	}
	token.User = *user

	return &token, nil
}

func (r *AuthRepository) RevokeRefreshToken(tokenHash string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"is_revoked": true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.refreshTokens.UpdateOne(ctx, filter, update)
	return err
}

func (r *AuthRepository) RevokeAllUserRefreshTokens(userID string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"user_id": userID,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"is_revoked": true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.refreshTokens.UpdateMany(ctx, filter, update)
	return err
}

func (r *AuthRepository) SavePasswordResetToken(token *models.PasswordResetToken) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	now := time.Now()

	if token.ID == "" {
		token.ID = uuid.New().String()
	}
	if token.CreatedAt.IsZero() {
		token.CreatedAt = now
	}
	token.UpdatedAt = now

	_, err := r.passwordResetTokens.InsertOne(ctx, token)
	return err
}

func (r *AuthRepository) FindPasswordResetToken(tokenHash string) (*models.PasswordResetToken, error) {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"used":       false,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	var token models.PasswordResetToken
	err := r.passwordResetTokens.FindOne(ctx, filter).Decode(&token)
	if err != nil {
		return nil, err
	}

	user, err := r.FindUserByID(token.UserID)
	if err != nil {
		return nil, err
	}
	token.User = *user

	return &token, nil
}

func (r *AuthRepository) MarkPasswordResetTokenUsed(tokenHash string) error {
	ctx, cancel := mongoCtx()
	defer cancel()

	filter := bson.M{
		"token_hash": tokenHash,
		"deleted_at": bson.M{
			"$exists": false,
		},
	}

	update := bson.M{
		"$set": bson.M{
			"used":       true,
			"updated_at": time.Now(),
		},
	}

	_, err := r.passwordResetTokens.UpdateOne(ctx, filter, update)
	return err
}
