package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/ooyeku/csv_parser/pkg"
)

func main() {
	// Read CSV file
	file, err := os.Open("data/employees.csv")
	if err != nil {
		log.Fatalf("Error opening file: %v", err)
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatalf("Error closing file: %v", err)
		}
	}(file)

	// Create table from CSV
	table, err := pkg.ReadTable(file, pkg.DefaultConfig())
	if err != nil {
		log.Fatalf("Error reading table: %v", err)
	}

	// Create different format options for different sections
	mainFormat := pkg.FormatOptions{
		Style:          pkg.RoundedStyle,
		HeaderStyle:    pkg.Bold,
		HeaderColor:    pkg.Cyan,
		BorderColor:    pkg.Blue,
		AlternateRows:  true,
		AlternateColor: pkg.Dim,
		NumberedRows:   true,
		MaxColumnWidth: 20,
		WrapText:       true,
		Alignment:      []string{"right", "left", "right", "left", "right", "left", "center"},
	}

	statsFormat := pkg.FormatOptions{
		Style:          pkg.FancyStyle,
		HeaderStyle:    pkg.Bold + pkg.Underline,
		HeaderColor:    pkg.Yellow,
		BorderColor:    pkg.Green,
		AlternateRows:  true,
		AlternateColor: pkg.Dim,
		Alignment:      []string{"left", "right", "right", "right"},
	}

	managerFormat := pkg.FormatOptions{
		Style:          pkg.DefaultStyle,
		HeaderStyle:    pkg.Bold,
		HeaderColor:    pkg.Magenta,
		BorderColor:    pkg.White,
		MaxColumnWidth: 30,
		Alignment:      []string{"left", "center", "right", "right"},
	}

	fmt.Println("=== Employee Data ===")
	fmt.Println(table.Format(mainFormat))

	// Department Statistics
	fmt.Println("\n=== Department Statistics ===")
	deptStats, err := table.GroupBy(
		[]string{"department"},
		map[string]string{
			"salary": "avg",
			"age":    "avg",
			"id":     "count",
		},
	)
	if err != nil {
		log.Fatalf("Error calculating department statistics: %v", err)
	}
	fmt.Println(deptStats.Format(statsFormat))

	// Manager Analysis with custom formatting
	fmt.Println("\n=== Manager vs Non-Manager Analysis ===")
	managerStats, err := table.GroupBy(
		[]string{"department", "is_manager"},
		map[string]string{
			"salary": "avg",
			"id":     "count",
		},
	)
	if err != nil {
		log.Fatalf("Error calculating manager statistics: %v", err)
	}
	fmt.Println(managerStats.Format(managerFormat))

	// Experience Analysis with gradient coloring
	fmt.Println("\n=== Experience Analysis ===")
	experienceTable := analyzeExperience(table)
	experienceFormat := pkg.FormatOptions{
		Style:          pkg.RoundedStyle,
		HeaderStyle:    pkg.Bold,
		HeaderColor:    pkg.BgBlue + pkg.White,
		BorderColor:    pkg.Cyan,
		AlternateRows:  false,
		MaxColumnWidth: 25,
		Alignment:      []string{"left", "right", "right", "right"},
	}
	fmt.Println(experienceTable.Format(experienceFormat))

	// Age Distribution with compact format
	fmt.Println("\n=== Age Distribution ===")
	ageGroups := createAgeGroups(table)
	ageFormat := pkg.FormatOptions{
		Style:          pkg.RoundedStyle,
		HeaderStyle:    pkg.Bold,
		HeaderColor:    pkg.BgGreen + pkg.Black,
		BorderColor:    pkg.Green,
		CompactBorders: true,
		Alignment:      []string{"center", "right", "right"},
	}
	fmt.Println(ageGroups.Format(ageFormat))
}

// Helper function to safely get column index
func getColIndex(t *pkg.Table, header string) int {
	for i, h := range t.Headers {
		if h == header {
			return i
		}
	}
	return -1
}

// analyzeExperience creates a new table with experience-based analysis
func analyzeExperience(t *pkg.Table) *pkg.Table {
	// Create new table for experience analysis
	expTable := pkg.NewTable([]string{"department", "experience_years", "employee_count", "avg_salary"})

	// Group employees by department
	deptMap := make(map[string][][]string)
	deptIdx := getColIndex(t, "department")
	dateIdx := getColIndex(t, "join_date")
	salaryIdx := getColIndex(t, "salary")

	for _, row := range t.Rows {
		dept := row[deptIdx]
		deptMap[dept] = append(deptMap[dept], row)
	}

	// Calculate experience and averages for each department
	for dept, rows := range deptMap {
		var totalYears float64
		var totalSalary float64

		for _, row := range rows {
			joinDate, _ := time.Parse("2006-01-02", row[dateIdx])
			years := time.Since(joinDate).Hours() / (24 * 365)
			salary, _ := strconv.ParseFloat(row[salaryIdx], 64)

			totalYears += years
			totalSalary += salary
		}

		avgYears := totalYears / float64(len(rows))
		avgSalary := totalSalary / float64(len(rows))

		err := expTable.AddRow([]string{
			dept,
			fmt.Sprintf("%.1f", avgYears),
			strconv.Itoa(len(rows)),
			fmt.Sprintf("%.2f", avgSalary),
		})
		if err != nil {
			return nil
		}
	}

	return expTable
}

// createAgeGroups creates age distribution analysis
func createAgeGroups(t *pkg.Table) *pkg.Table {
	ageTable := pkg.NewTable([]string{"age_group", "count", "avg_salary"})
	groups := make(map[string][]float64)

	ageIdx := getColIndex(t, "age")
	salaryIdx := getColIndex(t, "salary")

	// Group employees by age range
	for _, row := range t.Rows {
		age, _ := strconv.Atoi(row[ageIdx])
		salary, _ := strconv.ParseFloat(row[salaryIdx], 64)

		group := getAgeGroup(age)
		groups[group] = append(groups[group], salary)
	}

	// Calculate statistics for each age group
	for group, salaries := range groups {
		var total float64
		for _, salary := range salaries {
			total += salary
		}
		avg := total / float64(len(salaries))

		err := ageTable.AddRow([]string{
			group,
			strconv.Itoa(len(salaries)),
			fmt.Sprintf("%.2f", avg),
		})
		if err != nil {
			return nil
		}
	}

	return ageTable
}

// getAgeGroup returns the age group for a given age
func getAgeGroup(age int) string {
	switch {
	case age < 30:
		return "20-29"
	case age < 40:
		return "30-39"
	case age < 50:
		return "40-49"
	default:
		return "50+"
	}
}
