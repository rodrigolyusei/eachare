package number

import (
	"io/ioutil"
	"path/filepath"
	"strconv"
	"strings"
)

func GetNextPort() (int, error) {
	number, err := UpdateAndGetNumberFromFile()
	if err != nil { return 0, err }
	port, err := strconv.Atoi("80" + strconv.Itoa(number))
	if err != nil { return 0, err }
	return port, nil
}

func UpdateAndGetNumberFromFile() (int, error) {
	filepath := filepath.Join("data.txt")

	err := updateNumberInFile(filepath)
	if err != nil {
		return 0, err
	}

	number, err := getNumberFromFile(filepath)
	if err != nil {
		return 0, err
	}

	return number, nil
}

func updateNumberInFile(filename string) error {
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return err
	}

	// Convert the data to an integer
	number, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return err
	}

	// Increase the number by one
	number++

	// Write the updated number back to the file
	return ioutil.WriteFile(filename, []byte(strconv.Itoa(number)), 0644)
}

func getNumberFromFile(filename string) (int, error) {
	// Read the number from the file
	data, err := ioutil.ReadFile(filename)
	if err != nil {
		return 0, err
	}

	// Convert the data to an integer
	number, err := strconv.Atoi(strings.TrimSpace(string(data)))
	if err != nil {
		return 0, err
	}

	return number, nil
}
