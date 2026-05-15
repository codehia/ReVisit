package tui

import (
	"database/sql"

	ta "charm.land/bubbles/v2/textarea"
	tea "charm.land/bubbletea/v2"
"github.com/codehia/goflash/internal/store"
	"github.com/codehia/goflash/internal/types"
)

type TopicSummary = types.TopicSummary

type Screen int

const (
	ScreenTopicList Screen = iota
	ScreenCardQuestion
	ScreenCardAttempt
	ScreenEvalResult
	ScreenDone
)

// Card dimensions — fixed, never resize with terminal.
const (
	cardWidth  = 80
	cardHeight = 40
	cardInnerW = cardWidth - 2  // 78: inner content width
	cardInnerH = cardHeight - 2 // 38: inner content height
	headerH = 3
	footerH = 2
)

type RootModel struct {
	db            *sql.DB
	currentScreen Screen
	termWidth     int
	termHeight    int
	ready         bool
	// topic list
	topics          []types.TopicSummary
	selectedTopicID *string
	topicName       string
	cursor          int
	// cards
	cards     []store.Card
	cardIndex int
	// attempt
	textarea   ta.Model
	userAnswer string
	// eval
	evalResult types.EvalResult
	// session
	sessionScores []int
}

func NewRootModel(db *sql.DB) RootModel {
	return RootModel{db: db, currentScreen: ScreenTopicList}
}

func (m RootModel) Init() tea.Cmd {
	return tea.Batch(initTopicList(m), func() tea.Msg { return tea.RequestWindowSize() })
}

type TopicSelectedMsg struct {
	topicID   string
	topicName string
}

type CardSelectedMsg struct {
	cardID string
}

type EvalResultMsg struct {
	result types.EvalResult
	err    error
}

func (m RootModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.termWidth = msg.Width
		m.termHeight = msg.Height
		m.ready = true
		return m, nil

	case TopicSelectedMsg:
		m.selectedTopicID = &msg.topicID
		m.topicName = msg.topicName
		m.cursor = 0
		m.cardIndex = 0
		m.sessionScores = nil
		m.currentScreen = ScreenCardQuestion
		return m, InitCardList(m)

	case cardsLoadedMsg:
		m.cards = msg.cards
		return m, nil

	case CardSelectedMsg:
		m.textarea = buildTextarea()
		m.currentScreen = ScreenCardAttempt
		return m, nil

	case EvalResultMsg:
		m.evalResult = msg.result
		m.sessionScores = append(m.sessionScores, msg.result.Score)
		m.currentScreen = ScreenEvalResult
		return m, nil
	}

	switch m.currentScreen {
	case ScreenTopicList:
		return updateTopicList(msg, m)
	case ScreenCardQuestion:
		return updateCardQuestion(msg, m)
	case ScreenCardAttempt:
		return updateCardAttempt(msg, m)
	case ScreenEvalResult:
		return updateEvalResult(msg, m)
	case ScreenDone:
		return updateDone(msg, m)
	default:
		return m, nil
	}
}

func (m RootModel) View() tea.View {
	if !m.ready {
		return tea.NewView("")
	}

	out := renderCard(CardParams{
		TermW:       m.termWidth,
		TermH:       m.termHeight,
		BorderColor: colorPurple,
		BgColor:     colorBase,
		Header:      screenHeader(m),
		Body:        screenBody(m),
		Footer:      screenFooter(m),
	})
	return tea.NewView(out)
}

// screenHeader dispatches to the active screen's header fn.
func screenHeader(m RootModel) string {
	switch m.currentScreen {
	case ScreenTopicList:
		return topicListHeader(m)
	case ScreenCardQuestion:
		return cardQuestionHeader(m)
	case ScreenCardAttempt:
		return cardAttemptHeader(m)
	case ScreenEvalResult:
		return evalResultHeader(m)
	case ScreenDone:
		return doneHeader(m)
	default:
		return ""
	}
}

// screenBody dispatches to the active screen's body fn.
func screenBody(m RootModel) string {
	switch m.currentScreen {
	case ScreenTopicList:
		return topicListBody(m)
	case ScreenCardQuestion:
		return cardQuestionBody(m)
	case ScreenCardAttempt:
		return cardAttemptBody(m)
	case ScreenEvalResult:
		return evalResultBody(m)
	case ScreenDone:
		return doneBody(m)
	default:
		return ""
	}
}

// screenFooter dispatches to the active screen's footer fn.
func screenFooter(m RootModel) string {
	switch m.currentScreen {
	case ScreenTopicList:
		return topicListFooter(m)
	case ScreenCardQuestion:
		return cardQuestionFooter(m)
	case ScreenCardAttempt:
		return cardAttemptFooter(m)
	case ScreenEvalResult:
		return evalResultFooter(m)
	case ScreenDone:
		return doneFooter(m)
	default:
		return ""
	}
}

func buildTextarea() ta.Model {
	t := ta.New()
	t.Placeholder = "Answer to the selected card ...."
	t.DynamicHeight = true
	t.MinHeight = 4
	t.MaxHeight = 15
	t.SetWidth(contentWidth - 6)
	t.SetVirtualCursor(false)
	t.ShowLineNumbers = false
	t.Prompt = ""

	t.Focus()
	return t
}
