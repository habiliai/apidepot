package domain

import (
	"fmt"
	tclerrors "github.com/habiliai/apidepot/pkg/errors"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/habiliai/apidepot/pkg/internal/util/stringbuilder"
	"github.com/pkg/errors"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/plugin/soft_delete"
	"strings"
)

type DB struct {
	Name     string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type Storage struct {
	S3Bucket string `json:"s3_bucket"`
	TenantID string `json:"tenant_id"`
}

type Postgrest struct {
	Schemas []string `json:"schemas"`
}

type StackVapiEnvVar struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type Stack struct {
	Model
	DeletedAt soft_delete.DeletedAt

	ProjectID uint    `gorm:"uniqueIndex:stacks_name_idx_uniq,where:deleted_at=0"`
	Project   Project `gorm:"foreignKey:ProjectID"`

	GitRepo   string
	GitBranch string

	Name    string `gorm:"uniqueIndex:stacks_name_idx_uniq,where:deleted_at=0"`
	Hash    string `gorm:"uniqueIndex:stacks_hash_idx_uniq,where:deleted_at=0"`
	Domain  string
	Scheme  string
	SiteURL string

	Description       string
	LogoImageUrl      string
	ServiceTemplateID *uint
	ServiceTemplate   *ServiceTemplate `gorm:"foreignKey:ServiceTemplateID"`
	DefaultRegion     tcltypes.InstanceZone

	DB datatypes.JSONType[DB]

	AuthEnabled bool
	Auth        datatypes.JSONType[Auth]

	AdminApiKey string
	AnonApiKey  string

	StorageEnabled bool
	Storage        datatypes.JSONType[Storage]

	PostgrestEnabled bool
	Postgrest        datatypes.JSONType[Postgrest]

	Vapis     []StackVapi
	Instances []Instance

	VapiEnvVars datatypes.JSONSlice[StackVapiEnvVar]

	CustomVapis              []CustomVapi
	TelegramMiniappPromotion *TelegramMiniappPromotion `gorm:"foreignKey:StackID"`
}

type StackVapi struct {
	VapiID  uint        `gorm:"primarykey"`
	Vapi    VapiRelease `gorm:"foreignKey:VapiID"`
	StackID uint        `gorm:"primarykey"`
	Stack   Stack       `gorm:"foreignKey:StackID"`
}

type StackHistory struct {
	Model

	StackID uint
	Stack   Stack `gorm:"foreignKey:StackID"`

	StorageSize int
}

func (d DB) PostgresURI(host string, port int) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.Username,
		d.Password,
		host,
		port,
		d.Name,
	)
}

func (s *StackVapi) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(s).Error, "failed to save stack vapi")
}

func (s *StackVapi) Create(db *gorm.DB) error {
	return errors.Wrapf(db.Create(s).Error, "failed to create stack vapi")
}

func (s Stack) Namespace() string {
	return fmt.Sprintf("ns-%s", s.Hash)
}

func (s Stack) Endpoint() string {
	return fmt.Sprintf("%s://%s", s.Scheme, s.Domain)
}

func (s Stack) ServicePath(subPath string) string {
	if !strings.HasPrefix(subPath, "/") {
		subPath = "/" + subPath
	}
	return s.Endpoint() + subPath
}

func (s *Stack) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(s).Error, "failed to save stack")
}

func (s *Stack) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(s).Error, "failed to delete stack")
}

func (s *Stack) String() string {
	return stringbuilder.New("Stack").
		AddField("ID", s.ID).
		AddField("Name", s.Name).
		AddField("ProjectID", s.ProjectID).
		AddField("Domain", s.Domain).
		AddField("Scheme", s.Scheme).
		AddField("SiteURL", s.SiteURL).
		AddField("Hash", s.Hash).
		AddField("AuthEnabled", s.AuthEnabled).
		AddField("StorageEnabled", s.StorageEnabled).
		AddField("PostgrestEnabled", s.PostgrestEnabled).
		AddField("AdminApiKey", s.AdminApiKey).
		AddField("AnonApiKey", s.AnonApiKey).
		AddField("Vapis", s.Vapis).
		AddField("CustomVapis", s.CustomVapis).
		String()
}

