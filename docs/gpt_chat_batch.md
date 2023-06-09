## gpt chat batch

Chat complete a batch of answers

### Synopsis

Chat complete a batch of answers from a specified file.

```
gpt chat batch <outputFile> <promptFile> <systemFile> <answerFile> [questionFile] [flags]
```

### Options

```
  -a, --answer-field string     Answer field name (required)
  -b, --batch-size int          Batch size for concurrent requests (default 20)
  -h, --help                    help for batch
  -q, --question-field string   Question field name (optional)
  -Q, --question-id string      Question ID (optional, name | name=value)
  -s, --score-field string      Score field name (default "score")
  -S, --score-select string     Score selection: first | last | all | none (default "last")
```

### Options inherited from parent commands

```
  -t, --max-tokens int        Maximum number of tokens to generate
  -m, --model string          Model ID (default "gpt-4")
  -T, --temperature float32   Temperature for sampling (default 0.2)
```

### SEE ALSO

* [gpt chat](gpt_chat.md)	 - Complete a chat prompt

###### Auto generated by spf13/cobra on 16-Apr-2023
