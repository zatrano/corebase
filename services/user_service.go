package services

import (
	"context"
	"errors"
	"zatrano/models"
	"zatrano/pkg/constants"
	"zatrano/pkg/customerrors"
	"zatrano/pkg/logs"
	"zatrano/pkg/queryparams"
	"zatrano/repositories"

	"go.uber.org/zap"
)

const contextUserIDKey = "user_id"

type IUserService interface {
	GetAllUsers(params queryparams.ListParams) (*queryparams.PaginatedResult, error)
	GetUserByID(id uint) (*models.User, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, id uint, userData *models.User) error
	DeleteUser(ctx context.Context, id uint) error
	GetUserCount() (int64, error)
}

type UserService struct {
	repo repositories.IUserRepository
}

func NewUserService() IUserService {
	return &UserService{repo: repositories.NewUserRepository()}
}

func (s *UserService) GetAllUsers(params queryparams.ListParams) (*queryparams.PaginatedResult, error) {
	if params.Page <= 0 {
		params.Page = constants.DefaultPage
	}
	if params.PerPage <= 0 {
		params.PerPage = constants.DefaultPerPage
	} else if params.PerPage > constants.MaxPerPage {
		logs.Log.Warn("Sayfa başına istenen kayıt sayısı limiti aştı, varsayılana çekildi.",
			zap.Int("requested", params.PerPage),
			zap.Int("max", constants.MaxPerPage),
			zap.Int("default", constants.DefaultPerPage),
		)
		params.PerPage = constants.DefaultPerPage
	}
	if params.SortBy == "" {
		params.SortBy = constants.DefaultSortBy
	}
	if params.OrderBy == "" {
		params.OrderBy = constants.DefaultOrderBy
	}

	users, totalCount, err := s.repo.GetAllUsers(params)
	if err != nil {
		logs.Log.Error("GetAllUsersPaginated: Repository hatası", zap.Error(err))
		return nil, errors.New("kullanıcılar getirilirken bir hata oluştu")
	}

	totalPages := queryparams.CalculateTotalPages(totalCount, params.PerPage)

	result := &queryparams.PaginatedResult{
		Data: users,
		Meta: queryparams.PaginationMeta{
			CurrentPage: params.Page,
			PerPage:     params.PerPage,
			TotalItems:  totalCount,
			TotalPages:  totalPages,
		},
	}

	return result, nil
}

func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	user, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, customerrors.ErrRepoRecordNotFound) {
			logs.Log.Warn("Kullanıcı bulunamadı (ID ile arama)", zap.Uint("user_id", id))
			return nil, customerrors.ErrUserServiceUserNotFound
		}
		logs.Log.Error("Kullanıcı alınırken hata oluştu (ID ile arama)", zap.Uint("user_id", id), zap.Error(err))
		return nil, errors.New("kullanıcı bilgileri alınırken bir veritabanı hatası oluştu")
	}
	return user, nil
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.Password == "" {
		return customerrors.ErrPasswordRequired
	}

	if err := user.SetPassword(user.Password); err != nil {
		logs.Log.Error("Kullanıcı oluşturma: Şifre ayarlanamadı/hashlenemedi (SetPassword)", zap.String("account", user.Account), zap.Error(err))
		return customerrors.ErrPasswordHashingFailed
	}

	logs.Log.Info("Kullanıcı oluşturuluyor...",
		zap.String("account", user.Account),
		zap.Any("type", user.Type),
	)

	err := s.repo.CreateUser(ctx, user)
	if err != nil {
		logs.Log.Error("Kullanıcı oluşturulurken repository hatası",
			zap.String("account", user.Account),
			zap.Error(err),
		)
		return customerrors.ErrUserCreationFailed
	}

	logs.SLog.Infof("Kullanıcı başarıyla oluşturuldu: %s (ID: %d)", user.Account, user.ID)
	return nil
}

func (s *UserService) UpdateUser(ctx context.Context, id uint, userData *models.User) error {
	userIDValue := ctx.Value(contextUserIDKey)
	currentUserID, ok := userIDValue.(uint)
	if !ok || currentUserID == 0 {
		logs.Log.Error("UpdateUser: Context'te geçerli user_id bulunamadı veya 0.", zap.Any("value", userIDValue))
		return customerrors.ErrContextUserIDNotFound
	}

	_, err := s.repo.GetUserByID(id)
	if err != nil {
		if errors.Is(err, customerrors.ErrRepoRecordNotFound) {
			logs.Log.Warn("Kullanıcı güncellenemedi: Kullanıcı bulunamadı (ön kontrol)", zap.Uint("user_id", id))
			return customerrors.ErrUserServiceUserNotFound
		}
		logs.Log.Error("Kullanıcı güncellenemedi: Kullanıcı aranırken hata (ön kontrol)", zap.Uint("user_id", id), zap.Error(err))
		return errors.New("kullanıcı güncellenirken bir veritabanı hatası oluştu (ön kontrol)")
	}

	updateData := map[string]interface{}{
		"name":    userData.Name,
		"account": userData.Account,
		"status":  userData.Status,
		"type":    userData.Type,
	}

	passwordUpdated := false
	if userData.Password != "" {
		tempUserForHash := models.User{}
		if err := tempUserForHash.SetPassword(userData.Password); err != nil {
			logs.Log.Error("Kullanıcı güncelleme: Şifre ayarlanamadı/hashlenemedi (SetPassword)", zap.Uint("user_id", id), zap.Error(err))
			return customerrors.ErrPasswordHashingFailed
		}
		updateData["password"] = tempUserForHash.Password
		passwordUpdated = true
	}

	logs.Log.Info("Kullanıcı güncelleniyor (map ile)...",
		zap.Uint("target_user_id", id),
		zap.Bool("password_updated", passwordUpdated),
		zap.String("type", string(userData.Type)),
		zap.Uint("updated_by_user_id", currentUserID),
	)

	err = s.repo.UpdateUser(ctx, id, updateData, currentUserID)
	if err != nil {
		logs.Log.Error("Kullanıcı güncellenirken repository hatası",
			zap.Uint("user_id", id),
			zap.Error(err),
		)
		if errors.Is(err, customerrors.ErrRepoRecordNotFound) {
			return customerrors.ErrUserServiceUserNotFound
		}
		return customerrors.ErrUserUpdateFailed
	}

	logs.SLog.Infof("Kullanıcı başarıyla güncellendi (map ile): ID %d, Hesap: %s", id, userData.Account)
	return nil
}

func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	logs.Log.Info("Kullanıcı siliniyor...", zap.Uint("user_id", id))

	err := s.repo.DeleteUser(ctx, id)
	if err != nil {
		if errors.Is(err, customerrors.ErrRepoRecordNotFound) {
			logs.Log.Warn("Kullanıcı silinemedi: Kullanıcı bulunamadı", zap.Uint("user_id", id))
			return customerrors.ErrUserServiceUserNotFound
		}
		logs.Log.Error("Kullanıcı silinirken repository hatası", zap.Uint("user_id", id), zap.Error(err))
		return customerrors.ErrUserDeletionFailed
	}
	logs.SLog.Infof("Kullanıcı başarıyla silindi: ID %d", id)
	return nil
}

func (s *UserService) GetUserCount() (int64, error) {
	count, err := s.repo.GetUserCount()
	if err != nil {
		logs.Log.Error("Kullanıcı sayısı alınırken hata oluştu", zap.Error(err))
		return 0, errors.New("kullanıcı sayısı alınırken bir hata oluştu")
	}
	return count, nil
}

var _ IUserService = (*UserService)(nil)
