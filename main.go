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

// =======================
// Utility Functions
// =======================

// clearScreen clears the terminal screen based on the OS
func clearScreen() {
	var cmd *exec.Cmd

	switch runtime.GOOS {
	case "windows":
		// Windows clear screen command
		cmd = exec.Command("cmd", "/c", "cls")
	default:
		// Unix-like clear screen command (Linux, macOS, etc)
		cmd = exec.Command("clear")
	}

	cmd.Stdout = os.Stdout
	cmd.Run()

	// Print header after clearing
	fmt.Println("==========================")
	fmt.Println("======Mineplack ToDo======")
}

// ensureDir makes sure the directory for the tasks file exists,
// creates it (and any parents) if missing
func ensureDir() error {
	dir := filepath.Dir(getTaskFilePath())
	return os.MkdirAll(dir, 0755) // rwxr-xr-x permissions
}

// getTaskFilePath returns the full path to the tasks file,
// using the user's home directory and a "mineplacktodo" folder
func getTaskFilePath() string {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		// fallback: use current directory if home dir can't be found
		return "tasks.nasin"
	}
	// Construct path: ~/mineplacktodo/tasks.nasin or equivalent on Windows
	return filepath.Join(homeDir, "mineplacktodo", "tasks.nasin")
}

// =======================
// File Handling Functions
// =======================

// saveTasksCustom writes the tasks slice to the file in the custom format
func saveTasksCustom(tasks []Task) error {
	// Make sure directory exists before saving
	if err := ensureDir(); err != nil {
		return err
	}

	file, err := os.Create(getTaskFilePath())
	if err != nil {
		return err
	}
	defer file.Close()

	// Write each task line by line in the format:
	// T1 : "task name" : STRING : DONE:true/false
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
			// No saved tasks yet, return empty slice
			return []Task{}, nil
		}
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	var loadedTasks []Task

	for scanner.Scan() {
		line := scanner.Text()
		// Split the line by " : " delimiter to extract fields
		parts := strings.Split(line, " : ")
		if len(parts) < 4 {
			// skip malformed lines that don't have all fields
			continue
		}

		// parts breakdown:
		// parts[0] = task index like T1
		// parts[1] = task name in quotes
		// parts[2] = STRING (type, unused)
		// parts[3] = DONE:true or DONE:false

		// Remove surrounding quotes from the task name
		name := strings.Trim(parts[1], "\"")

		// Parse done status
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

// =======================
// Main Application Loop
// =======================

func main() {
	// Clear screen and print header on start
	clearScreen()

	// Load previously saved tasks from file
	var err error
	tasks, err = loadTasksCustom()
	if err != nil {
		fmt.Println("Error loading tasks:", err)
		tasks = []Task{}
	}

	reader := bufio.NewReader(os.Stdin)

	// Main REPL loop: prompt user, read input, parse commands
	for {
		fmt.Print("> ")
		input, _ := reader.ReadString('\n')
		input = strings.TrimSpace(input)

		// Skip empty input
		if input == "" {
			continue
		}

		// Split input into command and arguments
		tokens := strings.Fields(input)
		cmd := tokens[0]
		args := tokens[1:]

		switch cmd {
		case "help":
			// Display list of available commands
			fmt.Println(`
Available commands:
  add <task>         - Add a new task
  clear              - Clears Screen
  done <number>      - Mark a task as completed
  delete <number>    - Delete a task
  help               - Show this help message
  list               - List all tasks

  exit               - Exit the program`)

		case "add":
			// Join all arguments to form the task name
			taskName := strings.Join(args, " ")
			// Append new task to list
			tasks = append(tasks, Task{Name: taskName})
			fmt.Println("Added:", taskName)

			// Save updated tasks to file
			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "list":
			// Show all tasks with completion status
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
			// Mark specified task as done
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

			// Save changes
			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "delete":
			// Delete specified task
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

			// Save changes
			if err := saveTasksCustom(tasks); err != nil {
				fmt.Println("Error saving tasks:", err)
			}

		case "clear":
			// Clear terminal screen and print header again
			clearScreen()

		case "exit":
			// Exit the program gracefully
			fmt.Println("Goodbye!")
			return

		default:
			// Handle unknown commands
			fmt.Println("Unknown command:", cmd)
		}
	}
}
