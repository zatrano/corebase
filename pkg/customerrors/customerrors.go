package customerrors

import "errors"

var (
	ErrRepoRecordNotFound       = errors.New("kayıt bulunamadı")
	ErrInvalidUserContext       = errors.New("geçersiz kullanıcı context bilgisi")
	ErrUserServiceUserNotFound  = errors.New("kullanıcı bulunamadı")
	ErrPasswordHashingFailed    = errors.New("şifre oluşturulurken bir hata oluştu")
	ErrPasswordUpdateFailed     = errors.New("şifre güncellenirken bir hata oluştu")
	ErrUserCreationFailed       = errors.New("kullanıcı veritabanına kaydedilemedi")
	ErrUserUpdateFailed         = errors.New("kullanıcı veritabanında güncellenemedi")
	ErrUserDeletionFailed       = errors.New("kullanıcı silinirken bir veritabanı hatası oluştu")
	ErrPasswordRequired         = errors.New("şifre alanı boş olamaz")
	ErrContextUserIDNotFound    = errors.New("işlemi yapan kullanıcı kimliği context içinde bulunamadı")
	ErrInvalidCredentials       = errors.New("geçersiz kimlik bilgileri")
	ErrUserNotFound             = errors.New("kullanıcı bulunamadı")
	ErrUserInactive             = errors.New("kullanıcı aktif değil")
	ErrCurrentPasswordIncorrect = errors.New("mevcut şifre hatalı")
	ErrPasswordTooShort         = errors.New("yeni şifre en az 6 karakter olmalıdır")
	ErrPasswordSameAsOld        = errors.New("yeni şifre mevcut şifre ile aynı olamaz")
	ErrAuthGeneric              = errors.New("kimlik doğrulaması sırasında bir hata oluştu")
	ErrProfileGeneric           = errors.New("profil bilgileri alınırken hata")
	ErrUpdatePasswordGeneric    = errors.New("şifre güncellenirken bir hata oluştu")
	ErrHashingFailed            = errors.New("yeni şifre oluşturulurken hata")
	ErrDatabaseUpdateFailed     = errors.New("veritabanı güncellemesi başarısız oldu")
)
