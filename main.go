package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/eiannone/keyboard"
	"github.com/fatih/color"
	openai "github.com/sashabaranov/go-openai"
)

func printMenu(items []string, selection int) {
	for i, item := range items {
		if i == selection {
			color.Cyan("> %s <\n", item)
		} else {
			fmt.Printf("  %s\n", item)
		}
	}
}

func clearConsole() {
	cmd := exec.Command("cmd", "/c", "cls")
	cmd.Stdout = os.Stdout
	cmd.Run()
}

func selectChat() string {
	_, err := os.Stat("chats")
	if os.IsNotExist(err) {
		os.Mkdir("chats", 0755)
	}

	files, err := os.ReadDir("./chats")
	if err != nil {
		println(err)
	}

	menuItems := []string{"New Chat", "Exit"}
	for _, file := range files {
		if !file.IsDir() {
			conv := strings.TrimSuffix(file.Name(), ".txt")
			menuItems = append([]string{conv}, menuItems...)
		}
	}

	selection := 0
	clearConsole()
	printMenu(menuItems, selection)

	for {
		_, key, err := keyboard.GetSingleKey()

		if err != nil {
			panic(err)
		}

		if key == keyboard.KeyArrowUp {
			selection--
			if selection < 0 {
				selection = len(menuItems) - 1
			}
		}
		if key == keyboard.KeyArrowDown {
			selection++
			if selection > len(menuItems)-1 {
				selection = 0
			}
		}

		if key == keyboard.KeyArrowRight {
			if menuItems[selection] == "Exit" {
				os.Exit(0)
			} else if menuItems[selection] == "New Chat" {
				var name string
				fmt.Print("Name >> ")
				os.Stdout.Sync()
				fmt.Scanln(&name)
				clearConsole()
				return name
			} else {
				clearConsole()
				return menuItems[selection]
			}
		}

		clearConsole()
		printMenu(menuItems, selection)
	}
}

func readAPI() string {
	key, err := os.ReadFile("api-key.txt")
	if err != nil {
		if os.IsNotExist(err) {
			var api string
			fmt.Print("Enter API key >> ")
			fmt.Scanln(&api)
			os.WriteFile("api-key.txt", []byte(api), 0644)
			return api
		} else {
			fmt.Println("Error reading file:", err)
		}
	}
	return string(key)
}
func main() {
	chat := selectChat()
	green := color.New(color.FgGreen)
	cyan := color.New(color.FgCyan)
	client := openai.NewClient(readAPI())
	messages := make([]openai.ChatCompletionMessage, 0)
	reader := bufio.NewReader(os.Stdin)

	file, err := os.ReadFile("chats/" + chat + ".txt")
	if err == nil {
		err = json.Unmarshal(file, &messages)
		if err != nil {
			fmt.Println("Error decoding messages:", err)
			return
		}
	}

	for _, message := range messages {
		if message.Role == "user" {
			cyan.Print("You >> ")
			fmt.Print(message.Content)
		} else {
			green.Print("GPT >> ")
			fmt.Println(message.Content)
		}
	}

	for {
		cyan.Print("You >> ")
		message, _ := reader.ReadString('\n')

		if strings.TrimSpace(message) == ":m" {
			main()
			break
		} else if strings.TrimSpace(message) == ":q" {
			os.Exit(0)
		}

		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleUser,
			Content: message,
		})

		resp, err := client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:    openai.GPT3Dot5Turbo,
				Messages: messages,
			},
		)

		if err != nil {
			fmt.Printf("ChatCompletion error: %v\n", err)
			return
		}

		response := resp.Choices[0].Message.Content
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    openai.ChatMessageRoleAssistant,
			Content: response,
		})

		green.Print("GPT >> ")
		fmt.Println(response)

		file, err = json.Marshal(messages)
		if err != nil {
			fmt.Println("Error encoding messages:", err)
			return
		}
		err = os.WriteFile("chats/"+chat+".txt", file, 0644)
		if err != nil {
			fmt.Println(err)
		}
	}
}
