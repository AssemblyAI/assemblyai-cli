# AssemblyAI CLI

![Release](https://img.shields.io/github/v/release/assemblyai/assemblyai-cli)
![Build](https://img.shields.io/github/workflow/status/assemblyai/assemblyai-cli/Release%20workflow)
![License](https://img.shields.io/github/license/assemblyai/assemblyai-cli)

The AssemblyAI CLI helps you quickly test our latest AI models right from your terminal, with minimal installation required.

![Thumbnail](./assets/thumbnail.png)

## Installation

The CLI is simple to install, supports a wide range of operating systems like macOS, Windows, and Linux, and makes it more seamless to build with AssemblyAI.

### Homebrew

If you're on macOS, you can install it using Homebrew:

```bash
brew tap assemblyai/assemblyai
brew install assemblyai
```

### macOS or Linux

If you don't have Homebrew installed, or are running Linux:

```bash
/bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/AssemblyAI/assemblyai-cli/main/install.sh)"
```

### Windows

The CLI is available on Windows:

```bash
Set-ExecutionPolicy RemoteSigned -Scope CurrentUser
iex "$(curl https://raw.githubusercontent.com/AssemblyAI/assemblyai-cli/main/install.ps1)"
New-Alias -Name assemblyai -Value C:\\'./Program Files\'/AssemblyAI/assemblyai.exe
```

## Getting started

Get started by configuring the CLI with your AssemblyAI token. If you don't yet have an account, create one [here](https://app.assemblyai.com/).

```bash
assemblyai config [token]
```

This command will validate your account, and store your token safely in `~/.config/assemblyai/config.toml` later to be used when transcribing files.

## Usage

Installing the CLI provides access to the `assemblyai` command:

```bash
assemblyai [command] [--flags]
```

## Commands

### Transcribe

With the CLI, you can transcribe local files, remote URLs, and YouTube links.

```bash
assemblyai transcribe [local file | remote url | youtube links] [--flags]
```

<details>
  <summary>Flags</summary>
  
  > **-j, --json**  
  > default: false  
  > If true, the CLI will output the JSON.

> **-p, --poll**  
> default: true  
> The CLI will poll the transcription every 3 seconds until it's complete.

> **-s, --auto_chapters**  
> default: false  
> A "summary over time" for the audio file transcribed.

> **-a, --auto_highlights**  
> default: false  
> Automatically detect important phrases and words in the text.

> **-c, --content_moderation**  
> default: false  
> Detect if sensitive content is spoken in the file.

> **-d, --dual_channel**  
> default: false  
> Enable dual channel

> **-e, --entity_detection**  
> default: false  
> Identify a wide range of entities that are spoken in the audio file.

> **-f, --format_text**  
> default: true  
> Enable text formatting

> **-u, --punctuate**  
> default: true  
> Enable automatic punctuation

> **-r, --redact_pii**  
> default: false  
> Remove personally identifiable information from the transcription.

> **-i, --redact_pii_policies**  
> default: drug,number_sequence,person_name  
> The list of PII policies to redact ([source](https://www.assemblyai.com/docs/audio-intelligence#pii-redaction)), comma-separated. Required if the redact_pii flag is true.

> **-x, --sentiment_analysis**  
> default: false  
> Detect the sentiment of each sentence of speech spoken in the file.

> **-l, --speaker_labels**  
> default: true  
> Automatically detect the number of speakers in the file.

> **-t, --topic_detection**  
> default: false  
> Label the topics that are spoken in the file.

> **-w, --webhook_url**  
> Receive a webhook once your transcript is complete.

> **-b, --webhook_auth_header_name**  
> Containing the header's name which will be inserted into the webhook request.

> **-o, --webhook_auth_header_value**  
> Receive a webhook once your transcript is complete.

> **-n, --language_detection**  
> default: true  
> Automatic identify the dominant language that’s spoken in an audio file.
> [Here](https://www.assemblyai.com/docs/core-transcription#automatic-language-detection) you can view the ALD list for supported languages

> **-g, --language_code**  
> Manually secify the language of the speech in your audio file.
> Click [here](https://www.assemblyai.com/docs#supported-languages) to view all the supported languages

</details>

### Get

If you're not polling the transcription, you can fetch it later:

```bash
assemblyai get [id]
```

## Contributing

We're more than happy to welcome new contributors. If there's something you'd like to fix or improve, start by [creating an issue](https://github.com/AssemblyAI/assemblyai-cli/issues). Please make sure to follow our [code of conduct](https://github.com/AssemblyAI/assemblyai-cli/blob/main/CODE_OF_CONDUCT.md).

## Telemetry

The AssemblyAI CLI includes a telemetry feature that collects usage data, and is enabled by default.

To opt out of telemetry, set the telemetry variable in the `config.toml` file to false.

## Feedback

Please don't hesitate to [let us know what you think](https://forms.gle/oQgktMWyL7xStH2J8)!

## License

Copyright (c) AssemblyAI. All rights reserved.

Licensed under the [Apache License 2.0 license](https://github.com/AssemblyAI/assemblyai-cli/blob/main/LICENSE).
