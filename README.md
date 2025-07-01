# Codo

A Go-based CLI tool that creates an AI coding assistant powered by Claude via AWS Bedrock. The assistant can interact with your local filesystem through built-in tools.

## Features

- **Interactive Chat**: Chat with Claude AI directly from your terminal
- **File Management Tools**: Built-in tools for reading, listing, and editing files
- **AWS Bedrock Integration**: Uses AWS Bedrock to access Claude models
- **Persistent Conversations**: Maintains conversation context throughout the session

## Available Tools

- `read_file`: Read contents of files in your working directory
- `list_files`: List files and directories at any path
- `edit_file`: Make edits to text files (create new files if they don't exist)

## Setup

1. **AWS Configuration**: Set up your AWS credentials and ensure you have access to Bedrock
2. **Environment Variables**: Create a `.env` file based on `.env.example`
3. **Install Dependencies**:
   ```bash
   go mod download
   ```

## Usage

Run the assistant:
```bash
go run main.go
```

The tool will start an interactive chat session where you can:
- Ask questions about your codebase
- Request file modifications
- Get help with programming tasks
- Have Claude analyze and work with your local files

Type your messages and press Enter. Use `Ctrl+C` to quit.

## Requirements

- Go 1.24.3 or later
- AWS account with Bedrock access
- Valid AWS credentials configured
