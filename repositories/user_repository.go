package repositories

import (
	"context"
	"strings"

	"zatrano/configs"
	"zatrano/models"
	"zatrano/pkg/constants"
	"zatrano/pkg/logs"
	"zatrano/pkg/queryparams"
	"zatrano/pkg/turkishsearch"

	"go.uber.org/zap"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type IUserRepository interface {
	GetAllUsers(params queryparams.ListParams) ([]models.User, int64, error)
	GetUserByID(id uint) (*models.User, error)
	GetUserCount() (int64, error)
	CreateUser(ctx context.Context, user *models.User) error
	UpdateUser(ctx context.Context, id uint, data map[string]interface{}, updatedByID uint) error
	DeleteUser(ctx context.Context, id uint) error
}

type UserRepository struct {
	*BaseRepository[models.User]
	db *gorm.DB
}

func NewUserRepository() IUserRepository {
	db := configs.GetDB()
	return &UserRepository{
		BaseRepository: NewBaseRepository[models.User](db),
		db:             db,
	}
}

func (r *UserRepository) GetAllUsers(params queryparams.ListParams) ([]models.User, int64, error) {
	var users []models.User
	var totalCount int64

	query := r.db.Model(&models.User{})

	if params.Name != "" {
		sqlQueryFragment, queryParams := turkishsearch.SQLFilter("name", params.Name)
		query = query.Where(sqlQueryFragment, queryParams...)
	}

	err := query.Count(&totalCount).Error
	if err != nil {
		logs.Log.Error("Kullanıcı sayısı alınırken hata (GetAllUsers)", zap.Error(err))
		return nil, 0, err
	}

	if totalCount == 0 {
		return users, 0, nil
	}

	sortBy := params.SortBy
	orderBy := strings.ToLower(params.OrderBy)
	if orderBy != "asc" && orderBy != "desc" {
		orderBy = constants.DefaultOrderBy
	}
	allowedSortColumns := map[string]bool{"id": true, "name": true, "account": true, "created_at": true, "status": true, "type": true}
	if _, ok := allowedSortColumns[sortBy]; !ok {
		sortBy = constants.DefaultSortBy
	}
	orderClause := sortBy + " " + orderBy
	query = query.Order(orderClause)

	query = query.Preload(clause.Associations)
	offset := params.CalculateOffset()
	query = query.Limit(params.PerPage).Offset(offset)

	err = query.Find(&users).Error
	if err != nil {
		logs.Log.Error("Kullanıcılar çekilirken hata (GetAllUsers)", zap.Error(err))
		return nil, totalCount, err
	}

	return users, totalCount, nil
}

func (r *UserRepository) GetUserByID(id uint) (*models.User, error) {
	return r.GetByID(id)
}

func (r *UserRepository) GetUserCount() (int64, error) {
	return r.GetCount()
}

func (r *UserRepository) CreateUser(ctx context.Context, user *models.User) error {
	return r.Create(ctx, user)
}

func (r *UserRepository) UpdateUser(ctx context.Context, id uint, data map[string]interface{}, updatedByID uint) error {
	if updatedByID != 0 {
		data["updated_by"] = updatedByID
	} else {
		logs.Log.Warn("UserRepository.Update: updated_by eksik.", zap.Uint("target_user_id", id))
	}
	return r.Update(ctx, id, data)
}

func (r *UserRepository) DeleteUser(ctx context.Context, id uint) error {
	return r.Delete(ctx, id)
}
