package main

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
	"os"
	"encoding/json"
	"github.com/PuerkitoBio/goquery"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// --- Styling ---
var (
	titleStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("63")).Bold(true)
	helpStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	errorStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("9")).Bold(true)
	keyStyle       = lipgloss.NewStyle().Foreground(lipgloss.Color("252"))
	valueStyle     = lipgloss.NewStyle().Foreground(lipgloss.Color("86"))
	characterStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("213"))
	mainCharStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("86")).Bold(true)
)

const baseURL = "https://www.sa.playblackdesert.com"

// --- Data Structures ---

type FamilyInfo struct {
	Name          string
	CreationDate  string
	Guild         string
	PAPD          string
	Energy        string
	Contribution  string
}

type LifeSkill struct {
	Name   string
	LevelName  string
	LevelInt string
	Mastery string
}

type Character struct {
	Name   string
	Class  string
	Level  string
}

type FullProfile struct {
	FamilyInfo FamilyInfo
	LifeSkills []LifeSkill
	Characters []Character
}

// --- Application State ---
type appState int

const (
	stateSearch appState = iota
	stateLoading
	stateProfileView
	stateError
)

// --- Bubbletea Model ---

type model struct {
	state      appState
	textInput  textinput.Model
	spinner    spinner.Model
	viewport   viewport.Model
	profile    FullProfile
	errorMsg   string
}

func initialModel() model {
	ti := textinput.New()
	ti.Placeholder = "Family Name"
	ti.Focus()
	ti.CharLimit = 32
	ti.Width = 30

	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))

	vp := viewport.New(80, 20)

	return model{
		state:     stateSearch,
		textInput: ti,
		spinner:   s,
		viewport:  vp,
	}
}

// --- Bubbletea Messages ---

type profileURLMsg string
type profileResultMsg FullProfile
type errMsg struct{ err error }

// --- Scraper ---

func findProfileURL(familyName string) tea.Cmd {
	return func() tea.Msg {
		searchURL := fmt.Sprintf("%s/pt-BR/Adventure?checkSearchText=True&searchType=2&searchKeyword=%s", baseURL, familyName)
		
		doc, err := getDocument(searchURL)
		if err != nil {
			return errMsg{err}
		}
		selector := "div.box_list_area > ul > li > div.title > a"

		profilePath, exists := doc.Find(selector).First().Attr("href")
		if !exists {
			return errMsg{fmt.Errorf("Could not find profile link for '%s'. May not exist or profile is private", url.QueryEscape(familyName))}
		}

		return profileURLMsg(profilePath)
	}
}

func fetchProfileData(profileURL string) tea.Cmd {
	return func() tea.Msg {
		doc, err := getDocument(profileURL)
		if err != nil {
			return errMsg{err}
		}

		var profile FullProfile
		familyInfoBox := doc.Find("div.profile_detail")
		profile.FamilyInfo.Name = strings.TrimSpace(familyInfoBox.Find("div.nick_wrap > p.nick").Text())
		profile.FamilyInfo.CreationDate = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(1) > span.desc").Text())
		profile.FamilyInfo.Guild = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li > span.guild > a").Text())
		profile.FamilyInfo.PAPD = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(3) > span.desc").Text())
		profile.FamilyInfo.Energy = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(4) > span.desc").Text())
		profile.FamilyInfo.Contribution = strings.TrimSpace(familyInfoBox.Find("ul.line_list > li:nth-child(5) > span.desc").Text())
		
		lifeskillsBox := doc.Find("ul.character_data_box")
		lifeskillsBox.Find("li").Each(func(i int, s *goquery.Selection) {
			name := strings.TrimSpace(s.Find("span.spec_name").Text())

			rawLevel := strings.TrimSpace(s.Find("span.spec_level").Text())
			normalizedLevel := strings.Replace(rawLevel, "Nv.", " Nv.", 1)
			normalizedSplit := strings.SplitN(normalizedLevel, " ", 2)
			levelName := normalizedSplit[0]
			levelInt := normalizedSplit[1]

			mastery := strings.TrimSpace(s.Find("span.spec_stat").Text())

			profile.LifeSkills = append(profile.LifeSkills, LifeSkill{
				Name: name,
				LevelName: levelName,
				LevelInt: levelInt,
				Mastery: mastery,
			})
		})

		charactersBox := doc.Find("ul.character_list")
		charactersBox.Find("li").Each(func(i int, s *goquery.Selection) {

			isMain := strings.Contains(s.Find("p.character_name span.selected_label").Text(), "Personagem Principal")
			name := strings.TrimSpace(
				s.Find("p.character_name").Contents().FilterFunction(func(i int, sel *goquery.Selection) bool {
					return goquery.NodeName(sel) == "#text"
				}).Text(),
			)

			class := strings.TrimSpace(s.Find("span.character_symbol em:nth-child(2)").Text())
			level := strings.TrimSpace(s.Find("span.character_info span:nth-child(2)").Text())
			if isMain {
				level += " Personagem Principal"
			}

			profile.Characters = append(profile.Characters, Character{
				Name:   name,
				Class:  class,
				Level:  level,
			})
		})
		
		// Debug .json file. Can append a listener to this file to grab content on every update. Could export too.
		jsonData, err := json.MarshalIndent(profile, "", "  ")
		if err != nil {
			return errMsg{err}
		}
		err = os.WriteFile("debug_output.json", jsonData, 0644)
		if err != nil {
			return errMsg{err}
		}
		return profileResultMsg(profile)
	}
}

