package golsv

import (
	"fmt"
	"os"

	"github.com/crillab/gophersat/bf"
)

// WriteDIMACS writes a CNF formula to a file in DIMACS format.
// The formula should be provided as a gophersat bf.Formula.
// Returns an error if writing to the file fails.
func WriteDIMACS(formula bf.Formula, filePath string) error {
	// Convert formula to CNF
	cnf := bf.AsCNF(formula)
	
	// Open file for writing
	file, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("failed to create DIMACS file: %w", err)
	}
	defer file.Close()
	
	// Write the formula to the file in DIMACS format
	_, err = file.WriteString(cnf.CNF().String())
	if err != nil {
		return fmt.Errorf("failed to write to DIMACS file: %w", err)
	}
	
	return nil
}