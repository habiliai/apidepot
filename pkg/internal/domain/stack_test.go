package domain_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
)

func (s *DomainTestSuite) Test_WhenInsertingDuplicatedStacks_ShouldBeOk() {
	// Given
	user := domain.User{
		Name: "test-user",
	}

	project := domain.Project{
		Name:  "project-1",
		Owner: user,
	}
	s.Require().NoError(project.Save(s.db))

	stack := &domain.Stack{
		ProjectID: project.ID,
		Name:      "stack-1",
	}
	s.Require().NoError(stack.Save(s.db))

	// When
	stack1 := &domain.Stack{
		ProjectID: project.ID,
		Name:      "stack-1",
	}
	err1 := stack1.Save(s.db)

	// Then
	s.Error(err1)
}

func (s *DomainTestSuite) TestFindStackVapiByStackIDAndVapiID() {
	user := domain.User{
		Name: "test-user",
	}
	s.Require().NoError(user.Save(s.db))
	project := domain.Project{
		Name:  "project-1",
		Owner: user,
	}
	s.Require().NoError(project.Save(s.db))

	stack := domain.Stack{
		ProjectID: project.ID,
		Name:      "stack-1",
	}
	s.Require().NoError(stack.Save(s.db))

	vapiPackage := domain.VapiPackage{
		Name:    "vapi-1",
		OwnerId: user.ID,
	}
	s.Require().NoError(vapiPackage.Save(s.db))

	vapiRelease := domain.VapiRelease{
		PackageID: vapiPackage.ID,
		Version:   "1.0.0",
	}
	s.Require().NoError(vapiRelease.Save(s.db))

	stackVapi := domain.StackVapi{
		StackID: stack.ID,
		VapiID:  vapiRelease.ID,
	}
	s.Require().NoError(stackVapi.Save(s.db))
	s.Require().NoError(s.db.Preload("Vapis").First(&stack).Error)

	result, err := domain.GetStackVapiByStackIDAndVapiID(s.db, stack.ID, vapiRelease.ID)
	s.Require().NoError(err)

	s.NotNil(result)
	s.Equal(stack.ID, result.Stack.ID)
	s.Equal(stack.ProjectID, result.Stack.Project.ID)
	s.Equal(vapiRelease.ID, result.Vapi.ID)
	s.Equal(vapiPackage.ID, result.Vapi.Package.ID)
	s.T().Logf("%v", result)
}
