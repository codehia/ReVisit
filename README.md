# goflash

> AI-powered flashcards in your terminal, for any topic you want to learn.

Most flashcard tools make you rate yourself, which is unreliable and easy to game. goflash feeds your answers to an AI that scores them objectively and drives the spaced repetition schedule (SM-2), so cards you struggle with come back sooner. It runs entirely in the terminal and stores everything locally in SQLite.

## Prerequisites

- [Go](https://golang.org/dl/) 1.21 or higher
- A [DeepSeek API key](https://platform.deepseek.com/) (used for seeding cards and evaluating answers)

## Quick Start

Clone the repo and build the binary:

```bash
git clone https://github.com/codehia/goflash.git
cd goflash
go build ./...
```

Set your DeepSeek API key as an environment variable:

```bash
export DEEPSEEK_API_KEY=your_api_key_here
```

The repo includes a pre-seeded SQLite database covering system design topics, so you can start a session right away:

```bash
go run main.go
```

To study your own topics, see [Seeding](#seeding).

## Seeding

Seeding lets you populate goflash with any topic of your choice. It is a two-step process.

### Step 1: Generate cards

Prepare a JSON file describing your topic hierarchy. The format is a nested tree of topics, with leaf nodes containing a `notes` field:

```json
{
  "name": "Your Topic",
  "children": [
    {
      "name": "Subtopic",
      "children": [
        {
          "name": "Leaf Topic",
          "notes": "Your notes on this topic go here."
        }
      ]
    }
  ]
}
```

Run the seed script, passing your JSON file as an argument:

```bash
go run cmd/seed/main.go seedfile.json
```

This makes an AI request for each leaf node to generate a question and answer pair. The results are written to `output.json`.

### Step 2: Import into the database

Once `output.json` is ready, run the init script to populate the SQLite database:

```bash
go run cmd/init/main.go
```

The init script reads from `output.json` by default. It upserts records into the database: existing cards that have changed are updated, new cards are added, and nothing is deleted.

Your cards are now ready. Run `go run main.go` to start a session.

## License

MIT License. See [LICENSE](LICENSE) for details.
