package repository

import (
	"stock-monitor/internal/model"

	"gorm.io/gorm"
)

type UserRepo struct {
	db *gorm.DB
}

func NewUserRepo(db *gorm.DB) *UserRepo {
	return &UserRepo{db: db}
}

func (r *UserRepo) Create(user *model.User) error {
	return r.db.Create(user).Error
}

func (r *UserRepo) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	return &user, err
}

func (r *UserRepo) GetByOpenID(openID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("open_id = ?", openID).First(&user).Error
	return &user, err
}

func (r *UserRepo) GetByUnionID(unionID string) (*model.User, error) {
	var user model.User
	err := r.db.Where("union_id = ?", unionID).First(&user).Error
	return &user, err
}
