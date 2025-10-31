# LLM Data Analyzer

A command-line tool to analyze large text files using Large Language Models (LLMs).

## Overview

`llm-data-analyzer` is a Go-based command-line tool that analyzes large text or JSONL files. It breaks down the input file into smaller chunks that fit within the context window of a specified LLM. Each chunk is analyzed individually, and the results are then summarized to produce a final, consolidated report.

This tool is useful for tasks such as:
- Summarizing long documents or transcripts.
- Extracting key information from a large dataset.
- Performing sentiment analysis on a large number of reviews.

## Features

- **Large File Support:** Analyze files that are larger than the LLM's context window.
- **Parallel Processing:** Chunks are analyzed in parallel to speed up the process.
- **Recursive Summarization:** Intermediate summaries are recursively summarized until a final report is generated.
- **Configurable LLM Endpoints:** Supports any OpenAI-compatible API endpoint.
- **JSONL Support:** Can process JSONL files, treating each line as a separate document to be chunked.

## Installation

1.  Clone the repository:
    ```bash
    git clone https://github.com/your-username/llm-data-analyzer.git
    cd llm-data-analyzer
    ```
2.  Build the application:
    ```bash
    go build -o llm-data-analyzer .
    ```

## Configuration

The tool is configured using a `config.yaml` file. Here is an example:

```yaml
endpoints:
  - name: openai
    endpoint_url: "https://api.openai.com/v1"
    api_key_env: "OPENAI_API_KEY"
    model: "gpt-4"
    context_window_size: 8192
    chunk_size: 4096
  - name: local_mistral
    endpoint_url: "http://localhost:11434/v1"
    api_key_env: "LOCAL_API_KEY"
    model: "mistral"
    context_window_size: 4096
    chunk_size: 2048
```

- `name`: A unique name for the endpoint configuration.
- `endpoint_url`: The URL of the LLM API endpoint.
- `api_key_env`: The name of the environment variable that holds the API key.
- `model`: The name of the model to use.
- `context_window_size`: The maximum context window size of the model in tokens.
- `chunk_size`: The size of the chunks to split the data into, in tokens.

## Usage

```bash
./llm-data-analyzer [flags] <input_file_path>
```

**Flags:**

- `--config, -c` (string): Path to the config file (default: `config.yaml`).
- `--endpoint-name, -e` (string): Name of the LLM endpoint to use (defined in the config file).
- `--analysis-prompt-file` (string): Path to the data analysis prompt file (required).
- `--summary-prompt-file` (string): Path to the summary prompt file (required).
- `--output, -o` (string): Path to the output file (default is stdout).
- `--temp-dir` (string): Path to the temporary directory for intermediate files.
- `--keep-temp-dir` (bool): Keep the temporary directory after execution.
- `--verbose, -v` (bool): Enable verbose logging.
- `--jsonl` (bool): Treat the input file as JSONL.

### Example

1.  **Create an analysis prompt file (`analysis.txt`):**
    ```
    Summarize the following text:
    ```

2.  **Create a summary prompt file (`summary.txt`):**
    ```
    Summarize the following summaries:
    ```

3.  **Create an input file (`input.txt`):**
    ```
    (Your long text here)
    ```

4.  **Run the tool:**
    ```bash
    export OPENAI_API_KEY="your-api-key"
    ./llm-data-analyzer -e openai --analysis-prompt-file analysis.txt --summary-prompt-file summary.txt input.txt
    ```

## Development

This project is developed using Go and the Cobra CLI framework.

### Building from source

```bash
go build -o llm-data-analyzer .
```

### Running tests

```bash
go test ./...
```
