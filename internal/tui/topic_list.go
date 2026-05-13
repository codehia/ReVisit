package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"github.com/codehia/goflash/internal/store"
	"github.com/codehia/goflash/internal/types"
)

type topicsLoadedMsg struct {
	topics []types.Topic
	err    error
}

func initTopicList(m RootModel) tea.Cmd {
	return func() tea.Msg {
		tags, err := store.GetTopLevelTopics(m.db)
		if err != nil {
			return topicsLoadedMsg{err: err}
		}
		var topics []types.Topic
		for _, tag := range tags {
			topics = append(topics, types.Topic{Name: tag.Name, ID: tag.ID})
		}
		return topicsLoadedMsg{topics: topics}
	}
}

func updateTopicList(msg tea.Msg, m RootModel) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case topicsLoadedMsg:
		m.topics = msg.topics
		return m, nil

	case tea.KeyPressMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.topics)-1 {
				m.cursor++
			}
		case "enter", "space":
			id := m.topics[m.cursor].ID
			return m, func() tea.Msg {
				return TopicSelectedMsg{topicID: *id}

			}
		}
	}
	return m, nil
}

func topicListView(m RootModel) string {
	var s strings.Builder
	s.WriteString("Select the topic you want to work on:\n\n")
	for i, topic := range m.topics {
		cursor := " "
		if m.cursor == i {
			cursor = ">"
		}
		fmt.Fprintf(&s, "%s %s\n", cursor, topic.Name)
	}
	s.WriteString("\nPress q to quit.\n")
	return s.String()
}
