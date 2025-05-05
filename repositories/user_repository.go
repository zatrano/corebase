package repositories

import (
	"zatrano/models"

	"gorm.io/gorm"
)

type IUserRepository interface {
	IBaseRepository[models.User]
	FindUserByAccount(account string) (*models.User, error)
	FindUserByID(id uint) (*models.User, error)
	UpdateUser(user *models.User) error
}

type UserRepository struct {
	BaseRepository[models.User]
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{BaseRepository: BaseRepository[models.User]{db: db}}
}

func (r *UserRepository) FindUserByAccount(account string) (*models.User, error) {
	var user models.User
	err := r.db.Where("account = ?", account).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) FindUserByID(id uint) (*models.User, error) {
	var user models.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *UserRepository) UpdateUser(user *models.User) error {
	return r.db.Save(user).Error
}

var _ IUserRepository = (*UserRepository)(nil)
