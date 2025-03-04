package svctpl_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
	"github.com/habiliai/apidepot/pkg/internal/stack"
	tcltypes "github.com/habiliai/apidepot/pkg/internal/types"
	vapitest "github.com/habiliai/apidepot/pkg/internal/vapi/test"
	"github.com/stretchr/testify/mock"
	"gorm.io/datatypes"
	"os"
)

func (s *ServiceTestSuite) TestCreateStackFromServiceTemplate() {
	// Given
	githubAccessToken := os.Getenv("YOUR_GITHUB_TOKEN")
	if githubAccessToken == "" {
		s.T().Skip("github token is missing")
	}

	user := domain.User{
		GithubInstallationId: 1,
		GithubAccessToken:    githubAccessToken,
	}
	s.Require().NoError(user.Save(s.db))
	s.users.On("GetUser", mock.Anything).Return(&user, nil).Maybe()

	name := "test stack"
	description := "test description"
	region := tcltypes.InstanceZoneDefault
	logoImageUrl := "logoImageUrl"
	gitRepoName := "habiliai/service-template-example"

	_, rel2 := vapitest.RegisterVapis(s.T(), s, s.vapis)
	svctpl := domain.ServiceTemplate{
		Name:    "test service template",
		GitRepo: "habiliai/service-template-example",
		GitHash: "main",
		VapiIds: datatypes.NewJSONSlice([]uint{rel2.ID}),
	}
	s.Require().NoError(svctpl.Save(s.db))

	prj := domain.Project{
		Name:    "test project",
		OwnerID: user.ID,
		Owner:   user,
	}
	s.Require().NoError(prj.Save(s.db))

	// When
	stack, err := s.svctpls.CreateStackFromServiceTemplate(s.Context, svctpl.ID, stack.CreateStackInput{
		Name:          name,
		Description:   description,
		DefaultRegion: region,
		LogoImageUrl:  logoImageUrl,
		ProjectID:     prj.ID,
		SiteURL:       "localhost:3000",
		GitRepo:       gitRepoName,
		GitBranch:     "main",
	})

	// Then
	s.NoError(err)
	s.NotNil(stack)

	var st domain.Stack
	s.Require().NoError(s.db.First(&st, stack.ID).Error)
	s.Equal(name, st.Name)
	s.Equal(gitRepoName, st.GitRepo)
}