func (s *Stack) GetVapiReleases() []VapiRelease {
	var vapiReleases []VapiRelease
	for _, vapi := range s.Vapis {
		vapiReleases = append(vapiReleases, vapi.Vapi)
	}

	return vapiReleases
}

func (s *StackVapi) Delete(db *gorm.DB) error {
	return errors.Wrapf(db.Delete(s).Error, "failed to delete stack vapi")
}

func (s Stack) ToViewModel() Stack {
	auth := s.Auth.Data()
	auth.JWTSecret = ""

	s.Auth = datatypes.NewJSONType(auth)
	return s
}

func (s *Stack) GetVapiEnvVarsMap() map[string]string {
	envVars := make(map[string]string, len(s.VapiEnvVars))
	for _, envVar := range s.VapiEnvVars {
		envVars[envVar.Name] = envVar.Value
	}

	return envVars
}

func (s *Stack) SetVapiEnvVarsMap(envVars map[string]string) {
	vapiEnvVars := make([]StackVapiEnvVar, 0, len(envVars))
	for key, envVar := range envVars {
		vapiEnvVars = append(vapiEnvVars, StackVapiEnvVar{
			Name:  key,
			Value: envVar,
		})
	}

	s.VapiEnvVars = datatypes.NewJSONSlice(vapiEnvVars)
}

func (h *StackHistory) Save(db *gorm.DB) error {
	return errors.Wrapf(db.Save(h).Error, "failed to save stack history")
}

func FindStackByID(
	db *gorm.DB,
	id uint,
	options ...FindOptions,
) (*Stack, error) {
	opt := MergeFindOptions(options...)

	if opt.locking != nil {
		db = db.Clauses(*opt.locking)
	}

	var stack Stack
	if err := db.
		Preload(clause.Associations).
		Preload("Vapis.Vapi").
		Preload("Vapis.Vapi.Package").
		Preload("TelegramMiniappPromotion.Views").
		First(&stack, id).Error; err != nil {
		return nil, err
	}

	return &stack, nil
}

func FindStacks(db *gorm.DB) ([]Stack, error) {
	var stacks []Stack
	if err := db.
		Preload(clause.Associations).
		Preload("Vapis.Vapi").
		Find(&stacks).Error; err != nil {
		return nil, err
	}

	return stacks, nil
}

func DeleteStackByID(db *gorm.DB, id uint) error {
	return errors.Wrapf(db.Delete(&Stack{}, id).Error, "failed to delete project")
}

func GetStackVapiByStackIDAndVapiID(db *gorm.DB, stackId uint, vapiId uint) (StackVapi, error) {
	var stackVapi StackVapi
	if err := db.
		Preload("Vapi").
		Preload("Stack").
		Preload("Stack.Project").
		Preload("Vapi.Package").
		Where("stack_id = ? AND vapi_id = ?", stackId, vapiId).
		First(&stackVapi).
		Error; err != nil {
		return stackVapi, err
	}

	return stackVapi, nil
}

func (s *Stack) ValidateVapiNameUniqueness(
	tx *gorm.DB,
	name string,
) (err error) {
	var result int64
	if err = tx.
		Model(&StackVapi{}).
		InnerJoins("Vapi").
		InnerJoins("Vapi.Package", tx.Where(&VapiPackage{Name: name})).
		Where("stack_vapis.stack_id = ?", s.ID).
		Count(&result).Error; err != nil {
		return errors.Wrapf(err, "failed to get stack vapi package names")
	}

	if result > 0 {
		return errors.Wrapf(tclerrors.ErrBadRequest, "name('%s') is duplicated in the stack vapi(name='%s')", name, name)
	}

	if err = tx.
		Model(&CustomVapi{}).
		Select("stack_id = ? AND name = ?", s.ID, name).
		Count(&result).Error; err != nil {
		return errors.Wrapf(err, "failed to get custom vapi")
	}

	if result > 0 {
		return errors.Wrapf(tclerrors.ErrBadRequest, "name('%s') is duplicated in the stack custom vapi(name='%s')", name, name)
	}

	return nil
}
