package svctpl

import (
	"context"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/helpers"
	"github.com/pkg/errors"
	"strings"
)

type (
	SearchServiceTemplatesOutput struct {
		ServiceTemplates []domain.ServiceTemplate
		NextCursor       uint
		NumTotal         int64
	}
)

func (s *serviceImpl) SearchServiceTemplates(
	ctx context.Context,
	cursor uint,
	limit uint,
	searchQuery string,
) (*SearchServiceTemplatesOutput, error) {
	tx := helpers.GetTx(ctx)

	if limit == 0 {
		limit = 10
	}

	var o SearchServiceTemplatesOutput
	stmt := tx.Model(&domain.ServiceTemplate{})
	if searchQuery != "" {
		searchTermBuilder := strings.Builder{}
		for _, q := range strings.Split(searchQuery, " ") {
			if searchTermBuilder.Len() > 0 {
				searchTermBuilder.WriteString(" | ")
			}
			q = strings.TrimSpace(q) + ":*"
			searchTermBuilder.WriteString(q)
		}

		searchTerm := searchTermBuilder.String()
		stmt = stmt.Raw(`
			SELECT service_templates.*
			FROM service_templates, to_tsquery(?) query
			WHERE service_templates.tsv @@ query AND service_templates.deleted_at = 0
			ORDER BY ts_rank_cd(service_templates.tsv, query) DESC, service_templates.id DESC
		`, searchTerm)

		if err := tx.Table("(?) as A", stmt).Count(&o.NumTotal).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to count service templates")
		}
		stmt = stmt.Offset(int(cursor)).Limit(int(limit))
	} else {
		if err := stmt.Count(&o.NumTotal).Error; err != nil {
			return nil, errors.Wrapf(err, "failed to count service templates")
		}
		stmt = stmt.Order("service_templates.id ASC").Where("id > ?", cursor).Limit(int(limit))
	}

	if err := stmt.Find(&o.ServiceTemplates).Error; err != nil {
		return nil, errors.Wrapf(err, "failed to get service templates")
	}

	size := len(o.ServiceTemplates)

	if size > 0 {
		if searchQuery != "" {
			o.NextCursor = cursor + uint(size)
		} else {
			o.NextCursor = o.ServiceTemplates[size-1].ID
		}
	}

	return &o, nil
}

func (s *serviceImpl) GetServiceTemplateByID(ctx context.Context, id uint) (*domain.ServiceTemplate, error) {
	tx := helpers.GetTx(ctx)

	serviceTemplate, err := domain.FindServiceTemplateByID(tx, id)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get service template by id")
	}

	return serviceTemplate, nil
}
