package repositories

import (
	"zatrano/configs"
	"zatrano/models"

	"gorm.io/gorm"
)

type IAuthRepository interface {
	FindUserByAccount(account string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
}

type AuthRepository struct {
	*BaseRepository[models.User]
	db *gorm.DB
}

func NewAuthRepository() IAuthRepository {
	db := configs.GetDB()
	base := NewBaseRepository[models.User](db)
	return &AuthRepository{
		BaseRepository: base,
		db:             db,
	}
}

func (r *AuthRepository) FindUserByAccount(account string) (*models.User, error) {
	var user models.User
	err := r.db.Where("account = ?", account).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *AuthRepository) FindUserByID(id uint) (*models.User, error) {
	return r.GetByID(id)
}

func (r *AuthRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}
