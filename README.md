# OpenAI GPT API Client

This project provides a Go language OpenAI API client, along with a command-line
application that exercises a variety of API endpoints. You can use the application
to fine-tune models, create prompt completions, and process batches with template
files.

## Learning Resources

* [OpenAI API Reference](https://platform.openai.com/docs/api-reference)
* [Go Programming Language](https://go.dev/) Home Page
* [Go Standard Library](https://pkg.go.dev/std) package documentation
* [Learning Go](https://learning.oreilly.com/library/view/learning-go/9781492077206/),
  by Jon Bodner (highly recommended for learning modern, idiomatic Go)

## Developer Workstation Setup

* Install the [Go Language](https://golang.org/doc/install), and set up
  a [GOPATH environment variable](https://github.com/golang/go/wiki/SettingGOPATH).
* Install an IDE, such as [VSCode](https://code.visualstudio.com/) with
  the [Go extension](https://marketplace.visualstudio.com/items?itemName=golang.go).
* Install some Go tools used in the [Makefile](Makefile) in your GOPATH bin folder:

```bash
make tools
```

* Configure the following environment variables for your OpenAI account. For
  simplicity, you could put them in your `.bashrc`, `.zshrc`, or equivalent.
  Substitute the placeholders below with your own values, of course.

```bash
export OPENAI_API_KEY="sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx"
export OPENAI_ORG_ID="org-xxxxxxxxxxxxxxxxxxxxxxxx"
```

* Alternatively, you can create a file named `.env` and place the above environment
variables there. Example contents:

```text
OPENAI_API_KEY=sk-xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
OPENAI_ORG_ID=org-xxxxxxxxxxxxxxxxxxxxxxxx
```

* Build the application for running in your local development environment:

```bash
make build
```

## Install the Command-line Application

If you don't intend to modify the application, you can install it on your workstation
by simply downloading the appropriate [executable application](/dist/) and running it
locally from the command-line.

If your workstation is set up for Go language development, you can install the latest
version from source using the following command:

```bash
go install github.com/voxtechnica/gpt@latest
```

## Run the Command-line Application

To run the command-line application, be sure to either set the environment variables
identified above or create a `.env` file for your OpenAI account. To verify that the
application is configured properly, you can run the `gpt about` command. It will show
your organization ID and API key.

You can explore the available commands, subcommands, and optional flags using
the `-h` or `--help` flags with the executable application. For example:

```bash
./gpt -h
./gpt model -h
./gpt model list -h
./gpt model read -h
./gpt chat -h
./gpt chat prompt -h
./gpt chat random -h
./gpt chat batch -h
```

Listing the models is a convenient way to verify that you can access the OpenAI API
as expected.

Also, you can explore the [CLI documentation](/docs/gpt.md).

## Working with Text and CSV Files

Some of the commands (e.g. `chat random` and `chat batch`) use CSV files for data
inputs and outputs. The first line of these files is expected to be column labels,
and they need to be unique. If the column label is missing, it gets labeled as
"columnX" where X is the column number. The subsequent lines of the CSV file are
individual rows of data.

When you're using the `chat random` and `chat batch` commands, you can refer to
different fields in your dataset by using the column labels. They're case-sensitive,
and ideally, have no spaces. Good examples are `id`, `question`, and `answer`.
You can include extra columns in your CSV files, and they're either ignored or
reproduced faithfully in your output file.

If you want to refer to a specific field in a specific row of your dataset, you'll
be supplying two arguments: the name and value of an "id" field, and the name of
the field that contains the value you're referring to. The "id" field will be a
name=value pair, such as `pid=61324`.

## Using the chat Command

The [chat command](/docs/gpt_chat.md) has four subcommands: `prompt`, `random`,
`parallel`, and `batch`. The first two commands will typically be used when
engineering an effective prompt for use with a batch of data. The `prompt` command
works with just two text files: a prompt file and an optional system file. These
files are plain UTF-8 text. The system file can be used to inform GPT of its
identity and the role it is expected to play (e.g. [system.txt](/examples/system.txt)).
The prompt file then provides instructions for GPT.
The [limerick.txt](/examples/limerick.txt) file provides a simple example, and
the [spir_definition.txt](/examples/spir_definition.txt) file provides a more
complex example.

Note that for the `random`, `parallel`, and `batch` commands, the prompt file can
contain two template substitution variables: `{{question}}` and `{{answer}}`. These
will be replaced with the actual text of a provided (optional) question and
the (required) answer.

So, the `prompt` command is used for simple prompts, and when you're ready to
start experimenting with question(s) and answers in your CSV dataset, you can
use the `random` command to test your prompt with different values. If you want
to finesse the prompt with a specific answer, you can use the `--answer-id` flag
with the `random` command to force it to select the specified answer (which, of
course, is not random).

Once you're happy with the results you're seeing, you can use the `parallel` or
`batch` commands to process the entire dataset. The `parallel` command works
through your records in real time, sending multiple requests in parallel. The
`batch` command generates an asynchronous batch job, wherein OpenAI manages
the concurrency. Using `batch` is recommended, as it costs about half as much
as the real-time requests. Also, in testing, it appears to complete very quickly.

The `chat` commands can also parse "scores" (numbers) from the GPT response text.
The `--score-select` flag indicates whether you'd like the first number found in
the text, the last number, all the numbers, or none of the numbers (i.e. don't
bother parsing scores). You'll probably just want the last number, and that's
the default score selection.

When you use the `parallel` or `batch` commands, the output CSV file will contain
all the data provided in your input answer file, along with a few new columns. The
`chatID` column will contain a unique ID used with GPT, and the `completion`
column will contain the exact text of the GPT response. If you're also
parsing score(s) from the GPT response, those column(s) will be included too.
The default name for a score column is `score` or `scoreX` where X is a number
indicating the position of the score in the list, if multiple scores are being
parsed from the response text. Note that you can change the name of the score
column using the `--score-field` flag. This can be useful if you're storing the
results from multiple runs in the same CSV file. You can just give each run a
different score field name, reusing the output file as an input file. The new
score columns will be appended with each run.

One more tip on using questions and answers: if you have multiple questions,
and your answer file includes answers to different questions, you can include
a question ID column in your answers file. Then, when processing each answer,
the system will look up the appropriate question to use for that answer in
the questions CSV file. The question ID column name (e.g. `qid`) must be the
same in both the questions file and the answers file. The `--question-id` flag
controls this behavior. If you specify just a question ID field name, then it
will be assumed that the field exists in both files. If you specify a
`name=value` pair, then just that specific question will be used for the
entire set of answers. Also, note that questions are optional. If all you have
to process are "answers", then you can ignore the question bits.
