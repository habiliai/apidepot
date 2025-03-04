package svctpl_test

import (
	"github.com/habiliai/apidepot/pkg/internal/domain"
)

func (s *ServiceTestSuite) TestGetServiceTemplates() {
	// Given 3 service templates
	cursor := uint(0)
	limit := uint(10)
	expectedServiceTemplates := []domain.ServiceTemplate{
		{
			Name:   "hamster kombat",
			Detail: `The game is a 2D fighting game with gameplay mechanics similar to the Street Fighter series. The player can choose between eight characters, each with their own fighting style and special moves. The game features a single-player mode where the player must defeat a series of opponents to become the champion, as well as a two-player mode where two players can fight against each other. The game also features a training mode where the player can practice their moves and combos. The game is set in a fictional world where humans and anthropomorphic animals coexist, and the characters are all anthropomorphic animals. The game features colorful graphics and a catchy soundtrack.`,
		},
		{
			Name:   "blackjack",
			Detail: `The game is a card game where the player competes against the dealer to see who can get the closest to 21 without going over. The player is dealt two cards and can choose to "hit" to receive another card or "stand" to keep their current hand. The player can also choose to "double down" to double their bet and receive one more card, or "split" if they are dealt two cards of the same value. The dealer must hit until they reach 17 or higher. The player wins if they have a higher hand than the dealer without going over 21, or if the dealer busts. The game features realistic graphics and sound effects, as well as a variety of options to customize the gameplay experience.`,
		},
		{
			Name:   "token wallet",
			Detail: `The application is a digital wallet that allows users to store, send, and receive tokens on the blockchain. Users can create an account and securely store their tokens in the wallet. They can send tokens to other users by entering their wallet address and the amount of tokens to send. Users can also receive tokens by sharing their wallet address with others. The wallet supports multiple tokens and allows users to view their transaction history and token balances. The application is secure and easy to use, with a user-friendly interface and intuitive controls.`,
		},
	}

	s.Require().NoError(expectedServiceTemplates[0].Save(s.db))
	s.Require().NoError(expectedServiceTemplates[1].Save(s.db))
	s.Require().NoError(expectedServiceTemplates[2].Save(s.db))

	s.Run("when no search query, should be all service templates ordered by id", func() {
		expectedNextCursor := uint(3)

		// When
		output, err := s.svctpls.SearchServiceTemplates(s, cursor, limit, "")

		// Then
		s.NoError(err)
		s.Equalf(len(expectedServiceTemplates), len(output.ServiceTemplates), "expected %d service templates, got %d", len(expectedServiceTemplates), len(output.ServiceTemplates))
		for i := range expectedServiceTemplates {
			s.Equal(expectedServiceTemplates[i].ID, output.ServiceTemplates[i].ID, "expected ID %d, got %d", expectedServiceTemplates[i].ID, output.ServiceTemplates[i].ID)
			s.Equal(expectedServiceTemplates[i].Name, output.ServiceTemplates[i].Name, "expected Name %s, got %s", expectedServiceTemplates[i].Name, output.ServiceTemplates[i].Name)
		}
		s.Equal(expectedNextCursor, output.NextCursor)
	})

	s.Run("when search query is 'hamster', should return only service template with name 'hamster kombat'", func() {
		expectedNextCursor := uint(1)

		// When
		output, err := s.svctpls.SearchServiceTemplates(s, cursor, limit, "hamster")

		s.T().Logf("output: %+v", output)
		// Then
		s.NoError(err)
		s.Len(output.ServiceTemplates, 1)
		s.Equal(expectedServiceTemplates[0].ID, output.ServiceTemplates[0].ID)
		s.Equal(expectedServiceTemplates[0].Name, output.ServiceTemplates[0].Name)
		s.Equal(expectedNextCursor, output.NextCursor)
	})

	s.Run("when search query is 'game', should return 2 service templates", func() {
		expectedNextCursor := uint(2)

		// When
		output, err := s.svctpls.SearchServiceTemplates(s, cursor, limit, "game")

		// Then
		s.NoError(err)
		s.Len(output.ServiceTemplates, 2)
		s.Equal(expectedNextCursor, output.NextCursor)
	})

	s.Run("when search query is 'wallet', should return 1 service template", func() {
		expectedNextCursor := uint(1)

		// When
		output, err := s.svctpls.SearchServiceTemplates(s, cursor, limit, "wallet")

		// Then
		s.NoError(err)
		s.Len(output.ServiceTemplates, 1)
		s.Equal(expectedServiceTemplates[2].ID, output.ServiceTemplates[0].ID)
		s.Equal(expectedServiceTemplates[2].Name, output.ServiceTemplates[0].Name)
		s.Equal(expectedNextCursor, output.NextCursor)
	})

	s.Run("when search query is 'black ga', should return 1 service template", func() {
		expectedNextCursor := uint(2)

		// When
		output, err := s.svctpls.SearchServiceTemplates(s, cursor, limit, "black ga")

		// Then
		s.Require().NoError(err)
		s.Require().Len(output.ServiceTemplates, 2)
		s.Require().Equal(int64(2), output.NumTotal)
		s.Equal(expectedServiceTemplates[0].ID, output.ServiceTemplates[0].ID)
		s.Equal(expectedServiceTemplates[0].Name, output.ServiceTemplates[0].Name)
		s.Equal(expectedServiceTemplates[1].ID, output.ServiceTemplates[1].ID)
		s.Equal(expectedServiceTemplates[1].Name, output.ServiceTemplates[1].Name)
		s.Equal(expectedNextCursor, output.NextCursor)
	})
}

func (s *ServiceTestSuite) TestGetServiceTemplateByID() {
	// Given
	expectedServiceTemplate := domain.ServiceTemplate{
		Name: "service-template-1",
	}
	s.Require().NoError(expectedServiceTemplate.Save(s.db))

	// When
	output, err := s.svctpls.GetServiceTemplateByID(s, expectedServiceTemplate.ID)

	// Then
	s.NoError(err)
	s.Equal(expectedServiceTemplate.Name, output.Name)
}