func getDocument(url string) (*goquery.Document, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/108.0.0.0 Safari/537.36")

	client := &http.Client{Timeout: 10 * time.Second}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("server returned status: %s", res.Status)
	}

	return goquery.NewDocumentFromReader(res.Body)
}

// --- Model "Workflow" ---

func (m *model) Init() tea.Cmd {
	return textinput.Blink
}

func (m *model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch m.state {
		case stateSearch, stateError:
			switch msg.Type {
			case tea.KeyEnter:
				m.state = stateLoading
				m.errorMsg = ""
				return m, tea.Batch(m.spinner.Tick, findProfileURL(m.textInput.Value()))
			case tea.KeyCtrlC, tea.KeyEsc:
				return m, tea.Quit
			default:
				m.textInput, cmd = m.textInput.Update(msg)
				return m, cmd
			}
		
		case stateProfileView:
			switch msg.Type {
			case tea.KeyBackspace, tea.KeyEsc:
				m.state = stateSearch
				m.textInput.Reset()
				m.textInput.Focus()
				return m, nil
			case tea.KeyCtrlC:
				return m, tea.Quit
			}
		}

	// --- Asynchronous Results ---

	case profileURLMsg:
		return m, fetchProfileData(string(msg))

	case profileResultMsg:
		m.state = stateProfileView
		m.profile = FullProfile(msg)
		m.viewport.SetContent(m.formatProfile())
		m.viewport.GotoTop()
		return m, nil

	case errMsg:
		m.state = stateError
		m.errorMsg = msg.err.Error()
		return m, nil
	
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width - 4
		m.viewport.Height = msg.Height - 2
		
	case spinner.TickMsg:
		if m.state == stateLoading {
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
	}
	
	if m.state == stateProfileView {
		m.viewport, cmd = m.viewport.Update(msg)
		cmds = append(cmds, cmd)
	}
	
	return m, tea.Batch(cmds...)
}

func (m *model) View() string {
	var b strings.Builder

	b.WriteString(titleStyle.Render("BDO Family Profile Viewer") + "\n\n")

	switch m.state {
	case stateSearch:
		b.WriteString("Enter a family name to search:\n")
		b.WriteString(m.textInput.View() + "\n\n")
		b.WriteString(helpStyle.Render("Enter: search | Esc/Ctrl+C: quit"))
	
	case stateLoading:
		b.WriteString(fmt.Sprintf("%s Loading data for '%s'...", m.spinner.View(), m.textInput.Value()))

	case stateProfileView:
		b.WriteString(m.viewport.View() + "\n")
		b.WriteString(helpStyle.Render("↑/↓: scroll | backspace: back to search | Ctrl+C: quit"))
	
	case stateError:
		b.WriteString("An error occurred:\n")
		b.WriteString(errorStyle.Render(m.errorMsg) + "\n\n")
		b.WriteString("Press Enter to search again or Esc to quit.\n")
		b.WriteString(m.textInput.View())
	}
	
	return lipgloss.NewStyle().Margin(1, 2).Render(b.String())
}

func (m *model) formatProfile() string {
	var b strings.Builder

	// Family
	p := m.profile
	b.WriteString(keyStyle.Render("Family: ") + valueStyle.Render(p.FamilyInfo.Name) + "\n")
	b.WriteString(keyStyle.Render("Guild: ") + valueStyle.Render(p.FamilyInfo.Guild) + "\n")
	b.WriteString(keyStyle.Render("Created: ") + valueStyle.Render(p.FamilyInfo.CreationDate) + "\n")
	b.WriteString(fmt.Sprintf("%s | %s | %s\n",
		keyStyle.Render("PAPD: ") + valueStyle.Render(p.FamilyInfo.PAPD),
		keyStyle.Render("Energy: ") + valueStyle.Render(p.FamilyInfo.Energy),
		keyStyle.Render("CP: ") + valueStyle.Render(p.FamilyInfo.Contribution),
	))
	b.WriteString("\n" + titleStyle.Render("Characters") + "\n")
	b.WriteString(strings.Repeat("-", 40) + "\n")
	
	// Characters
	for _, char := range p.Characters {
		line := fmt.Sprintf("• %-18s %-12s %s", char.Name, char.Class, char.Level)
		b.WriteString(characterStyle.Render(line) + "\n")
	}
	
	b.WriteString("\n" + titleStyle.Render("Life Skills") + "\n")
	b.WriteString(strings.Repeat("-", 40) + "\n")

	// Life Skills
	for _, skill := range p.LifeSkills {
		b.WriteString(fmt.Sprintf("• %-34s %-28s %s\n", keyStyle.Render(skill.Name), valueStyle.Render(skill.LevelName), valueStyle.Render(skill.LevelInt))) 
	}

	return b.String()
}


func main() {
	initialModel := initialModel()
	p := tea.NewProgram(&initialModel, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Alas, there's been an error: %v", err)
	}
}