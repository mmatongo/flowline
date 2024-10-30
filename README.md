<div align="center">

[![Go Report Card](https://goreportcard.com/badge/github.com/mmatongo/flowline)](https://goreportcard.com/report/github.com/mmatongo/flowline)
[![GoDoc](https://godoc.org/github.com/mmatongo/flowline?status.svg)](https://pkg.go.dev/github.com/mmatongo/flowline)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](./LICENSE)
</div>

> <p align="center">A golang utility to help migrate your knwoledge base from one confluence</p>

## About <a id="about"></a>

*Flowline* is a golang utility to help migrate your knowledge base from confluence. It is designed to be easy to use and flexible enough to be extended to support any platform. (Hopefully)

## Installation <a id="installation"></a>

```bash
go install github.com/mmatongo/flowline
```

## Features <a id="features"></a>

Right now, *Flowline* only supports migrating from Confluence to Outline.

- [x] Text content and formatting such as italic, bold, underline
- [x] Links
- [x] Lists, numbered lists, check lists
- [ ] ~~Notices (Info / error / etc) (I am working on this)~~ (Will work on this if needed)
- [x] Code blocks
- [x] File attachments (kind of)
- [x] Embedded images
- [x] Document nesting (Document nesting is now supported for outline and markdown migrations)
- [x] Emojis
- [x] Simple tables

## Usage <a id="usage"></a>

Here's a basic example of how to use *Flowline* to migrate your knowledge base from Confluence to Outline.

Flowline has a simple ui.

```bash
flowline

A tiny little tool built in a few hours out of frustration to migrate a confluence knowledge base

Usage:
  flowline [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  help        Help about any command
  markdown    Convert Confluence HTML export to markdown files
  outline     Process a confluence HTML export and import it into Outline

Flags:
  -h, --help   help for flowline

Use "flowline [command] --help" for more information about a command.
```

```bash
flowline outline

Process and convert individual pages to markdown, while aiming to preserve document structure

Usage:
  flowline outline [flags]

Flags:
  -c, --collection string   collection id to be populated
  -G, --get-collections     retrieve a list of all the collections
  -h, --help                help for outline
  -i, --input string        path to the confluence HTML export
  -o, --output string       desired output path for the processed documents
  -r, --verify              verify the contents of each page before upload
```

```bash
flowline markdown
Error: required flag(s) "input", "output" not set
Usage:
  flowline markdown [flags]

Flags:
  -h, --help            help for markdown
  -i, --input string    path to the confluence HTML export
  -o, --output string   output path for the markdown files
  -r, --verify          verify before proceeding with conversion

exit status 1
```

## Example 1 <a id="example-1"></a>

Before you can use Flowline to migrate your knowledge base, you need to export your knowledge base from Confluence. You can do this by following the instructions [here](https://confluence.atlassian.com/doc/export-content-to-word-pdf-html-and-xml-139475.html) to export your knowledge base to HTML.

You then need to craete a collection in Outline where you want to import your knowledge base. You can do this by following the instructions [here](https://docs.getoutline.com/s/guide/doc/collections-l9o3LD22sV).

Then create an API key in Outline. You can do this by going to your settings then to the API section.
```bash
# Get a list of all the collections
flowline outline -G

[
  {
    "name": "Wiki",
    "id": "c0df2bd9-8b16-4169-b4ea-ecea5038be1d",
    "url": "https://wiki.example.com/api/collection/digital-PxkVzcux4V"
  },
  {
    "name": "Welcome",
    "id": "daf00c41-05c8-403f-a665-2a27b736c5cb",
    "url": "https://wiki.example.com/api/collection/welcome-SH4HCAvCWl"
  }
]
```
From the output above, you can see that we have two collections. We can use the `id` to specify the collection we want to populate.

Note that you need to export your BASE_URL and API_KEY as environment variables.
i.e
```bash
export BASE_URL=https://wiki.example.com/api
export API_KEY=your_api_key
```

Now you can run the following command to start the migration process.

```bash
flowline outline -i /path/to/confluence-export -o /path/to/output -c c0df2bd9-8b16-4169-b4ea-ecea5038be1d
```

## Example 2 <a id="example-2"></a>

You can also convert the confluence HTML export to markdown files.

```bash
flowline markdown -i /path/to/confluence-export -o /path/to/output
```

## Caveats <a id="caveats"></a>

- Flowline is still in its early stages and may not support all the features you need.
- Flowline is not perfect and may not work as expected.
- Flowline is not affiliated with any of the platforms it supports.
- In the case of Outline, Flowline uses the Outline API to upload documents and attachments. This means that you need to have an internet connection to upload your documents. If you have a large knowledge base, this may take some time as rate limiting is enforced by Outline. (turns out can be byassed but would not recommend it)

## Contributing <a id="contributing"></a>

Contributions are welcome! Feel free to open an issue or submit a pull request if you have any suggestions or improvements.

## License <a id="license"></a>

This project is licensed under the MIT License - see the [LICENSE](./LICENSE) file for details.

## Known Issues <a id="known-issues"></a>

- Flowline does not support all the features of Confluence. If you need a feature that is not supported, feel free to open an issue or submit a pull request.
- Some complex tables may not be converted correctly, I've tried to fix this but it's still a work in progress.
