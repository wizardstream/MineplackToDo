package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
)

// =======================
// Data Structures
// =======================

// Task represents a single todo task with a name and completion status
type Task struct {
	Name string
	Done bool
}

// Global slice to hold all tasks in memory
var tasks []Task

// Set file to load
var currentFilePath string = "tasks.nasin"

// =======================
// Utility Functions
// =======================

// clearScreen clears the terminal screen based on the OS
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()

	// Print header after clearing
	fmt.Println("==========================")
	fmt.Println("======Mineplack ToDo======")
	fmt.Println("==========================")
}

// ensureDir makes sure the directory for the tasks file exists,
// creates it (and any parents) if missing
func ensureDir() error {
	dir := filepath.Dir(getTaskFilePath())
	return os.MkdirAll(dir, 0755)
}

// getTaskFilePath returns the full path to the tasks file,
// using the user's home directory and a "mineplacktodo" folder
func getTaskFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return currentFilePath // fallback to current dir
	}
	return filepath.Join(homeDir, "mineplacktodo", currentFilePath)
}

// =======================
// File Handling Functions
// =======================

// saveTasksCustom writes the tasks slice to the file in the custom format
func saveTasksCustom(tasks []Task) error {
	if err := ensureDir(); err != nil {
		return err
	}

	file, err := os.Create(getTaskFilePath())
	if err != nil {
		return err
	}
	defer file.Close()

	for i, t := range tasks {
		doneStr := "false"
		if t.Done {
			doneStr = "true"
		}
		line := fmt.Sprintf("T%d : \"%s\" : STRING : DONE:%s\n", i+1, t.Name, doneStr)
		if _, err := file.WriteString(line); err != nil {
			return err
		}
	}

	return nil
}

// loadTasksCustom reads the custom formatted file into a slice of Tasks
func loadTasksCustom() ([]Task, error) {
	file, err := os.Open(getTaskFilePath())
	if err != nil {
		if os.IsNotExist(err) {
			return []Task{}, nil // file doesn't exist yet
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var loadedTasks []Task

	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.Split(line, " : ")
		if len(parts) < 4 {
			continue // skip malformed lines
		}

		name := strings.Trim(parts[1], "\"")

		done := false
		if strings.HasPrefix(parts[3], "DONE:") {
			doneStr := strings.TrimPrefix(parts[3], "DONE:")
			done = (doneStr == "true")
		}

		loadedTasks = append(loadedTasks, Task{Name: name, Done: done})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return loadedTasks, nil
}

// clearAllTasks clears the in-memory tasks slice and saves an empty list to file
func clearAllTasks() error {
	tasks = []Task{}
	return saveTasksCustom(tasks)
}

// =======================
// Main Application Loop
// =======================

func main() {
	clearScreen()

	// Load tasks from file on startup
	var err error
	tasks, err = loadTasksCustom()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		tasks = []Task{}
	}

	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		if input == "" {
			continue
		}

		tokens := strings.Fields(input)
		cmd := tokens[0]
		args := tokens[1:]

		switch cmd {
		case "help":
			fmt.Println(`
Available commands:
  add <task>         - Add a new task
  list               - List all tasks
  done <number>      - Mark a task as completed
  delete <number>    - Delete a task
  delAll             - Delete ALL tasks in the current file
  file <file>        - Change to a different task file (loads its tasks)
  clear              - Clear the terminal screen
  help               - Show this help message
  exit               - Exit the program`)

		case "add":
			taskName := strings.Join(args, " ")
			tasks = append(tasks, Task{Name: taskName})
			fmt.Println("Added:", taskName)

			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "list":
			if len(tasks) == 0 {
				fmt.Println("No tasks.")
				continue
			}
			for i, t := range tasks {
				status := " "
				if t.Done {
					status = "x"
				}
				fmt.Printf("%d. [%s] %s\n", i+1, status, t.Name)
			}

		case "done":
			if len(args) < 1 {
				fmt.Println("Usage: done <number>")
				continue
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 1 || i > len(tasks) {
				fmt.Println("Invalid task number")
				continue
			}
			tasks[i-1].Done = true
			fmt.Println("Marked done:", tasks[i-1].Name)

			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "delete":
			if len(args) < 1 {
				fmt.Println("Usage: delete <number>")
				continue
			}
			i, err := strconv.Atoi(args[0])
			if err != nil || i < 1 || i > len(tasks) {
				fmt.Println("Invalid task number")
				continue
			}
			deleted := tasks[i-1].Name
			tasks = append(tasks[:i-1], tasks[i:]...)
			fmt.Println("Deleted:", deleted)

			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "delAll":
			if err := clearAllTasks(); err != nil {
				fmt.Println("Error clearing tasks:", err)
			} else {
				fmt.Println("All tasks deleted from current file.")
			}

		case "file":
			if len(args) < 1 {
				fmt.Println("Usage: file <filename>")
				continue
			}
			newFile := args[0]
			currentFilePath = newFile

			tasks, err = loadTasksCustom()
			if err != nil {
				fmt.Println("Error loading tasks from new file:", err)
				tasks = []Task{}
			}
			fmt.Println("Switched to file:", newFile, "with", len(tasks), "tasks loaded.")

		case "clear":
			clearScreen()

		case "exit":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Unknown command:", cmd)
		}
	}
}
