## gpt text batch

Text complete a batch of answers

### Synopsis

Text  complete a batch of answers from a specified file.

```
gpt text batch <outputFile> <promptFile> <answerFile> [questionFile] [flags]
```

### Options

```
  -a, --answer-field string     Answer field name (required)
  -b, --batch-size int          Batch size for concurrent requests (default 15)
  -h, --help                    help for batch
  -q, --question-field string   Question field name (optional)
  -Q, --question-id string      Question ID (optional, name=value)
```

### Options inherited from parent commands

```
  -t, --max-tokens int        Maximum number of tokens to generate (default 256)
  -m, --model string          Model ID (default "text-davinci-003")
  -T, --temperature float32   Temperature for sampling (default 0.2)
```

### SEE ALSO

* [gpt text](gpt_text.md)	 - Complete a text prompt

###### Auto generated by spf13/cobra on 16-Apr-2023
