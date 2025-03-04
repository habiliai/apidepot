package vapitest

import (
	"context"
	"github.com/google/uuid"
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/vapi"
	"github.com/stretchr/testify/require"
	"testing"
)

func RegisterVapis(
	t *testing.T,
	ctx context.Context,
	vapiService vapi.Service,
) (*domain.VapiRelease, *domain.VapiRelease) {
	t.Logf("-- enable vapi")

	rel1, err := vapiService.Register(
		ctx,
		"habiliai/vapi-helloworld-sns",
		"main",
		"sns",
		"",
		[]string{},
		uuid.NewString(),
		"",
	)
	require.NoError(t, err)

	rel2, err := vapiService.Register(
		ctx,
		"habiliai/vapi-helloworld",
		"main",
		"helloworld",
		"",
		[]string{},
		uuid.NewString(),
		"",
	)
	require.NoError(t, err)

	return rel1, rel2
}
