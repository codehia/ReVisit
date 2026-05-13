package tui

import (
	"database/sql"

	tea "charm.land/bubbletea/v2"
	"github.com/codehia/goflash/internal/types"
)

type Screen int

const (
	ScreenTopicList Screen = iota
	ScreenCardList
	ScreenCardAttempt
	ScreenEvalResult
)

type RootModel struct {
	db            *sql.DB
	currentScreen Screen
	// topic list state
	topics          []types.Topic
	selectedTopicID *string
	cursor          int
	// card list state
	cards     []types.Card
	cardIndex int
	// card attempt state
	userAnswer string
	// eval result state
	evalResult types.EvalResult
}

func NewRootModel(db *sql.DB) RootModel {
	return RootModel{db: db, currentScreen: ScreenTopicList}
}

func (m RootModel) Init() tea.Cmd {
	return initTopicList(m)
}

type TopicSelectedMsg struct {
	topicID string
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TopicSelectedMsg:
		m.selectedTopicID = &msg.topicID
		m.currentScreen = ScreenCardList
		return m, InitCardList(m)
	}

	switch m.currentScreen {
	case ScreenTopicList:
		return updateTopicList(msg, m)
	default:
		return m, nil
	}
}

func (m RootModel) View() tea.View {
	switch m.currentScreen {
	case ScreenTopicList:
		return tea.NewView(topicListView(m))
	default:
		return tea.NewView("")
	}
}
