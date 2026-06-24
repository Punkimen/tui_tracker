package ui

import (
	"fmt"
	"time"
)

func GetAllDaysInMonth(m time.Month) []int {
	now := time.Now()
	lastDay := time.Date(now.Year(), m+1, 0, 0, 0, 0, 0, time.Local).Day()
	days := make([]int, lastDay)

	for i := range days {
		days[i] = i + 1
	}

	fmt.Printf("%v, %v, %v\n", lastDay, time.Month(m), days)

	return make([]int, 10)
}

// import tea "charm.land/bubbletea/v2"
//
// type AppModel struct {
// 	currentView string
// 	form        FormModel
// }
//
// func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch m.currentView {
// 	case "form":
// 		updated, cmd: = m.form.Update(msg)
// 		m.form = updated(FormModel)
// 		return m, smd
// 	}
// }
