package services

import (
	"errors"
	"time"

	"newyear-api/dto"
	"newyear-api/models"
	"newyear-api/repositories"

	"go.mongodb.org/mongo-driver/v2/mongo"
)

type ProfileService struct {
	authRepo *repositories.AuthRepository
	fileRepo *repositories.FileRepository
}

func NewProfileService(authRepo *repositories.AuthRepository, fileRepo *repositories.FileRepository) *ProfileService {
	return &ProfileService{
		authRepo: authRepo,
		fileRepo: fileRepo,
	}
}

func (s *ProfileService) GetProfile(userID string) (*models.User, error) {
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, errors.New("user not found")
		}
		return nil, err
	}
	return user, nil
}

func (s *ProfileService) UpdateProfile(userID string, input dto.UpdateProfileDTO) (*models.User, error) {
	user, err := s.authRepo.FindUserByID(userID)
	if err != nil {
		return nil, err
	}

	if input.DisplayName != nil {
		user.DisplayName = *input.DisplayName
	}
	if input.Bio != nil {
		user.Bio = *input.Bio
	}
	if input.AvatarFileID != nil {
		file, err := s.fileRepo.FindByID(*input.AvatarFileID)
		if err != nil {
			if errors.Is(err, mongo.ErrNoDocuments) {
				return nil, errors.New("avatar file not found")
			}
			return nil, err
		}
		if file.UserID != userID {
			return nil, errors.New("file does not belong to user")
		}
		user.AvatarFileID = input.AvatarFileID
	}

	user.UpdatedAt = time.Now()

	if err := s.authRepo.UpdateUser(user); err != nil {
		return nil, err
	}

	return user, nil
}
