package services

import (
	"context"
	"github.com/habiliai/apidepot/pkg/config"
	"github.com/habiliai/apidepot/pkg/internal/digo"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	"github.com/jackc/pgx/v5"
	"github.com/pkg/errors"
	"strings"
	"text/template"
)

type RuntimeSchema struct {
	dbConfig config.DBConfig

	stackSQLTemplates    map[string][]*template.Template
	installRoleTemplates []*template.Template
	afterCreateUserTmpl  *template.Template
}

func NewRuntimeSchema(
	dbConfig config.DBConfig,
) (*RuntimeSchema, error) {
	rs := &RuntimeSchema{
		dbConfig:          dbConfig,
		stackSQLTemplates: map[string][]*template.Template{},
	}

	{
		stackSqls := map[string][]string{
			"forward": {
				"CREATE USER {{ .Username }} WITH ENCRYPTED PASSWORD '{{ .Password }}' IN ROLE anon, authenticated, service_role",
				"CREATE DATABASE {{ .DBName }} WITH OWNER {{ .Username }}",
			},
			"backward": {
				"DROP USER {{ .Username }}",
				"DROP DATABASE {{ .DBName }}",
			},
		}

		for key, sqls := range stackSqls {
			for _, sql := range sqls {
				tmpl, err := template.New("").Parse(sql)
				if err != nil {
					return nil, errors.Wrapf(err, "failed to parse sql template. key: %s", key)
				}
				rs.stackSQLTemplates[key] = append(rs.stackSQLTemplates[key], tmpl)
			}
		}
	}
	{
		const sqlQueries = `
CREATE ROLE anon NOLOGIN NOINHERIT;
CREATE ROLE authenticated NOLOGIN NOINHERIT;
CREATE ROLE service_role NOLOGIN NOINHERIT BYPASSRLS;
`
		for _, query := range strings.Split(sqlQueries, "\n") {
			query = strings.TrimSpace(query)
			if query == "" {
				continue
			}

			tmpl, err := template.New("").Parse(query)
			if err != nil {
				return nil, errors.Wrapf(err, "failed to parse sql template")
			}

			rs.installRoleTemplates = append(rs.installRoleTemplates, tmpl)
		}
	}
	{
		const afterCreateUser = `
-- Alter public schema change owner of the schema.
ALTER SCHEMA public OWNER TO {{ .Username }};
`
		tmpl, err := template.New("").Parse(afterCreateUser)
		if err != nil {
			return nil, errors.Wrapf(err, "failed to parse sql template")
		}

		rs.afterCreateUserTmpl = tmpl
	}

	return rs, nil
}

const ServiceKeyRuntimeSchema digo.ObjectKey = "runtimeSchema"

func init() {
	digo.ProvideService(ServiceKeyRuntimeSchema, func(ctx *digo.Container) (any, error) {
		switch ctx.Env {
		case digo.EnvProd:
			return NewRuntimeSchema(ctx.Config.DB)
		case digo.EnvTest:
			return NewRuntimeSchema(config.DBConfig{
				Seoul: config.RegionalDBConfig{
					Host:     "localhost",
					User:     "postgres",
					Password: "postgres",
					Name:     "test",
					Port:     6543,
				},
				Singapore: config.RegionalDBConfig{
					Host:     "localhost",
					User:     "postgres",
					Password: "postgres",
					Name:     "test",
					Port:     6543,
				},
			})
		default:
			return nil, errors.New("unknown env")
		}
	})
}

func exec(ctx context.Context, conn *pgx.Conn, sqlTemplate *template.Template, values any) error {
	var sql strings.Builder
	if err := sqlTemplate.Execute(&sql, values); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}

	if _, err := conn.Exec(ctx, sql.String()); err != nil {
		return errors.Wrapf(err, "failed to execute sql")
	}

	return nil
}

func (rs *RuntimeSchema) forward(ctx context.Context, region tcltypes.InstanceZone, values any) error {
	conn, err := pgx.Connect(ctx, rs.dbConfig.GetRegionalConfig(region).GetURI())
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}
	defer conn.Close(ctx)

	for i, sqlTemplate := range rs.stackSQLTemplates["forward"] {
		if err := exec(ctx, conn, sqlTemplate, values); err != nil {
			for j := i - 1; j >= 0; j-- {
				if err := exec(ctx, conn, rs.stackSQLTemplates["backward"][j], values); err != nil {
					logger.Error("failed to rollback", "err", err)
				}
			}
			return errors.Wrapf(err, "failed to execute sql. i: %d", i)
		}
	}

	return nil
}

func (rs *RuntimeSchema) backward(ctx context.Context, region tcltypes.InstanceZone, values any) error {
	conn, err := pgx.Connect(ctx, rs.dbConfig.GetRegionalConfig(region).GetURI())
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}
	defer conn.Close(ctx)

	for i := len(rs.stackSQLTemplates["backward"]) - 1; i >= 0; i-- {
		sqlTemplate := rs.stackSQLTemplates["backward"][i]
		if err := exec(ctx, conn, sqlTemplate, values); err != nil {
			return errors.Wrapf(err, "failed to execute sql. i: %d", i)
		}
	}
	return nil
}

func (rs *RuntimeSchema) CreateUserAndDB(ctx context.Context, zone tcltypes.InstanceZone, username, password, dbname string) error {
	values := struct {
		Username string
		Password string
		DBName   string
	}{username, password, dbname}
	if err := rs.forward(ctx, zone, values); err != nil {
		return err
	}

	var buf strings.Builder
	if err := rs.afterCreateUserTmpl.Execute(&buf, values); err != nil {
		return errors.Wrapf(err, "failed to execute template")
	}

	conn, err := pgx.Connect(ctx, rs.dbConfig.GetRegionalConfig(zone).WithDBName(dbname).GetURI())
	if err != nil {
		return errors.Wrapf(err, "failed to connect to db")
	}
	defer conn.Close(ctx)
	if _, err := conn.Exec(ctx, buf.String()); err != nil {
		return err
	}

	return nil
}

func (rs *RuntimeSchema) DropUserAndDB(ctx context.Context, region tcltypes.InstanceZone, username, dbname string) error {
	return rs.backward(ctx, region, struct {
		Username string
		DBName   string
	}{username, dbname})
}
