## gpt tune create

Create a fine-tuned model

### Synopsis

Create a fine-tuned model from the provided training file ID.

```
gpt tune create <trainingFileID> [validationFileID] [flags]
```

### Options

```
  -b, --base string     Base model (default: curie) (default "curie")
  -h, --help            help for create
  -s, --suffix string   Name suffix of the fine-tuned model
```

### Options inherited from parent commands

```
  -r, --raw   Raw OpenAI Response?
```

### SEE ALSO

* [gpt tune](gpt_tune.md)	 - Manage fine-tuned models

###### Auto generated by spf13/cobra on 16-Apr-2023
