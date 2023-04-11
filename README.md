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

* Build the applications for running in your local development environment:

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
```

Listing the models is a convenient way to verify that you can access the OpenAI API
as expected.

Also, you can explore the [CLI documentation](/docs/gpt.md).
