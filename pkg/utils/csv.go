package utils

import (
	"encoding/csv"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/mergestat/timediff"
)

type Todo struct {
	Id     int
	Name   string
	Date   time.Time
	Status bool
}

func AppendTodo(args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("no todo item provided")
	}

	// First read todos before acquiring any file locks
	todos, err := ReadTodos()
	if err != nil {
		return fmt.Errorf("error reading existing todos: %v", err)
	}

	nextId := 1
	if len(todos) > 0 {
		nextId = todos[len(todos)-1].Id + 1
	}

	// Create the new todo
	newTodo := Todo{
		Id:     nextId,
		Name:   args[0],
		Date:   time.Now(),
		Status: false,
	}

	// Now open the file for appending after ReadTodos has released its lock
	file, err := LoadFile(FilePath)
	if err != nil {
		return err
	}
	defer CloseFile(file)

	// Seek to end of file for appending
	_, err = file.Seek(0, 2) // Seek to end of file
	if err != nil {
		return fmt.Errorf("error seeking to end of file: %v", err)
	}

	writer := csv.NewWriter(file)
	defer writer.Flush()

	record := newTodo.ToRecord()
	if err := writer.Write(record); err != nil {
		return fmt.Errorf("error writing todo: %v", err)
	}

	fmt.Printf("Todo added successfully with ID: %d\n", nextId)
	return nil
}

func ReadTodos() ([]Todo, error) {
	if _, err := os.Stat(FilePath); os.IsNotExist(err) {
		return []Todo{}, nil
	}

	file, err := LoadFile(FilePath)
	if err != nil {
		return nil, err
	}

	defer CloseFile(file)

	reader := csv.NewReader(file)

	records, err := reader.ReadAll()
	if err != nil {
		return nil, fmt.Errorf("error Reading csv : %v ", err)
	}
	todos := make([]Todo, 0, len(records))
	for _, record := range records {
		todo, err := FromRecord(record)
		if err != nil {
			// Skip invalid records
			continue
		}
		todos = append(todos, todo)
	}
	return todos, nil
}

func ListTodo() error {
	todos, err := ReadTodos()
	if err != nil {
		return err
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)

	fmt.Fprintln(w, "ID\tNAME\tDATE\tSTATUS\t")
	fmt.Fprintln(w, "--\t----\t----\t------\t")

	for _, todo := range todos {
		status := "Pending"

		if todo.Status {
			status = "Completed"
		}

		age := timediff.TimeDiff(todo.Date)

		fmt.Fprintf(w, "%d\t%s\t%s\t%s\t\n",
			todo.Id,
			todo.Name,
			age,
			status)

	}
	w.Flush()
	return nil
}

func DeleteTodo(id int) error {
	todos, err := ReadTodos()
	if err != nil {
		return fmt.Errorf("error reading csv: %v", err)
	}

	found := false
	filtredTodos := make([]Todo, 0, len(todos))
	for _, todo := range todos {
		if todo.Id == id {
			found = true
		} else {
			filtredTodos = append(filtredTodos, todo)
		}
	}

	if !found {
		return fmt.Errorf("couldnt find todo with Id : %d", id)
	}

	file, err := os.OpenFile(
		FilePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		os.ModePerm,
	)
	if err != nil {
		return fmt.Errorf("error oppening file for writing : %v", err)
	}
	defer CloseFile(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, todo := range filtredTodos {
		record := todo.ToRecord()
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writting  in file : %v", err)
		}
	}
	fmt.Printf("Todo with ID %d deleted successfully", id)

	return nil
}

func CompleteTodo(id int) error {
	todos, err := ReadTodos()
	if err != nil {
		return fmt.Errorf("error reading csv : %v", err)
	}

	for _, todo := range todos {
		if todo.Id == id {
			todo.Status = !todo.Status
		}
	}

	file, err := os.OpenFile(
		FilePath,
		os.O_WRONLY|os.O_CREATE|os.O_TRUNC,
		os.ModePerm,
	)
	if err != nil {
		return fmt.Errorf("error oppening file for writing : %v", err)
	}
	defer CloseFile(file)

	writer := csv.NewWriter(file)
	defer writer.Flush()

	for _, todo := range todos {
		record := todo.ToRecord()
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("error writting  in file : %v", err)
		}
	}

	fmt.Printf("Todo with ID %d is completed successfully\n", id)

	return nil
}

func (t Todo) ToRecord() []string {
	return []string{
		strconv.Itoa(t.Id),
		t.Name,
		t.Date.Format(time.RFC3339),
		strconv.FormatBool(t.Status),
	}
}

func FromRecord(record []string) (Todo, error) {
	if len(record) != 4 {
		return Todo{}, fmt.Errorf(
			"invalid record: expected 4 fields got %d",
			len(record),
		)
	}

	id, err := strconv.Atoi(record[0])
	if err != nil {
		return Todo{}, fmt.Errorf("invalid ID: %v", err)
	}
	date, err := time.Parse(time.RFC3339, record[2])
	if err != nil {
		return Todo{}, fmt.Errorf("invalid Date format: %v", err)
	}

	status, err := strconv.ParseBool(record[3])
	if err != nil {
		status = false // Default to false if status parsing fails
	}

	return Todo{
		Id:     id,
		Name:   record[1],
		Date:   date,
		Status: status,
	}, nil
}
