package warewulfd

import (
	"bytes"
	"testing"
)

func Test_SumOne(t *testing.T) {
	firstText := `Scalable. Flexible. Today, Warewulf unites the ecosystem with the ability to provision containers directly to the bare metal hardware at massive scale, simplistically while retaining massive flexibility.`
	secondText := `Being open source for over two-decades, and pioneering the concept of stateless node management, Warewulf is among the most successful HPC cluster platforms in the industry with support from OpenHPC, contributors around the world, and usage from every industry.`
	DBAddImage("n01", "firstText", bytes.NewReader([]byte(firstText)))
	DBAddImage("n01", "secondText", bytes.NewReader([]byte(secondText)))
	if sum_n02 := DBGetSum("n02"); sum_n02 != [32]byte{} {
		t.Errorf("Sum of second entry must be zero")
	}
	if sum_n01 := DBGetSum("n01"); sum_n01 == [32]byte{} {
		t.Errorf("Sum of entry must not be zero")
	}
	DBResetAll()
	if sum_n01 := DBGetSum("n01"); sum_n01 != [32]byte{} {
		t.Errorf("Sum after reset must be zero")
	}
}
