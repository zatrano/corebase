package services

import (
	"zatrano/models"
	"zatrano/pkg/customerrors"
	"zatrano/pkg/logs"
	"zatrano/repositories"

	"go.uber.org/zap"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type ServiceError string

func (e ServiceError) Error() string {
	return string(e)
}

type IAuthService interface {
	Authenticate(account, password string) (*models.User, error)
	GetUserProfile(id uint) (*models.User, error)
	UpdatePassword(userID uint, currentPass, newPassword string) error
}

type AuthService struct {
	repo repositories.IAuthRepository
}

func NewAuthService() IAuthService {
	return &AuthService{repo: repositories.NewAuthRepository()}
}

func (s *AuthService) Authenticate(account, password string) (*models.User, error) {
	user, err := s.repo.FindUserByAccount(account)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Log.Warn("Kimlik doğrulama başarısız: Kullanıcı bulunamadı", zap.String("account", account))
			return nil, customerrors.ErrInvalidCredentials
		}
		logs.Log.Error("Kimlik doğrulama hatası (DB)",
			zap.String("account", account),
			zap.Error(err),
		)
		return nil, customerrors.ErrAuthGeneric
	}

	if !user.Status {
		logs.Log.Warn("Kimlik doğrulama başarısız: Kullanıcı aktif değil",
			zap.String("account", account),
			zap.Uint("user_id", user.ID),
		)
		return nil, customerrors.ErrUserInactive
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		logs.Log.Warn("Kimlik doğrulama başarısız: Geçersiz parola",
			zap.String("account", account),
			zap.Uint("user_id", user.ID),
		)
		return nil, customerrors.ErrInvalidCredentials
	}

	logs.Log.Info("Kimlik doğrulama başarılı",
		zap.String("account", account),
		zap.Uint("user_id", user.ID),
	)
	return user, nil
}

func (s *AuthService) GetUserProfile(id uint) (*models.User, error) {
	user, err := s.repo.FindUserByID(id)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Log.Warn("Profil alınamadı: Kullanıcı bulunamadı", zap.Uint("user_id", id))
			return nil, customerrors.ErrUserNotFound
		}
		logs.Log.Error("Profil alma hatası (DB)",
			zap.Uint("user_id", id),
			zap.Error(err),
		)
		return nil, customerrors.ErrProfileGeneric
	}
	return user, nil
}

func (s *AuthService) UpdatePassword(userID uint, currentPass, newPassword string) error {
	user, err := s.repo.FindUserByID(userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			logs.Log.Warn("Parola güncelleme başarısız: Kullanıcı bulunamadı", zap.Uint("user_id", userID))
			return customerrors.ErrUserNotFound
		}
		logs.Log.Error("Parola güncelleme hatası: Kullanıcı bulunurken DB hatası",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return customerrors.ErrUpdatePasswordGeneric
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(currentPass)); err != nil {
		logs.Log.Warn("Parola güncelleme başarısız: Mevcut parola hatalı", zap.Uint("user_id", userID))
		return customerrors.ErrCurrentPasswordIncorrect
	}

	if len(newPassword) < 6 {
		logs.Log.Warn("Parola güncelleme başarısız: Yeni parola çok kısa", zap.Uint("user_id", userID))
		return customerrors.ErrPasswordTooShort
	}
	if currentPass == newPassword {
		logs.Log.Warn("Parola güncelleme başarısız: Yeni parola eskiyle aynı", zap.Uint("user_id", userID))
		return customerrors.ErrPasswordSameAsOld
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		logs.Log.Error("Parola güncelleme hatası: Yeni parola hashlenemedi",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return customerrors.ErrHashingFailed
	}

	user.Password = string(hashedPassword)
	if err := s.repo.UpdateUser(user); err != nil {
		logs.Log.Error("Parola güncelleme hatası: Kullanıcı güncellenirken DB hatası",
			zap.Uint("user_id", userID),
			zap.Error(err),
		)
		return customerrors.ErrDatabaseUpdateFailed
	}

	logs.Log.Info("Parola başarıyla güncellendi", zap.Uint("user_id", userID))
	return nil
}

var _ IAuthService = (*AuthService)(nil)
