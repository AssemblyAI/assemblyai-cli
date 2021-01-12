# Assembly AI CLI

### Command Line Interface (CLI) for Assembly AI

## Installation

1. Follow these instructions to ensure Python is installed on your system: https://www.python.org/downloads/.

2. Make sure `pip` is installed: https://pip.pypa.io/en/stable/installing/

3. Install AssemblyAI CLI: `pip install assemblyai-cli`

## Usage 

### On MacOS/Linux

Login to Assembly AI

```shell
assemblyai login --api_token <MY_API_TOKEN>
```

Transcribe an audio file

```shell
assemblyai transcribe --audio_url "https://s3-us-west-2.amazonaws.com/blog.assemblyai.com/audio/8-7-2018-post/7510.mp3"
```

Show all available methods

```shell
assemblyai help
```

### On Windows

Login to Assembly AI

```shell
assemblyai.exe login --api_token <MY_API_TOKEN>
```

Transcribe an audio file

```shell
assemblyai.exe transcribe --audio_url "https://s3-us-west-2.amazonaws.com/blog.assemblyai.com/audio/8-7-2018-post/7510.mp3"
```

Show all available methods

```shell
assemblyai.exe help
```

## Available methods

```shell
AssemblyAI CLI

positional arguments:
  {login,help,feedback,transcribe}
                        CLI Command to perform (login, help, feedback,
                        transcribe).

optional arguments:
  -h, --help            show this help message and exit
  --api_token API_TOKEN
                        API Token to authenticate with AssemblyAI.
  --message MESSAGE     Feedback to send
  --audio_file AUDIO_FILE
                        Path to local audio file
                        (https://docs.assemblyai.com/guides/uploading-audio-
                        files-for-transcription)
  --audio_url AUDIO_URL
                        URL of remote audio file
                        (https://docs.assemblyai.com/guides/transcribing-an-
                        audio-file-recording)
  --output_to OUTPUT_TO
                        File to output the transcription results (prints to
                        stdout if blank)
  --quiet               If present, CLI will suppress status messages
  --text_only           If present, only the text of the transcription will be
                        returned
  --word_boost WORD_BOOST
                        List of words to boost accuracy for; eg: "foo,bar,foo
                        bar" *optional*
                        (https://docs.assemblyai.com/guides/boosting-accuracy-
                        for-keywords-or-phrases)
  --boost_weight {low,default,high}
                        Boost weight to use *optional*
                        (https://docs.assemblyai.com/guides/boosting-accuracy-
                        for-keywords-or-phrases)
  --model {assemblyai_default,assemblyai_en_au,assemblyai_en_uk}
                        Model to use *optional*
                        (https://docs.assemblyai.com/guides/transcribing-with-
                        a-different-acoustic-or-custom-language-model)
```