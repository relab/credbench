package testconfig

import (
	"encoding/json"
	"os"
)

type TestConfig struct {
	TotalAccounts       int    `json:"total_accounts"`
	AccountDistribution string `json:"account_distribution"`
	Faculties           int    `json:"faculties"`
	FacultyMembers      int    `json:"faculty_members"`
	Semesters           int    `json:"semesters"`
	Courses             int    `json:"courses"`
	Evaluators          int    `json:"evaluators"`
	Exams               int    `json:"exams"`
	Students            int    `json:"students"`
}

func LoadConfig(filename string) (config TestConfig, err error) {
	file, err := os.Open(filename)
	defer file.Close()
	if err != nil {
		return
	}
	err = json.NewDecoder(file).Decode(&config)
	return
}
