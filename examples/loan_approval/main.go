package main

import (
	"fmt"

	"github.com/loian/fuzzylib/fis"
)

func main() {
	fmt.Println("=== Loan Approval Fuzzy System ===")
	fmt.Println()

	// Load the loan approval FIS
	system, err := fis.LoadFIS("testdata/loan_approval.fis")
	if err != nil {
		fmt.Printf("Error loading FIS: %v\n", err)
		return
	}

	fmt.Println("System loaded: LoanApprovalSystem")
	fmt.Println("Inputs: CreditScore, Income, DebtToIncomeRatio, LoanTerm,")
	fmt.Println("        CreditUtilization, EmploymentDuration, MaritalStatus, LTV, NumberOfDependents")
	fmt.Println("Output: ApprovalChance [0-100]")
	fmt.Println()

	// Define test applicants
	applicants := []struct {
		name               string
		creditScore        float64
		income             float64
		debtToIncomeRatio  float64
		loanTerm           float64
		creditUtilization  float64
		employmentDuration float64
		maritalStatus      float64
		ltv                float64
		numberOfDependents float64
		expectedOutcome    string
	}{
		{
			name:               "Excellent Applicant",
			creditScore:        9.0,
			income:             85000,
			debtToIncomeRatio:  0.75,
			loanTerm:           15,
			creditUtilization:  25,
			employmentDuration: 8,
			maritalStatus:      1.5,
			ltv:                0.5,
			numberOfDependents: 2,
			expectedOutcome:    "Approved",
		},
		{
			name:               "Poor Applicant",
			creditScore:        2.5,
			income:             25000,
			debtToIncomeRatio:  0.15,
			loanTerm:           5,
			creditUtilization:  85,
			employmentDuration: 2,
			maritalStatus:      0.5,
			ltv:                1.2,
			numberOfDependents: 4,
			expectedOutcome:    "Rejected",
		},
		{
			name:               "Average Applicant",
			creditScore:        5.5,
			income:             50000,
			debtToIncomeRatio:  0.5,
			loanTerm:           25,
			creditUtilization:  50,
			employmentDuration: 5,
			maritalStatus:      1.5,
			ltv:                0.75,
			numberOfDependents: 2,
			expectedOutcome:    "Review",
		},
		{
			name:               "High Income Poor Credit",
			creditScore:        3.0,
			income:             90000,
			debtToIncomeRatio:  0.25,
			loanTerm:           30,
			creditUtilization:  70,
			employmentDuration: 6,
			maritalStatus:      2.5,
			ltv:                0.95,
			numberOfDependents: 5,
			expectedOutcome:    "Review",
		},
		{
			name:               "Excellent Credit Low Income",
			creditScore:        9.0,
			income:             30000,
			debtToIncomeRatio:  0.65,
			loanTerm:           20,
			creditUtilization:  30,
			employmentDuration: 7,
			maritalStatus:      1.5,
			ltv:                0.45,
			numberOfDependents: 1,
			expectedOutcome:    "Review or Approved",
		},
	}

	fmt.Println("Evaluating Loan Applicants:")
	fmt.Println(repeatChar("=", 80))

	for i, applicant := range applicants {
		fmt.Printf("\n%d. %s\n", i+1, applicant.name)
		fmt.Println(repeatChar("-", 80))

		inputs := map[string]float64{
			"CreditScore":        applicant.creditScore,
			"Income":             applicant.income,
			"DebtToIncomeRatio":  applicant.debtToIncomeRatio,
			"LoanTerm":           applicant.loanTerm,
			"CreditUtilization":  applicant.creditUtilization,
			"EmploymentDuration": applicant.employmentDuration,
			"MaritalStatus":      applicant.maritalStatus,
			"LTV":                applicant.ltv,
			"NumberOfDependents": applicant.numberOfDependents,
		}

		fmt.Printf("   Credit Score:        %.1f/10\n", applicant.creditScore)
		fmt.Printf("   Income:              $%.0f\n", applicant.income)
		fmt.Printf("   Debt-to-Income:      %.2f\n", applicant.debtToIncomeRatio)
		fmt.Printf("   Loan Term:           %.0f years\n", applicant.loanTerm)
		fmt.Printf("   Credit Utilization:  %.0f%%\n", applicant.creditUtilization)
		fmt.Printf("   Employment Duration: %.1f\n", applicant.employmentDuration)
		fmt.Printf("   Marital Status:      %.1f (0=Single, 1.5=Married, 2.5=Divorced)\n", applicant.maritalStatus)
		fmt.Printf("   LTV (Loan-to-Value): %.2f (%.0f%%)\n", applicant.ltv, applicant.ltv*100)
		fmt.Printf("   Number of Dependents: %.0f\n", applicant.numberOfDependents)

		outputs, err := system.Infer(inputs)
		if err != nil {
			fmt.Printf("   Error: %v\n", err)
			continue
		}

		approvalChance := outputs["ApprovalChance"]
		decision := getDecision(approvalChance)
		symbol := getSymbol(decision)

		fmt.Printf("\n   Approval Chance: %.1f/100\n", approvalChance)
		fmt.Printf("   Decision: %s %s\n", symbol, decision)
		fmt.Printf("   Expected: %s\n", applicant.expectedOutcome)
	}

	fmt.Println()
	fmt.Println(repeatChar("=", 80))
	fmt.Println("\nDecision Ranges:")
	fmt.Println("  Rejected: 0-40")
	fmt.Println("  Review:   30-70")
	fmt.Println("  Approved: 60-100")
}

func getDecision(score float64) string {
	switch {
	case score < 35:
		return "REJECTED"
	case score < 65:
		return "NEEDS REVIEW"
	default:
		return "APPROVED"
	}
}

func getSymbol(decision string) string {
	switch decision {
	case "APPROVED":
		return "✓"
	case "REJECTED":
		return "✗"
	default:
		return "⚠"
	}
}

func repeatChar(char string, count int) string {
	result := ""
	for i := 0; i < count; i++ {
		result += char
	}
	return result
}
